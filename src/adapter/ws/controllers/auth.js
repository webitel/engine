/**
 * Created by Admin on 04.08.2015.
 */
'use strict';

var WebitelCommandTypes = require(__appRoot + '/const').WebitelCommandTypes,
    authService = require(__appRoot + '/services/auth'),
    getCommandResponseJSON = require('./responseTemplate').getCommandResponseJSON,
    getCommandResponseJSONError = require('./responseTemplate').getCommandResponseJSONError,
    aclService = require(__appRoot + '/services/acl'),
    jwt = require('jwt-simple'),
    tokenSecretKey = require(__appRoot + '/utils/token'),
    log = require(__appRoot +  '/lib/log')(module);


module.exports = authCtrl();

function authCtrl () {
    var controller = {};
    controller[WebitelCommandTypes.Auth.name] = auth;
    controller[WebitelCommandTypes.Logout.name] = logout;
    controller[WebitelCommandTypes.Login.name] = login;
    return controller;
};

function auth(caller, execId, args, ws) {
    if (args.key && args.token) {
        var decoded;
        try {
            decoded = jwt.decode(args.token, tokenSecretKey);
        } catch (e) {
            return getCommandResponseJSONError(ws, execId, new Error("Invalid Token or Key"));
        }
        authService.validateUser(args.key, (e, dbUser) => {
            try {
                if (e) {
                    return getCommandResponseJSONError(ws, execId, e)
                }
                if (!dbUser || args.token !== dbUser.token) {
                    return getCommandResponseJSONError(ws, execId, new Error("Invalid Token or Key"));
                }
                if (decoded.exp <= Date.now()) {
                    return getCommandResponseJSONError(ws, execId, new Error("Token Expired"));
                }
                aclService._whatResources(dbUser.roleName, (e, acl) => {
                    if (e)
                        return getCommandResponseJSONError(ws, execId, e);

                    let webitelId = dbUser.username,
                        userData = {
                            'domain': dbUser.domain,
                            'state': "",
                            'status': "",
                            'cc-agent': "",
                            'acl': acl,
                            'roleName': dbUser.roleName,
                            // TODO obj role to string
                            'role': {name: dbUser.roleName}
                        };

                    ws.emit('webitelAuth', webitelId, userData, execId);
                });
            } catch (e) {
                log.warn('User socket close:', e.message);
            }
        })
    } else {
        authService.checkUser(
            args['account'],
            args['secret'],
            args['code'],
            (err, userParam) => {
                if (err) {
                    try {
                        ws.send(JSON.stringify({
                            'exec-uuid': execId,
                            'exec-complete': '+ERR',
                            'exec-response': {
                                'login': err
                            }
                        }));
                    } catch (e) {
                        log.warn('User socket close:', e.message);
                    }
                } else {
                    try {
                        var webitelId = args['account'];
                        ws.emit('webitelAuth', webitelId, userParam, execId);
                    } catch (e) {
                        log.warn('User socket close:', e.message);
                    }
                }
            }
        )
    }
};

function logout(caller, execId, args, ws) {
    getCommandResponseJSON(ws, execId, {body: "+OK: logged: " + caller.setLogged(false)});
};

function login(caller, execId, args, ws) {
    getCommandResponseJSON(ws, execId, {body: "+OK: logged: " + caller.setLogged(true)});
};