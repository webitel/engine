/**
 * Created by Igor Navrotskyj on 04.09.2015.
 */

'use strict';

var bookService = require(__appRoot + '/services/contactBook');

module.exports = {
    addRoutes: addRoutes
};

/**
 * Adds routes to the api.
 */
function addRoutes(api) {
    api.post('/api/v2/contacts/', createBook);
    // ?name=&phone=&tag=&limit=
    api.get('/api/v2/contacts', listBook);
    api.get('/api/v2/contacts/:id', itemBook);
    api.post('/api/v2/contacts/searches', searches);
    api.put('/api/v2/contacts/:id', updateItem);
    api.delete('/api/v2/contacts/:id', deleteItem);
};

function listBook (req, res, next) {
    var options = {
        "name": req.query['name'],
        "phone": req.query['phone'],
        "tag": req.query['tag'],
        "limit": req.query['limit']
    };
    bookService.list(req.webitelUser, req.query['domain'], options, function (err, result) {
        if (err) {
            return next(err);
        };

        var _r = {
            "status": "OK",
            "info": result['_id'],
            "data": result
        };

        return res
            .status(200)
            .json(_r);
    });
};

function itemBook (req, res, next) {
    bookService.getById(req.webitelUser, req.query['domain'], req.params['id'], function (err, result) {
        if (err) {
            return next(err);
        };

        var _r = {
            "status": "OK",
            "data": result
        };

        return res
            .status(200)
            .json(_r);
    });
};

function createBook (req, res, next) {

    bookService.create(req.webitelUser, req.query['domain'] || req.body['domain'], req.body, function (err, result) {
        if (err) {
            return next(err);
        };

        var _r = {
            "status": "OK",
            "info": result['_id'],
            "data": result
        };

        return res
            .status(200)
            .json(_r);
    });
};

function searches (req, res, next) {
    bookService.search(req.webitelUser, req.query['domain'], req.body,
        function (err, result) {
            if (err) {
                return next(err);
            };

            return res
                .status(200)
                .json({
                    "status": "OK",
                    "data": result
                });
        }
    );
};

function updateItem (req, res, next) {
    bookService.updateItem(req.webitelUser, req.query['domain'], req.params['id'], req.body,
        function (err, result) {
            if (err) {
                return next(err);
            };

            return res
                .status(200)
                .json({
                    "status": "OK",
                    "data": result
                });
        }
    );
};

function deleteItem (req, res, next) {
    bookService.removeItem(req.webitelUser, req.query['domain'], req.params['id'],
        function (err, result) {
            if (err) {
                return next(err);
            };

            if (!result) {
                next(new Error("Bad response db"));
            };

            return res
                .status(200)
                .json({
                    "status": result.ok == 1 ? "OK" : "error",
                    "info": result.n
                });
        }
    );
};