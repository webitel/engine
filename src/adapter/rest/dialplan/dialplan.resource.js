/**
 * Created by Igor Navrotskyj on 27.08.2015.
 */

'use strict';

const dialplanService = require(__appRoot + '/services/dialplan');
const getRequest = require(__appRoot + '/utils/helper').getRequest;

module.exports = {
    addRoutes: addRoutes
};

/**
 * Adds routes to the api.
 */
function addRoutes(api) {
    api.get('/api/v2/routes/extensions', listExtension);
    api.get('/api/v2/routes/extensions/:id', itemExtension);
    api.put('/api/v2/routes/extensions/:id', updateExtension);
    api.delete('/api/v2/routes/extensions/:id', removeExtension);

    api.get('/api/v2/routes/variables', listDomainVariables);
    api.post('/api/v2/routes/variables', insertOrUpdateDomainVariable);
    api.put('/api/v2/routes/variables', insertOrUpdateDomainVariable);


    api.post('/api/v2/routes/public', postPublic);
    api.post('/api/v2/routes/public/:id/debug', debugPublic);
    api.get('/api/v2/routes/public', listPublic);
    api.get('/api/v2/routes/public/:id', itemPublic);
    api.delete('/api/v2/routes/public/:id', deletePublic);
    api.put('/api/v2/routes/public/:id', updatePublic);

    api.post('/api/v2/routes/default', postDefault);
    api.post('/api/v2/routes/default/:id/debug', debugDefault);
    api.put('/api/v2/routes/default/:id/up', moveUpDefault);
    api.put('/api/v2/routes/default/:id/down', moveDownDefault);
    api.get('/api/v2/routes/default', listDefault);
    api.get('/api/v2/routes/default/:id', itemDefault);
    api.delete('/api/v2/routes/default/:id', deleteDefault);
    api.put('/api/v2/routes/default/:id', updateDefault);
    // todo deprecated
    // api.put('/api/v2/routes/default/:id/setOrder', setOrderDefault);
    // api.put('/api/v2/routes/default/:domainName/incOrder', incOrderDefault);
}

function removeExtension(req, res, next) {
    const options = {
        id: req.params.id,
        domain: req.query.domain
    };

    dialplanService.removeExtension(req.webitelUser, options, function (err, result) {
        if (err)
            return next(err);

        return res.status(200).json({
            "status": "OK",
            "data": result
        });
    })
}

function listExtension (req, res, next) {
    dialplanService.listExtension(req.webitelUser, getRequest(req), function (err, result) {
        if (err)
            return next(err);

        return res.status(200).json({
            "status": "OK",
            "data": result
        });
    });
}

function itemExtension(req, res, next) {
    const options = {
        id: req.params.id,
        domain: req.query.domain
    };

    dialplanService.itemExtension(req.webitelUser, options, function (err, result) {
        if (err)
            return next(err);

        return res.status(200).json({
            "status": "OK",
            "data": result
        });
    });
}

function updateExtension (req, res, next) {
    const options = req.body || {};
    options.id = req.params.id;
    options.domain = req.query.domain;

    dialplanService.updateExtension(req.webitelUser, options,
        function (err, result) {
            if (err) {
                return next(err);
            }

            return res
                .status(200)
                .json({
                    "status": "OK",
                    "data": result
                });
        }
    );
}

function postPublic (req, res, next) {
    dialplanService.createPublic(req.webitelUser, req.query['domain'] || req.body['domain'], req.body, function (err, result) {
        if (err) {
            return next(err);
        }

        return res
            .status(200)
            .json({
                "status": "OK",
                "info": result['id'],
                "data": result
            });
    });
}


function deletePublic (req, res, next) {
    const options = {
        id: req.params.id,
        domain: req.query.domain
    };

    dialplanService.removePublic(req.webitelUser, options, function (err, result) {
        if (err) {
            return next(err);
        }

        return res
            .status(200)
            .json(result);
    });
}

function updatePublic (req, res, next) {
    const options = req.body || {};
    options.id = req.params.id;
    options.domain = req.query.domain;

    dialplanService.updatePublic(req.webitelUser, options,
        function (err, result) {
            if (err) {
                return next(err);
            }

            return res
                .status(200)
                .json({
                    "status": "OK",
                    "data": result
                });
        }
    );
}

function postDefault (req, res, next) {
    dialplanService.createDefault(req.webitelUser, req.query['domain'] || req.body['domain'], req.body, function (err, result) {
        if (err) {
            return next(err);
        }

        return res
            .status(200)
            .json({
                "status": "OK",
                "info": result['id'],
                "data": result
            });
    });
}

function listPublic (req, res, next) {
    dialplanService.listPublic(req.webitelUser, getRequest(req), function (err, result) {
        if (err)
            return next(err);

        return res.status(200).json({
            "status": "OK",
            "data": result
        });
    });
}

function itemPublic (req, res, next) {
    const options = {
        id: req.params.id,
        domain: req.query.domain
    };

    dialplanService.itemPublic(req.webitelUser, options, function (err, result) {
        if (err)
            return next(err);

        return res.status(200).json({
            "status": "OK",
            "data": result
        });
    });
}

function listDefault (req, res, next) {
    dialplanService.listDefault(req.webitelUser, getRequest(req), function (err, result) {
        if (err)
            return next(err);

        return res.status(200).json({
            "status": "OK",
            "data": result
        });
    });
}

function itemDefault (req, res, next) {
    const options = {
        id: req.params.id,
        domain: req.query.domain
    };

    dialplanService.itemDefault(req.webitelUser, options, function (err, result) {
        if (err)
            return next(err);

        return res.status(200).json({
            "status": "OK",
            "data": result
        });
    });
}

function deleteDefault (req, res, next) {
    const options = {
        id: req.params.id,
        domain: req.query.domain
    };

    dialplanService.removeDefault(req.webitelUser, options, function (err, result) {
        if (err) {
            return next(err);
        }

        return res
            .status(200)
            .json(result);
    });
}

function updateDefault (req, res, next) {
    const options = req.body || {};
    options.id = req.params.id;
    options.domain = req.query.domain;

    dialplanService.updateDefault(req.webitelUser, options,
        function (err, result) {
            if (err) {
                return next(err);
            }

            return res
                .status(200)
                .json({
                    "status": "OK",
                    "data": result
                });
        }
    );
}

function listDomainVariables (req, res, next) {
    dialplanService.listDomainVariables(req.webitelUser, getRequest(req), function (err, result) {
        if (err) {
            return next(err);
        }

        return res
            .status(200)
            .json({
                "status": "OK",
                "data": result
            });
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

function moveUpDefault(req, res, next) {
    const options = {
        id: req.params.id,
        domain: req.query.domain,
        up: true
    };

    dialplanService.move(req.webitelUser, options, (err) => {
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

function moveDownDefault(req, res, next) {
    const options = {
        id: req.params.id,
        domain: req.query.domain,
        up: false
    };

    dialplanService.move(req.webitelUser, options, (err) => {
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