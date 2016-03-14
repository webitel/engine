/**
 * Created by i.navrotskyj on 13.03.2016.
 */
'use strict';
var Collection = require(__appRoot + '/lib/collection');

var dbData = {
    "event": "CHANNEL_CREATE",
    "enable": true,
    "description": "Test",
    "actions": [
        {
            "type": "web",
            "url": "http://my.com/blabla"
        }
    ],
    "domain": "10.10.10.144",
    "filter": {
        "variable_user_name": {
            "operation": "==",
            "value": "100"
        }
    }
};

class Hook {
    constructor(event, domain, url, filter) {
        this.event = event;
        this.domain = domain;
        this.url = url;
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
            if (!obj.hasOwnProperty(key)
                || !Operations.hasOwnProperty(this._filters[key].operation)
                || !Operations[this._filters[key].operation](obj[key], this._filters[key].value))
                return false;
        };
        return true;
    };
};

const Operations = {
    "==": function (a, b) {
        return a == b;
    },
    "!=": function (a, b) {
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
    }
};

class Message {
    constructor(eventName, message) {
        this.action = eventName;
        this.data = message;
    }

    toJSON() {
        return {
            "action": this.action,
            "data": this.message
        }
    }
};

class WebHooks {
    constructor (app) {
        this.hooks = new Collection('id');
        this._app = app;
        this._fsEvents = [];

        var scope = this;
        app.on('sys::eslConnect', ()=> {
            scope.subscribeEsl('CHANNEL_CREATE')
        });
        // test
        var hook = new Hook(dbData.event, dbData.domain, dbData.url, dbData.filter);
        this.hooks.add(hook.getId(), [hook]);
    };
    subscribeEsl (eventName) {
        if (~this._fsEvents.indexOf(eventName))
            return true;
        this._app.Esl.filter('Event-Name', eventName, (e) => {
            console.log('Subscribe ', eventName);
        });
    };

    toId (domain, eventName) {
        return domain + '/' + eventName;
    };

    emit (eventName, domain, e) {
        if (!eventName || !(e instanceof Object))
            return false;
        let hooks = this.hooks.get(this.toId(domain, eventName));

        if (hooks instanceof Array) {
            for (let hook of hooks)
                if (hook.check(e))
                    this.send(hook, eventName, e);
        }
    };

    send (hook, name, e) {
        console.log('FIRE');
        var message = new Message(name, e);

    }
};

module.exports = WebHooks;