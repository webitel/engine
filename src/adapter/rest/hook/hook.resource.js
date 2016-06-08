/**
 * Created by i.navrotskyj on 31.03.2016.
 */
'use strict';

var hookService = require(__appRoot + '/services/hook')
    ;

module.exports = {
    addRoutes: addRoutes
};

function addRoutes (api) {
    api.get('/api/v2/hooks', list);
    api.post('/api/v2/hooks', create);
    api.get('/api/v2/hooks/:id', item);
    api.put('/api/v2/hooks/:id', update);
    api.delete('/api/v2/hooks/:id', remove);
    //api.post('/api/v2/hooks/searches', searches);
};

function list (req, res, next) {
    let options = {
        limit: req.query.limit,
        pageNumber: req.query.page,
        domain: req.query.domain,
        columns: {}
    };
    if (req.query.columns)
        req.query.columns.split(',')
            .forEach( (i) => options.columns[i] = 1 );

    hookService.list(req.webitelUser, options, (err, result) => {
        if (err)
            return next(err);

        return res.status(200).json({
            "status": "OK",
            "data": result
        });
    })
};

function item(req, res, next) {
    let options = {
        domain: req.query.domain,
        id: req.params.id
    };

    hookService.item(req.webitelUser, options, (err, result) => {
        if (err)
            return next(err);

        return res.status(200).json({
            "status": "OK",
            "data": result
        });
    })
};

function update(req, res, next) {
    let options = {
        doc: req.body,
        id: req.params.id,
        domain: req.query.domain
    };

    hookService.update(req.webitelUser, options, (err, result) => {
        if (err)
            return next(err);

        return res.status(200).json({
            "status": "OK",
            "data": result
        });
    })
};

function create(req, res, next) {
    let options = {
        doc: req.body,
        domain: req.query.domain
    };

    hookService.create(req.webitelUser, options, (err, result) => {
        if (err)
            return next(err);

        return res.status(200).json({
            "status": "OK",
            "data": result
        });
    });
};

function remove(req, res, next) {
    let options = {
        id: req.params.id,
        domain: req.query.domain
    };

    hookService.remove(req.webitelUser, options, (err, result) => {
        if (err)
            return next(err);

        return res.status(200).json({
            "status": "OK",
            "data": result
        });
    })
}

function searches(req, res, next) {

};