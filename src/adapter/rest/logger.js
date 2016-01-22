/**
 * Created by Igor Navrotskyj on 25.08.2015.
 */

'use strict';

var log = require(__appRoot + '/lib/log')(module);

module.exports = {
    addRoutes: addRoutes
};

function addRoutes(api) {
    api.use(function(req, res, next) {
        log.trace('Method: %s, url: %s, path: %s, ip:', req.method, req.url, req.path, req.ip, req.ips);
        next();
    });
};