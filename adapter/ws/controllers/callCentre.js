/**
 * Created by Igor Navrotskyj on 14.09.2015.
 */

'use strict';

var WebitelCommandTypes = require(__appRoot + '/const').WebitelCommandTypes,
    getCommandResponseJSON = require('./responceTemplate').getCommandResponseJSON,
    callCentreService = require(__appRoot + '/services/callCentre'),
    log = require(__appRoot +  '/lib/log')(module);


module.exports = callCentreCtrl();

function callCentreCtrl () {
    var controller = {};
    controller[WebitelCommandTypes.CallCenter.Login.name] = login;
    controller[WebitelCommandTypes.CallCenter.Logout.name] = logout;
    controller[WebitelCommandTypes.CallCenter.Tier.List.name] = tiersFromUser;
    return controller;
};

function login(caller, execId, args, ws) {
    callCentreService.login(caller, args, function (err, res) {
        let result = res;
        if (err)
            result = '-ERR: ' + err.message;

        getCommandResponseJSON(ws, execId, result);
    });
};

function logout(caller, execId, args, ws) {
    callCentreService.logout(caller, args, function (err, res) {
        let result = res;
        if (err)
            result = '-ERR: ' + err.message;

        getCommandResponseJSON(ws, execId, result);
    });
};

function tiersFromUser (caller, execId, args, ws) {
    callCentreService.getTiersFromCaller(caller, args, function (err, res) {
        let result = res;
        if (err)
            result = '-ERR: ' + err.message;

        getCommandResponseJSON(ws, execId, result);
    });
};