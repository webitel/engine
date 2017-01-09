/**
 * Created by i.navrotskyj on 25.01.2016.
 */
'use strict';

const licenseService = require(__appRoot + '/services/license'),
    conf = require(__appRoot + '/conf'),
    url = require('url'),
    log = require(__appRoot + '/lib/log')(module),
    LICENSE_HOST = conf.get('licenseServer:host'),
    USE_LICENSE_API = `${conf.get('licenseServer:enabled')}` === 'true'
    ;


let licenseHostInfo,
    http;

if (USE_LICENSE_API) {
    licenseHostInfo = url.parse(LICENSE_HOST);
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

    if (USE_LICENSE_API)
        api.patch('/api/license/:cid/:sid', genLicense);
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
        path: `/api/v1/license/${option.cid}/${option.sid}`
    }, result => {
        res.status(result.statusCode);
        result.pipe(res);
    });

    request.on('error', (e) => {
        log.error(e);
    });

    request.end();
}