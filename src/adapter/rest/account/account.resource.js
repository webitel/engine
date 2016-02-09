/**
 * Created by Admin on 03.08.2015.
 */
'use strict';

var accountService = require(__appRoot + '/services/account');

module.exports = {
    addRoutes: addRoutes
};

/**
 * Adds routes to the api.
 */
function addRoutes(api) {
    api.post('/api/v2/accounts', create);
    api.get('/api/v2/accounts/:name', item);
    api.get('/api/v2/accounts?:domain', list);
    api.put('/api/v2/accounts/:name', update);
    api.delete('/api/v2/accounts/:name', remove);

    // V1
    api.post('/api/v1/accounts?', createV1);
    api.delete('/api/v1/accounts?/:name', removeV1);
};

function createV1 (req, res, next) {
    let option = req.body || {};
    if (req.query['domain']) {
        option['domain'] = req.query['domain'];
    };

    accountService.create(req.webitelUser, option,
        function (err, result) {
            if (err) {
                return res
                    .status(200)
                    .send('-ERR: ' + err.message)
                    ;
            };

            return res
                .status(200)
                .send(result)
                ;
        }
    );
};

function removeV1 (req, res, next) {
    let option = {
        "name": req.params['name'],
        "domain": req.query['domain']
    };

    accountService.remove(req.webitelUser, option,
        function (err, result) {
            if (err) {
                return res
                    .status(200)
                    .send(err.message);
            };

            return res
                .status(200)
                .send(result);
        }
    );
};

function create (req, res, next) {
    let option = req.body || {};
    if (req.query['domain']) {
        option['domain'] = req.query['domain'];
    };

    accountService.create(req.webitelUser, option,
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

function item (req, res, next) {
    let option = {
        "name": req.params['name'],
        "domain": req.query['domain']
    };

    accountService.item(req.webitelUser, option,
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

function list (req, res, next) {
    let option = {
        "domain": req.query['domain']
    };

    accountService.accountList(req.webitelUser, option,
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

function remove (req, res, next) {
    let option = {
        "name": req.params['name'],
        "domain": req.query['domain']
    };

    accountService.remove(req.webitelUser, option,
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

function update (req, res, next) {
    let option = req.body;
    accountService.update(req.webitelUser, req.params['name'], req.query['domain'], option,
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