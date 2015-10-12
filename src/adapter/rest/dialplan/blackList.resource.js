/**
 * Created by Igor Navrotskyj on 28.08.2015.
 */

'use strict';

var blacklistService = require(__appRoot + '/services/blacklist');

module.exports = {
    addRoutes: addRoutes
};

function addRoutes (api) {
    api.get('/api/v2/routes/blacklists', getNames);
    api.post('/api/v2/routes/blacklists/searches', searches);
    api.post('/api/v2/routes/blacklists/:name', post);
    //?domain=&page=&order=&orderValue=1&limit=40
    api.get('/api/v2/routes/blacklists/:name', getFromName);
    api.get('/api/v2/routes/blacklists/:name/:number', getNumberFromName);
    api.delete('/api/v2/routes/blacklists/:name', removeName);
    api.delete('/api/v2/routes/blacklists/:name/:number', removeNumber);
};

function getNames (req, res, next) {
    blacklistService.getNames(req.webitelUser, req.query['domain'],
        function (err, result) {
            if (err) {
                return next(err);
            };

            return res
                .status(200)
                .json(result);
        }
    );
};

function searches (req, res, next) {
    blacklistService.search(req.webitelUser, req.query['domain'], req.body,
        function (err, result) {
            if (err) {
                return next(err);
            };

            return res
                .status(200)
                .json(result);
        }
    );
};

function post (req, res, next) {
    var option = req.body;
    option['name'] = req.params['name'];
    option['domain'] = req.query['domain'];

    blacklistService.create(req.webitelUser, option,
        function (err, result) {
            if (err) {
                return next(err);
            };

            return res
                .status(200)
                .json(result);
        }
    );
};

function getFromName (req, res, next) {
    blacklistService.getFromName(
        req.webitelUser,
        req.params['name'],
        req.query['domain'],
        req.query,
        function (err, result) {
            if (err) {
                return next(err);
            };

            return res
                .status(200)
                .json(result);
        }
    );
};

function getNumberFromName (req, res, next) {
    var option = {
        "name": req.params['name'],
        "number": req.params['number'],
        "domain": req.query['domain']
    };

    blacklistService.getNumberFromName(req.webitelUser, option,
        function (err, result) {
            if (err) {
                return next(err);
            };

            return res
                .status(200)
                .json(result);
        }
    );
};

function removeName (req, res, next) {
    var option = {
        "name": req.params['name']
    };
    blacklistService.remove(req.webitelUser, req.query['domain'], option,
        function (err, result) {
            if (err) {
                return next(err);
            };

            return res
                .status(200)
                .json(result);
        }
    );
};

function removeNumber (req, res, next) {
    var option = {
        "name": req.params['name'],
        "number": req.params['number']
    };
    blacklistService.remove(req.webitelUser, req.query['domain'], option,
        function (err, result) {
            if (err) {
                return next(err);
            };

            return res
                .status(200)
                .json(result);
        }
    );
};