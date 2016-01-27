/**
 * Created by Igor on 30.09.2015.
 */

'use strict';

var WebitelCommandTypes = require(__appRoot + '/const').WebitelCommandTypes,
    getCommandResponseJSON = require('./responseTemplate').getCommandResponseJSON,
    getCommandResponseJSONError = require('./responseTemplate').getCommandResponseJSONError,
    gatewayService = require(__appRoot + '/services/gateway'),
    log = require(__appRoot +  '/lib/log')(module)
    ;

module.exports = gatewayCtrl();

function gatewayCtrl () {
    var controller = {};
    controller[WebitelCommandTypes.Gateway.List.name] = list;
    controller[WebitelCommandTypes.Gateway.Create.name] = create;
    controller[WebitelCommandTypes.Gateway.Remove.name] = remove;
    controller[WebitelCommandTypes.Gateway.Change.name] = change;
    controller[WebitelCommandTypes.Gateway.Up.name] = up;
    controller[WebitelCommandTypes.Gateway.Down.name] = down;
    controller[WebitelCommandTypes.Gateway.Kill.name] = kill;
    controller[WebitelCommandTypes.SipProfile.List.name] = listProfile;
    controller[WebitelCommandTypes.SipProfile.Rescan.name] = rescanProfile;
    return controller;
};

function list(caller, execId, args, ws) {
    var domain = args.domain;
    gatewayService.listGateway(caller, domain, function (err, result) {
        if (err)
            return getCommandResponseJSONError(ws, execId, err);
        // TODO new Response
        //getCommandResponseJSON(ws, execId, result);

        return getCommandResponseJSON(ws, execId, {response: result});
    }, 'plain');
};

function create(caller, execId, args, ws) {
    gatewayService.createGateway(caller, args, function (err, result) {
        if (err)
            return getCommandResponseJSONError(ws, execId, err);
        // TODO parse json ?
        //getCommandResponseJSON(ws, execId, result);

        // TODO new Response
        return getCommandResponseJSON(ws, execId, JSON.stringify({status: 'OK', gateway: result}));
    });
};

function remove(caller, execId, args, ws) {
    gatewayService.deleteGateway(caller, args['name'], function (err, result) {
        if (err)
            return getCommandResponseJSONError(ws, execId, err);

        getCommandResponseJSON(ws, execId, result);
    });
};

function change(caller, execId, args, ws) {
    gatewayService.changeGateway(caller, args['name'], args['type'], args['params'] || {}, function (err, result) {
        if (err)
            return getCommandResponseJSONError(ws, execId, err);

        // TODO new Response
        //getCommandResponseJSON(ws, execId, result);

        return getCommandResponseJSON(ws, execId, {response: result});
    }, 'plain');
};

function up(caller, execId, args, ws) {
    gatewayService.upGateway(caller, args['name'], args['profile'], function (err, result) {
        if (err)
            return getCommandResponseJSONError(ws, execId, err);

        // TODO new Response
        //getCommandResponseJSON(ws, execId, result);
        return getCommandResponseJSON(ws, execId, {response: result});
    });
};

function down(caller, execId, args, ws) {
    gatewayService.downGateway(caller, args['name'], function (err, result) {
        if (err)
            return getCommandResponseJSONError(ws, execId, err);

        // TODO new Response
        //getCommandResponseJSON(ws, execId, result);
        return getCommandResponseJSON(ws, execId, {response: result});
    });
};

function kill(caller, execId, args, ws) {
    var option = {
        "profile": args['profile'] || "",
        "gateway": args['gateway'] || ""
    };

    gatewayService.killGateway(caller, option, function (err, result) {
        if (err)
            return getCommandResponseJSONError(ws, execId, err);

        getCommandResponseJSON(ws, execId, result);
    });
};

function listProfile(caller, execId, args, ws) {
    var option = args['domain'];

    gatewayService.listSipProfile(caller, option, function (err, result) {
        if (err)
            return getCommandResponseJSONError(ws, execId, err);

        getCommandResponseJSON(ws, execId, result);
    });
};

function rescanProfile(caller, execId, args, ws) {
    var option = {
        "profile": args['profile']
    };

    gatewayService.rescanSipProfile(caller, option, function (err, result) {
        if (err)
            return getCommandResponseJSONError(ws, execId, err);

        getCommandResponseJSON(ws, execId, result);
    });
};