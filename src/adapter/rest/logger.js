/**
 * Created by Igor Navrotskyj on 25.08.2015.
 */

'use strict';

var log = require(__appRoot + '/lib/log')(module),
    getIp = require(__appRoot + '/utils/ip');

module.exports = {
    addRoutes: addRoutes
};

function addRoutes(api) {
    api.use(function(req, res, next) {
        log.trace('Method: %s, url: %s, path: %s, ip:', req.method, req.url, req.path, getIp(req));
        next();
    });
};