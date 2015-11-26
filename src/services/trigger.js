/**
 * Created by i.navrotskyj on 30.10.2015.
 */
'use strict';

var log = require(__appRoot + '/lib/log')(module),
    conf = require(__appRoot + '/conf'),
    _ = require('underscore'),
    USER_EVENTS = conf.get('application:freeSWITCHEvents')
;

var app = require(__appRoot + '/application');

class Events {
    constructor () {
        this._events = {}
    }

    add (name, fields, fn) {
        let _name = name || 'all';
        if (this._events[_name]) {

        }
    }
}

class Trigger {
    constructor (option) {
        this.events = [];
        this._events = [];
        this._subscribed = {};
    }

    initEvents (cb) {
        // todo DB
        var esl = app.Esl;
        this.events.push({
            "filters": {
                "eventName": "CHANNEL_EXECUTE",
                "fields": {
                    "Channel-Presence-ID": "102@10.10.10.144"
                }
            },
            "exec": "test"
        });
        
        this._events.forEach( (item) => {

            /**
             * Subscribe if not subscribed Event-Name
             */
            let eventName = item.filters && item.filters.eventName;
            if (eventName && USER_EVENTS.indexOf(eventName) == -1 && !this._subscribed.hasOwnProperty(eventName)) {
                esl.subscribe(eventName);
                this._subscribed[eventName] = true;
                log.debug('Subscribe %s', eventName);
                if (!item.fields) {
                    esl.filter('Event-Name', eventName);
                }
            }
        });

        if (cb)
            cb();
    }

    onEvent (e) {
        console.dir(e.type);
    }

    exec (name) {
        console.log(name);
    }
}

module.exports = Trigger;