/**
 * Created by Igor Navrotskyj on 26.08.2015.
 */

'use strict';

var jwt = require('jwt-simple'),
    config = require(__appRoot + '/conf'),
    CodeError = require(__appRoot + '/lib/error'),
    authService = require(__appRoot + '/services/auth'),
    aclService = require(__appRoot + '/services/acl'),
    cdrSrv = config.get('cdrServer'),
    tokenSecretKey = config.get('application:auth:tokenSecretKey');

module.exports = {
    addRoutes: addRoutes
};

/**
 * Adds routes to the api.
 */
function addRoutes(api) {
    api.all('/api/v1/*', validateRequestV1);
    api.all('/api/v2/*', validateRequestV2);
    api.get('/api/v2/whoami', whoami);
    api.post('/login', login);
    api.post('/logout', logout);
};

function login (req, res, next) {
    var username = req.body.username || '';
    var password = req.body.password || '';

    if (username == '') {
        res.status(401);
        res.json({
            "status": 401,
            "message": "Invalid credentials"
        });
        return;
    }

    authService.login({
        username: username,
        password: password
    }, function (err, result) {
        if (err) {
            return next(err);
        }

        if (result) {
            result.cdr = cdrSrv;
            return res
                .json(result);
        }

        return res
            .status(500)
            .json({
                "status": "error"
            });
    });
};

function logout (req, res, next) {
    try {
        var key = (req.body && req.body.x_key) || (req.query && req.query.x_key) || req.headers['x-key'];
        var token = (req.body && req.body.access_token) || (req.query && req.query.access_token) || req.headers['x-access-token'];
        if (!key || !token) {
            res.status(401);
            res.json({
                "status": 401,
                "message": "Invalid credentials"
            });
            return;
        };

        authService.logout({
            key: key,
            token: token
        }, function (err) {
            if (err) {
                return next(err);
            };

            res.status(200).json({
                "status": "OK",
                "info": "Successful logout."
            });
        });
    } catch (e) {
        next(e);
    }
};

function validateRequestV1(req, res, next) {
    try {
        var header = req.headers['authorization'] || '',
            token = header.split(/\s+/).pop() || '',
            auth = new Buffer(token, 'base64').toString(),
            parts = auth.split(/:/),
            username = parts[0],
            password = parts[1];

        return authService.baseAuth({
            "username": username,
            "password": password
        }, (err) => {
            if (err) return next(err);
            req['webitelUser'] = {
                id: 'root',
                domain: null,
                role: 'root',
                roleName: 'root'
            };
            return next();
        });

    } catch (err) {
        res.status(500);
        return res.json({
            "status": 500,
            "message": "Oops something went wrong",
            "error": err
        });
    }
};

function validateRequestV2(req, res, next) {
    var token = (req.body && req.body.access_token) || (req.query && req.query.access_token) || req.headers['x-access-token'];
    var key = (req.body && req.body.x_key) || (req.query && req.query.x_key) || req.headers['x-key'];

    if (token && key) {
        try {
            var decoded = jwt.decode(token, tokenSecretKey);
        } catch (e) {
            return next(new CodeError(401, "Invalid Token or Key"));
        };

        if (decoded.exp <= Date.now()) {
            return next(new CodeError(400, "Token Expired"));
        };

        // Authorize the user to see if s/he can access our resources

        authService.validateUser(key, function (err, dbUser) {
            if (dbUser && dbUser.token == token) {
                req['webitelUser'] = {
                    id: dbUser.username,
                    domain: dbUser.domain,
                    role: dbUser.role,
                    roleName: dbUser.roleName,
                    expires: dbUser.expires,
                    acl: dbUser.acl
                    //testLeak: new Array(1e6).join('X')
                };
                next(); // To move to next middleware
            } else {
                // No user with this name exists, respond back with a 401
                return next(new CodeError(401, "Invalid User"));
            }
        });
    } else {
        return next(new CodeError(401, "Invalid Token or Key"));
    };
};

function whoami(req, res, next) {
    aclService._whatResources(req.webitelUser.roleName, (e, acl) => {
        if (e)
            return next(e);

        var _user = req.webitelUser;
        return res.json({
            'acl': acl,
            'id': _user.id,
            'domain': _user.domain,
            'roleName': _user.roleName,
            'expires': _user.expires,
            'cdr': cdrSrv
        });
    });
};