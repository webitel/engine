/**
 * Created by Igor Navrotskyj on 31.08.2015.
 */

'use strict';

var HashCollection = require(__appRoot + '/lib/Collection'),
    eventsCollection = new HashCollection('id'),
    log = require(__appRoot + '/lib/log')(module)
    ;


var _eventsModule = {
    registered: function (eventName) {
        if (!eventName || eventName == '')
            return;
        eventsCollection.add(eventName, {
            domains: new HashCollection('id')
        });
        log.info('Registered event %s', eventName);
    },

    unRegistered: function (eventName) {
        eventsCollection.remove(eventName);
        log.trace('Unregistered event %s', eventName);
    },

    addListener: function (eventName, caller, sessionId, cb) {
        if (typeof eventName != 'string' || !caller || !sessionId) {
            if (cb)
                cb(new Error('-ERR: Bad parameters.'));
            return;
        };

        var _event = eventsCollection.get(eventName);
        if (!_event) {
            if (cb)
                cb(new Error('-ERR: Event unregistered'));
            return;
        };

        var _domainId = caller.domain || 'root',
            domainSubscribes = _event.domains.get(_domainId);

        if (!domainSubscribes) {
            var _user = {};
            _user[sessionId] = caller.id;
            _event.domains.add(_domainId, _user);
        } else {
            if (domainSubscribes.hasOwnProperty(sessionId)) {
                if (cb)
                    cb(new Error('-ERR: event subscribed!'));
                return;
            } else {
                domainSubscribes[sessionId] = caller.id;
            };
        };

        if (cb)
            cb(null, '+OK: subscribe ' + eventName);
    },

    removeListener: function (eventName, caller, sessionId, cb) {
        if (!caller || typeof eventName != 'string') {
            if (cb)
                cb(new Error('-ERR: Bad parameters'));
            return;
        };

        var _event = eventsCollection.get(eventName);
        if (!_event) {
            if (cb)
                cb(new Error('-ERR: Event unregistered'));
            return;
        };

        var _domainId = caller.domain || 'root';

        var domainSubscribes = _event.domains.get(_domainId);

        if (domainSubscribes && domainSubscribes.hasOwnProperty(sessionId)) {
            delete domainSubscribes[sessionId];
        };

        if (cb)
            cb(null, '+OK: unsubscribe ' + eventName);
    },
    // TODO existsCb переделать
    fire: function (eventName, domainId, event, cb, existsFn) {
        if (typeof eventName != 'string' || !(event instanceof Object)) {
            if (cb)
                cb(new Error('-ERR: Bad parameters'));
            return;
        };

        var _event = eventsCollection.get(eventName);
        if (!_event) {
            if (cb)
                cb(new Error('-ERR: Event unregistered'));
            return;
        };

        var _domain = _event.domains.get(domainId),
            user
            ;

        if (!_domain) {
            if (cb)
                cb(new Error('-ERR: Not subscribes'));
            return;
        };
        event['webitel-event-name'] = 'server';

        var _iterator = 0;
        for (var key in _domain) {
            user = application.Users.get(_domain[key]);
            if (!user) {
                delete _domain[key];
                continue;
            };

            // TODO !!!
            if (existsFn && !existsFn(user, event)) {
                return;
            };

            if (!user.sendSessionObject(event, key)) {
                delete _domain[key];
            } else {
                _iterator++;
            };
        };
        if (_iterator == 0) {
            _event.domains.remove(domainId);
            log.trace('[%s] Remove subscribed domain %s', eventName, domainId);
        };
    }
};

module.exports = _eventsModule;