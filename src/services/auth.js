/**
 * Created by i.n on 04.08.2015.
 */
'use strict';

var conf = require('../conf'),
    RootName = conf.get("webitelServer:account"),
    jwt = require('jwt-simple'),
    RootPassword = conf.get("webitelServer:secret"),
    crypto = require('crypto'),
    ACCOUNT_ROLE = require( __appRoot + '/const').ACCOUNT_ROLE,
    generateUuid = require('node-uuid'),
    CodeError = require(__appRoot + '/lib/error')
    ;


module.exports = {
    checkUser: checkUser,
    validateUser: validateUser,
    baseAuth: baseAuth,
    login: login,
    logout: logout,
    getTokenMaxExpires: getTokenMaxExpires,
    removeFromUserName: removeFromUserName,
    _removeDomainsTokens: _removeDomainsTokens
};

function logout(option, cb) {
    var key = option['key'] || '',
        token = option['token'] || '';

    if (key == '') {
        return cb(new CodeError(400, 'Bad key.'));
    };

    validateUser(key, function (err, user) {
        if (err) {
            return cb(err);
        };

        if (user && user['token'] == token) {
            return removeKey(key, cb);
        } else {
            return cb(new CodeError(401, 'Invalid credentials.'));
        };
    });
};

function login (option, cb) {
    var username = option['username'] || '',
        password = option['password'] || ''
        ;

    if (username == '') {
        return cb(new CodeError(401, 'Invalid credentials'));
    };

    return getTokenObject(username, password, cb);
};

function baseAuth(option, cb) {
    var username = option['username'] || '',
        password = option['password'] || ''
        ;

    if (RootName != username || RootPassword != password) {
        return cb(new CodeError(401, "Invalid credentials"));
    } else {
        return cb();
    }
};

function getTokenObject (username, password, cb) {
    var _id = generateUuid.v4();
    return validate(username, password, _id, cb);
};

function removeKey(key, cb) {
    var authDb = application.DB._query.auth;
    authDb.remove(key, cb);
};

function validateUser(key, cb) {
    try {
        var authDb = application.DB._query.auth;
        authDb.getByKey(key, cb);
    } catch (e){
        cb(e);
    };
};

function checkUser (login, password, cb) {
    try {
        login = login || '';
        password = password || '';
        if (login === RootName) {
            if (password === RootPassword) {
                cb(null, {
                    'role': ACCOUNT_ROLE.ROOT,
                    'roleName': 'root'
                });
            } else {
                cb(new CodeError(401, 'secret incorrect'));
            };
            return
        };

        application.WConsole.userDara(login, 'global', ['a1-hash', 'account_role', 'cc-agent', 'status', 'state', 'description'], function (res) {
            try {
                var resJson = JSON.parse(res['body']);
            } catch (e) {
                cb(new CodeError(401, res['body'] || e.message));
                return;
            };
            var a1Hash = md5(login.replace('@', ':') + ':' + password);
            var registered = (a1Hash == resJson['a1-hash']);

            if (registered) {
                cb(null, {
                    'role': ACCOUNT_ROLE.getRoleFromName(resJson['account_role']),
                    'roleName': resJson['account_role'],
                    'status': resJson['status'],
                    'state': resJson['state'],
                    'domain': login.split('@')[1],
                    'cc-agent': resJson['cc-agent'],
                    'description': decodeURI(resJson['description'] || "")
                });
            } else {
                cb(new CodeError(401, 'Bad password.'));
            };
        });

    } catch (e) {
        cb(e);
    };
};

function validate (username, password, _id, cb) {
    checkUser(username, password, function (err, user) {
        if (err) {
            return cb(err);
        };

        var tokenObj = genToken(username),
            userObj = {
                "key": _id,
                "domain": user.domain,
                "username": username,
                "expires": tokenObj.expires,
                "token": tokenObj.token,
                "role": user.role.val,
                "roleName": user.role.name
            };
        var authDb = application.DB._query.auth;
        authDb.add(userObj, function (err) {
            return cb(err, userObj)
        });

    });
};

function getTokenMaxExpires (caller, diff, cb) {
    if (!caller) {
        return cb(new CodeError(401, "Bad caller."));
    };
    var expires = new Date().getTime();
    if (diff)
        expires += diff;

    var authDb = application.DB._query.auth;
    authDb.getByUserName(caller['id'], expires, function (err, res) {
        if (err) {
            return cb(err);
        };

        if (res && res.length > 0) {
            return cb(null, res[0]);
        };

        try {
            var _id = generateUuid.v4();
            var tokenObj = genToken(caller['id']),
                userObj = {
                    "key": _id,
                    "domain": caller['domain'],
                    "username": caller['id'],
                    "expires": tokenObj.expires,
                    "token": tokenObj.token,
                    "role": caller['role'] && caller['role']['val'],
                    "roleName": caller['roleName']
                };
            var authDb = application.DB._query.auth;
            authDb.add(userObj, function (err) {
                return cb(err, userObj)
            });
        } catch (e) {
            return cb(e);
        }

    });
};

function removeFromUserName (username, domain, cb) {
    if (!username) {
        return cb(new CodeError(400, 'User name is required.'));
    };
    var authDb = application.DB._query.auth;
    return authDb.removeUserTokens(username, domain, cb);
};

function _removeDomainsTokens (domain, cb) {
    if (!domain) {
        return cb(new CodeError(400, 'Domain is required.'));
    };
    var authDb = application.DB._query.auth;
    return authDb.removeDomainTokens(domain, cb);
}

var md5 = function (str) {
    var hash = crypto.createHash('md5');
    hash.update(str);
    return hash.digest('hex');
};

function genToken(user) {
    var expires = expiresIn(conf.get('application:auth:expiresDays'));
    var token = jwt.encode({
        exp: expires
    }, conf.get('application:auth:tokenSecretKey'));

    return {
        token: token,
        expires: expires,
        user: user
    };
};

function expiresIn(numDays) {
    var dateObj = new Date();
    return dateObj.setDate(dateObj.getDate() + numDays);
};