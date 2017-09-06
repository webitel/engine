/**
 * Created by Igor Navrotskyj on 31.08.2015.
 */

'use strict';

const HashCollection = require(__appRoot + '/lib/collection'),
    eventsCollection = new HashCollection('id'),
    log = require(__appRoot + '/lib/log')(module)
    ;


const _eventsModule = {
    registered: function (eventName) {
        if (!eventName || eventName === '')
            return;

        let e = eventsCollection.add(eventName, {
            domains: new HashCollection('id'),
            name: eventName
        });
        log.info('Registered event %s', eventName);
        return e;
    },

    addListener: function (eventName, caller, sessionId, options, cb) {
        if (typeof eventName !== 'string' || !caller || !sessionId) {
            if (cb)
                cb(new Error('-ERR: Bad parameters.'));
            return;
        }

        const _event = eventsCollection.get(eventName);
        if (!_event) {
            log.error('-ERR: Event unregistered');
            if (cb)
                cb(new Error('-ERR: Event unregistered'));
            return;
        }

        const _domainId = caller.domain || 'root',
            domainSubscribes = _event.domains.get(_domainId);

        if (!domainSubscribes) {
            log.trace('subscribe', sessionId, eventName);
            _event.domains.add(_domainId, {
                [sessionId]: {
                    name: caller.id,
                    args: options
                }
            });
        } else {
            if (domainSubscribes.hasOwnProperty(sessionId)) {
                log.error('subscribe', sessionId, eventName);
                if (cb)
                    cb(new Error('-ERR: event subscribed!'));
                return;
            } else {
                log.trace('subscribe', sessionId, eventName);
                domainSubscribes[sessionId] = {
                    name: caller.id,
                    args: options
                };
            }
        }

        application.emit(`subscribe::${eventName}`, options, caller, eventName, sessionId);
        if (cb)
            cb(null, '+OK: subscribe ' + eventName);
    },

    removeListener: function (eventName, caller, sessionId, cb) {
        if (!caller || typeof eventName !== 'string') {
            if (cb)
                cb(new Error('-ERR: Bad parameters'));
            return;
        }

        const _event = eventsCollection.get(eventName);
        if (!_event) {
            if (cb)
                cb(new Error('-ERR: Event unregistered'));
            return;
        }

        const domainSubscribes = _event.domains.get(caller.domain || 'root');

        if (domainSubscribes && domainSubscribes.hasOwnProperty(sessionId)) {
            application.emit(`unsubscribe::${eventName}`, domainSubscribes[sessionId], caller, eventName);
            delete domainSubscribes[sessionId];
        }

        log.trace(`Session ${sessionId} un subscribe event ${eventName}`);
        if (cb)
            cb(null, '+OK: unsubscribe ' + eventName);
    },
    // TODO existsCb переделать (FAH)
    fire: function (eventName, domainId, event, cb, existsFn) {

        if (typeof eventName !== 'string' || !(event instanceof Object)) {
            if (cb)
                cb(new Error('-ERR: Bad parameters'));
            return;
        }

        const _event = eventsCollection.get(eventName);

        if (!_event) {
            if (cb)
                cb(new Error('-ERR: Event unregistered'));
            return;
        }

        const _domain = _event.domains.get(domainId);
        let user;

        if (!_domain) {
            if (cb)
                cb(new Error('-ERR: Not subscribes'));
            return;
        }

        event['webitel-event-name'] = 'server';

        let _iterator = 0;
        for (let key in _domain) {
            user = application.Users.get(_domain[key].name);
            if (!user) {
                log.debug('REMOVE DOMAIN session', key);
                delete _domain[key];
                continue;
            }

            // TODO !!!
            if (existsFn && !existsFn(user, event)) {
                _iterator++;
                log.trace(`Skip fire ${key} - exists false`);
                continue
            }

            if (!user.sendSessionObject(event, key)) {
                log.warn('REMOVE DOMAIN session', key);
                delete _domain[key];
            } else {
                log.debug('Emit server event %s --> %s [%s]', eventName, user.id, key);
                _iterator++;
            }
        }

        if (_iterator === 0) {
            log.debug('REMOVE DOMAIN', domainId);
            _event.domains.remove(domainId);
            log.trace('[%s] Remove subscribed domain %s', eventName, domainId);
        } else {
            log.trace('send: ', _iterator)
        }
    }
};

module.exports = _eventsModule;