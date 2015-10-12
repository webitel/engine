/**
 * Created by Igor Navrotskyj on 31.08.2015.
 */

'use strict';

var WebitelCommandTypes = require(__appRoot + '/const').WebitelCommandTypes,
    getCommandResponseJSON = require('./responceTemplate').getCommandResponseJSON,
    log = require(__appRoot +  '/lib/log')(module),
    eventService = require(__appRoot +  '/services/events')
    ;

module.exports = eventsCtrl();

function eventsCtrl () {
    var controller = {};
    controller[WebitelCommandTypes.Event.On.name] = subscribe;
    controller[WebitelCommandTypes.Event.Off.name] = unSubscribe;
    return controller;
};

function subscribe (caller, execId, args, ws) {
    eventService.addListener(args['event'], caller, caller.getSession(ws), function (err, result) {
        var _result = {
            "body": (err && err.message) || result
        };
        getCommandResponseJSON(ws, execId, _result);
    });
};

function unSubscribe (caller, execId, args, ws) {
    eventService.removeListener(args['event'], caller, caller.getSession(ws), function (err, result) {
        var _result = {
            "body": (err && err.message) || result
        };
        getCommandResponseJSON(ws, execId, _result);
    });
};