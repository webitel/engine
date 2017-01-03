/**
 * Created by igor on 30.12.16.
 */

"use strict";

var vmailService = require(__appRoot + '/services/vmail');

module.exports = {
    addRoutes: addRoutes
};

/**
 * Adds routes to the api.
 */
function addRoutes(api) {
    api.get('/api/v2/vmail/:id', list);
    api.put('/api/v2/vmail/:id/:uuid', update);
    api.delete('/api/v2/vmail/:id/:uuid', remove);
}

function list(req, res, next) {
    const option = {
        id: req.params.id,
        domain: req.query.domain
    };

    vmailService.list(req.webitelUser, option, (err, result) => {
        if (err)
            return next(err);

        return res
            .status(200)
            .json({
                status: "OK",
                data: result
            });
    })
}

function update(req, res, next) {
    const option = {
        id: req.params.id,
        domain: req.query.domain,
        uuid: req.params.uuid,
        state: req.body.state
    };

    vmailService.setState(req.webitelUser, option, (err, result) => {
        if (err)
            return next(err);

        return res
            .status(200)
            .json({
                status: "OK",
                data: result
            });
    })
}

function remove(req, res, next) {
    const option = {
        id: req.params.id,
        domain: req.query.domain,
        uuid: req.params.uuid
    };

    vmailService.remove(req.webitelUser, option, (err, result) => {
        if (err)
            return next(err);

        return res
            .status(200)
            .json({
                status: "OK",
                data: result
            });
    })
}