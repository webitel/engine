/**
 * Created by Admin on 04.08.2015.
 */
'use strict';

var WebitelCommandTypes = require(__appRoot + '/const').WebitelCommandTypes,
    authService = require(__appRoot + '/services/auth'),
    getCommandResponseJSON = require('./responceTemplate').getCommandResponseJSON,
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
    authService.checkUser(
        args['account'],
        args['secret'],
        function (err, userParam) {
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
                };
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
};

function logout(caller, execId, args, ws) {
    getCommandResponseJSON(ws, execId, {body: "+OK: logged: " + caller.setLogged(false)});
};

function login(caller, execId, args, ws) {
    getCommandResponseJSON(ws, execId, {body: "+OK: logged: " + caller.setLogged(true)});
};