/**
 * Created by i.navrotskyj on 20.03.2016.
 */
'use strict';

var log = require(__appRoot + '/lib/log')(module),
    EventEmitter2 = require('eventemitter2').EventEmitter2,
    async = require('async'),
    generateUuid = require('node-uuid'),
    gatewayService = require(__appRoot + '/services/gateway'),
    Amqp = require('amqplib/callback_api');

const HOOK_QUEUE = 'hooks',
    GATEWAY_QUEUE = 'engine.gateway',
    STORAGE_QUEUE = 'engine-storage',
    TELEGRAM_QUEUE = 'telegram-notification';

const _onReturnedMessage = Symbol('onReturnedMessage'),
      _onRequestFsApi = Symbol('onRequestFsApi'),
    _stackCb = Symbol('stackCb')
    ;

class WebitelAmqp extends EventEmitter2 {

    constructor (amqpConf, app) {
        super();
        this.bgapi = this.api;
        this.config = amqpConf;
        this.app = app;
        this.connect();
        this.queue = null;
        this.customChannelQueue = null;
        this.channel = null;
        this._instanceId = generateUuid.v4();
        this[_stackCb] = {};
    };

    get Exchange () {
        return {
            FS_EVENT: this.config.eventsExchange.channel,
            FS_CC_EVENT: this.config.eventsExchange.cc,
            FS_COMMANDS: this.config.eventsExchange.commands,
            STORAGE_COMMANDS: this.config.storageExchange.commands
        };
    };

    connect () {
        let scope = this,
            timerId;

        //setInterval(()=> {
        //    scope.api('status', (res) => {
        //        //console.log(res.body);
        //    });
        //}, 0);

        function start () {
            if (timerId)
                timerId = clearTimeout(timerId);

            let closeChannel = function() {
                scope.queue = null;

                if (scope.channel) {
                    //scope.channel.close();
                    scope.channel.removeAllListeners('return');
                    scope.channel = null;
                    scope[_stackCb] = {};
                }
            };
            try {

                Amqp.connect(scope.config.uri, (err, conn) => {
                    if (err) {
                        // TODO Docker no reconnect...
                        log.error(err);
                        closeChannel();
                        timerId = setTimeout(start, 5000);
                        return;
                    }

                    log.info('[AMQP] connect: OK');
                    conn.on('error', (err) => {
                        if (err.message !== "Connection closing") {
                            log.error("conn error", err);
                        }
                        conn.close();
                    });
                    conn.on('close', (err)=> {
                        log.error(err);
                        closeChannel();
                        timerId = setTimeout(start, 5000);
                    });

                    conn.createChannel((err, channel) => {
                        if (err) {
                            log.error(err);
                            closeChannel();
                            timerId = setTimeout(start, 5000);
                            return;
                        }
                        channel.on('error', (e) => {
                            log.error(e);
                        });
                        channel.on('return', scope[_onReturnedMessage].bind(scope));

                        scope.init(channel);
                    });
                });

            } catch (e) {
                log.error(e);
            }
        }
        start();
    };

    publish (exchange, rk, content, cb) {
        let ch = this.channel;
        if (!ch)
            return cb && cb(new Error(`No live connect.`));

        try {
            if (!exchange || !content)
                return cb && cb(new Error(`Bad parameters.`));

            if (content instanceof Object) {
                content = new Buffer(JSON.stringify(content));
            }
            log.trace(`publish ${rk}`);
            ch.publish(exchange, rk, content, {contentType: "text/json"});
            return cb && cb();
        } catch (e) {
            log.error(e);
        }
    };

    bindChannelEvents (caller, cb) {
        let ch = this.channel;
        if (!ch || !this.queue) return cb && cb(new Error('No connect.'));
        try {
            ch.bindQueue(this.queue, this.Exchange.FS_EVENT, getPresenceRoutingFromCaller(caller), {}, cb);
        } catch (e) {
            log.error(e);
        }
    };

    unBindChannelEvents (caller) {
        try {
            if (this.channel && this.queue)
                this.channel.unbindQueue(this.queue, this.Exchange.FS_EVENT, getPresenceRoutingFromCaller(caller))
        } catch (e) {
            log.error(e);
        }
    };

    bindDomainEvent (eventName, domainName, cb) {
        let ch = this.channel;
        if (!ch || !this.customChannelQueue) return cb && cb(new Error('No connect.'));
        try {
            ch.bindQueue(this.customChannelQueue, this.Exchange.FS_EVENT, `*.${encodeRK(eventName)}.*.*.${encodeRK(domainName)}`, {}, cb);
        } catch (e) {
            log.error(e);
        }
    }

    unBindDomainEvent(eventName, domainName) {
        try {
            if (this.channel && this.customChannelQueue)
                this.channel.unbindQueue(this.customChannelQueue, this.Exchange.FS_EVENT, `*.${encodeRK(eventName)}.*.*.${encodeRK(domainName)}`)
        } catch (e) {
            log.error(e);
        }
    }

    _parseHookEventName (eventName) {
        let _e = eventName.split('->').map((value) => encodeRK(value) );

        if (~eventName.indexOf('->webitel::')) {
            return {
                ex: this.Exchange.FS_CC_EVENT,
                rk: `*.${_e[1]}.*.*.*`
            };
        } else if (~eventName.indexOf('callcenter')) {
            //FreeSWITCH-Hostname.callcenter%3A%3Ainfo.member-queue-start.kkk%4010%2E10%2E10%2E144.6e4249b7-5595-40be-a819-42b467a2843b
            return {
                ex: this.Exchange.FS_CC_EVENT,
                rk: `*.callcenter%3A%3Ainfo.${_e[1]}.*.*`
            };
        } else {
            return {
                ex: this.Exchange.FS_EVENT,
                rk: `*.${_e[0]}.${_e[1] || '*' }.*.*`
            }
        }
    };

    bindHook (event, cb) {
        try {
            if (!event)
                return cb && cb(new Error("Bad event name"));

            let opt = this._parseHookEventName(event);

            if (this.channel)
                this.channel.bindQueue(HOOK_QUEUE, opt.ex, opt.rk, {}, cb);
        } catch (e) {
            log.error(e);
        }
    };

    unBindHook (event, cb) {
        try {
            if (!event)
                return cb && cb(new Error("Bad event name"));

            let opt = this._parseHookEventName(event);

            if (this.channel)
                this.channel.unbindQueue(HOOK_QUEUE, opt.ex, opt.rk, {}, cb);
        } catch (e) {
            log.error(e);
        }
    };

    isConnect () {
        return !!this.channel;
    };

    // TODO
    disconnect () {};

    subscribe (queueName = '', params = {autoDelete: true, durable: false, exclusive: true}, handler, cb) {
        if (!this.channel)
            return cb(new Error(`No connect to amqp`));

        this.channel.assertQueue(queueName, params, (err, qok) => {
            if (err)
                return cb(err);
            this.channel.consume(qok.queue, handler, {noAck: true});
            return cb(null, qok.queue);
        });
    };

    bind (queueName, exchange, rk, cb) {
        if (!this.channel)
            return cb(new Error(`No connect to amqp`));

        if (!queueName || !exchange || !rk)
            return cb(new Error(`Bad parameters`));

        this.channel.bindQueue(queueName, exchange, rk);
        return cb()
    };

    unbind (queueName, exchange, rk, cb) {
        if (!this.channel)
            return cb(new Error(`No connect to amqp`));

        if (!queueName || !exchange || !rk)
            return cb(new Error(`Bad parameters`));

        this.channel.unbindQueue(queueName, exchange, rk, {}, cb);
    }

    init (channel) {
        let scope = this;

        async.waterfall(
            [
                // init channel event exchange
                function (cb) {
                    log.debug('Try init channel event exchange');
                    channel.assertExchange(scope.Exchange.FS_EVENT, 'topic', {durable: true}, cb);
                },
                // init custom event exchange
                function (_, cb) {
                    log.debug('Try init custom event exchange');
                    channel.assertExchange(scope.Exchange.FS_CC_EVENT, 'topic', {durable: true}, cb);
                },
                // init storage exchange
                function (_, cb) {
                    log.debug('Try init storage exchange');
                    channel.assertExchange(scope.Exchange.STORAGE_COMMANDS, 'topic', {durable: true}, cb);
                },

                //init channel queue
                function (_, cb) {
                    log.debug('Try init channel event queue');
                    channel.assertQueue('', {autoDelete: true, durable: false, exclusive: true}, (err, qok) => {
                        scope.queue = qok.queue;

                        channel.consume(scope.queue, (msg) => {
                            try {
                                let json = JSON.parse(msg.content.toString());
                                // TODO https://freeswitch.org/jira/browse/FS-8817
                                if (json['Event-Name'] == 'CUSTOM') return;
                                scope.emit('callEvent', json);
                            } catch (e) {
                                log.error(e);
                            }
                        }, {noAck: true});

                        let activeUsers = scope.app.Users.getKeys();
                        activeUsers.forEach((userName) => {
                            scope.bindChannelEvents({id: userName});
                        });
                        return cb(null, null);
                    });
                },
                //init custom channel queue
                function (_, cb) {
                    log.debug('Try init custom channel event queue');
                    channel.assertQueue('', {autoDelete: true, durable: false, exclusive: true}, (err, qok) => {
                        scope.customChannelQueue = qok.queue;
                        channel.consume(scope.customChannelQueue, (msg) => {
                            try {
                                let json = JSON.parse(msg.content.toString());
                                // TODO https://freeswitch.org/jira/browse/FS-8817
                                if (json['Event-Name'] == 'CUSTOM') return;
                                scope.emit('callDomainEvent', json);
                            } catch (e) {
                                log.error(e);
                            }
                        }, {noAck: true});
                        return cb(null, null);
                    });
                },
                // function (_, cb) {
                //     log.debug('Try init telegram events queue');
                //     channel.assertQueue(TELEGRAM_QUEUE, {autoDelete: false, durable: false, exclusive: false}, (err, qok) => {
                //         channel.bindQueue(qok.queue, scope.Exchange.FS_CC_EVENT, "*.webitel%3A%3Atelegram.*.*.*");
                //         channel.consume(qok.queue, (msg) => {
                //             try {
                //                 let e = JSON.parse(msg.content.toString());
                //
                //                 scope.emit('telegramEvent', e);
                //             } catch (e) {
                //                 log.error(e);
                //             }
                //         }, {noAck: true});
                //
                //
                //         return cb(null, null);
                //     });
                // },

                // init call center events
                function (_, cb) {
                    log.debug('Try init call center events queue');
                    channel.assertQueue('', {autoDelete: true, durable: false, exclusive: true}, (err, qok) => {

                        channel.bindQueue(qok.queue, scope.Exchange.FS_CC_EVENT, "*.callcenter%3A%3Ainfo.*.*.*");

                        channel.consume(qok.queue, (msg) => {
                            try {
                                scope.emit('ccEvent', JSON.parse(msg.content.toString()));
                            } catch (e) {
                                log.error(e);
                            }
                        }, {noAck: true});

                        return cb(null, null);
                    });
                },

                // init hooks queue
                function (_, cb) {
                    log.debug('Try init hooks events queue');
                    channel.assertQueue(HOOK_QUEUE, {autoDelete: false, durable: false, exclusive: false}, (err, qok) => {

                        channel.consume(qok.queue, (msg) => {
                            try {
                                let e = JSON.parse(msg.content.toString()),
                                    domain = getDomain(e);

                                if (!domain) {
                                    log.debug(`No found domain`, domain);
                                    log.trace(e);
                                    return;
                                }

                                scope.emit('hookEvent', e['Event-Name'], domain, e);
                            } catch (e) {
                                log.error(e);
                            }
                        }, {noAck: true});

                        scope.emit(`init:hook`);

                        return cb(null, null);
                    });
                },

                //init gateway queue
                function (_, cb) {
                    log.debug('Try init gateway events queue');
                    channel.assertQueue(GATEWAY_QUEUE, {autoDelete: false, durable: true, exclusive: false}, (err, qok) => {

                        channel.consume(qok.queue, (msg) => {
                            try {
                                gatewayService._onChannel(JSON.parse(msg.content.toString()));
                                channel.ack(msg);
                            } catch (e) {
                                log.error(e);
                            }
                        }, {noAck: false});

                        channel.bindQueue(qok.queue, scope.Exchange.FS_EVENT, `*.${encodeRK('CHANNEL_CREATE')}.*.*.*`);
                        channel.bindQueue(qok.queue, scope.Exchange.FS_EVENT, `*.${encodeRK('CHANNEL_DESTROY')}.*.*.*`);

                        return cb(null, null);
                    });
                },

                // init console event
                function (_, cb) {
                    log.debug('Try init console event queue');
                    channel.assertQueue('', {autoDelete: true, durable: false, exclusive: true}, (err, qok) => {

                        channel.bindQueue(qok.queue, scope.Exchange.FS_CC_EVENT, "*.webitel%3A%3Aaccount_status.*.*.*");
                        channel.bindQueue(qok.queue, scope.Exchange.FS_CC_EVENT, "*.webitel%3A%3Auser_create.*.*.*");
                        channel.bindQueue(qok.queue, scope.Exchange.FS_CC_EVENT, "*.webitel%3A%3Auser_destroy.*.*.*");
                        channel.bindQueue(qok.queue, scope.Exchange.FS_CC_EVENT, "*.webitel%3A%3Adomain_create.*.*.*");
                        channel.bindQueue(qok.queue, scope.Exchange.FS_CC_EVENT, "*.webitel%3A%3Adomain_destroy.*.*.*");
                        channel.bindQueue(qok.queue, scope.Exchange.FS_CC_EVENT, "*.webitel%3A%3Auser_managed.*.*.*");

                        channel.consume(qok.queue, (msg) => {
                            try {
                                let event = JSON.parse(msg.content.toString());
                                scope.emit('webitelEvent', parseConsoleEvent(event));
                            } catch (e) {
                                log.error(e);
                            }
                        }, {noAck: true});
                    });
                    return cb(null, null);
                }
                
                //init commands

                // function (_, cb) {
                //     log.debug('Try init switch commands queue response');
                //     channel.assertQueue('', {autoDelete: true, durable: false, exclusive: true}, (err, qok) => {
                //
                //         channel.bindQueue(qok.queue, scope.Exchange.FS_COMMANDS, `engine.${scope._instanceId}.#`);
                //
                //         channel.consume(qok.queue, scope[_onRequestFsApi].bind(scope), {noAck: true}, cb);
                //     });
                // }
            ],
            (err) => {
                if (err)
                    return log.error(err);
                log.info('Init AMQP: OK');

                scope.channel = channel;
                scope.emit(`init:broker`);
            });
    };

    [_onReturnedMessage] (msg) {
        if (!msg) return;

        let rk = msg.properties.headers['x-fs-api-resp-key'],
            jobId = getLastKey(rk), //rk.substring(rk.lastIndexOf('.') + 1);
            _cb = this[_stackCb][jobId]
            ;
        if (_cb)
            delete this[_stackCb][jobId];
        else log.error('bad jobId: ', jobId);

        if (typeof _cb == 'function')
            _cb({body: "-ERR FreeSWITCH not found."});

        log.error('Bad gateway');
    };

    [_onRequestFsApi] (msg) {
        try {

            let rk = msg.fields.routingKey,
                jobId = getLastKey(rk),
                _msg = msg.content.toString(),
                _cb = this[_stackCb][jobId]
                ;
            if (this[_stackCb][jobId])
                delete this[_stackCb][jobId];
            else log.error('Bad job', jobId);

            //console.log(`Count stack ${Object.keys(this[_stackCb]).length}`);
            log.trace(`Response fs-api: ${jobId}`);

            if (msg.properties.contentType == "text/json")
                _msg = JSON.parse(_msg);

            if (typeof _cb === "function")
                _cb({body: _msg.output}); //TODO new Api response

        } catch (e) {
            log.error(e);
        }
    };

    api (command, args, jobid, cb) {

        if(typeof args === 'function') {
            cb = args;
            args = '';
            jobid = null;
        }

        if(typeof jobid === 'function') {
            cb = jobid;
            jobid = null;
        }

        if (!this.channel)
            return cb & cb({body: "Channel not open"}); // TODO new response API

        args = args || '';

        if(args instanceof Array)
            args = args.join(' ');

        command += ' ' + args;

        jobid = jobid || generateUuid.v4();


        this[_stackCb][jobid] = cb;
        log.trace(`Execute fs-api: ${command} -> ${jobid}`);

        this.channel.publish(
            this.Exchange.FS_COMMANDS,
            'commandBindingKey', //TODO move config
            new Buffer(command),
            {
                headers: {
                    "x-fs-api-resp-exchange": this.Exchange.FS_COMMANDS,
                    "x-fs-api-resp-key": `engine.${this._instanceId}.${jobid}`
                },
                mandatory: true
                //deliveryMode: 2
            }
        );

    };

    show (item, format, cb) {
        if(typeof format === 'function') {
            cb = format;
            format = null;
        }

        format = format || 'json';

        this.api('show ' + item + ' as ' + format, (e) => {
            var data = e.body || "", parsed = {};

            //if error send them that
            if(data.indexOf('-ERR') !== -1) {
                if(cb) cb(new Error(data));
                return;
            }

            switch(format) {
                case 'json':
                    try { parsed = JSON.parse(data); }
                    catch(e) { if(cb) cb(e); return; }

                    if(!parsed.rows) parsed.rows = [];

                    break;

                case 'xml':
                    // TODO
                    break;

                default: //delim seperated values, custom parsing
                    if(format.indexOf('delim')) {
                        var delim = format.replace('delim ', ''),
                            lines = data.split('\n'),
                            cols = lines[0].split(delim);

                        parsed = { rowCount: lines.length - 1, rows: [] };

                        for(var i = 1, len = lines.length; i < len; ++i) {
                            var vals = lines[i].split(delim),
                                o = {};
                            for(var x = 0, xlen = vals.length; x < xlen; ++x) {
                                o[cols[x]] = vals[x];
                            }

                            parsed.rows.push(o);
                        }
                    }
                    break;
            }

            if(cb) cb(null, parsed, data);
        })
    }
}

const WEBITEL_EVENT = {
    "webitel::account_status": "ACCOUNT_STATUS",
    "webitel::user_create": "USER_CREATE",
    "webitel::user_destroy": "USER_DESTROY",
    "webitel::domain_create": "DOMAIN_CREATE",
    "webitel::domain_destroy": "DOMAIN_DESTROY",
    "webitel::user_managed": "USER_MANAGED"
};

const ALLOW_CONSOLE_HEADER = ['Account-Domain', 'Account-Role', 'Account-Status', 'Account-User', 'Account-User-State',
    'Event-Account', 'Event-Date-Timestamp', 'Event-Domain', 'Domain-Name', 'variable_customer_id', 'User-Domain', 'User-ID',
    'User-State', 'Account-Skills', 'Account-Status-Descript', 'Account-Agent-State', 'Account-Agent-Status', 'variable_skills'];

function parseConsoleEvent (e) {
    let event = {
        "Event-Name": WEBITEL_EVENT[e['Event-Subclass']]
    };

    for (let h of ALLOW_CONSOLE_HEADER) {
        if (e.hasOwnProperty(h))
            event[h] = e[h];
    }

    return event
}

function encodeRK (rk) {
    try {
        if (rk)
            return encodeURIComponent(rk)
                .replace(/\./g, '%2E')
                .replace(/\:/g, '%3A')
    } catch(e) {
        log.error(e);
        return null;
    }
}

function getPresenceRoutingFromCaller (caller) {
    try {
        let callerId = encodeRK(caller.id);
        return `*.*..${callerId}.*`;
    } catch (e) {
        log.error(e);
        return null;
    }
}

function getDomain (data) {
    if (!data)
        return null;

    if (data.variable_domain_name)
        return data.variable_domain_name;

    if (data.variable_w_domain)
        return data.variable_w_domain;

    if (data['Channel-Presence-ID'])
        return data['Channel-Presence-ID'].substring(data['Channel-Presence-ID'].indexOf('@') + 1);

    if (data['variable_presence_id'])
        return data['variable_presence_id'].substring(data['variable_presence_id'].indexOf('@') + 1);
}

function getLastKey (rk) {
    let arr = (rk || "").split('.');
    return arr[arr.length - 1];
}

module.exports = WebitelAmqp;