/**
 * Created by i.navrotskyj on 13.03.2016.
 */
'use strict';
var Collection = require(__appRoot + '/lib/collection'),
    generateUuid = require('node-uuid'),
    url = require('url'),
    http = require('http'),
    request = require('request'),
    _ = require('underscore'),
    log = require(__appRoot + '/lib/log')(module);

class Hook {
    constructor(option) {
        let filter = option.filter;
        this.event = option.event;
        this.domain = option.domain;
        this.action = option.action;
        this.fields = option.fields;
        this.map = option.map;
        this._filters = {};
        this._fields = [];
        for (let key in filter) {
            this._fields.push(key);
            this._filters[key] = {
                "operation": filter[key].operation || null,
                "value": filter[key].value || null
            }
        };
    };

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
        };
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
            if (typeof b != 'string')
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
    constructor(eventName, message, fields, map) {
        this.action = eventName;
        this.data = fields ? _.pick(message, fields) : message;
        if (map instanceof Object) {
            for (let key in map)
                if (this.data.hasOwnProperty(key)) {
                    this.data[map[key]] = this.data[key];
                    delete this.data[key];
                }
        };
        this.id = generateUuid.v4();
    };

    toRequest (uri, method) {
        let _method = method && method.toUpperCase();
        let request = {
            method: _method,
            uri: uri
        };
        if ( _method == 'GET') {
            request.qs = this.data;
        } else {
            request.json = this.toJson();
        }

        return request;
    };

    toJson () {
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
            brokerConnected = false;
        
        var init = function () {
            if (dbConnected && brokerConnected) {
                scope._events.length = 0;
                scope._init()
            }
        };

        app.Broker.on('init:hook', () => {
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
            };
            return cb(null, []);
        });
    };

    _init () {
        let scope = this;
        this.Db.list({enable: true}, (err, res) => {
            if (err)
                return scope.stop(err);

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
                        log.debug(`skipp ${hook.event}`);
                        log.trace(hook);
                        log.trace(e);
                    }
                }
            });
        } catch (e) {
            log.error(e);
        };
    };

    send (hook, name, e) {
        let message = new Message(name, e, hook.fields, hook.map);

        log.trace(`Try send message: ${message.id}`);
        let id = message.id;
        switch (hook.action.type) {
            case TYPES.WEB:
                request(
                    message.toRequest(hook.action.url, hook.action.method),
                    (err, response) => {
                        if (err)
                            return log.warn(err);

                        if (response.statusCode === 200)
                            log.trace(`Send message: ${id}`);
                        else log.warn(`Send message: ${id} status code: ${response.statusCode}`);
                    }
                );

                break;
            default:
                log.warn('Bad hook action: ', hook);
        };

    };

};

const TYPES = {
    WEB: 'web'
};

module.exports = Trigger;