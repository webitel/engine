/**
 * Created by i.navrotskyj on 09.12.2015.
 */
'use strict';

var configureService = require(__appRoot + '/services/configure');

module.exports = {
    addRoutes: addRoutes
};

function addRoutes (api) {
    api.put('/api/v2/system/reload/xml', reloadXml);
    api.put('/api/v2/system/reload/:modName', reloadFsModule);
    api.put('/api/v2/system/cache/clear', cache);

    //  V1
    api.get('/api/v1/reloadxml', reloadXml);
}

function reloadXml (req, res, next) {
    configureService.reloadXml(req.webitelUser, (err, result) => {
        if (err)
            return next(err);

        return res
            .status(200)
            .json({
                "status": "OK",
                "info": result
            })
    });
}

function reloadFsModule(req, res, next) {
    configureService.reloadMod(req.webitelUser, req.params.modName,  (err, result) => {
        if (err)
            return next(err);

        return res
            .status(200)
            .json({
                "status": "OK",
                "info": result
            })
    });
}

function cache(req, res, next) {
    configureService.cache(req.webitelUser, {},  (err, result) => {
        if (err)
            return next(err);

        return res
            .status(200)
            .json({
                "status": "OK",
                "info": result
            })
    });
}