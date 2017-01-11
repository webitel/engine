/**
 * Created by i.navrotskyj on 25.01.2016.
 */
'use strict';

const licenseService = require(__appRoot + '/services/license'),
    conf = require(__appRoot + '/conf'),
    url = require('url'),
    log = require(__appRoot + '/lib/log')(module),
    CodeError = require(__appRoot + '/lib/error'),
    LICENSE_HOST = conf.get('licenseServer:host'),
    IS_MASTER = `${conf.get('licenseServer:master')}` === 'true',
    USE_LICENSE_API = `${conf.get('licenseServer:enabled')}` === 'true'
    ;


let licenseHostInfo,
    http;

if (USE_LICENSE_API) {
    licenseHostInfo = url.parse(LICENSE_HOST);
    licenseHostInfo.pathname = '/' + licenseHostInfo.pathname.replace(/^\/|\/$/g,'');
    if (licenseHostInfo.pathname.length === 1)
        licenseHostInfo.pathname = "";
    http = (licenseHostInfo.protocol === 'http:') ? require('http') : require('https')
}

module.exports = {
    addRoutes: addRoutes
};

function addRoutes (api) {
    api.get('/api/v2/license', list);
    api.get('/api/v2/license/', list);
    api.get('/api/v2/license/:id', item);
    api.put('/api/v2/license/', upload);
    api.delete('/api/v2/license/:id', remove);

    if (IS_MASTER && USE_LICENSE_API) {
        api.get('/api/v2/customers', proxyCustomers);
        api.get('/api/v2/customers/:cid', proxyCustomers);
        api.post('/api/v2/customers', proxyCustomers);
        api.put('/api/v2/customers/:cid', proxyCustomers);
        api.delete('/api/v2/customers/:cid', proxyCustomers);
    }

    if (USE_LICENSE_API) {
        api.patch('/api/license/:cid/:sid', genLicense);
        api.get('/api/license/:cid', getLicenseInfo);
    }
}

function proxyCustomers(req, res, next) {

    if (req.webitelUser.id !== 'root')
        return next(new CodeError(401, "Forbidden."));

    const request = http.request({
        method: req.method,
        host: licenseHostInfo.hostname,
        port: licenseHostInfo.port,
        path: licenseHostInfo.pathname + req.url,
        headers: {
            "content-type": req.headers['content-type']
        }
    }, result => {
        _setHeader(result, res);

        res.status(result.statusCode);
        result.pipe(res);
    });

    request.on('error', (e) => {
        log.error(e);
    });

    if (req.method !== 'GET' && +req.headers['content-length'] > 0 && Object.keys(req.body).length > 0 ) {
        request.write(JSON.stringify(req.body))
    }

    request.end();
}

function list (req, res, next) {
    var addSid = req.query['sid'] === "true";
    licenseService.list(req.webitelUser, (e, data) => {
        if (e)
            return next(e);

        let json = {
            "status": "OK",
            "info": data
        };
        if (addSid)
            json.sid = application.WConsole._serverId
        ;

        return res.json(json)
    });
};

function item (req, res, next) {
    let option = {
        'cid': req.params.id
    };
    licenseService.item(req.webitelUser, option, (e, data) => {
        if (e)
            return next(e);

        return res.json({
            "status": "OK",
            "info": data
        })
    });
};

function upload (req, res, next) {
    let option = {
        'token': req.body.token
    };
    licenseService.upload(req.webitelUser, option, (e, data) => {
        if (e)
            return next(e);

        return res.json({
            "status": "OK",
            "info": data
        })
    });
};

function remove (req, res, next) {
    let option = {
        'cid': req.params.id
    };
    licenseService.remove(req.webitelUser, option, (e, data) => {
        if (e)
            return next(e);

        return res.json({
            "status": "OK",
            "info": data
        })
    });
};

function genLicense(req, res, next) {
    let option = {
        'cid': req.params.cid,
        'sid': req.params.sid
    };

    const request = http.request({
        method: "PATCH",
        host: licenseHostInfo.hostname,
        port: licenseHostInfo.port,
        path: `${licenseHostInfo.pathname}/api/license/${option.cid}/${option.sid}`
    }, result => {

        _setHeader(result, res);

        res.status(result.statusCode);
        result.pipe(res);
    });

    request.on('error', (e) => {
        log.error(e);
    });

    request.end();
}

function getLicenseInfo(req, res, next) {
    let option = {
        'cid': req.params.cid
    };

    const request = http.request({
        method: "GET",
        host: licenseHostInfo.hostname,
        port: licenseHostInfo.port,
        path: `${licenseHostInfo.pathname}/api/license/${option.cid}`
    }, result => {
        
        _setHeader(result, res);

        res.status(result.statusCode);
        result.pipe(res);
    });

    request.on('error', (e) => {
        log.error(e);
    });

    request.end();
}


function _setHeader(result, res) {
    if (result.headers.hasOwnProperty('content-type'))
        res.header('content-type', result.headers['content-type']);

    if (result.headers.hasOwnProperty('content-length'))
        res.header('content-length', result.headers['content-length']);

    if (result.headers.hasOwnProperty('transfer-encoding'))
        res.header('transfer-encoding', result.headers['transfer-encoding']);
}