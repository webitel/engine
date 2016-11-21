/**
 * Created by Admin on 04.08.2015.
 */
'use strict';

var log = require(__appRoot + '/lib/log')(module);

module.exports = {
    addRoutes: addRoutes
};

function addRoutes(api) {
    api.use(function(req, res, next){
        res.status(404);
        res.json({
            "status": "error",
            "info": req.originalUrl + ' not found.'
        });
        return;
    });

    api.use(function(err, req, res, next){
        res.status(err.status || 500);
        return res.json({
            "status": "error",
            "info": err.message,
            "code": err.code
        });
    });
};