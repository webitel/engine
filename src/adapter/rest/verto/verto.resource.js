/**
 * Created by i.navrotskyj on 07.10.2015.
 */
'use strict';
var conf = require(__appRoot + '/conf'),
    FS_HOST = conf.get('freeSWITCH:host'),
    conferenceService = require(__appRoot + '/services/conference')
    ;

module.exports = {
    addRoutes: addRoutes
};

/**
 * Adds routes to the api.
 */
function addRoutes(api) {
    api.post('/config.json', getConfig);
    api.get('/validate', validate);
};

function getConfig (req, res, next) {
    conferenceService.getConfig(req.body, function (err, result) {
        if (err)
            return next(err);

        return res
            .status(200)
            .json(result)
        ;
    });

    /*
    let dbConference = application.DB._query.conference;
    if (eMail && password) {
        dbConference.getByEMail(eMail, password, function (err, result) {
            if (err)
                return next(err);

            // User not found
            let config = result;
            if (!config) {
                config = {
                    "autologin": false,
                    "password": "",
                    "login": "",
                    "hostname": CONFERENCE_DOMAIN,
                    "wsURL": "wss://" + "pre.webitel.com" + ":8082"
                };
            } else {
                config['valid'] = true;
                config['hostname'] = CONFERENCE_DOMAIN;
                config['wsURL'] = "wss://" + "pre.webitel.com" + ":8082";
                config['autologin'] = true;
            }

            return res
                .status(200)
                .json(config)
                ;
        });
    } else {
        var _c = {
            "autologin": "false",
            "password": "",
            "login": "",
            "hostname": CONFERENCE_DOMAIN,
            "wsURL": "wss://" + "pre.webitel.com" + ":8082"
        };

        return res
            .status(200)
            .json(_c);
    };
    */
};

function validate (req, res, next) {
    let id = req.query['_id'];
    let join = req.query['join'];
    conferenceService.validate(id, function (err, result) {
        if (err)
            return next(err);
        let url = '/';
        if (join)
            url += '#?join=' + join;

        return res
            .redirect(url);
    });
};