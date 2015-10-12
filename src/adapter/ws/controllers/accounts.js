/**
 * Created by Igor Navrotskyj on 06.08.2015.
 */

'use strict';

var WebitelCommandTypes = require(__appRoot + '/const').WebitelCommandTypes,
    getCommandResponseJSON = require('./responceTemplate').getCommandResponseJSON,
    getCommandResponseJSONError = require('./responceTemplate').getCommandResponseJSONError,
    accountService = require(__appRoot + '/services/account'),
    log = require(__appRoot +  '/lib/log')(module)
    ;


module.exports = accountCtrl();

function accountCtrl () {
    var controller = {};
    controller[WebitelCommandTypes.Account.List.name] = accountList;
    controller[WebitelCommandTypes.Account.Create.name] = create;
    controller[WebitelCommandTypes.ListUsers.name] = userList;
    //TODO
    controller[WebitelCommandTypes.Account.Change.name] = update;
    controller[WebitelCommandTypes.Account.Remove.name] = remove;
    controller[WebitelCommandTypes.Account.Item.name] = item;
    return controller;
};

function accountList(caller, execId, args, ws) {
    accountService.accountList(caller, args, function (err, result) {
        if (err)
            return getCommandResponseJSONError(ws, execId, err);

        getCommandResponseJSON(ws, execId, result);
    });
};

function userList(caller, execId, args, ws) {
    accountService.list(caller, args, function (err, result) {
        if (err)
            return getCommandResponseJSONError(ws, execId, err);

        getCommandResponseJSON(ws, execId, result);
    });
};

function create(caller, execId, args, ws) {
    try {
        var param = args['param'];
        if (param) {
            let _t = param.split('@'),
                _u = _t[0].split(':');
            args['domain'] = _t[1];
            args['login'] = _u[0];
            args['password'] = _u[1];
        };
        if (args.attribute) {
            for (var key in args.attribute) {
                if (args.attribute.hasOwnProperty(key))
                    args[key] = args.attribute[key];
            };
        };

        accountService.create(caller, args, function (err, result) {
            if (err)
                return getCommandResponseJSONError(ws, execId, err);

            getCommandResponseJSON(ws, execId, result);
        });
    } catch (e) {
        log.error(e);
        getCommandResponseJSONError(ws, execId, e);
    };
};

function update (caller, execId, args, ws) {
    let _t = args['user'] && args['user'].split('@'),
        param = args.param,
        option
        ;
    if (param instanceof Object) {
        option = param
    } else {
        option = {
            "parameters": [param + '=' + args.value]
        };
    };

    accountService.update(caller, _t[0], _t[1], option, function (err, result) {
        if (err)
            return getCommandResponseJSONError(ws, execId, err);

        getCommandResponseJSON(ws, execId, result);
    });
};

function remove(caller, execId, args, ws) {
    let _t = args.user && args.user.split('@');
    let option = {
        "name": _t[0],
        "domain": _t[1]
    };
    accountService.remove(caller, option, function (err, result) {
        if (err)
            return getCommandResponseJSONError(ws, execId, err);

        getCommandResponseJSON(ws, execId, result);
    });
};

function item(caller, execId, args, ws) {
    let _t = (args.user && args.user.split('@')) || [];
    let option = {
        "name": _t[0],
        "domain": _t[1]
    };
    accountService.item(caller, option, function (err, result) {
        if (err)
            return getCommandResponseJSONError(ws, execId, err);
        // TODO del JSON
        getCommandResponseJSON(ws, execId, JSON.stringify(result));
    });
};