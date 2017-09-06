/**
 * Created by Igor Navrotskyj on 31.08.2015.
 */

'use strict';

var WebitelCommandTypes = require(__appRoot + '/const').WebitelCommandTypes,
    getCommandResponseJSON = require('./responseTemplate').getCommandResponseJSON,
    log = require(__appRoot +  '/lib/log')(module)
    ;

module.exports = eventsCtrl();

function eventsCtrl () {
    const controller = {};
    controller[WebitelCommandTypes.Event.On.name] = subscribe;
    controller[WebitelCommandTypes.Event.Off.name] = unSubscribe;
    return controller;
}

function subscribe (caller, execId, args, ws) {
    let _all = args.all,
        eventName = args['event'];
    // TODO add ACL commands
    if (!caller)
        return getCommandResponseJSON(ws, execId, {"body": "-ERR: Authentication required!"});

    const sessionId = caller.getSession(ws);

    caller.subscribe(eventName, sessionId, args, (err, result) => {
        let _result = {
            "body": ""
        };
        if (err) {
            _result.body = err.message
        } else {
            _result.body = result;
            // application.emit(`subscribe::${eventName}`, args, caller, eventName, sessionId);
            if (_all)
                caller._subscribeEvent[eventName] = true;
        }

        getCommandResponseJSON(ws, execId, _result);
    });
}

function unSubscribe (caller, execId, args, ws) {
    const eventName = args['event'];
    // TODO add ACL commands
    if (!caller)
        return getCommandResponseJSON(ws, execId, {"body": "-ERR: Authentication required!"});

    caller.unSubscribe(eventName, caller.getSession(ws), function (err, result) {
        if (caller._subscribeEvent.hasOwnProperty(eventName))
            delete caller._subscribeEvent[eventName];

        // application.emit(`unsubscribe::${eventName}`, args, caller, eventName);

        getCommandResponseJSON(ws, execId, {
            "body": (err && err.message) || result
        });
    });
}