/**
 * Created by I. Navrotskyj on 30.10.17.
 */

"use strict";

const metadataService = require(__appRoot + '/services/metadata');

module.exports = {
    addRoutes: addRoutes
};

function addRoutes(api) {
    api.get('/api/v2/metadata/:object_name', get);
    api.post('/api/v2/metadata/:object_name', createOrReplace);

    api.delete('/api/v2/metadata/:object_name', remove);
}

function get(req, res, next) {
    let options = {
        object_name: req.params.object_name,
        domain: req.query.domain
    };

    metadataService.item(req.webitelUser, options, (err, result) => {
        if (err)
            return next(err);

        return res.status(200).json({
            "status": "OK",
            "data": result
        });
    });
}

function createOrReplace(req, res, next) {
    let options = {
        object_name: req.params.object_name,
        domain: req.query.domain,
        data: req.body
    };

    metadataService.createOrReplace(req.webitelUser, options, (err, result) => {
        if (err)
            return next(err);

        return res.status(200).json({
            "status": "OK",
            "data": result
        });
    });
}

function remove(req, res, next) {
    let options = {
        object_name: req.params.object_name,
        domain: req.query.domain
    };

    metadataService.remove(req.webitelUser, options, (err, result) => {
        if (err)
            return next(err);

        return res.status(200).json({
            "status": "OK",
            "data": result
        });
    });
}