/**
 * Created by Igor Navrotskyj on 26.08.2015.
 */

'use strict';

var emailService = require(__appRoot + '/services/email');

module.exports = {
    addRoutes: addRoutes
};

/**
 * Adds routes to the api.
 */
function addRoutes(api) {
    api.get('/api/v2/email/settings', getEmailSettings);
    api.post('/api/v2/email/settings', postEmailSettings);
    api.post('/api/v2/email/settings/test/:to?', sendTestEmail);
    api.put('/api/v2/email/settings', putEmailSettings);
    api.delete('/api/v2/email/settings', deleteEmailSettings);
};

function getEmailSettings(req, res, next) {
    // TODO permission
    var caller = req.webitelUser || {},
        domain = caller['domain'] || req.query['domain'];

    if (!domain) {
        return res.status(400).json({
            "status": "error",
            "info": "Bad request (domain required)."
        });
    };

    emailService.get(domain, function (err, result) {
        if (err) {
            return next(err);
        };

        if (!result) {
            return res.status(404).json({
                "status": "error",
                "info": "Not found!"
            });
        };

        return res.status(200).json(result);
    });
};

function postEmailSettings(req, res, next) {
    var caller = req.webitelUser || {};
    var domain = caller['domain'] || req.query['domain'];
    var settings = req.body;
    settings["domain"] = domain;
    
    emailService.set(settings, function (err, result) {
        if (err)
            return next(err);

        return res
            .status(200)
            .json(result);
    });
};

function putEmailSettings(req, res, next) {
    var caller = req.webitelUser || {};
    var domain = caller['domain'] || req.query['domain'];
    var settings = req.body;
    settings["domain"] = domain;

    emailService.update(settings, function (err, result) {
        if (err)
            return next(err);

        return res
            .status(200)
            .json(result);
    });
};

function deleteEmailSettings(req, res, next) {
    var caller = req.webitelUser || {};
    var domain = caller['domain'] || req.query['domain'];

    emailService.remove(domain, function (err, result) {
        if (err)
            return next(err);

        return res
            .status(200)
            .json({
                "status": "OK",
                "Info": "Removed " + result.result.n + " record."
            });
    });
};

function sendTestEmail (req, res, next) {
    var caller = req.webitelUser || {};
    var domain = caller['domain'] || req.query['domain'];

    emailService.send({
            "to": req.params['to'] || req.body['to'],
            "subject": "Webitel",
            "html": "<h1>Helo from <img href=\"webitel.com\" src=\"cid:logoID\"/> </h1>",
            "attachments": [{
                "filename": "logo768.png",
                "path": __appRoot + "/public/static/logo1024.png",
                "cid": "logoID"
            }]
        },
        domain,
        function (err, info) {
            if (err) {
                return next(err);
            };

            res.status(200).json({
                "status": "OK",
                "info": info
            });
        }
    );
};