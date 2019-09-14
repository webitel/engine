/**
 * Created by i.n. on 13.04.2015.
 */

'use strict';

var handleWebSocketError = require(__appRoot + '/middleware/handleWebSocketError'),
    log = require(__appRoot + '/lib/log')(module),
    eventService = require(__appRoot +  '/services/events'),
    hotdeskService = require(__appRoot +  '/services/hotdesk'),
    aclServices = require(__appRoot + '/services/acl')
    ;


const User = function (id, sessionId, ws, params) {
    this.id = id;
    this.ws = {};
    this.domain = id.split('@')[1];
    this.sessionLength = 1;

    this.hotdesk = {
        id: null,
        sessions: []
    };

    ws._mapUserEvents = [];

    this.ws[sessionId] = ws;
    this.state = '';
    this.status = '';
    this._subscribeEvent = {};

    const variable = params;
    for (let key in variable) {
        if (variable.hasOwnProperty(key) && !this[key]) {
            this[key] = variable[key];
        }
    }

    this.logged = params['logged'] || false;
};

User.prototype.addHotdeskingSession = function(id, sessionId) {
    if (this.flushHotdesking(id)) {
        this.hotdesk.id = id;
    }
    this.hotdesk.sessions.push(sessionId);
    log.debug(`hotdesk add session: ${sessionId}`);
};

User.prototype.checkHotdeskSession = function(id, sessionId) {
    if (this.hotdesk.id !== id) {
        return false;
    }
    return !!~this.hotdesk.sessions.indexOf(sessionId)
};

User.prototype.flushHotdesking = function(id) {
    if (!this.hotdesk.id || this.hotdesk.id !== id) {
        this.hotdesk.id = null;
        this.hotdesk.sessions = [];
        return true
    }
    return false
};

User.prototype.removeHotdeskSession = function(sessionId) {
    const idx = this.hotdesk.sessions.indexOf(sessionId);
    if (~idx) {
        log.debug(`hotdesk ${this.hotdesk.id} destroy session: ${this.hotdesk.sessions.splice(idx, 1)}`);
        if (!this.hotdesk.sessions.length) {
            log.trace(`hotdesk ${this.hotdesk.id} added task sign out ${this.id}`);
            hotdeskService.addTask(this.id, this.hotdesk.id, this.domain);
        }
    }
};

User.prototype.getLogged = function () {
    return this.logged;
};

User.prototype.getSession = function (ws) {
    for (let key in this.ws) {
        if (this.ws.hasOwnProperty(key) && this.ws[key] == ws) {
            return key;
        }
    }

    return null;
};

User.prototype.setLogged = function (logged) {
    this.logged = logged;
    return this.logged;
};

User.prototype.addSession = function (sessionId, ws, params) {
    ws._mapUserEvents = [];

    this.ws[sessionId] = ws;
    this.sessionLength ++;

    if (this.sessionLength > 20) {
        log.warn(`user ${this.id} opened ${this.sessionLength} sockets`)
    } else if (this.sessionLength > 100) {
        log.error(`user ${this.id} opened ${this.sessionLength} sockets`)
    }

    return this.sessionLength;
};

User.prototype.removeSession = function (sessionId) {
    if (this.ws[sessionId]) {
        if (this.ws[sessionId]._mapUserEvents.length > 0) {
            this.ws[sessionId]._mapUserEvents.forEach(eventName => this.unSubscribe(eventName, sessionId, e => {
                if (e) {
                    log.error(e)
                }
            }))
        }
        this.removeHotdeskSession(sessionId);
        delete this.ws[sessionId];
        this.sessionLength--;
        log.trace('Pear disconnect: %s [%s]', this.id, sessionId);
        return this.sessionLength;
    } else {
        return false;
    }
};

User.prototype.getSessionLength = function () {
   return this.sessionLength;
};

User.prototype.sendSessionObject = function (obj, sessionId) {
    if (!obj || !sessionId || !this.ws[sessionId]) {
        return false;
    }

    try {
        this.ws[sessionId].send(JSON.stringify(obj));
        return true;
    } catch (e) {
        log.error(e);
        return false;
    }
};

User.prototype.sendObject = function (obj) {
    if (!this.hasPermitNotifyAccountStatus(obj, this.acl && this.acl['account'])) {
        return
    }

    for (let key in this.ws) {
        if (this.ws.hasOwnProperty(key)) {
            this.sendSessionObject(obj, key);
        }
    }
};

//TODO
User.prototype.hasPermitNotifyAccountStatus = function (e, aclAccount = []) {
    if (e && e['Event-Name'] === "ACCOUNT_STATUS") {
        return (e['presence_id'] === this.id ||
            (e['presence_id'] !== this.id && (~aclAccount.indexOf('*') || ~aclAccount.indexOf('r'))))
    }
    return true
};

User.prototype.__broadcastInDomain = function (obj, domainId) {
    try {
        domainId = domainId || this.domain;
        var domain = application.Domains.get(domainId);
        if (domain && domain['users']) {
            var users = domain['users'],
                user;
            for (var key in users) {
                if (users.hasOwnProperty(key)) {
                    user = application.Users.get(users[key]);
                    if (user) {
                        user.sendObject(obj);
                    };
                };
            };
        }
    } catch (e) {
        log.error(e);
    }
};

User.prototype.broadcastInDomain = function (event) {
    if (this.domain)
        this.__broadcastInDomain(event, this.domain);
};

User.prototype.disconnect = function () {
    try {
        const ws = this.ws;
        for (let key in ws) {
            if (ws.hasOwnProperty(key)) {
                ws[key].terminate();
            }
        }
    } catch (e) {
        log.error(e);
        return false;
    }
};

User.prototype.setState = function (state, status, description) {
    if (status) {
        this.status = status;
    }

    if (state) {
        this.state = state;
    }

    if (description) {
        this.description = description;
    }

    log.debug('User %s status: %s, state: %s', this.id, this.status, this.state);
};

User.prototype.changeRole = function (role) {
    if (!role)
        return log.error('Bad set role %s', this.id);

    aclServices._whatResources(role, (e, aclResource) => {
        if (e)
            return log.error(e);

        this.roleName = role;
        this.acl = aclResource;
    });
};

User.prototype.subscribe = function (eventName, sessionId, args, cb) {
    if (this.ws[sessionId]) {
        this.ws[sessionId]._mapUserEvents.push(eventName)
    }
    eventService.addListener(eventName, this, sessionId, args, cb)
};

User.prototype.unSubscribe = function (eventName, sessionId, cb) {
    if (this.ws[sessionId] && ~this.ws[sessionId]._mapUserEvents.indexOf(eventName)) {
        this.ws[sessionId]._mapUserEvents.splice(
            this.ws[sessionId]._mapUserEvents.indexOf(eventName),
            1
        )
    }
    eventService.removeListener(eventName, this, sessionId, (err, res) => {
        if (!err) {
            if (this.ws[sessionId]) {
                const ws = this.ws[sessionId];
                if (typeof ws._observeEvents === "object") {
                    delete ws._observeEvents[eventName]
                }
            }
        }
        return cb(err, res)
    })
};

User.prototype.observeEvent = function (eventName, sessionId, exec) {
    if (this.ws[sessionId]) {
        const ws = this.ws[sessionId];
        if (!ws._observeEvents) {
            ws._observeEvents = {}
        }

        if (ws._observeEvents[eventName] instanceof Array) {
            ws._observeEvents[eventName].push(exec)
        } else {
            ws._observeEvents[eventName] = [exec]
        }
    }
};


User.prototype.closeObserveEvent = function (ws, sessionId) {
    let arr;
    let fn;

    if (ws._observeEvents) {
        for (let  key in ws._observeEvents) {
            arr = ws._observeEvents[key];
            if (arr instanceof Array) {
                while (arr.length) {
                    fn = arr.shift();
                    if (typeof fn === 'function') {
                        fn(sessionId, key)
                    }
                }
            }
        }
    }
};

module.exports = User;