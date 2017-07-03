/**
 * Created by Igor Navrotskyj on 25.08.2015.
 */

'use strict';

var conf = require(__appRoot + '/conf'),
    http = require('http'),
    https = require('https'),
    url = require('url'),
    log = require(__appRoot +  '/lib/log')(module),
    CDR_SERVER_HOST = conf.get('cdrServer:host'),
    middlewareRequest = `${conf.get('cdrServer:useProxy')}` === 'true' ? proxyToCdr : redirectToCdr;
    ;

module.exports = {
    addRoutes: addRoutes
};

/**
 * Adds routes to the api.
 */

function addRoutes(api) {
    api.all(/^\/api\/v2\/r\/(cdr|files|media)/, getRedirectUrl);
    api.all(/^\/api\/v2\/(cdr|files|media)/, middlewareRequest);
};

if (CDR_SERVER_HOST) {
    CDR_SERVER_HOST = CDR_SERVER_HOST.replace(/\/$/g, '');
};

var cdrHostInfo = url.parse(CDR_SERVER_HOST);

var client = cdrHostInfo.protocol == 'http:' ? http.request : https.request;

var CDR_SERVER = {
    path: cdrHostInfo.path.length == 1 ? '' : cdrHostInfo.path,
    hostName: cdrHostInfo.hostname,
    port: parseInt(cdrHostInfo.port)
};

function redirectToCdr(request, response, next) {
    response.redirect(307, CDR_SERVER_HOST + request.originalUrl);
};

var httpProxy = require('http-proxy');
var proxy = httpProxy.createProxyServer({});
proxy.on('error', (e) => {
    log.error(e);
});

function proxyToCdr(req, res, next) {
    proxy.web(req, res, { target: CDR_SERVER_HOST}, (e) => {
        if (e) {
            log.error(e);
        }
    });
};

function getRedirectUrl(req, res, next) {
    if (!CDR_SERVER_HOST) {
        return res.status(500).json({
            "status": "error",
            "info": "Not config CDR_SERVER_HOST"
        });
    };
    res.status(200).json({
        "status": "OK",
        "info": CDR_SERVER_HOST + req.originalUrl.replace(/(\/api\/v2\/)(r\/)/, '$1')
    });
};

function getPort(req) {
    var matches = req.headers.host.match(/:(\d+)/);
    if (matches)
        return matches[1];
    return req.connection.pair ? '443' : '80';
};