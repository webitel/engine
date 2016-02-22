/**
 * Created by Igor Navrotskyj on 31.08.2015.
 */

'use strict';

var WebitelCommandTypes = require(__appRoot + '/const').WebitelCommandTypes,
    getCommandResponseJSON = require('./responseTemplate').getCommandResponseJSON,
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
    var _all = args.all,
        eventName = args['event'];
    // TODO add ACL commands
    if (!caller)
        return getCommandResponseJSON(ws, execId, {"body": "-ERR: Authentication required!"});
    eventService.addListener(eventName, caller, caller.getSession(ws), function (err, result) {
        let _result = {
            "body": ""
        };
        if (err) {
            _result.body = err.message
        } else {
            _result.body = result;
            if (_all)
                caller._subscribeEvent[eventName] = true;
        }

        getCommandResponseJSON(ws, execId, _result);
    });
};

function unSubscribe (caller, execId, args, ws) {
    var eventName = args['event'];
    // TODO add ACL commands
    if (!caller)
        return getCommandResponseJSON(ws, execId, {"body": "-ERR: Authentication required!"});
    eventService.removeListener(eventName, caller, caller.getSession(ws), function (err, result) {
        var _result = {
            "body": (err && err.message) || result
        };
        if (caller._subscribeEvent.hasOwnProperty(eventName))
            delete caller._subscribeEvent[eventName];

        getCommandResponseJSON(ws, execId, _result);
    });
};