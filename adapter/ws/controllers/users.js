/**
 * Created by Igor Navrotskyj on 07.09.2015.
 */

'use strict';

var WebitelCommandTypes = require(__appRoot + '/const').WebitelCommandTypes,
    getCommandResponseJSON = require('./responceTemplate').getCommandResponseJSON,
    authService = require(__appRoot + '/services/auth'),
    chatService = require(__appRoot + '/services/chat'),
    log = require(__appRoot +  '/lib/log')(module);


module.exports = usersCtrl();

function usersCtrl () {
    var controller = {};
    controller[WebitelCommandTypes.WhoAmI.name] = whoami;
    controller[WebitelCommandTypes.Token.Generate.name] = generateToken;
    controller[WebitelCommandTypes.Sys.Message.name] = sendInternalMessage;
    return controller;
};

function whoami(caller, execId, args, ws) {
    if (!caller) {
        return getCommandResponseJSON(ws, execId, {body: "-ERR: user not auth."});
    };
    return getCommandResponseJSON(ws, execId, {body: JSON.stringify({
        "id": caller['id'],
        "domain": caller['domain'],
        "logged": caller['logged'],
        "sessionLength": caller['sessionLength'],
        "roleName": caller['roleName']
    })});
};

function generateToken (caller, execId, args, ws) {
    var diff = 24 * 60 * 60 * 1000; // + day
    authService.getTokenMaxExpires(caller, diff, function (err, res) {
        try {
            if (err)
                return getCommandResponseJSON(ws, execId, {body: "-ERR: " + err.message});
            return getCommandResponseJSON(ws, execId, {body: JSON.stringify(res)});
        } catch (e) {
            log.error(e);
        }
    });
};

function sendInternalMessage (caller, execId, args, ws) {
    chatService.sendInternalWSMessage(caller, args, function (err, result) {
        if (err)
            return getCommandResponseJSON(ws, execId, {body: "-ERR: " + err.message});

        return getCommandResponseJSON(ws, execId, {body: "+OK: " + result});
    });
};