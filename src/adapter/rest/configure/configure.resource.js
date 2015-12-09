/**
 * Created by i.navrotskyj on 09.12.2015.
 */
'use strict';

var configureService = require(__appRoot + '/services/configure');

module.exports = {
    addRoutes: addRoutes
};

function addRoutes (api) {
    api.get('/api/v2/reloadxml', reloadXml);

    //  V1
    api.get('/api/v1/reloadxml', reloadXml);
};

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
};