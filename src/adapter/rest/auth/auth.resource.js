/**
 * Created by Igor Navrotskyj on 26.08.2015.
 */

'use strict';

const jwt = require('jwt-simple'),
    config = require(__appRoot + '/conf'),
    request = require('request'),
    log = require(__appRoot + '/lib/log')(module),
    CodeError = require(__appRoot + '/lib/error'),
    authService = require(__appRoot + '/services/auth'),
    aclService = require(__appRoot + '/services/acl'),
    cdrSrv = config.get('cdrServer'),
    vertoSocket = config.get('freeSWITCH:verto'),
    tokenSecretKey = require(__appRoot + '/utils/token');

const BASE_URL = config.get("server:baseUrl").replace(/\/$/, '');

function genGuid(data = []) {
    if (!data || data.length !== 16 ) {
        return null;
    }

    return data.slice(0,4).reverse().toString('hex') + "-" + data.slice(4,6).reverse().toString('hex') + "-"
        + data.slice(6,8).reverse().toString('hex') + "-" + data.slice(8,10).toString('hex')
        + "-" + data.slice(10).toString('hex');
}

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
    api.get( '/login/:domain', loginOAuth);
    api.post('/logout', logout);
}

function loginOAuth(req, res, next) {
    const {code} = req.query;
    const {domain} = req.params;

    if (!code || !domain) {
        return next(new CodeError(403, `Code or domain is required`))
    }

    let db = application.DB._query.domain;
    db.getAuthSettings(domain, (err, auth) => {
        if (err) {
            return next(err);
        }

        if (!auth) {
            return next(new CodeError(401, "Invalid credentials"))
        }

        request.post(auth.host, {
                rejectUnauthorized: false,
                form:{
                    code,
                    client_id: auth.client_id,
                    grant_type: "authorization_code",
                    redirect_uri: `${BASE_URL}/login/${domain}`
                }},
            (err, _, data) => {
                if (err) {
                    log.error(err);
                    return next(new CodeError(401, "Invalid credentials"))
                }
                try {
                    const token = jwt.decode(JSON.parse(data).access_token, auth.signature);
                    if (auth.user_id && token[auth.user_id]) {
                        return validateOAuthToken({domain, filter: {
                                    [auth.user_id]: auth.user_id_guid ? genGuid(Buffer.from(token[auth.user_id], 'base64')) : token[auth.user_id]
                            }},
                            res,
                            next
                        )
                    }

                    return next(new CodeError(401, "Invalid credentials"))
                } catch (e) {
                    log.error(e);
                    next(new CodeError(401, "Invalid credentials"));
                }
            });
    });
}

function validateOAuthToken(options, res, next) {
    authService.checkUserByFilter(options, (err, result) => {
        if (err)
            return next(err);

        if (result) {
            result.cdr = cdrSrv;
            result.verto = vertoSocket;
            return res
                .json(result);
        }

        return res
            .status(500)
            .json({
                "status": "error"
            });
    })
}

function login (req, res, next) {
    var username = req.body.username || '';
    var password = req.body.password || '';

    if (username === '') {
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
            result.verto = vertoSocket;
            return res
                .json(result);
        }

        return res
            .status(500)
            .json({
                "status": "error"
            });
    });
}

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
}

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
}

function decodeToken(token) {
    try {
        return jwt.decode(token, tokenSecretKey);
    } catch (e) {
        return null
    }
}

function validateRequestV2(req, res, next) {
    const token = (req.body && req.body.access_token) || (req.query && req.query.access_token) || req.headers['x-access-token'];
    const key = (req.body && req.body.x_key) || (req.query && req.query.x_key) || req.headers['x-key'];

    if (!token)
        return next(new CodeError(401, "Invalid token"));

    const decoded = decodeToken(token);

    if (!decoded)
        return next(new CodeError(401, "Invalid token"));

    if (decoded.exp <= Date.now()) {
        return next(new CodeError(401, "Token Expired"));
    }

    if (decoded.v === 2 && decoded.t === 'domain') {
        authService.validateDomainKey(decoded.d, decoded.id, (err, data) => {
            if (err)
                return next(err);

            const tokenDb = data && data.tokens && data.tokens[0];

            if (!tokenDb)
                return next(new CodeError(401, "Not found token"));

            req['webitelUser'] = {
                id: `${decoded.id}@${decoded.d}`,
                domain: decoded.d,
                role: tokenDb.roleName,
                roleName: tokenDb.roleName
            };
            next(); // To move to next middleware
        });

    } else if (key) {

        // Authorize the user to see if s/he can access our resources

        authService.validateUser(key, function (err, dbUser) {
            if (dbUser && dbUser.token === token) {
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
    }
}

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
            'cdr': cdrSrv,
            'verto': vertoSocket
        });
    });
}