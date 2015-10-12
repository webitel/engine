/**
 * Created by i.n. on 13.04.2015.
 */

'use strict';

var handleWebSocketError = require(__appRoot + '/middleware/handleWebSocketError'),
    log = require(__appRoot + '/lib/log')(module),
    outQueryService = require(__appRoot + '/services/outboundQueue')
    ;

const SCRAP_RATE = 1.02; //  - > %

var User = function (id, sessionId, ws, params) {
    this.id = id;
    this.ws = {};
    this.domain = id.split('@')[1];
    this.sessionLength = 1;
    this.ws[sessionId] = ws;
    this.state = '';
    this.status = '';
    var variable = params;
    for (var key in variable) {
        if (variable.hasOwnProperty(key) && !this[key]) {
            this[key] = variable[key];
        };
    };
    this.logged = params['logged'] || false;

    /**
     TODO outbound company
    outQueryService.getAvgBillSecUser(this, function (err, res) {
            if (err)
                return log.error(err);

            this.avgBillSec = (Math.round(res * SCRAP_RATE) || 0);
        }
        .bind(this)
    );
     **/
};

User.prototype.getLogged = function () {
    return this.logged;
};

User.prototype.getSession = function (ws) {
    for (var key in this.ws) {
        if (this.ws.hasOwnProperty(key) && this.ws[key] == ws) {
            return key;
        };
    };

    return null;
};

User.prototype.setLogged = function (logged) {
    this.logged = logged;
    return this.logged;
};

User.prototype.addSession = function (sessionId, ws, params) {
    this.ws[sessionId] = ws;
    this.sessionLength ++;
    return this.sessionLength;
};

User.prototype.removeSession = function (sessionId) {
    if (this.ws[sessionId]) {
        delete this.ws[sessionId];
        this.sessionLength --;
        //log.trace('Pear disconnect: %s [%s]', this.id, sessionId);
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
    };

    try {
        this.ws[sessionId].send(JSON.stringify(obj));
        return true;
    } catch (e) {
        log.error(e);
        return false;
    };
};

User.prototype.sendObject = function (obj) {
    for (var key in this.ws) {
        if (this.ws.hasOwnProperty(key)) {
            this.sendSessionObject(obj, key);
        }
    }
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
        var ws = this.ws;
        for (var key in ws) {
            if (ws.hasOwnProperty(key)) {
                ws[key].close();
            };
        };
    } catch (e) {
        log.error(e);
        return false;
    }
};

User.prototype.setState = function (state, status) {
    if (status) {
        this.status = status;
    };

    if (state) {
        this.state = state;
    };
    log.debug('User %s status: %s, state: %s', this.id, this.status, this.state);

/**
    TODO outbound company
    try {
        if (this.state === 'ONHOOK') {
            outQueryService.getCallee(this, {}, function (err, res) {

            });
        }
    } catch (e) {
        log.error(e);
    } finally {
        return {
            "state": this.state,
            "status": this.status
        };
    };

**/
};

module.exports = User;