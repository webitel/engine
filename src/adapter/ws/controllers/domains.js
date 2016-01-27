/**
 * Created by Igor on 30.09.2015.
 */

'use strict';

var WebitelCommandTypes = require(__appRoot + '/const').WebitelCommandTypes,
    getCommandResponseJSON = require('./responseTemplate').getCommandResponseJSON,
    getCommandResponseJSONError = require('./responseTemplate').getCommandResponseJSONError,
    domainService = require(__appRoot + '/services/domain'),
    log = require(__appRoot +  '/lib/log')(module)
    ;


module.exports = domainCtrl();

function domainCtrl () {
    var controller = {};
    controller[WebitelCommandTypes.Domain.List.name] = list;
    controller[WebitelCommandTypes.Domain.Create.name] = create;
    controller[WebitelCommandTypes.Domain.Remove.name] = remove;
    controller[WebitelCommandTypes.Domain.Item.name] = item;
    controller[WebitelCommandTypes.Domain.Update.name] = update;
    return controller;
};

function list(caller, execId, args, ws) {
    // TODO del
    args['type'] = 'plain';
    domainService.list(caller, args, function (err, result) {
        if (err)
            return getCommandResponseJSONError(ws, execId, err);
        // TODO new Response
        //return getCommandResponseJSON(ws, execId, result);

        // Delete after update lib;
        return getCommandResponseJSON(ws, execId, {response: result});
    });
};

function create(caller, execId, args, ws) {
    var param = args['parameters'];
    for (var key in param) {
        if (param.hasOwnProperty(key))
            args[key] = param[key];
    };
    domainService.create(caller, args, function (err, result) {
        if (err)
            return getCommandResponseJSONError(ws, execId, err);

        getCommandResponseJSON(ws, execId, result);
    });
};

function remove(caller, execId, args, ws) {
    domainService.remove(caller, args, function (err, result) {
        if (err)
            return getCommandResponseJSONError(ws, execId, err);

        getCommandResponseJSON(ws, execId, result);
    });
};

function item(caller, execId, args, ws) {
    domainService.item(caller, args, function (err, result) {
        if (err)
            return getCommandResponseJSONError(ws, execId, err);

        getCommandResponseJSON(ws, execId, result);
    });
};

function update(caller, execId, args, ws) {
    let option = {
            "name": args['name']
        },
        params = args['params'];
    if (params) {
        option['type'] = params.type;
        option['params'] = params.attribute;
    };
    domainService.update(caller, option, function (err, result) {
        if (err)
            return getCommandResponseJSONError(ws, execId, err);

        getCommandResponseJSON(ws, execId, result);
    });
};