
const hotdeskService = require(__appRoot + '/services/hotdesk');

module.exports = {
    addRoutes: addRoutes
};


/**
 * Adds routes to the api.
 */
function addRoutes(api) {
    api.get('/api/v2/hotdesk', list);
    api.get('/api/v2/hotdesk/:id', item);

    api.post('/api/v2/hotdesk', create);
    api.put('/api/v2/hotdesk/:id', update);
    api.delete('/api/v2/hotdesk/:id', remove);
}

function list(req, res, next) {
    hotdeskService.list(req.webitelUser, req.query, (err, result) => {
        if (err)
            return next(err);

        return res.status(200).json({
            "status": "OK",
            "data": result
        });
    });
}

function item(req, res, next) {
    const options = {
        domain: req.query.domain,
        id: req.params.id
    };

    hotdeskService.item(req.webitelUser, options, (err, result) => {
        if (err)
            return next(err);

        return res.status(200).json({
            "status": "OK",
            "data": result
        });
    });
}

function create(req, res, next) {
    const options = {
        domain: req.query.domain,
        params: req.body,
        id: req.body.id
    };
    delete req.body.id;

    hotdeskService.create(req.webitelUser, options, (err, result) => {
        if (err)
            return next(err);

        return res.status(200).json({
            "status": "OK",
            "data": result
        });
    });
}

function update(req, res, next) {
    const options = {
        domain: req.query.domain,
        id: req.params.id,
        params: req.body
    };

    hotdeskService.update(req.webitelUser, options, (err, result) => {
        if (err)
            return next(err);

        return res.status(200).json({
            "status": "OK",
            "data": result
        });
    });
}

function remove(req, res, next) {
    const options = {
        domain: req.query.domain,
        id: req.params.id
    };

    hotdeskService.remove(req.webitelUser, options, (err, result) => {
        if (err)
            return next(err);

        return res.status(200).json({
            "status": "OK",
            "data": result
        });
    });
}