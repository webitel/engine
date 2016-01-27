/**
 * Created by i.navrotskyj on 25.01.2016.
 */
'use strict';

var WebitelCommandTypes = require(__appRoot + '/const').WebitelCommandTypes,
    getCommandResponseJSON = require('./responseTemplate').getCommandResponseJSON,
    getCommandResponseJSONError = require('./responseTemplate').getCommandResponseJSONError,
    licenseService = require(__appRoot + '/services/license'),
    log = require(__appRoot +  '/lib/log')(module);


module.exports = licenseCtrl();

function licenseCtrl () {
    var controller = {};
    controller[WebitelCommandTypes.License.List.name] = list;
    controller[WebitelCommandTypes.License.Item.name] = item;
    controller[WebitelCommandTypes.License.Upload.name] = upload;
    controller[WebitelCommandTypes.License.Remove.name] = remove;
    return controller;
};

function list (caller, execId, args, ws) {
    licenseService.list(caller, function (err, result) {
        if (err)
            return getCommandResponseJSONError(ws, execId, err);

        return getCommandResponseJSON(ws, execId, result);
    });
};

function item (caller, execId, args, ws) {
    licenseService.item(caller, args, function (err, result) {
        if (err)
            return getCommandResponseJSONError(ws, execId, err);

        return getCommandResponseJSON(ws, execId, result);
    });
};

function upload (caller, execId, args, ws) {
    licenseService.upload(caller, args, function (err, result) {
        if (err)
            return getCommandResponseJSONError(ws, execId, err);

        return getCommandResponseJSON(ws, execId, result);
    });
};

function remove (caller, execId, args, ws) {
    licenseService.remove(caller, args, function (err, result) {
        if (err)
            return getCommandResponseJSONError(ws, execId, err);

        return getCommandResponseJSON(ws, execId, result);
    });
};