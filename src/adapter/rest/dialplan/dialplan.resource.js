/**
 * Created by Igor Navrotskyj on 27.08.2015.
 */

'use strict';

const dialplanService = require(__appRoot + '/services/dialplan');

module.exports = {
    addRoutes: addRoutes
};

/**
 * Adds routes to the api.
 */
function addRoutes(api) {
    api.get('/api/v2/routes/extensions', getExtensions);
    // TODO get by ID !!
    // api.get
    api.put('/api/v2/routes/extensions/:id', putExtension);

    api.get('/api/v2/routes/variables', getDomainVariables);
    api.post('/api/v2/routes/variables', insertOrUpdateDomainVariable);
    api.put('/api/v2/routes/variables', insertOrUpdateDomainVariable);


    api.post('/api/v2/routes/public', postPublic);
    api.post('/api/v2/routes/public/:id/debug', debugPublic);
    api.get('/api/v2/routes/public', getPublic);
    api.delete('/api/v2/routes/public/:id', deletePublic);
    api.put('/api/v2/routes/public/:id', putPublic);

    api.post('/api/v2/routes/default', postDefault);
    api.post('/api/v2/routes/default/:id/debug', debugDefault);
    api.get('/api/v2/routes/default', getDefault);
    api.delete('/api/v2/routes/default/:id', deleteDefault);
    api.put('/api/v2/routes/default/:id', putDefault);
    api.put('/api/v2/routes/default/:id/setOrder', setOrderDefault);
    api.put('/api/v2/routes/default/:domainName/incOrder', incOrderDefault);
}

function getExtensions (req, res, next) {
    dialplanService.getExtensions(req.webitelUser, req.query['domain'], function (err, result) {
        if (err) {
            return next(err);
        };

        return res
            .status(200)
            .json(result);
    });
}

function putExtension (req, res, next) {
    var option = req.body || {};
    dialplanService.updateExtension(req.webitelUser, req.params['id'], option,
        function (err, result) {
            if (err) {
                return next(err);
            };

            return res
                .status(200)
                .json(result);
        }
    );
}

function postPublic (req, res, next) {
    dialplanService.createPublic(req.webitelUser, req.query['domain'] || req.body['domain'], req.body, function (err, result) {
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
}

function getPublic (req, res, next) {
    dialplanService.getPublic(req.webitelUser, req.query['domain'], function (err, result) {
        if (err) {
            return next(err);
        };

        return res
            .status(200)
            .json(result);
    });
}

function deletePublic (req, res, next) {
    dialplanService.removePublic(req.webitelUser, req.params['id'], function (err, result) {
        if (err) {
            return next(err);
        };

        return res
            .status(200)
            .json(result);
    });
}

function putPublic (req, res, next) {
    var option = req.body || {};
    dialplanService.updatePublic(req.webitelUser, req.params['id'], option,
        function (err, result) {
            if (err) {
                return next(err);
            };

            return res
                .status(200)
                .json(result);
        }
    );
}

function postDefault (req, res, next) {
    dialplanService.createDefault(req.webitelUser, req.query['domain'] || req.body['domain'], req.body, function (err, result) {
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
}

function getDefault (req, res, next) {
    dialplanService.getDefault(req.webitelUser, req.query['domain'], function (err, result) {
        if (err) {
            return next(err);
        };

        return res
            .status(200)
            .json(result);
    });
}

function deleteDefault (req, res, next) {
    dialplanService.removeDefault(req.webitelUser, req.params['id'], function (err, result) {
        if (err) {
            return next(err);
        };

        return res
            .status(200)
            .json(result);
    });
}

function putDefault (req, res, next) {
    var option = req.body || {};
    dialplanService.updateDefault(req.webitelUser, req.params['id'], option,
        function (err, result) {
            if (err) {
                return next(err);
            };

            return res
                .status(200)
                .json(result);
        }
    );
}

function setOrderDefault (req, res, next) {
    var option = req.body || {};
    dialplanService.setOrderDefault(req.webitelUser, req.params['id'], option,
        function (err, result) {
            if (err) {
                return next(err);
            };

            return res
                .status(200)
                .json(result);
        }
    );
}

function incOrderDefault (req, res, next) {
    var option = {
        inc: parseInt(req.body['inc']),
        start: parseInt(req.body['start'])
    };
    dialplanService.incOrderDefault(req.webitelUser, req.params['domainName'], option,
        function (err, result) {
            if (err) {
                return next(err);
            };

            return res
                .status(200)
                .json(result);
        }
    );
}

function getDomainVariables (req, res, next) {
    dialplanService.getDomainVariable(req.webitelUser, req.query['domain'], function (err, result) {
        if (err) {
            return next(err);
        };

        return res
            .status(200)
            .json(result);
    });
}

function insertOrUpdateDomainVariable (req, res, next) {
    dialplanService.insertOrUpdateDomainVariable(req.webitelUser, req.query['domain'], req.body, function (err, result) {
        if (err) {
            return next(err);
        }

        return res
            .status(200)
            .json(result);
    });
}

function debugPublic(req, res, next) {
    const options = {
        domain: req.query.domain,
        number: req.body.number,
        uuid: req.body.uuid,
        from: req.body.from
    };

    dialplanService.debugPublic(req.webitelUser, options, (err, result) => {
        if (err) {
            return next(err);
        }

        return res
            .status(200)
            .json({
                "status": "OK"
            });
    })
}

function debugDefault(req, res, next) {
    const options = {
        domain: req.query.domain,
        number: req.body.number,
        uuid: req.body.uuid,
        from: req.body.from
    };

    dialplanService.debugDefault(req.webitelUser, options, (err, result) => {
        if (err) {
            return next(err);
        }

        return res
            .status(200)
            .json({
                "status": "OK"
            });
    })
}