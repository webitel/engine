'use strict';

var EventEmitter2 = require('eventemitter2').EventEmitter2,
    util = require('util'),
    Parser = require('./Parser'),
    log = require('../log')(module),
    net = require('net'),
    PERMISSION_DENIED = '-ERR permission denied!',
    ACCOUNT_ROLE = require('../../const').ACCOUNT_ROLE,
    Controller = require(__appRoot + '/services/dialplan'), //{} ,//require('../mod/dialplan/controller'),
    authServices = require(__appRoot + '/services/auth'),
    async = require('async'),
    CodeError = require(__appRoot + '/lib/error'),
    Event = require('./Event'),
    COMMAND_TYPES = require('../../const').WebitelCommandTypes;

var VARIABLE_EXTENSION_NAME = 'webitel-extensions';

var Webitel = module.exports = function (parameters) {
    EventEmitter2.call(this, {
        wildcard: true,
        delimiter: '::',
        maxListeners: 25
    });
    this.host = parameters['server'];
    this.port = parameters['port'];
    this.account = parameters['account'];
    this.password = parameters['secret'];
    this.authed = false;

    this.reconnect = parameters['reconnect'] || -1;
    this._status = ConnectionStatus.Disconnected;
    this.socket = null;
    this._parser = null;
    this.cmdCallbackQueue = [];
    this.apiCallbackQueue = [];
};

util.inherits(Webitel, EventEmitter2);

Webitel.prototype.version = '3.2.0';

Webitel.prototype.connect = function () {
    log.trace("Host: " + this.host);
    var self = this;
    // TODO нужно тестировать будет ли ответ когда упадет соединение
    this.cmdCallbackQueue.length = 0;
    this.apiCallbackQueue.length = 0;

    this.socket = net.connect({
        port: this.port,
        host: this.host
    }, this._onConnect.bind(this));

    this.socket.on('close', this._onSocketClose.bind(this));

    this.socket.on('error', this._onSocketError.bind(this));

    this.socket.on('end', function() {
        self._status = ConnectionStatus.Disconnected;
        self.emit('webitel::end');
        self.socket = null;
    });

    var self = this;

    this.on('webitel::event::command::reply', function() {
        if(self.cmdCallbackQueue.length === 0) return;

        var fn = self.cmdCallbackQueue.shift();

        if(fn && typeof fn === 'function')
            fn.apply(self, arguments);
    });

    this.on('webitel::event::api::response', function() {
        if(self.apiCallbackQueue.length === 0) return;

        var fn = self.apiCallbackQueue.shift();

        if(fn && typeof fn === 'function')
            fn.apply(self, arguments);
    });

};

Webitel.prototype._onSocketClose = function () {
    this.emit('webitel::socket::close', new Error('socket close!'));
};

Webitel.prototype.send = function(command, args) {
    var self = this,
        cmdStr;
    var _command = [];
    try {
        _command.push(command);
        for (var key in args) {
            _command.push(args[key]);
        };
        cmdStr = _command.join(' ');

        self.socket.write(cmdStr + '\n\n');
        log.debug('Execute: ' + cmdStr);
    }
    catch(e) {
        self.emit('error', e);
    }
};

Webitel.prototype.sendRecv = function(command, args, cb) {
    if(typeof args === 'function') {
        cb = args;
        args = null;
    }

    //queue callback for command reply
    this.cmdCallbackQueue.push(cb);

    this.send(command, args);
};

Webitel.prototype._onError = function(err) {
    this.emit('error', err);
};

Webitel.prototype._onSocketError = function (err) {
    this.emit('error::socket', err);
};

Webitel.prototype._onConnect = function() {
    this.parser = new Parser(this.socket);

    this.parser.on('webitel::event', this._onEvent.bind(this));

    //on parser error
    this.parser.on('error', this._onError.bind(this));

    //emit that we conencted
    this.emit('webitel::connect');

    //wait for auth request
    this.on('webitel::event::auth::request', this.auth.bind(this));
};

Webitel.prototype.auth = function (cb) {
    var self = this;

    //send auth command
    self.sendRecv('auth ' + self.account + ' ' + self.password, function(evt) {
        if(evt.getHeader('Webitel-Reply-OK')) {
            self.authed = true;
            self._status = ConnectionStatus.Connected;
//            self.subscribe(self.reqEvents);

            self.emit('webitel::event::auth::success', evt);
            self.emit('webitel::ready');

            if(cb && typeof cb === 'function') cb(null, evt);
        } else {
            self._status = ConnectionStatus.Disconnected;
            self.authed = false;
            self.emit('webitel::event::auth::fail', evt);

            if(cb && typeof cb === 'function') cb(new Error('Authentication Failed'), evt);
        };
    });
};

Webitel.prototype.disconnect = function () {
    if (this.socket) {
        this.send('exit');
        this.socket.end();
        this.socket = null;
    }
}

Webitel.prototype.api = function (command, args, cb) {
    if(typeof args === 'function') {
        cb = args;
        args = '';
    }
    if (!this.socket || this.socket.destroyed || this._status !== ConnectionStatus.Connected)
        return cb(new Event({}, "-ERR: Console connect error"));

    if(args instanceof Array)
        args = args.join(' ');

    args = (args ? ' ' + args : '');

    //queue callback for api response
    this.apiCallbackQueue.push(cb);

    this.send('api ' + command + args);
};

Webitel.prototype.subscribe = function (param) {
    try {
        this.sendRecv('event json ' + (param instanceof Array ? param.join(' ') : param), function (res) {
            if (res.getHeader('Webitel-Reply-OK')) {
                log.debug(res.getHeader('Reply-Text'));
            } else if (res.getHeader('Webitel-Reply-ERR')) {
                log.error(res.getHeader('Reply-Text'));
            } else {
                log.error('Subscribe Webitel error!');
            }
        });
    } catch (e) {
        log.error(e.message);
    }
};

Webitel.prototype._onEvent = function(event, headers, body) {
    var emit = 'webitel::event';

    switch(headers['Content-Type']) {
        case 'auth/request':
            emit += '::auth::request';
            break;

        case 'command/reply':
            emit += '::command::reply';
            break;

        case 'text/disconnect-notice':
            emit += '::disconnect::notice';
            break;

        case 'api/response':
            emit += '::api::response';
            break;

        case 'text/event-json':
        case 'text/event-plain':
        case 'text/event-xml':
            emit += '::event::' + event.getHeader('Event-Name');
            break;

        default:
            emit += '::raw::' + headers['Content-Type'];
    }

    this.emit(emit, event, headers, body);
};

var ConnectionStatus = {
    Connected: 1,
    Disconnected: 2
};

Webitel.prototype.domainCreate = function(_caller, name, customerId, option, cb) {
    var param = '';
    if (option && option['parameters'] instanceof Array) {
        param += '[' + option['parameters'].join(',') + ']';
    };
    if (option && option['variables'] instanceof Array) {
        param += '{' + option['variables'].join(',') + '}';
    };

    this.api(WebitelCommandTypes.Domain.Create, [
        param + name,
        customerId || ''
    ], function (res) {
        var err = checkBodyError(res);
        if (err) {
            cb(err);
            return;
        };

        cb(null, res.body);
    });
    /*var command = new WebitelCommand(WebitelCommandTypes.Domain.Create, {
        name: '\"' + name + '\"',
        customerId: customerId
    }, cb);
    command.execute(); */
};

Webitel.prototype.domainItem = function (_caller, name, cb) {
    this.api(WebitelCommandTypes.Domain.Item, [
        name || ''
    ], function (res) {
        var err = checkBodyError(res);
        if (err) {
            cb(err);
            return;
        };

        cb(null, res.body);
    });
};

Webitel.prototype.domainList = function(_caller, customerId, cb) {

    this.api(WebitelCommandTypes.Domain.List, [
        customerId || ''
    ], function (res) {
        var err = checkBodyError(res);
        if (err) {
            cb(err);
            return;
        };

        cb(null, res.body);
    });
    /*
    var command = new WebitelCommand(WebitelCommandTypes.Domain.List, {
        customerId: _customerId
    }, _cb);
    command.execute();
    */
};

Webitel.prototype.domainRemove = function(_caller, name, cb) {
    this.api(WebitelCommandTypes.Domain.Remove, [
        name || ''
    ], function (res) {
        var err = checkBodyError(res);
        if (err) {
            cb(err);
            return;
        };

        cb(null, res.body);
    });
    /*var command = new WebitelCommand(WebitelCommandTypes.Domain.Remove, {
        name: name
    }, cb);
    command.execute(); */
};

Webitel.prototype.updateDomain = function(_caller, name, option, cb) {
   var param = ''
        ;

    if ('var param'.indexOf(option['type']) === -1) {
        cb(new CodeError(400, "Bad update type."));
        return;
    };

    param += ' ' + option['type'] + ' ';
    option['params'].forEach(function(item, index) {
        if (index > 0)
            param += ',';
        if (item instanceof Object) {
            param += item['key'] + '=' + (item['value'] || '');
        } else {
            param += item;
        };
    });

    this.api(WebitelCommandTypes.Domain.Item, [
        name || '',
        param
    ], function (res) {
        var err = checkBodyError(res);
        if (err) {
            cb(err);
            return;
        };

        cb(null, res.body);
    });
};

Webitel.prototype.list_users = function(_caller, domain, cb, format) {
    // Для ивентов, чтобы заполнить online
    this.api(WebitelCommandTypes.ListUsers, [
        domain || ''
    ], function (res) {
        var err = checkBodyError(res);
        if (err) {
            cb(err);
            return;
        };

        cb(null, res.body);
    });

    /*var cmd = new WebitelCommand(WebitelCommandTypes.ListUsers, {
        param: _domain
    }, _cb);
    cmd.execute();*/
};

Webitel.prototype.userList = function(_caller, domain, cb) {
    this.api(WebitelCommandTypes.Account.List, [
        domain || ''
    ], function (res) {
        var err = checkBodyError(res);
        if (err) {
            cb(err);
            return;
        }

        cb(null, res.body);
    });
/*
    var cmd = new WebitelCommand(WebitelCommandTypes.Account.List, {
        param: _domain
    }, _cb);
    cmd.execute();*/
};

// TODO mod_cc

Webitel.prototype.userCreate = function(_caller, args, cb) {
    try {
        var _id = args['param'] || '',
            role = '',
            _str = '',
            parameters = [],
            extensions,
            variables,
            scope = this
            ;

        if (args['attribute'] instanceof Object) {
            parameters = args.attribute['parameters'] || [];
            variables = args.attribute['variables'];
        }
        ;

        var _domain = _id.split('@')[1];

        role = (args['role'] instanceof Array ? args['role'].join(',') : args['role'])  + ',webrtc';

        var _ext;
        for (var key in parameters) {
            _ext = parameters[key];
            if (_ext && _ext.indexOf(VARIABLE_EXTENSION_NAME + '=') == 0) {
                extensions = _ext.replace(VARIABLE_EXTENSION_NAME + '=', '');
                break;
            }
        };

        var _refUser = _id.split(/\:|@/)[0];
        var number = extensions || _refUser;

        // TODO возможность задавать масив номера для пользователя
        if (typeof number !== 'string') {
            return cb(new CodeError(400, "bad request (number)"));
        }
        ;
        parameters.push(VARIABLE_EXTENSION_NAME + '=' + number);

        if (parameters instanceof Array) {
            _str += '[' + parameters + ']';
        }
        ;
        if (variables  instanceof Array) {
            _str += '{' + variables + '}';
        }
        ;

        role = _str + role;

        Controller._existsExtension(number, _domain, function (err, exists) {
            try {
                if (err) {
                    return cb(err);
                }
                ;
                if (exists) {
                    log.debug('Add number: %s in extension collection exists.', number);
                    return cb(new CodeError(400, "Number exists"));
                }
                ;

                scope.api(WebitelCommandTypes.Account.Create, [
                    role,
                    _id
                ], function (res) {
                    var err = checkBodyError(res);
                    if (err) {
                        cb(err);
                        return;
                    };

                    cb(null, res.body);
                });

            } catch (e) {
                log.error(e['message'])
            }
        });
    } catch (e) {
        cb(e);
        log.error(e['message']);
    };

   /* var cmd = new WebitelCommand(WebitelCommandTypes.Account.Create, {
        role: role,
        param: _param
    }, cb);
    cmd.execute(); */
};


Webitel.prototype.userUpdateV2 = function (_caller, user, domain, option, cb) {
    try {
        var _domain = domain;

        var params = option['parameters'];
        var variables = option['variables'];
        var extensions;
        var scope = this;
        var roleName;

        if (!_domain || ( !(params instanceof Array) && !(variables instanceof Array) )) {
            cb(new CodeError(400, "Bad request."));
            return;
        };
        
        var setParamsOrVars = function (callback) {
            var cmd = '';

            if (params instanceof Array && params.length > 0) {
                cmd += '[' + params + ']';
            };

            if (variables instanceof Array && variables.length > 0) {
                cmd += '{' + variables + '}';
            };

            scope.api(WebitelCommandTypes.Account.Change, [
                user + '@' + _domain,
                cmd
            ], function (res) {

                if (res && res['body'] && res['body'].indexOf('+OK') > 0) {
                    scope._parsePlainCollectionToJSON(res['body'], function (err, resJSON) {
                        if (err) {
                            log.error(err);
                            callback(err);
                            return;
                        };
                        callback(null, resJSON);
                    });
                } else {
                    callback(new Error(res['body']));
                };
            });
        };
        
        var setExtensions = function (callback) {
            Controller._existsExtension(extensions, _domain, function (err, exists) {
                if (err) {
                    callback(err);
                }
                ;
                if (exists) {
                    log.debug('Add number: %s in extension collection exists.', user);
                    callback(new Error('-ERR Number reserved.'));
                    return;
                }
                ;

                Controller._updateOrInsertExtension(user, extensions, _domain, function (err) {
                    if (err) {
                        callback(err);
                        return log.error(err['message']);
                    };
                    callback(null, 'User update DB: ' + user);
                    log.debug('User update DB: %s', user);
                });
            });
        };

        var resetToken = function (callback) {
            try {
                log.debug('Reset token');
                authServices.removeFromUserName(user + '@' + _domain, _domain, callback);
                if (typeof roleName == 'string') {
                    let _user = application.Users.get(user + '@' + _domain);
                    if (_user) {
                        _user.changeRole(roleName.replace(/'/g, ''));
                    };
                }
            } catch (e) {
                log.error(e);
            };
        };

        var task = [];

        let _resetToken = false;

        if (params instanceof Array) {

            for (let item of params) {
                // if (_pushExt && _pushPass) break;

                if (/^password=/.test(item)) {
                    _resetToken = true;
                } else if (/^webitel-extensions=/.test(item)) {
                    extensions = item.replace(VARIABLE_EXTENSION_NAME + '=', '').replace(/\D/g, '');
                    if (extensions == '') {
                        cb(new CodeError(400, "Bad request (webitel-extensions)."));
                        return;
                    }
                    task.push(setExtensions);
                } else if (/^role=/.test(item)) {
                    _resetToken = true;
                }
            }
        }

        if (variables instanceof Array) {
            for (let item of variables) {
                if (/^account_role=/.test(item)) {
                    roleName = item.replace('account_role=', '');
                    _resetToken = true;
                    break;
                }
            }
        }

        if (_resetToken) {
            if (_caller.id === `${user}@${domain}`) {
                return cb(new CodeError(400, `Woow! Slow down!`))
            }
            task.push(resetToken);
        }


        if (variables || params) {
            task.push(setParamsOrVars);
        };

        async.series(task, function (err, res) {
            if (err) {
                cb(err);
                return;
            };

            cb(null, (res[0] instanceof Object) ? res[0] : res[1])
        });

    } catch (e) {
        cb(e);
    };
};

Webitel.prototype.setAccountStatus = function (id, status, cb) {
    if (!id || !status)
        return cb && cb(new Error('Bad parameters'));

    this.api(WebitelCommandTypes.Account.Change, [
        id,
        `[status=${status}]`
    ], cb);
};

Webitel.prototype.userUpdate = function(_caller, user, paramName, paramValue, cb) {
    var _domain = user.split('@')[1];

    if (!_caller || (_caller['attr']['role'].val < COMMAND_TYPES.Account.Change.perm ||
        ((_caller['attr']['domain'] != _domain || (_caller['attr']['role'].val == ACCOUNT_ROLE.USER.val &&
            user != _caller['id']))&& _caller['attr']['role'].val != ACCOUNT_ROLE.ROOT.val))) {
        cb({
            body: PERMISSION_DENIED
        });
        return;
    };

    if (paramName == '' || paramValue == '') {
        cb({
            body: '-ERR Bad request!'
        });
        return;
    };

    if (paramName == 'role' && user == _caller['id']) {
        cb({
            body: '-ERR Woow! Slow down!' // (c) srg
        });
        return
    } else if (paramName == 'extensions') {
        Controller.existsNumber(paramValue, _domain, function (err, exists) {
            if (err) {
                return cb({
                    "body": "-ERR: " + err['message']
                });
            }
            ;
            if (exists) {
                log.debug('Add number: %s in extension collection exists.', user);
                return cb({
                    "body": "-ERR: Number reserved."
                });
            }
            ;

            Controller.updateOrInsertNumber(user.split('@')[0], paramValue, _domain, function (err) {
                if (err) {
                    cb({
                        body: '-ERR: ' + err['message']
                    });
                    return log.error(err['message']);
                };
                cb({
                    "body": "+OK: User updated."
                });
                log.debug('User update DB: %s', user);
            });
        });

        return;
    };

    this.api(WebitelCommandTypes.Account.Change, [
            user || '',
            (paramName || '') + '=' + (paramValue || '')
    ], cb);
    /*var cmd = new WebitelCommand(WebitelCommandTypes.Account.Change, {
        user: user,
        param: paramName,
        value: paramValue
    }, cb);
    cmd.execute(); */
};

Webitel.prototype.userRemove = function(_caller, user, cb) {

    this.api(WebitelCommandTypes.Account.Remove, [
            user || ''
    ], function (res) {
        var err = checkBodyError(res);
        if (err) {
            cb(err);
            return;
        };

        cb(null, res.body);
    });
};

Webitel.prototype.userItem = function (_caller, user, domain, cb) {
    var _id = user + '@' + domain;

    this.api(WebitelCommandTypes.Account.Change, [
        _id
    ], function (res) {
        var err = checkBodyError(res);
        if (err) {
            cb(err);
            return;
        };

        cb(null, res.body);
    });
};

Webitel.prototype.queueList = function (_caller, args, cb) {
    args = args || {};
    var _domain = args['domain'];


    if (!_domain || typeof _domain !== 'string') {
        cb({
            body: "-ERR Bad request: domain is required!"
        });
        return;
    };

    var _params = WebitelCommandTypes.CallCenter.List + ' '.concat('@', _domain),
        self = this
        ;

    this.api(WebitelCommandTypes.CallCenter.Root, [
        _params
    ], function (res) {
        var err = checkBodyError(res);
        if (err) {
            cb(err);
            return;
        };
        self._parsePlainTableToJSONArray(res.getBody(), function (err, resJSON) {
            if (err) {
                log.error(err);
                cb(err);
                return;
            }
            return cb(null, resJSON);
        });
    });

};

Webitel.prototype.queueCreate = function (_caller, args, cb) {

    var _params = WebitelCommandTypes.CallCenter.Create + ' '
        ;

    if (args.hasOwnProperty('params') && args['params'] instanceof Array) {
        _params = _params.concat('[', args['params'].join(','), ']');
    };
    _params = _params.concat(args['name'] + '@' + args['domain']);

    this.api(WebitelCommandTypes.CallCenter.Root, [_params], function (res) {
        var body = res.getBody() || '';
        var err = checkBodyError(res);
        if (err) {
            cb(err);
            return;
        };
        return cb(null, body);
    });
};

Webitel.prototype.queueItem = function (_caller, args, cb) {

    var _params = 'queue ' + `'${args['name']}` + '@' + args['domain'] + '\'',
        self = this
        ;
    this.api(WebitelCommandTypes.CallCenter.Root, [_params], function (res) {
        var err = checkBodyError(res);
        if (err) {
            cb(err);
            return;
        };

        self._parsePlainCollectionToJSON(res.getBody(), function (err, resJSON) {
            if (err) {
                log.error(err);
                cb(err);
                return;
            }
            cb(null, resJSON);
        });
    });
};

Webitel.prototype.queueUpdateItem = function (_caller, args, cb) {
    var _params = 'queue ' + args['name'] + '@' + args.domain + ' ',
        self = this
        ;

    for (var key in args['params']) {
        _params = _params.concat(key, '=' , args['params'][key] instanceof Object
                ? ''
                : String(args['params'][key])
            , ',');
    };

    this.api(WebitelCommandTypes.CallCenter.Root, [_params], function (res) {
        var err = checkBodyError(res);
        if (err) {
            cb(err);
            return;
        };

        self._parsePlainCollectionToJSON(res.getBody(), function (err, resJSON) {
            if (err) {
                log.error(err);
                cb(err);
                return;
            }
            cb(null, resJSON);
        });
    });
};

Webitel.prototype.queueDelete = function (_caller, args, cb) {
    var _params = WebitelCommandTypes.CallCenter.Delete + ' ' + args['name'] + '@' + args['domain']
        ;
    this.api(WebitelCommandTypes.CallCenter.Root, [_params], function (res) {
        if (res['body'].indexOf('-ERR') == 0) {
            cb(new CodeError(500, res.body));
            return;
        };
        cb(null, res.body);
    });
};

Webitel.prototype.queueUpdateItemState = function (_caller, args, cb) {
    if (['enable', 'disable'].indexOf(args['state']) == -1) {
        cb(new CodeError(400, "-ERR Bad request: state " + args['state'] + ' inexplicable!'));
        return;
    };

    var _params = 'queue ' + args['state'] + ' ' + args['name'] + '@' + args['domain']
        ;
    this.api(WebitelCommandTypes.CallCenter.Root, [_params], function (res) {
        var err = checkBodyError(res);
        if (err) {
            cb(err);
            return;
        };

        cb(null, res.body);
    });
};

Webitel.prototype.tierList = function (_caller, args, cb) {
        /// callcenter_config tiers list @domain
    var _params = [
        'tier list',
        (args.type == 'agent' ? 'agent' : 'queue') + ':' + args['data'] + '@' + args['domain']
    ];

    this.api(WebitelCommandTypes.CallCenter.Root, _params, function (res) {
        var err = checkBodyError(res);
        if (err) {
            cb(err);
            return;
        };

        cb(null, res.body);
    });
};

Webitel.prototype.tierCreate = function (_caller, args, cb) {

    var _params = 'tier add '.concat(args['queue'], '@', args['domain'], ' ', args['agent'],
        '@', args['domain']);
    if (args['level'])
        _params += ' ' + args['level']
        ;

    if (args['position'])
        _params += ' ' + args['position']
        ;
    this.api(WebitelCommandTypes.CallCenter.Root, [_params], function (res) {
        var err = checkBodyError(res);
        if (err) {
            cb(err);
            return;
        };

        cb(null, res.body);
    });
};

Webitel.prototype.tierSetLvl = function (_caller, args, cb) {

    var _params = 'tier set level '.concat(args['queue'], '@', args['domain'], ' ', args['agent'],
        '@', args['domain'], ' ', args['level']);

    this.api(WebitelCommandTypes.CallCenter.Root, [_params], function (res) {
        var err = checkBodyError(res);
        if (err) {
            cb(err);
            return;
        };

        cb(null, res.body);
    });
};

Webitel.prototype.tierSetPos = function (_caller, args, cb) {
    var _params = 'tier set position '.concat(args['queue'], '@', args['domain'], ' ', args['agent'],
        '@', args['domain'], ' ', args['position']);

    this.api(WebitelCommandTypes.CallCenter.Root, [_params], function (res) {
        var err = checkBodyError(res);
        if (err) {
            cb(err);
            return;
        };

        cb(null, res.body);
    });
};

Webitel.prototype.tierRemove = function (_caller, args, cb) {
    var _params = 'tier del '.concat(args['queue'], '@', args['domain'], ' ', args['agent'],
        '@', args['domain']);

    this.api(WebitelCommandTypes.CallCenter.Root, [_params], function (res) {
        var err = checkBodyError(res);
        if (err) {
            cb(err);
            return;
        };

        cb(null, res.body);
    });
};

// TODO mod_cc END

Webitel.prototype.deviceList = function(_caller, domain, cb) {
    var _cb, _domain;
    if (typeof arguments[1] == "function") {
        _cb = arguments[1];
        _domain = null
    } else {
        _cb = cb;
        _domain = domain || _caller['attr']['domain'];
    };

    if (!_caller || (_caller['attr']['role'].val < COMMAND_TYPES.Device.List.perm ||
        (_caller['attr']['domain'] != _domain && _caller['attr']['role'].val != ACCOUNT_ROLE.ROOT.val))) {
        cb({
            body: PERMISSION_DENIED
        });
        return;
    };

    this.api(WebitelCommandTypes.Device.List, [
            _domain || ''
    ], _cb);
    /*
    var cmd = new WebitelCommand(WebitelCommandTypes.Device.List, {
        param: _domain
    }, _cb);
    cmd.execute();*/
};

Webitel.prototype.deviceCreate = function(_caller, type, _param, cb) {
    _param = _param || '';
    var _domain = _param.split('@')[1];

    if (!_caller || (_caller['attr']['role'].val < COMMAND_TYPES.Device.Create.perm ||
        (_caller['attr']['domain'] != _domain && _caller['attr']['role'].val != ACCOUNT_ROLE.ROOT.val))) {
        cb({
            body: PERMISSION_DENIED
        });
        return;
    }
    this.api(WebitelCommandTypes.Device.Create, [
            type || type,
            _param
    ], cb);

    /*
    var cmd = new WebitelCommand(WebitelCommandTypes.Device.Create, {
        type: type,
        param: param
    }, cb);
    cmd.execute(); */
};

Webitel.prototype.deviceUpdate = function(_caller, device, paramName, paramValue, cb) {

    var _domain = device.split('@')[1];
    // TODO юзер можеть только себе параметры менять, админ в домене, рут у всех!!!!
    if (!_caller || (_caller['attr']['role'].val < COMMAND_TYPES.Device.Change.perm ||
        (_caller['attr']['domain'] != _domain && _caller['attr']['role'].val != ACCOUNT_ROLE.ROOT.val))) {
        cb({
            body: PERMISSION_DENIED
        });
        return;
    };
    this.api(WebitelCommandTypes.Device.Change, [
            device || '',
            paramName || '',
            paramValue || ''
    ], cb);

    /*
    var cmd = new WebitelCommand(WebitelCommandTypes.Device.Change, {
        device: device,
        param: paramName,
        value: paramValue
    }, cb);
    cmd.execute();*/
};

Webitel.prototype.deviceRemove = function(_caller, device, cb) {
    var _domain = device.split('@')[1];
    if (!_caller || (_caller['attr']['role'].val < COMMAND_TYPES.Device.Remove.perm ||
        (_caller['attr']['domain'] != _domain && _caller['attr']['role'].val != ACCOUNT_ROLE.ROOT.val))) {
        cb({
            body: PERMISSION_DENIED
        });
        return;
    };
    this.api(WebitelCommandTypes.Device.Remove, [
            device || ''
    ], cb);

    /*
    var cmd = new WebitelCommand(WebitelCommandTypes.Device.Remove, {
        device: device
    }, cb);
    cmd.execute();*/
};

Webitel.prototype.whoami = function (_caller, cb) {
    // TODO вернуть _caller
    this.api(WebitelCommandTypes.Whoami, cb);
   /* var that = WebitelConnection;
    var cmd = new WebitelCommand(WebitelCommandTypes.Whoami, {
    }, function (res) {
        if (res.status == WebitelCommandResponseTypes.Success) {
            cb(that.parseCurrentAccount(res.response))
        };
    });
    cmd.execute(); */
};

Webitel.prototype.reloadAgents = function (_caller, cb) {
    this.api(WebitelCommandTypes.ReloadAgents, cb);
    /*
    var command = new WebitelCommand(WebitelCommandTypes.ReloadAgents, {}, cb);
    command.execute();*/
};

Webitel.prototype.reloadXml = function (_caller, cb) {
    this.api(WebitelCommandTypes.ReloadXml, (res) => {
        var err = checkBodyError(res);
        if (err) {
            cb(err);
            return;
        };

        cb(null, res.body);
    });
};

Webitel.prototype.showSipGateway = function (_caller, domain, cb) {
    // api sip_gateway
    var _t = '';
    if (domain) {
        _t = '@' + domain
    };
    this.api(WebitelCommandTypes.Gateway.Index, [
        _t
    ], function (res) {
        var err = checkBodyError(res);
        if (err) {
            cb(err);
            return;
        };

        cb(null, res.body);
    });
};

Webitel.prototype.createSipGateway = function (_caller, gateway, cb) {
    if (typeof gateway !== 'object' || !gateway['name'] || typeof gateway['username'] !== 'string' ||
        /\s/g.test(gateway['username'])) {
        cb(new CodeError(500, '-ERR Invalid arguments'));
        return;
    };

    var _domain = gateway['domain'],
        _params = gateway['params'],
        _var = gateway['var'],
        _ivar = gateway['ivar'],
        _ovar = gateway['ovar'],
        _commandsLine = 'create ';

    if (_params instanceof Array) {
        _commandsLine = _commandsLine.concat('[');
        _commandsLine = parseArrayToCommandLine(_params, _commandsLine);
        _commandsLine = _commandsLine.concat(']');
    };

    _commandsLine = _commandsLine.concat('{');

    if (_var instanceof Array) {
        _commandsLine = parseArrayToCommandLine(_var, _commandsLine);
    };

    if (_ivar instanceof Array) {
        _commandsLine = parseArrayToCommandLine(_ivar, _commandsLine, 'inbound');
    };

    if (_ovar instanceof Array) {
        _commandsLine = parseArrayToCommandLine(_ovar, _commandsLine, 'outbound');
    };
    _commandsLine = _commandsLine.concat('}');

    if (typeof gateway['template'] == 'string' && gateway['template'] != '') {
        _commandsLine = _commandsLine.concat(gateway['template'].replace(/\s/g,''), "::");
    };

    _commandsLine = _commandsLine.concat(gateway['name']);
    if (_domain) {
        _commandsLine = _commandsLine.concat('@',_domain.replace(/\s/g,''));
    };
    _commandsLine = _commandsLine.concat(' ', gateway['username']);

    if (typeof gateway['password'] == 'string' && gateway['password'] != '') {
        _commandsLine = _commandsLine.concat(':', gateway['password'].replace(/\s/g,''));
    };

    if (typeof gateway['realm'] == 'string' && gateway['realm'] != '') {
        _commandsLine = _commandsLine.concat('@', gateway['realm'].replace(/\s/g,''));
    };

    if (typeof gateway['profile'] == 'string' && gateway['profile']) {
        _commandsLine = _commandsLine.concat(' ', 'up ', gateway['profile']);
    };

    this.api(WebitelCommandTypes.Gateway.Index, [
        _commandsLine
    ], function (res) {
            var err = checkBodyError(res);
            if (err) {
                cb(err);
                return;
            };

        cb(null, res.body);
    }
    );
};

Webitel.prototype.gatewayVars = function (id, varName, cb) {
    this.api(WebitelCommandTypes.Gateway.Index, [
        id,
        varName
    ], function (res) {
        var err = checkBodyError(res);
        if (err) {
            cb(err);
            return;
        };

        cb(null, res.body);
    });
};

Webitel.prototype.changeSipGateway = function (_caller, gateway_id, type, params, cb) {
    var _cl = gateway_id + ' ';
    switch (type) {
        case WebitelGatevayTypeAttribute.VAR:
            _cl = _cl.concat(WebitelGatevayTypeAttribute.VAR);
            break;

        case WebitelGatevayTypeAttribute.IVAR:
            _cl = _cl.concat(WebitelGatevayTypeAttribute.IVAR);
            break;
        case WebitelGatevayTypeAttribute.OVAR:
            _cl = _cl.concat(WebitelGatevayTypeAttribute.OVAR);
            break;
    };

    if (params instanceof Array) {
        _cl = _cl.concat(' ');
        for (var i = 0, len = params.length; i < len; i++) {
            if (params[i]['name'])
                _cl = _cl.concat(params[i]['name'], '=');
            if (params[i]['value'])
                _cl = _cl.concat(params[i]['value']);
            _cl = _cl.concat(',');
        };
    };
    this.api(WebitelCommandTypes.Gateway.Index, [
        _cl
    ], function (res) {
        var err = checkBodyError(res);
        if (err) {
            cb(err);
            return;
        };

        cb(null, res.body);
    });
};

Webitel.prototype.removeSipGateway = function (_caller, gateway_id, cb) {
    this.api(WebitelCommandTypes.Gateway.Index, [
        'remove ' + gateway_id
    ], function (res) {
        var err = checkBodyError(res);
        if (err) {
            cb(err);
            return;
        };

        cb(null, res.body);
    });
};

Webitel.prototype.upSipGateway = function (_caller, gateway_id, profile, cb) {
    if (!gateway_id)
        return cb(new CodeError(400, 'Gateway name is required.'));

    this.api(WebitelCommandTypes.Gateway.Index, [
        gateway_id + ' up ' + (profile || 'external')
    ], function (res) {
        var err = checkBodyError(res);
        if (err) {
            cb(err);
            return;
        };

        cb(null, res.body);
    });
};

Webitel.prototype.downSipGateway = function (_caller, gateway_id, cb) {
    if (!gateway_id)
        return cb(new CodeError(400, 'Gateway name is required.'));

    this.api(WebitelCommandTypes.Gateway.Index, [
        gateway_id + ' down'
    ], function (res) {
        var err = checkBodyError(res);
        if (err) {
            cb(err);
            return;
        };

        cb(null, res.body);
    });
};


Webitel.prototype.userDara = function (userId, type, paramsArray, cb) {
    this.api(WebitelCommandTypes.UserData, [].concat(userId, type || 'global', paramsArray.join(','), 'as json'), cb);
};

Webitel.prototype._getServerId = function (cb) {
    this.api('global_getvar server_uuid', (res) => {
        var err = checkBodyError(res);
        if (err) {
            cb(err);
            return;
        };

        cb(null, res.body);
    });
};

function parseArrayToCommandLine (_arr, _cl, direction) {
    var _d = direction
        ? '[direction=' + direction + ']'
        : '';

    for (var i = 0, len = _arr.length; i < len; i++) {
        if (!_arr[i]['name']) continue;

        _cl = _cl.concat(_arr[i]['name'] + _d + '=');
        if (_arr[i]['value'] || typeof _arr[i]['value'] == "boolean")
            _cl = _cl.concat(_arr[i]['value']);
        _cl = _cl.concat(',');
    };
    return _cl;
};

var WebitelCommandTypes = {

// COMMANDS
    Auth: 'auth',
    Event: 'event',
    NixEvent: 'nixevent',

// API
    AgentList: 'account_list',
    Domain: {
        List: 'domain list',
        Create: 'domain create',
        Remove: 'domain remove',
        Item: 'domain '
    },
    Account: {
        List: 'account list', //
        Create: 'account create',
        Change: 'account  ',
        Remove: 'account remove'
    },
    Device: {
        List: 'device list',
        Create: 'device create',
        Change: 'device change',
        Remove: 'device remove'
    },
    ListUsers: 'list_users',
    Whoami: 'api whoami',
    ReloadXml: 'reloadxml',

    ReloadAgents: 'favbet reload agents',

    Gateway: {
        Index: "sip_gateway"
    },
    UserData: 'user_data',

    CallCenter: {
        Root: 'callcenter_config',
        List: 'queue list',
        Create: 'queue create',
        Delete: 'queue delete',
        Enable: 'enable',
        Disable: 'disable'
    }
};

var WebitelGatevayTypeAttribute = {
    PARAM: 'param',
    VAR: 'var',
    OVAR: 'ovar',
    IVAR: 'ivar'
};

Webitel.prototype.doSendCommand = function (res) {
    if (!webitel.authed) {
        try {
            res.writeHead(500, {'Content-Type': 'text/plain'});
            res.write("Error: Webitel server disconnect!");
            res.end();
            return false;
        } catch (e) {
            log.warn('Write message:', e.message);
            return false;
        };
    };
    return true;
};

Webitel.prototype.doSendCommandV2 = function (res) {
    if (!webitel.authed) {
        try {
            res.status(500).json({
                "status": "error",
                "info": "Error: Webitel server disconnect!"
            });
            return false;
        } catch (e) {
            log.warn('Write message:', e.message);
            return false;
        };
    };
    return true;
};

/* Parse table */
var const_DataSeparator = '=================================================================================================';
Webitel.prototype._parsePlainTableToJSON = function(data, domain, cb) {
    if (!data) {
        cb('Data is undefined!');
        return
    };
    try {
        domain = domain || '_undef_';
        var _line,
            _head,
            _json = {},
            _id,
            _user;

        _line = data.split('\n');
        _head = _line[0].split('\t');
        for (var i = 2; i < _line.length && _line[i] != const_DataSeparator; i++) {
            _id = '';
            _line[i].split('\t').reduce(function (_json, line, index) {
                if (index == 0) {
                    _id = line.trim(); // + '@' + domain;
                    _json[_id] = {
                        id: _id
                    };
                } else {
                    if (_head[index] === 'online') {
                        _user = Users.get(_json[_id]['id'] + '@' + domain);

                        _json[_id]['online'] = ((_user && _user.logged)
                            ? true
                            : false);
                    } else {
                        _json[_id][_head[index].trim()] = line.trim();
                    };
                };
                return _json;
            }, _json);
        };
        cb(null, _json);
    } catch (e) {
        cb(e);
    };
};
/*  */

Webitel.prototype._parsePlainTableToJSONArray = function(data, cb, _separator) {
    if (!data) {
        cb('Data is undefined!');
        return
    };
    try {
        var _line,
            _head,
            _json = [],
            _item,
            _lineItems,
            _headCounts,
            separator = _separator || '\t'
            ;

        _line = data.split('\n');
        _head = _line[0].split(separator);
        _headCounts = _head.length;
        for (var i = 1; i < _line.length; i++) {
            _lineItems = _line[i].split(separator);
            if (_line[i] == "" || _line[i] == const_DataSeparator || _lineItems.length != _headCounts) continue;
            _item = {};
            _lineItems.reduce(function (_item, line, index) {
                _item[_head[index].trim()] = line.trim();
                return _item;
            }, _item);

            _json.push(_item);
        };
        cb(null, _json);
    } catch (e) {
        cb(e);
    };
};
/*  */

Webitel.prototype._parsePlainCollectionToJSON = function (data, cb) {
    if (!data) {
        cb('Data is undefined!');
        return
    };

    try {

        var _json = {},
            lines = data.split('\n'),
            line,
            attribute,
            separatorId;

        for (var i = 0, len = lines.length; i < len; i++) {
            line = lines[i];
            separatorId = line.indexOf('=');
            attribute = line.substring(0, separatorId);
            if (attribute === '')
                continue;
            _json[attribute] = line.substring(separatorId + 1);
        };

        cb(null, _json);
    } catch (e) {
        cb(e['message']);
    }
};

function checkBodyError(response) {
    try {
        if (!response || !response.body)
            return new CodeError(500, "Bad response console server.");
        if (response.body.indexOf('-ERR') === 0 || response.body.indexOf('-USAGE:') === 0) {
            return new CodeError(/not found/.test(response.body) ? 404 : 500, response.body);
        }
        return null;
    } catch (e) {
        return e;
    }
};

function parseArrayVariables (arr) {
    var _a = arr instanceof Array ? arr : [];
    return _a.map( (i) => {
        return /^.+='.*'$/.test(i) ? i : i.replace("=", "='") + "'"
    });
}

/*
    [].map( (i) => {
        return /^.+='.*'$/.test(i) ? i : i.replace("=", "='") + "'"
    })
*/