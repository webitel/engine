'use strict';

var conf = require('../../conf'),
    socketTimeUnauthorized = conf.get('application:socketTimeUnauthorized') || 10,
    log = require('../../lib/log')(module),
    handleError = require(__appRoot + '/middleware/handleWebSocketError'),
    generateUuid = require('node-uuid'),
    Controller = require('./controllers'),
    ACCOUNT_EVENTS = require(__appRoot + '/const').ACCOUNT_EVENTS,
    User = require(__appRoot + '/models/user'),
    handleStatusDb = require(__appRoot + '/services/userStatus').insert,
    DIFF_AGENT_LOGOUT_SEC = conf.get('application:callCentre:diffAgentLogoutTimeSec') || 60,
    SCHEDULE_TIME_SEC = conf.get('application:callCentre:scheduleLogoutSec') || 60,
    getIp = require(__appRoot + '/utils/ip')
    ;

module.exports = Handler;

function Handler(wss, application) {
    var controller = Controller(application);

    wss.on('connection', function(ws) {

        var caller = null,
            sessionId = generateUuid.v4(),
            timerId
            ;

        function onAuth (id, params, execId) {
            clearTimeout(timerId);
            var user = application.Users.get(id)
                ;
            if (!user) {
                user = new User(id, sessionId, ws, params);
                application.Users.add(id, user);
            } else {
                user.addSession(sessionId, ws, params);
            };

            caller = user;
            // test leak;
            //caller.bigData = new Array(1e6).join('XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX\n');
            var response = {
                'exec-uuid': execId,
                'exec-complete': '+OK',
                'exec-response': {
                    'login': id,
                    'role': params.role.name,
                    'domain': params.domain,
                    'state': params.state,
                    'status': params.status,
                    'cc-agent': params['cc-agent'],
                    'session': sessionId,
                    'ws-count': caller.getSessionLength(),
                    'cc-logged': !!caller['cc-logged'],
                    'acl': params.acl
                }
            };
            caller.sendSessionObject(response, sessionId);
            log.trace('New pear %s [%s]', caller && caller.id, getIp(ws.upgradeReq));
        };

        ws.once('webitelAuth', onAuth);

        if (socketTimeUnauthorized > 0) {
            timerId = setTimeout(function () {
                if (!caller) {
                    log.trace('Disconnect session %s, remoteAddress: %s [unauth]', sessionId,
                        ws.upgradeReq.connection.remoteAddress);
                    try {
                        ws.send(JSON.stringify({
                            "webitel-event-name": "account",
                            "Event-Name": "DISCONNECT"
                        }));
                        ws.close();
                    } catch (e) {
                        log.warn('User socket close:', e.message);
                    }
                }
            }, socketTimeUnauthorized * 1000);
        };

        ws.on('message', function(message, flags) {
            try {
                if (flags.binary) {
                    log.warn('[%s->%s] Bad message type binary', caller && caller.id, ws.upgradeReq.connection.remoteAddress);
                    return;
                };

                log.trace('[%s->%s] received: %s', caller && caller.id,
                    getIp(ws.upgradeReq), message
                );

                var msg = JSON.parse(message);
                var execId = msg['exec-uuid'];
                var args = msg['exec-args'] || {};

                if (typeof controller[msg['exec-func']] === 'function') {
                    controller[msg['exec-func']](caller, execId, args, ws);
                } else {
                    ws.send(JSON.stringify({
                        'exec-uuid': execId,
                        'exec-complete': '+ERR',
                        'exec-response': msg['exec-func'] + ' - not found.'
                    }));
                    log.warn(msg['exec-func'] + ' - not found --> ' + (caller && caller.id));
                };
            } catch (e) {
                handleError(ws);
                log.error('Command error:', e.message);
            };
        });

        ws.on('close', function () {
            try {
                ws.removeListener('webitelAuth', onAuth);
                //ws.close();
                log.trace('Close session %s', sessionId);
                if (caller && caller.removeSession(sessionId) === 0) {
                    application.Users.remove(caller.id);
                    log.trace('disconnect: ', caller.id);
                    log.debug('Users session: ', application.Users.length());
                };

            } catch (e) {
                log.error(e);
            };
        });

        ws.on('error', function(e) {
            log.error('Socket error:', e);
        });
    });
    
    application._getWSocketSessions = function () {
        return wss.clients.size
    };
    
    application.broadcast = function (event, user) {
        if (user) {
            user.broadcastInDomain(event);
        };

        var root = this.Users.get('root');
        if (root) {
            root.sendObject(event);
        }
    };
    
    application.broadcastInDomain = function (event, domainId) {
        if (!event || !domainId) {
            return log.error('broadcastInDomain bad parameters.');
        };

        try {
            var domain = application.Domains.get(domainId);
            if (domain && domain['users']) {
                var users = domain['users'],
                    user;
                for (var key in users) {
                    if (users.hasOwnProperty(key)) {
                        user = application.Users.get(users[key]);
                        if (user) {
                            user.sendObject(event);
                        };
                    };
                };
            }
        } catch (e) {
            log.error(e);
        }
    };
    
    application.Schedule(DIFF_AGENT_LOGOUT_SEC * 1000, function () {
        log.debug('Schedule logout agents.');
        if (application.loggedOutAgent.length() > 0) {
            var collection = application.loggedOutAgent.collection,
                currentTime = Date.now();
            for (let key in collection) {
                if (collection[key] < currentTime) {
                    application.loggedOutAgent.remove(key);
                    application.Esl.bgapi('callcenter_config agent set status ' + key + " 'Logged Out'", function (res) {
                        log.debug('Logout agent %s [%s]', key, res.body && res.body.trim());
                    });
                };
            };
        };
    });
    
    application.Users.on('removed', function (user) {
        try {
            application.Broker.unBindChannelEvents(
                user,
                (e) => {
                    if (e)
                        log.error(e);
                }
            );

            var _id = user.id.split('@'),
                _domain = _id[1] || _id[0],
                domain = application.Domains.get(_domain),
                jsonEvent;
            try {
                jsonEvent = getJSONUserEvent(ACCOUNT_EVENTS.OFFLINE, _domain, _id[0]);
                log.debug(jsonEvent['Event-Name'] + ' -> ' + user.id);
                application.broadcast(jsonEvent, user);
            } catch (e) {
                log.warn('Broadcast account event: ', domain);
            };

            if (domain) {
                var _index = domain.users.indexOf(user.id);
                if (_index !== -1) {
                    domain.users.splice(_index, 1);
                    if (domain.users.length === 0) {
                        application.Domains.remove(_domain);
                        log.debug('Domains session: ', application.Domains.length());
                    };
                };
            };

            if (user["cc-logged"]) {
                var agent = application.loggedOutAgent.get(user.id);
                if (!agent) {
                    application.loggedOutAgent.add(user.id, addMinutes(DIFF_AGENT_LOGOUT_SEC));
                } else {
                    agent = addMinutes(DIFF_AGENT_LOGOUT_SEC);
                };
            };

            insertSession (_id[0], user.domain, user.state, user.status, user.description, false);
        } catch (e) {
            log.warn('On remove domain error: ', e.message);
        }
    });
    application.Users._maxSession = 0;
    application.Users.on('added', function (user) {
        try {

            application.Broker.bindChannelEvents(
                user,
                (e) => {
                    if (e)
                        log.error(e);
                }
            );


            if (this._maxSession < this.length())
                this._maxSession = this.length();

            var _id = user.id.split('@'),
                _domain = _id[1] || _id[0],
                domain = application.Domains.get(_domain),
                jsonEvent;

            if (!domain) {
                application.Domains.add(_domain, {
                    id: _domain,
                    users: [user.id]
                });
                log.debug('Domains session: ', application.Domains.length());
            } else {
                if (domain.users.indexOf(user.id) == -1) {
                    domain.users.push(user.id);
                };
            };

            try {
                jsonEvent = getJSONUserEvent(ACCOUNT_EVENTS.ONLINE, _domain, _id[0]);
                log.debug(jsonEvent['Event-Name'] + ' -> ' + user.id);
                application.broadcast(jsonEvent, user);
            } catch (e) {
                log.warn('Broadcast account event: ', domain);
            }
            insertSession (_id[0], user.domain, user.state, user.status, user.description, true);
        } catch (e) {
            log.warn('On add domain error: ', e.message);
        }
    });
};

var getJSONUserEvent = function (eventName, domainName, userId) {
    return {
        "Event-Name": eventName,
        "Event-Domain": domainName,
        "User-ID": userId,
        "User-Domain": domainName,
        "User-Scheme":"account",
        "Content-Type":"text/event-json",
        "webitel-event-name":"user"
    };
};

function insertSession (account, domain, state, status, description, online) {
    if (account !== 'root') {
        handleStatusDb({
            "domain": domain,
            "account": account,
            "status": (status || "").toUpperCase(),
            "state": (state || "").toUpperCase(),
            "description": (description || ""),
            "online": online,
            "date": Date.now()
        }, (err) => {
            if (err)
                log.error(err);
        })
    }
};

function addMinutes(diff) {
    return Date.now() + diff * 1000;
};