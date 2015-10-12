/**
 * Created by Igor Navrotskyj on 07.09.2015.
 */

'use strict';
var WebitelCommandTypes = require(__appRoot + '/const').WebitelCommandTypes,
    getCommandResponseJSON = require('./responceTemplate').getCommandResponseJSON,
    cdrService = require(__appRoot + '/services/cdr'),
    log = require(__appRoot +  '/lib/log')(module);


module.exports = cdrCtrl();

function cdrCtrl () {
    var controller = {};
    controller[WebitelCommandTypes.CDR.RecordCall.name] = existsRecordFile;
    return controller;
};

function existsRecordFile(caller, execId, args, ws) {
    var uuid = args['uuid'];
    cdrService.getRecordFile(caller, uuid, function (err, res) {
        if (err)
            return getCommandResponseJSON(ws, execId, {body: "-ERR: " + err.message});

        return getCommandResponseJSON(ws, execId, res);
    });
};