/**
 * Created by i.navrotskyj on 25.01.2016.
 */
'use strict';

var licenseService = require(__appRoot + '/services/license')
    ;

module.exports = {
    addRoutes: addRoutes
};

function addRoutes (api) {
    api.get('/api/v2/license', list);
    api.get('/api/v2/license/:id', item);
    api.put('/api/v2/license/', upload);
    api.delete('/api/v2/license/:id', remove);
};

function list (req, res, next) {
    licenseService.list(req.webitelUser, (e, data) => {
        if (e)
            return next(e);

        return res.json({
            "status": "OK",
            "info": data
        })
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