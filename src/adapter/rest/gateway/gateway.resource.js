/**
 * Created by Igor Navrotskyj on 24.09.2015.
 */

'use strict';

var gatewayService = require(__appRoot + '/services/gateway');

module.exports = {
    addRoutes: addRoutes
};

/**
 * Adds routes to the api.
 */
function addRoutes(api) {
    api.post('/api/v2/gateway', createGateway);
    api.get('/api/v2/gateway', listGateway);
    api.get('/api/v2/gateway/:name', itemGateway);
    api.patch('/api/v2/gateway/:name/up', upGateway);
    api.put('/api/v2/gateway/:name/up', upGateway);
    api.patch('/api/v2/gateway/:name/down', downGateway);
    api.put('/api/v2/gateway/:name/down', downGateway);

    api.put('/api/v2/gateway/:name/:type', changeGateway);
    api.get('/api/v2/gateway/:name/:type', varGateway);
    api.delete('/api/v2/gateway/:name', deleteGateway);
};

function varGateway (req, res, next) {
    var option = {
        name: req.params['name'],
        direction: req.params['type']
    };
    gatewayService.varGateway(req.webitelUser, option,
        function (err, result) {
            if (err) {
                return next(err);
            };

            return res
                .status(200)
                .json({
                    "status": "OK",
                    "info": result
                });
        }
    );
};

function createGateway (req, res, next) {
    var option = req.body;
    option['domain'] = option['domain'] || req.query['domain'];

    gatewayService.createGateway(req.webitelUser, option,
        function (err, result) {
            if (err) {
                return next(err);
            };

            return res
                .status(200)
                .json({
                    "status": "OK",
                    "info": result
                });
        }
    );
};

function listGateway (req, res, next) {
    gatewayService.listGateway(req.webitelUser, req.query['domain'],
        function (err, result) {
            if (err) {
                return next(err);
            };

            return res
                .status(200)
                .json({
                    "status": "OK",
                    "info": result
                });
        }
    );
};

function itemGateway (req, res, next) {
    gatewayService.itemGateway(req.webitelUser, req.params['name'],
        function (err, result) {
            if (err) {
                return next(err);
            };

            return res
                .status(200)
                .json({
                    "status": "OK",
                    "info": result
                });
        }
    );
};

function changeGateway (req, res, next) {
    gatewayService.changeGateway(req.webitelUser, req.params['name'], req.params['type'], req.body,
        function (err, result) {
            if (err) {
                return next(err);
            };

            return res
                .status(200)
                .json({
                    "status": "OK",
                    "info": result
                });
        }
    );
};

function deleteGateway (req, res, next) {
    gatewayService.deleteGateway(req.webitelUser, req.params['name'],
        function (err, result) {
            if (err) {
                return next(err);
            };

            return res
                .status(200)
                .json({
                    "status": "OK",
                    "info": result
                });
        }
    );
};

function upGateway (req, res, next) {
    gatewayService.upGateway(req.webitelUser, req.params['name'], req.query['profile'],
        function (err, result) {
            if (err) {
                return next(err);
            };

            return res
                .status(200)
                .json({
                    "status": "OK",
                    "info": result
                });
        }
    );
};

function downGateway (req, res, next) {
    gatewayService.downGateway(req.webitelUser, req.params['name'],
        function (err, result) {
            if (err) {
                return next(err);
            };

            return res
                .status(200)
                .json({
                    "status": "OK",
                    "info": result
                });
        }
    );
};