/**
 * Created by i.navrotskyj on 13.03.2016.
 */
'use strict';
const Collection = require(__appRoot + '/lib/collection'),
    generateUuid = require('node-uuid'),
    url = require('url'),
    http = require('http'),
    request = require('request'),
    _ = require('underscore'),
    async = require('async'),
    conf = require(__appRoot + '/conf'),
    log = require(__appRoot + '/lib/log')(module),
    MAX_RETRIES = conf.get('application:hook:maxRetries') || Infinity,
    DEF_DELAY_SEC = conf.get('application:hook:defaultDelaySec') || 60
;

class Hook {
    constructor(option) {
        let filter = option.filter;
        this.event = option.event;
        this.domain = option.domain;
        this.retries = 0;
        this.delaySec = DEF_DELAY_SEC;//option.delay;
        if (option.retries > 0 && option.retries <= MAX_RETRIES) {
            this.retries = option.retries;
        } else {
            log.debug(`Hook ${option._id}@${option.domain} bad options retries, skip retries`);
        }

        if (option.delay > 0) {
            this.delaySec = option.delay;
        } else {
            log.debug(`Hook ${option._id}@${option.domain} bad options delay, use default ${this.delaySec}`);
        }

        this.action = option.action;
        this.fields = option.fields;
        this.headers = option.headers;
        this.map = option.map;
        this.auth = option.auth;
        if (option.customBody === true) {
            this._rawBody = option.rawBody;
        }
        this._filters = {};
        this._fields = [];
        for (let key in filter) {
            this._fields.push(key);
            this._filters[key] = {
                "operation": filter[key].operation || null,
                "value": filter[key].value || null
            }
        }
    };

    useAuth () {
        return this.auth && this.auth.enabled && this.auth.url;
    }

    useCookie () {
        return this.auth && this.auth.cookie;
    }

    toAuthRequest (e) {
        let auth = {};
        [
            auth.method = 'GET',
            auth.url = ''
        ] = [
            this.auth.method && this.auth.method.toUpperCase(),
            this.auth.url
        ];

        let request = {
            method: auth.method,
            uri: auth.url,
            headers: _setHeadersFromEvent(this.auth.headers, e)
        };

        if ( request.method === 'GET') {
            request.qs = _setHeadersFromEvent(this.auth.map, e);
        } else {
            request.json = _setHeadersFromEvent(this.auth.map, e);
        }

        return request;
    }

    getId () {
        return this.domain + '/' + this.event;
    };
    check(obj) {
        if (!(obj instanceof Object))
            return false;

        for (let key in this._filters) {
            if (!Operations.hasOwnProperty(this._filters[key].operation)
                || !Operations[this._filters[key].operation](obj[key], this._filters[key].value))
                return false;
        }
        return true;
    };
}

const Operations = {
    "==": function (a, b) {
        if (b === 'null')
            b = undefined;

        return a == b;
    },
    "!=": function (a, b) {
        if (b === 'null')
            b = undefined;

        return a != b;
    },
    "<": function (a, b) {
        return a < b
    },
    ">": function (a, b) {
        return a > b;
    },
    "<=": function (a, b) {
        return a <= b;
    },
    ">=": function (a, b) {
        return a >= b
    },
    "reg": function (a, b) {
        try {
            if (typeof b !== 'string')
                return;
            let flags = b.match(new RegExp('^/(.*?)/([gimy]*)$'));
            if (!flags)
                flags = [null, b];

            let regex = new RegExp(flags[1], flags[2]);
            return regex.test(a);
        } catch (e) {
            log.error(e);
            return false;
        }
    }
};

class Message {
    constructor(eventName, message, fields, map, rawBody) {
        this.action = eventName;
        this._rawData = false;
        if (typeof rawBody === 'string') {
            let val;
            rawBody = rawBody.replace(/\$\{([\s\S]*?)\}/gi, function (a, b) {
                if (~b.indexOf('.')) {
                    val = message;
                    b.split('.').forEach(function (token) {
                        val = val && val[token];
                    });
                } else {
                    val = message[`variable_${b}`] || message[b]
                }

                return val || ""
            });
            try {
                this.data = JSON.parse(rawBody);
                this._rawData = true;
            } catch (e) {
                this.data = {};
                log.error(e);
            }
        } else {
            this.data = fields && fields.length > 0 ? _.pick(message, fields) : message;
            if (map instanceof Object) {
                for (let key in map)
                    if (this.data.hasOwnProperty(key)) {
                        this.data[map[key]] = this.data[key];
                        delete this.data[key];
                    }
            }
        }
        if (!message.webitel_hook_msg_id) {
            message.webitel_hook_msg_id = generateUuid.v4();
        }
        this.id = message.webitel_hook_msg_id;
    };

    toRequest (uri, method, eventObj, hook) {
        let _method = method && method.toUpperCase();
        let request = {
            method: _method,
            uri: uri,
            headers: {}
        };

        if (hook && hook.headers) {
            request.headers = _setHeadersFromEvent(hook.headers, eventObj);
        }

        if (hook.useCookie() && eventObj['auth:header:set-cookie']) {
            request.headers.Cookie = eventObj['auth:header:set-cookie'];
        }

        if ( _method === 'GET') {
            request.qs = this.data;
        } else {
            request.json = this.toJson();
        }

        return request;
    };

    toJson () {
        if (this._rawData) {
            return this.data;
        }
        return {
            "id": this.id,
            "action": this.action,
            "data": this.data
        }
    }

    toString() {
        return JSON.stringify(this.toJson());
    }
}

class Trigger {
    constructor (app) {
        this.hooks = new Collection('id');
        this._app = app;
        this._events = [];

        var scope = this;

        let dbConnected = false,
            initHookQueue = false,
            brokerConnected = false;
        
        var init = function () {
            if (dbConnected && brokerConnected && initHookQueue) {
                scope._events.length = 0;
                scope._init()
            }
        };

        app.Broker.on('init:hook', () => {
            initHookQueue = true;
            init();
        });

        app.Broker.on('init:broker', () => {
            brokerConnected = true;
            init();
        });

        app.once('sys::connectDb', (db)=> {
            scope.Db = db._query.hook;
            dbConnected = true;
            init();
        });

        app.Broker.on('hookEvent', scope.emit.bind(scope));
    };

    find (eventName, domainName, cb) {
        this.Db.list({enable: true, domain: domainName, event: eventName}, (err, res) => {
            if (err)
                return cb(err);

            if (res.length > 0) {
                let result = [];
                res.forEach((item) => {
                    result.push(new Hook(item));
                });
                return cb(null, result);
            }
            return cb(null, []);
        });
    };

    _init () {
        let scope = this;
        this.Db.list({enable: true}, (err, res) => {
            if (err)
                return scope._app.stop(err);

            if (res.length > 0) {
                res.forEach((item) => {
                    if (item.event && item.domain && item.action) {
                        scope.subscribe(item.event);
                    } else {
                        log.warn('Bad hook: ', item);
                    }
                })
            } else {
                log.info("No hook.");
            }

        })
    };

    subscribe (eventName) {
        if (~this._events.indexOf(eventName))
            return true;
        let scope = this;

        this._app.Broker.bindHook(eventName, (e) => {
            if (e)
                return log.error(e);
            log.debug(`subscribe ${eventName}`);
            scope._events.push(eventName);
        });
    };

    toId (domain, eventName) {
        return domain + '/' + eventName;
    };

    emit (eventName, domain, e) {
        try {
            if (!eventName || !(e instanceof Object))
                return false;

            if (e.hasOwnProperty('Event-Subclass'))
                eventName += '->' + e['Event-Subclass'];

            this.find(eventName, domain, (err, hooks) => {
                if (err)
                    return log.error(err);

                for (let hook of hooks) {
                    if (hook.check(e)) {
                        this.send(hook, eventName, e);
                    } else {
                        log.debug(`skip ${hook.event}`);
                        log.trace(hook);
                        log.trace(e);
                    }
                }
            });
        } catch (e) {
            log.error(e);
        }
    };

    send (hook, name, e) {
        switch (hook.action.type) {
            case TYPES.WEB:
                async.waterfall(
                    [
                        (cb) => {
                            if (hook.useAuth()) {
                                request(
                                    hook.toAuthRequest(e),
                                    (err, res) => {
                                        if (err)
                                            return cb(err);

                                        if (res.statusCode !== 200) {
                                            return cb(new Error(`Bad auth status code ${res.statusCode}`));
                                        }

                                        for (let k in res.body) {
                                            if (res.body.hasOwnProperty(k) && typeof res.body[k] === 'string')
                                                e[`auth:body:${k}`] = res.body[k];
                                        }

                                        for (let h in res.headers) {
                                            if (res.headers.hasOwnProperty(h)) {
                                                if (res.headers[h] instanceof Array) {
                                                    let indexArrayHeadSeparator = null;
                                                    let indexArrayValSeparator = null;
                                                    for (let arrHead of res.headers[h]) {
                                                        indexArrayHeadSeparator = arrHead.indexOf('=');
                                                        indexArrayValSeparator = arrHead.indexOf(';');
                                                        if (indexArrayValSeparator !== -1) {
                                                            e[`auth:header:${arrHead.substring(0, indexArrayHeadSeparator)}`] =
                                                                arrHead.substring(indexArrayHeadSeparator + 1, indexArrayValSeparator);
                                                        }  else {
                                                            e[`auth:header:${arrHead.substring(0, indexArrayHeadSeparator)}`] =
                                                                arrHead.substring(indexArrayHeadSeparator + 1);
                                                        }
                                                    }

                                                    e[`auth:header:${h}`] = res.headers[h].join(';');
                                                } else {
                                                    e[`auth:header:${h}`] = res.headers[h];
                                                }
                                                // console.dir(e, {colors: true, depth: 10});
                                            }
                                        }
                                        return cb(null);
                                    }
                                )
                            } else {
                                cb(null)
                            }
                        },

                        (cb) => {
                            let message = new Message(name, e, hook.fields, hook.map, hook._rawBody);

                            log.trace(`Try send message: ${message.id}`);
                            let id = message.id;

                            request(
                                message.toRequest(hook.action.url, hook.action.method, e, hook),
                                (err, res) => {
                                    cb(err, res, id);
                                }
                            );
                        }
                    ],
                    (err, response, id) => {
                        if (err)
                            return responseError(err, response, e, id, hook);

                        if (response.statusCode === 200)
                            log.trace(`Send message: ${id}`);
                        else responseError(err, response, e, id, hook);
                    }
                );


                break;
            default:
                log.warn('Bad hook action: ', hook);
        }

    };

}

function responseError(err, response = {}, event = {}, id, hook) {
    if (isNaN(+event.webitel_hook_retries)) {
        event.webitel_hook_retries = 1;
    } else {
        event.webitel_hook_retries++;
    }

    if (hook.retries < event.webitel_hook_retries) {
        log.warn(`Hook ${id} max retries ${hook.retries} last status code ${response.statusCode}`);
        return; //TODO rem db
    }
    log.trace(`Message ${id} retry send ${event.webitel_hook_retries} max ${hook.retries} after ${hook.delaySec} sec`);
    application.Broker.publishWithArgs(
        application.Broker.Exchange.HOOK_DELAY_EXCHANGE,
        `hook`,
        event,
        {
            persistent: true,
            contentType: "text/json",
            headers: {'x-delay': hook.delaySec * 1000}
        }
    );
}

function _setHeadersFromEvent(mapHeaders, eventObj) {
    let headers = {};
    for (let head in mapHeaders) {
        if (mapHeaders.hasOwnProperty(head) && typeof mapHeaders[head] === 'string' && eventObj) {
            headers[head] = mapHeaders[head].replace(/\$\{([\s\S]*?)\}/gi, (a, b) => {
                return eventObj[b] || ""
            });
        }
    }
    return headers;
}

const TYPES = {
    WEB: 'web'
};

module.exports = Trigger;