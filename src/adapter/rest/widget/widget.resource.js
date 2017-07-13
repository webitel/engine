/**
 * Created by igor on 05.07.17.
 */

'use strict';

module.exports = {
    addRoutes: addRoutes
};

const widgetService = require(__appRoot + '/services/widget');
const getRequest = require(__appRoot + '/utils/helper').getRequest;

function addRoutes(api) {
    api.get('/api/v2/widget', list);
    api.get('/api/v2/widget/:id', get);
    api.post('/api/v2/widget', create);
    api.put('/api/v2/widget/:id', update);
    api.delete('/api/v2/widget/:id', del);
}

function list(req, res, next) {
    widgetService.list(req.webitelUser, getRequest(req), (err, result) => {
        if (err)
            return next(err);

        return res.status(200).json({
            "status": "OK",
            "data": result
        });
    })
}
function get(req, res, next) {
    let options = {
        id: req.params.id,
        domain: req.query.domain
    };

    widgetService.get(req.webitelUser, options, (err, result) => {
        if (err)
            return next(err);

        if (!result)
            return res.status(404).json({
                "status": "error",
                "info": `Not found ${options.id}`
            });

        return res.status(200).json({
            "status": "OK",
            "data": result
        });
    });
}
function create(req, res, next) {
    const options = req.body;

    if (req.query['domain'])
        options.domain = req.query['domain'];

    widgetService.create(req.webitelUser, options, (err, result) => {
        if (err) {
            return next(err);
        }

        return res
            .status(200)
            .json({
                "status": "OK",
                "data": result,
            });
    });
}
function update(req, res, next) {
    let options = {
        id: req.params.id,
        domain: req.query.domain,
        data: req.body
    };

    widgetService.update(req.webitelUser, options, (err, result) => {
        if (err)
            return next(err);

        return res.status(200).json({
            "status": "OK",
            "data": result
        });
    });
}
function del(req, res, next) {
    let options = {
        id: req.params.id,
        domain: req.query.domain
    };

    widgetService.remove(req.webitelUser, options, (err, result) => {
        if (err)
            return next(err);

        return res.status(200).json({
            "status": "OK",
            "data": result
        });
    });
}