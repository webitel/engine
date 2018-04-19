/**
 * Created by Igor Navrotskyj on 27.09.2015.
 */

'use strict';
var CodeError = require(__appRoot + '/lib/error'),
    validateCallerParameters = require(__appRoot + '/utils/validateCallerParameters'),
    plainTableToJSON = require(__appRoot + '/utils/parse').plainTableToJSON,
    plainTableToJSONArray = require(__appRoot + '/utils/parse').plainTableToJSONArray,
    plainCollectionToJSON = require(__appRoot + '/utils/parse').plainCollectionToJSON,
    checkPermissions = require(__appRoot + '/middleware/checkPermissions'),
    log = require(__appRoot + '/lib/log')(module)
    ;

var Service = {

    /**
     *
     * @param caller
     * @param option
     * @param cb
     */
    create: function (caller, option, cb) {
        checkPermissions(caller, 'account', 'c', function (err) {
            if (err)
                return cb(err);

            if (!option) {
                return cb(new CodeError(400, 'Bad request.'));
            };

            let login = option['login'],
                domain = validateCallerParameters(caller, option['domain']),
                role = option['role'],
                password = option['password'],
                extensions = option['extensions'],
                parameters = option['parameters'],
                variables = option['variables']
                ;

            if (!domain || !login || !role) {
                return cb(new CodeError(400, "Domain, login is require."));
            };

            let _param =[];
            _param.push(login);

            if (password && password != '')
                _param.push(':' + password);

            _param.push('@' + domain);

            let q = {
                "role": role,
                "param": _param.join(''),
                "attribute": {
                    "parameters": parameters,
                    "variables": variables,
                    "extensions": extensions
                }
            };

            application.WConsole.userCreate(caller, q, function (err, res) {
                if (err)
                    return cb(err);

                return cb(null, res);
            });
        });
    },

    /**
     *
     * @param caller
     * @param option
     * @param cb
     */
    item: function (caller, option, cb) {
        checkPermissions(caller, 'account', 'r', function (err) {
            if (err)
                return cb(err);

            if (!option) {
                return cb(new CodeError(400, 'Bad request.'));
            }

            let name = option['name'],
                domain = validateCallerParameters(caller, option['domain'])
                ;

            if (!domain || !name) {
                return cb(new CodeError(400, "Domain, login is require."));
            }

            application.WConsole.userItem(caller, name, domain, function (err, res) {
                if (err)
                    return cb(err);
                try {
                    return plainCollectionToJSON(res, cb);
                } catch (e) {
                    log.error(e);
                    cb(e);
                }
            });

        });
    },

    /**
     * 
     * @param caller
     * @param option
     * @param cb
     */
    list: function (caller, option, cb) {
        checkPermissions(caller, 'account', 'r', function (err) {
            if (err)
                return cb(err);

            if (!option) {
                return cb(new CodeError(400, 'Bad request.'));
            }

            let domain = validateCallerParameters(caller, option['domain']);

            if (!domain) {
                return cb(new CodeError(400, "Domain is require."));
            }

            application.WConsole.list_users(caller, domain, function (err, res) {
                if (err)
                    return cb(err);

                try {
                    return plainTableToJSON(res, null, cb);
                } catch (e) {
                    log.error(e);
                    cb(e);
                }
            });

        });
    },

    /**
     *
     * @param caller
     * @param option
     * @param cb
     */
    accountList: function (caller, option, cb) {
        checkPermissions(caller, 'account', 'r', function (err) {
            if (err)
                return cb(err);

            if (!option) {
                return cb(new CodeError(400, 'Bad request.'));
            }

            option['domain'] = validateCallerParameters(caller, option['domain']);

            if (!option['domain']) {
                return cb(new CodeError(400, "Domain is require."));
            }

            return application.WConsole.userList2(caller, option, function (err, res) {
                if (err)
                    return cb(err);

                try {
                    return parseAccount(res, option['domain'], cb);
                } catch (e) {
                    log.error(e);
                    cb(e);
                }
            });

        });
    },

    /**
     *
     * @param caller
     * @param userName
     * @param domain
     * @param option
     * @param cb
     */
    update: function (caller, userName, domain, option, cb) {
        let perm = caller.id !== userName + '@' + domain ? 'u' : 'uo';

        checkPermissions(caller, 'account', perm, function (err) {
            if (err)
                return cb(err);

            if (!option ||!userName) {
                return cb(new CodeError(400, 'Bad request.'));
            }

            let _domain = validateCallerParameters(caller, domain);

            if (!_domain) {
                return cb(new CodeError(400, "Domain is require."));
            }

            application.WConsole.userUpdateV2(caller, userName, _domain, option, cb);

        });
    },

    /**
     *
     * @param caller
     * @param option
     * @param cb
     */
    remove: function (caller, option, cb) {
        checkPermissions(caller, 'account', 'd', function (err) {
            if (err)
                return cb(err);

            if (!option) {
                return cb(new CodeError(400, 'Bad request.'));
            }

            let name = option['name'],
                domain = validateCallerParameters(caller, option['domain'])
                ;

            if (!domain || !name) {
                return cb(new CodeError(400, "Domain, login is require."));
            }

            let _id = name + '@' + domain;

            if (_id === caller.id) {
                return cb(new CodeError(400, "Easy! it's YOU !!!"));
            }

            application.WConsole.userRemove(caller, _id, cb);
        });
    },
    
    
    _listByDomain: function (domainName, cb) {
        application.WConsole.userList({}, domainName, function (err, res) {
            if (err)
                return cb(err);

            try {
                return parseAccount(res, domainName, cb);
            } catch (e) {
                log.error(e);
                cb(e);
            }
        });
    }
};

module.exports = Service;

function parseAccount (data, domain, cb) {
    if (typeof data  !== "string") {
        cb('Data is undefined!');
        return
    }

    const result = {};
    const lines = data.split('\n');
    lines.pop();
    lines.pop();
    lines.pop();
    const columns = lines.shift().split('|');

    let user, ws;

    lines.forEach( item => {
        user = item.split('|').reduce((res, val, key) => {
            res[columns[key]] = val;
            return res;
        }, {});

        user.id = user.user;
        delete user.user;

        ws = application.Users.get(user.id + '@' + domain);
        if (ws) {
            user.online = ws.logged;
            user.cc_logged = !!ws['cc-logged'];
        } else {
            user.online = user.cc_logged = false;

        }

        user.domain = domain;

        result[user.id] = user;
    });
    return cb(null, result);
}