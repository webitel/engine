/**
 * Created by Igor Navrotskyj on 09.09.2015.
 */

'use strict';
var WebitelCommandTypes = require(__appRoot + '/const').WebitelCommandTypes,
    getCommandResponseJSON = require('./responseTemplate').getCommandResponseJSON,
    log = require(__appRoot +  '/lib/log')(module),
    blacklistService = require(__appRoot + '/services/blacklist');


module.exports = blackListCtrl();

function blackListCtrl () {
    var controller = {};
    controller[WebitelCommandTypes.BlackList.GetNames.name] = getNames;
    controller[WebitelCommandTypes.BlackList.List.name] = listBlackList;
    controller[WebitelCommandTypes.BlackList.Create.name] = create;
    controller[WebitelCommandTypes.BlackList.Search.name] = search;
    controller[WebitelCommandTypes.BlackList.GetFromName.name] = getFromName;
    controller[WebitelCommandTypes.BlackList.RemoveNumber.name] = removeNumber;
    controller[WebitelCommandTypes.BlackList.RemoveName.name] = removeName;
    return controller;
};

function getNames(caller, execId, args, ws) {
    blacklistService.getNames(caller, args['domain'],
        function (err, result) {
            var body;
            if (err) {
                body = '-ERR: ' + err.message
            } else {
                body = result
            };

            getCommandResponseJSON(ws, execId, body);
        }
    );
};

function listBlackList(caller, execId, args, ws) {
    blacklistService.getFromName(
        caller,
        args['name'],
        args['domain'],
        args['option'],
        function (err, result) {
            var body;
            if (err) {
                body = '-ERR: ' + err.message
            } else {
                body = result
            };

            getCommandResponseJSON(ws, execId, body);
        }
    );
};

function create(caller, execId, args, ws) {

    blacklistService.create(caller, args,
        function (err, result) {
            var body;
            if (err) {
                body = '-ERR: ' + err.message
            } else {
                body = result
            };

            getCommandResponseJSON(ws, execId, body);
        }
    );
};

function search(caller, execId, args, ws) {
    blacklistService.search(caller, args['domain'], args,
        function (err, result) {
            var body;
            if (err) {
                body = '-ERR: ' + err.message
            } else {
                body = result
            };

            getCommandResponseJSON(ws, execId, body);
        }
    );
};

function getFromName(caller, execId, args, ws) {
    blacklistService.getFromName(
        caller,
        args['name'],
        args['domain'],
        args['parameters'],
        function (err, result) {
            var body;
            if (err) {
                body = '-ERR: ' + err.message
            } else {
                body = result
            };

            getCommandResponseJSON(ws, execId, body);
        }
    );
};

function removeNumber(caller, execId, args, ws) {
    if (!args['name'] || !args['number']) {
        return getCommandResponseJSON(ws, execId, '-ERR: bad parameters.');
    };

    var option = {
        "name": args['name'],
        "number": args['number']
    };
    blacklistService.remove(caller, args['domain'], option,
        function (err, result) {
            var body;
            if (err) {
                body = '-ERR: ' + err.message
            } else {
                body = result
            };

            getCommandResponseJSON(ws, execId, body);
        }
    );
};

function removeName(caller, execId, args, ws) {
    if (!args['name']) {
        return getCommandResponseJSON(ws, execId, '-ERR: bad parameters.');
    };

    var option = {
        "name": args['name']
    };
    blacklistService.remove(caller, args['domain'], option,
        function (err, result) {
            var body;
            if (err) {
                body = '-ERR: ' + err.message
            } else {
                body = result
            };

            getCommandResponseJSON(ws, execId, body);
        }
    );
};