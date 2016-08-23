/**
 * Created by Admin on 03.08.2015.
 */

'use strict';

const os = require('os'),
    statsService = require(__appRoot + '/services/stats'),
    conf = require(__appRoot + '/conf');

module.exports = {
    addRoutes: addRoutes
};

/**
 * Adds routes to the api.
 */
function addRoutes(api) {
    api.get('/api/v2/status', applicationStatus);
    api.get('/api/v2/status/config', applicationStatusConfig);
    api.get('/api/v1/status', applicationStatus);
}

function applicationStatusConfig (req, res) {
    // TODO add config
    res.status(200).end();
}

function applicationStatus(req, res) {
    statsService.allStats((err, stats) => {
        res.json(stats)
    })
}