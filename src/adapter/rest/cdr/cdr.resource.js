/**
 * Created by Igor Navrotskyj on 25.08.2015.
 */

'use strict';

const conf = require(__appRoot + '/conf'),
    log = require(__appRoot +  '/lib/log')(module),
    middlewareRequest = `${conf.get('cdrServer:useProxy')}` === 'true' ? proxyToCdr : redirectToCdr;

let CDR_SERVER_HOST = conf.get('cdrServer:host');

const proxy = require('http-proxy').createProxyServer({});
proxy.on('error', (e) => {
    log.error(e);
});

module.exports = {
    addRoutes: addRoutes
};

/**
 * Adds routes to the api.
 */

function addRoutes(api) {
    api.all(/^\/api\/v2\/r\/(cdr|files|media|tcp_dump|statistic)/, getRedirectUrl);
    api.all(/^\/api\/v2\/(cdr|files|media|tcp_dump|statistic)/, middlewareRequest);
}

if (CDR_SERVER_HOST) {
    CDR_SERVER_HOST = CDR_SERVER_HOST.replace(/\/$/g, '');
}

function redirectToCdr(request, response, next) {
    response.redirect(307, CDR_SERVER_HOST + request.originalUrl);
}

function proxyToCdr(req, res, next) {
    proxy.web(req, res, { target: CDR_SERVER_HOST}, (e) => {
        if (e) {
            log.error(e);
        }
    });
}

function getRedirectUrl(req, res, next) {
    if (!CDR_SERVER_HOST) {
        return res.status(500).json({
            "status": "error",
            "info": "Not config CDR_SERVER_HOST"
        });
    }
    res.status(200).json({
        "status": "OK",
        "info": CDR_SERVER_HOST + req.originalUrl.replace(/(\/api\/v2\/)(r\/)/, '$1')
    });
}