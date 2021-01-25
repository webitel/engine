'use strict';

const conf = require('../../conf'),
    socketTimeUnauthorized = conf.get('application:socketTimeUnauthorized') || 10,
    log = require('../../lib/log')(module),
    handleError = require(__appRoot + '/middleware/handleWebSocketError'),
    generateUuid = require('node-uuid'),
    Controller = require('./controllers'),
    ACCOUNT_EVENTS = require(__appRoot + '/const').ACCOUNT_EVENTS,
    User = require(__appRoot + '/models/user'),
    handleStatusDb = require(__appRoot + '/services/userStatus').insert,
    DIFF_AGENT_LOGOUT_SEC = conf.get('application:callCentre:diffAgentLogoutTimeSec') || 60,
    PING_INTERVAL = +conf.get('server:socket:pingInterval') || 30000,
    getIp = require(__appRoot + '/utils/ip');

let maxUniqueOnline = 0;
let maxOpenedSocketPerUser = 10;

if (+conf.get('application:maxSocketPerUser') > 0) {
    maxOpenedSocketPerUser = +conf.get('application:maxSocketPerUser');
    log.info(`Maximum opened socket per user ${maxOpenedSocketPerUser}`)
}

if (+conf.get('application:auth:maxUniqueOnline')) {
    maxUniqueOnline = +conf.get('application:auth:maxUniqueOnline');
    log.info(`Maximum license connections ${maxUniqueOnline}`)
}

module.exports = Handler;

function sendMaxUser(params, execId, ws) {
    try {
        log.info(`Maximum license connections ${maxUniqueOnline}`);
        ws.send(JSON.stringify({
            'exec-uuid': execId,
            'exec-complete': '-ERR',
            'exec-response': 'Maximum license connections'
        }));
        ws.terminate();
    } catch (e) {
        log.warn('User socket close:', e.message);
    }
}

function sendMaxOpenedSocket(params, execId, ws, userId) {
    try {
        log.warn(`user ${userId} opened maximum sockets ${maxOpenedSocketPerUser}`);
        ws.send(JSON.stringify({
            'exec-uuid': execId,
            'exec-complete': '-ERR',
            'exec-response': 'Maximum opened socket ' + maxOpenedSocketPerUser
        }));
        ws.terminate();
    } catch (e) {
        log.warn('User socket close:', e.message);
    }
}


function Handler(wss, application) {
    const controller = Controller(application);

    wss.on('connection', function(ws, req) {

        const ipAddr = getIp(req);
        ws.ipAddr = ipAddr;
        const sessionId = generateUuid.v4();

        ws.isAlive = true;

        log.trace(`Receive new connection from IP: ${ipAddr} Origin: ${req.headers.origin} Agent: ${req.headers['user-agent']}`);

        let caller = null,
            timerId;

        ws.webitelUserId = null;

        function onAuth (id, params, execId) {
            clearTimeout(timerId);
            ws.webitelUserId = id;
            let user = application.Users.get(id);

            if (!user) {
                if (maxUniqueOnline && id !== 'root' && application.Users.lengthUsers() >= maxUniqueOnline) {
                    return sendMaxUser(params, execId, ws)
                }
                user = new User(id, sessionId, ws, params);
                application.Users.add(id, user);
            } else {
                if (user.getSessionLength() >= maxOpenedSocketPerUser) {
                    if (id !== 'root') {
                        return sendMaxOpenedSocket(params, execId, ws, id)
                    }
                }
                user.addSession(sessionId, ws, params);
            }

            caller = user;
            // test leak;
            //caller.bigData = new Array(1e6).join('XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX\n');
            const response = {
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
            log.trace('New pear %s [%s]', caller && caller.id, ipAddr);
        }

        ws.once('webitelAuth', onAuth);

        if (socketTimeUnauthorized > 0) {
            timerId = setTimeout(function () {
                if (!caller) {
                    log.trace('Disconnect session %s, remoteAddress: %s [unauth]', sessionId, ipAddr);
                    try {
                        ws.send(JSON.stringify({
                            "webitel-event-name": "account",
                            "Event-Name": "DISCONNECT"
                        }));
                        ws.terminate();
                    } catch (e) {
                        log.warn('User socket close:', e.message);
                    }
                }
            }, socketTimeUnauthorized * 1000);
        }

        ws.on('message', function(message) {
            try {
                log.trace('[%s->%s] received: %s', caller && caller.id, ipAddr, message);

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
                }
            } catch (e) {
                handleError(ws);
                log.error('Command error:', e.message);
            }
        });

        ws.on('close', function () {
            try {
                ws.removeListener('webitelAuth', onAuth);
                log.trace('Close session %s', sessionId);
                if (caller && caller.removeSession(sessionId) === 0) {
                    application.Users.remove(caller.id);
                    log.trace('disconnect: ', caller.id);
                    log.debug('Users session: ', application.Users.length());
                }

            } catch (e) {
                log.error(e);
            }
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
                if (application.Users.existsKey(key)) {
                    application.loggedOutAgent.remove(key);
                    continue
                }

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
    application.Users.lengthUsers = function () {
        const allLength = this.length();
        if (this.existsKey('root')) {
            return allLength - 1;
        }
        return allLength;
    };
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
    //TODO
    return;
    if (account !== 'root') {
        handleStatusDb({
            "Account-Domain": domain,
            "Account-User": account,
            "Account-Status": (status || "").toUpperCase(),
            "Account-User-State": (state || "").toUpperCase(),
            "Account-Status-Descript": (description || ""),
            "ws": online
        }, (err) => {
            if (err)
                log.error(err);
        })
    }
};

function addMinutes(diff) {
    return Date.now() + diff * 1000;
};
