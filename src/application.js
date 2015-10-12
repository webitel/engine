'use strict';

var EventEmitter2 = require('eventemitter2').EventEmitter2,
    util = require('util'),
    fs = require('fs'),
    log = require('./lib/log')(module),
    WConsole = require('./lib/Console'),
    conf = require('./conf'),
    Esl = require('./lib/modesl'),
    Collection = require('./lib/Collection'),
    api = require('./adapter/rest'),
    Acl = require('acl'),
    ACL_CONF = require('./conf/acl'),
    ws = require('./adapter/ws'),
    httpSrv = (conf.get('ssl:enabled').toString() == 'true') ? require('https') : require('http'),
    initDb = require('./db'),
    emailService = require('./services/email'),
    plainTableToJSONArray = require('./utils/parse').plainTableToJSONArray,
    outQueryService = require('./services/outboundQueue');

function APPLICATION() {
    EventEmitter2.call(this, {
        wildcard: false,
        //delimiter: '::',
        maxListeners: 25
    });

    this.DB = null;
    this.WConsole = null;
    this.Esl = null;
    this.Users = new Collection('id');
    this.Domains = new Collection('id');
    this.Agents = new Collection('id');
    this.OutboundQuery = new Collection('id');
    this.loggedOutAgent = new Collection('id');

    var self = this;
    self.connectDb();
    return this;
};

util.inherits(APPLICATION, EventEmitter2);

APPLICATION.prototype.Schedule = function (time, fn, arg) {
    var timerId = setTimeout(function tick() {
        fn(arg);
        timerId = setTimeout(tick, time);
    }, time);
};

APPLICATION.prototype.initAcl = function (cb) {
    var acl;

    this.acl = acl = new Acl(new Acl.memoryBackend());
    this.acl.allow(ACL_CONF, function (err) {
        if (err)
            return cb(err);
    });

    ACL_CONF.forEach(function (item) {
        acl.addUserRoles(item.roles, item.roles, function (err) {
            if (err) {
                return cb(err);
            };
            log.debug('Register role %s', item.roles);
        });
    });
    log.info('Load roles.');
    if (cb) {
        cb();
    };
    return 1;
};

APPLICATION.prototype.connectDb = function() {
    var scope = this;
    scope.once('sys::connectDb', function (db) {
        scope.DB = db;
        scope.connectToEsl();
        scope.connectToWConsole();
        scope.initAcl();
    });

    this.once('sys::connectEsl', function () {
        scope.configureExpress();
        /**
         * Init outbound
         */
        //outQueryService._init(scope);
    });

    initDb(scope);

    if (typeof gc == 'function') {
        setInterval(function () {
            gc();
            console.log('----------------- GC -----------------');
        }, 5000);
    };
};

APPLICATION.prototype.connectToEsl = function () {
    var waitTimeReconnectFreeSWITCH = conf.get('freeSWITCH:reconnect') * 1000,
        scope = this;
    if (this.Esl && this.Esl.connected) {
        return;
    };

    var esl = this.Esl = new Esl.Connection(conf.get('freeSWITCH:host'),
        conf.get('freeSWITCH:port'),
        conf.get('freeSWITCH:pwd'),
        function() {
            log.info('Connect freeSWITCH: %s:%s', conf.get('freeSWITCH:host'), conf.get('freeSWITCH:port'));
            this.apiCallbackQueue.length = 0;
            scope.emit('sys::eslConnect');

            //TODO
            log.info('Load tiers');
            this.bgapi('callcenter_config tier list', function (res) {
                let body = res && res['body'];
                if (!body) {
                    return log.error('Load tiers response undefined !!!');
                };
                plainTableToJSONArray(body, function (err, result) {
                    if (err) {
                        return log.error(err);
                    };
                    scope.Agents.removeAll();
                    if (result instanceof Array) {
                        result.forEach(function (item) {
                            let agent = scope.Agents.get(item['agent']);
                            if (!agent) {
                                scope.Agents.add(item['agent'], [item]);
                            } else {
                                agent.push(item);
                            };
                        });
                    };

                }, '|')
            });
        });

    esl.on('error', function(e) {
        log.error('freeSWITCH connect error:', e);
        esl.connected = false;

        setTimeout(function () {
            scope.connectToEsl();
        }, waitTimeReconnectFreeSWITCH);
    });

    esl.on('esl::event::auth::success', function () {
        esl.connected = true;

        var ev = conf.get('application:freeSWITCHEvents');
        esl.subscribe(ev);
        esl.filter('Event-Name', 'CHANNEL_PROGRESS_MEDIA');
        //for (var key in ev) {
        //    esl.filter('Event-Name', ev[key]);
        //};
        esl.filter('Event-Subclass', 'callcenter::info');

        scope.emit('sys::connectEsl');
    });

    esl.on('esl::event::auth::fail', function () {
        esl['authed'] = false;
        log.error('esl::event::auth::fail');
        scope.stop(new Error('Auth freeSWITH fail, please enter the correct password.'));
    });

    esl.on('esl::end', function () {
        esl.connected = false;

        log.error('FreeSWITCH: socket close.');
        setTimeout(function () {
            scope.connectToEsl();
        }, waitTimeReconnectFreeSWITCH);
    });

    esl.on('esl::event::disconnect::notice', function() {
        log.error('esl::event::disconnect::notice');
        this.apiCallbackQueue.length = 0;
        this.cmdCallbackQueue.length = 0;
        esl.connected = false;

        setTimeout(function () {
            scope.connectToEsl();
        }, waitTimeReconnectFreeSWITCH);
    });

};

APPLICATION.prototype.connectToWConsole = function () {
    var scope = this,
        waitTimeReconnectConsole = conf.get('webitelServer:reconnect') * 1000;

    var wconsole = this.WConsole = new WConsole.Connection({
        server: conf.get('webitelServer:host'),
        port: conf.get('webitelServer:port'),
        account: conf.get('webitelServer:account'),
        secret: conf.get('webitelServer:secret')
    });

    wconsole.on('webitel::socket::close', function (e) {
        log.error('Webitel error:', e.toString());
        setTimeout(function () {
            scope.connectToWConsole();
        }, waitTimeReconnectConsole);
    });

    wconsole.on('error', function (err) {
        log.warn('Webitel warn:', err);
    });

    wconsole.on('webitel::event::auth::success', function () {
        log.info('Connect Webitel: %s:%s', this.host, this.port);
        wconsole.subscribe('all');
    });

    wconsole.on('webitel::event::auth::fail', function () {
        wconsole.authed = false;
        log.error('webitel::event::auth::fail');
        log.trace('Reconnect to webitel...');
        setTimeout(function () {
            scope.connectToWConsole();
        }, waitTimeReconnectConsole);
    });

    wconsole.on('webitel::end', function () {
        wconsole.authed = false;
        log.error('Webitel: socket close.');
    });

    wconsole.on('webitel::event::disconnect::notice', function () {
        log.error('webitel::event::disconnect::notice');
    });

    wconsole.on('webitel::event::event::**', function () {

    });

    if (conf.get('application:sleepConnectToWebitel')) {
        setTimeout(function () {
            wconsole.connect();
        }, conf.get('application:sleepConnectToWebitel'));
    } else {
        wconsole.connect();
    };
};

APPLICATION.prototype.configureExpress = function() {
    this.startServer(api);
};

APPLICATION.prototype.startServer = function(api) {
    try {
        var scope = this,
            server;

        if (conf.get('ssl:enabled').toString() == 'true') {
            var https_options = {
                key: fs.readFileSync(conf.get('ssl:ssl_key')),
                cert: fs.readFileSync(conf.get('ssl:ssl_cert'))
            };
            server = httpSrv.createServer(https_options, api).listen(conf.get('server:port'), conf.get('server:host'), function() {
                log.info('Server (https) listening on port ' + this.address().port);
                scope.emit('sys::serverStart', this, true);
            });
        } else {
            server = httpSrv.createServer(api).listen(conf.get('server:port'), conf.get('server:host'), function() {
                log.info('Server (http) listening on port ' + this.address().port);
                scope.emit('sys::serverStart', this, false);
            });
        };
        ws(server, this);

    } catch (e) {
        log.error('Server create:' + e.message);
        this.stop(e);
    }
};

APPLICATION.prototype.stop = function(err) {
    log.warn("Stop server \n" + (err || ''));
    if (this.DB) {
        this.DB.close();
        log.info('Disconnect DB...');
    };

    if (this.Esl) {
        this.Esl.disconnect();
        log.info('Disconnect ESL...');
    };

    if (this.WConsole) {
        this.WConsole.disconnect();
        log.info('Disconnect WConsole...');
    };

    process.exit(1);
};

process.on('uncaughtException', function (err) {
    log.error('UncaughtException:', err.message);
    log.error(err.stack);

    var _fnStop = function() {
        if (application) {
            application.stop();
        }
        process.exit(1);
    };
    
    emailService._report(err, function () {
        _fnStop();
    });
});

module.exports = APPLICATION;