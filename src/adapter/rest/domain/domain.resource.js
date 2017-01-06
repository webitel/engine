/**
 * Created by Igor Navrotskyj on 27.09.2015.
 */

'use strict';
var domainService = require(__appRoot + '/services/domain');

module.exports = {
    addRoutes: addRoutes
};


/**
 * Adds routes to the api.
 */
function addRoutes(api) {
    api.post('/api/v2/domains', create);
    api.get('/api/v2/domains', list);
    api.get('/api/v2/domains/:name', item);
    api.get('/api/v2/domains/:name/settings', getSettings);
    api.put('/api/v2/domains/:name/settings', updateOrInsert);

    api.post('/api/v2/domains/:name/settings/token', genToken);
    api.delete('/api/v2/domains/:name/settings/token/:uuid', removeToken);
    api.patch('/api/v2/domains/:name/settings/token/:uuid', setStateToken);

    api.put('/api/v2/domains/:name/:type', update);
    api.delete('/api/v2/domains/:name', remove);

    // V1
    api.post('/api/v1/domains?', createV1);
    api.delete('/api/v1/domains?/:name', removeV1);
}

function createV1 (req, res, next) {
    var option = {
        "name": req.body['domain_name'],
        "customerId": req.body['customer_id'],
        "parameters": req.body['parameters'],
        "variables": req.body['variables']
    };

    domainService.create(req.webitelUser, option, function (err, result) {
        if (err) {
            return res
                .status(200)
                .send(err.message);
        };

        return res
            .status(200)
            .send(result)
            ;
    });
}

function removeV1 (req, res, next) {
    var option = {
        "name": req.params['name']
    };

    domainService.remove(req.webitelUser, option, function (err, result) {
        if (err) {
            return res
                .status(200)
                .send(err.message);
        };

        return res
            .status(200)
            .send(result);
    });
}

function create (req, res, next) {
    var option = {
        "name": req.body['domain_name'],
        "customerId": req.body['customer_id'],
        "parameters": req.body['parameters'],
        "variables": req.body['variables']
    };

    domainService.create(req.webitelUser, option, function (err, result) {
        if (err) {
            return next(err);
        };

        return res
            .status(200)
            .json({
                "status": "OK",
                "info": result
            });
    });
}

function list (req, res, next) {
    var option = {
        "customerId": req.query['customerId']
    };

    domainService.list(req.webitelUser, option, function (err, result) {
        if (err) {
            return next(err);
        };

        return res
            .status(200)
            .json({
                "status": "OK",
                "info": result
            });
    });
}

function item (req, res, next) {
    var option = {
        "name": req.params['name']
    };

    domainService.item(req.webitelUser, option, function (err, result) {
        if (err) {
            return next(err);
        };

        return res
            .status(200)
            .json({
                "status": "OK",
                "info": result
            });
    });
}

function update (req, res, next) {
    var option = {
        "name": req.params['name'],
        "type": req.params['type'],
        "params": req.body
    };

    domainService.update(req.webitelUser, option, function (err, result) {
        if (err) {
            return next(err);
        };

        return res
            .status(200)
            .json({
                "status": "OK",
                "info": result
            });
    });
}

function remove (req, res, next) {
    var option = {
        "name": req.params['name']
    };

    domainService.remove(req.webitelUser, option, function (err, result) {
        if (err) {
            return next(err);
        };

        return res
            .status(200)
            .json({
                "status": "OK",
                "info": result
            });
    });
}

function getSettings(req, res, next) {
    let option = {
        "name": req.params['name']
    };

    domainService.settings.get(req.webitelUser, option, function (err, result) {
        if (err)
            return next(err);

        if (!result) //code error 404
            return next();

        return res
            .status(200)
            .json({
                "status": "OK",
                "info": result
            });
    });
}

function updateOrInsert(req, res, next) {
    let option = req.body;
    option.name = req.params.name;

    domainService.settings.updateOrInsert(req.webitelUser, option, function (err, result) {
        if (err)
            return next(err);

        if (!result) //code error 404
            return next();

        return res
            .status(200)
            .json({
                "status": "OK",
                "info": result
            });
    });
}

function genToken (req, res, next) {
    const option = {
        "name": req.params['name'],
        "expire": req.body.expire,
        "role": req.body.role
    };

    domainService.settings.genToken(req.webitelUser, option, function (err, result) {
        if (err)
            return next(err);

        if (!result) //code error 404
            return next();

        return res
            .status(200)
            .json({
                "status": "OK",
                "info": result
            });
    });
}
function removeToken (req, res, next) {
    const option = {
        name: req.params.name,
        uuid: req.params.uuid
    };

    domainService.settings.removeToken(req.webitelUser, option, function (err, result) {
        if (err)
            return next(err);

        if (!result) //code error 404
            return next();

        return res
            .status(200)
            .json({
                "status": "OK",
                "info": result
            });
    });
}

function setStateToken(req, res, next) {
    const option = {
        name: req.params.name,
        uuid: req.params.uuid,
        state: req.body.state === true
    };

    domainService.settings.setStateToken(req.webitelUser, option, function (err, result) {
        if (err)
            return next(err);

        if (!result) //code error 404
            return next();

        return res
            .status(200)
            .json({
                "status": "OK",
                "info": result
            });
    });
}