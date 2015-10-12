/**
 * Created by i.navrotskyj on 08.10.2015.
 */
'use strict';

var CodeError = require(__appRoot + '/lib/error'),
    conf = require(__appRoot + '/conf'),
    accountService = require(__appRoot + '/services/account'),
    SmtpOption = conf.get('application:mailReport:smtp'),
    From = conf.get('application:mailReport:from'),
    log = require(__appRoot + '/lib/log')(module),
    FS_HOST = conf.get('freeSWITCH:host'),
    mailService = require('./email'),
    generateUuid = require('node-uuid'),
    fs = require('fs'),
    MAIL_TEMPLATE = fs.readFileSync(__appRoot + '/conf/conferenceMailTemplate.html', {encoding: "utf8"}),
    CONFERENCE_DOMAIN = conf.get('conference:domain'),
    VALIDATE_URI = conf.get('conference:validateUri'),
    VALIDATE_MAIL = conf.get('conference:validateMail'),
    ALIVE_USER_HOUR = conf.get('conference:aliveUserHour'),
    SCHEDULE_USER_HOUR = conf.get('conference:scheduleHour'),
    ENABLE_CONFERENCE = conf.get('conference:enable')
    ;


var Service = {
    /**
     *
     * @param option
     * @param cb
     * @returns {*}
     */
    getConfig: function (option, cb) {
        let eMail = option['email'],
            password = option['password'],
            join = option['join'],
            name = option['name']
            ;

        if (!eMail || !password) {
            return cb(new CodeError(200, "EMail and password is required."));
        };

        let dbConference = application.DB._query.conference;
        dbConference.getByEMail(eMail, function (err, result) {
            if (err)
                return cb(err);

            var config = result;
            if (config) {
                if (config['password'] === password && config['confirmed'] === true) {
                    return cb(null, Service._setConfigAttribute(config, true, join));
                };
                return cb(new CodeError(401, 'Invalid credentials.'));
            } else {
                Service.register(eMail, password, option, function (err, data) {
                    if (err)
                        return cb(err);

                    config = VALIDATE_MAIL
                        ? {"validateEmail": eMail}
                        : data
                    ;

                    cb(null, Service._setConfigAttribute(config, !VALIDATE_MAIL, join));
                    // TODO
                });
            };

        });
    },

    /**
     *
     * @param id
     * @param data
     * @param join
     * @returns {*}
     * @private
     */
    _getMailTemplate: function (id, data, join) {
        var _link  = VALIDATE_URI + 'validate?_id=' + id;
        if (join)
            _link += '&join=' + join
        ;

        let _t = MAIL_TEMPLATE
                .replace(/##LINK##/g, _link || '')
                .replace(/##ID##/g, id)
            ;
        for (let key in data) {
            if (data.hasOwnProperty(key) && !key.indexOf('_') == 0)
                _t = _t.replace(new RegExp('##' + key.toUpperCase() + '##', 'g'), data[key])
        };
        return _t;
    },

    /**
     *
     * @param config
     * @param valid
     * @param join
     * @returns {*}
     * @private
     */
    _setConfigAttribute: function (config, valid, join) {
        config['valid'] = valid || false;
        config['hostname'] = CONFERENCE_DOMAIN;
        config['wsURL'] = "wss://" + "pre.webitel.com" + ":8082";
        config['autologin'] = valid || false;
        config['autocall'] = join || config['meetingId'];
        return config;
    },

    /**
     *
     * @param email
     * @param password
     * @param userData
     * @param cb
     * @returns {*}
     */
    register: function (email, password, userData, cb) {
        if (!email || !password) {
            return cb(new CodeError(400, "EMail and password is required."));
        };

        let dbConference = application.DB._query.conference;
        dbConference.existsEmail(email, function (err, exists) {
            if (err)
                return cb(err);

            if (exists)
                return cb(new Error('EMail exists.'));

            var option = {
                "meetingId": generateUuid.v4().replace(/-/g, '').substring(12),
                "login": _generate(),
                "confirmed": !VALIDATE_MAIL,
                "name": userData['name'],
                "domain": CONFERENCE_DOMAIN
            };
            dbConference.insert(email, password, option, function (err, data) {
                if (err) {
                    log.error(err);
                    return cb(err);
                };
                var id = '';
                try {
                    id = data['_id'].toString();
                } catch (e) {
                    log.error(e);
                };

                if (VALIDATE_MAIL) {
                    var mail = {
                        "html": Service._getMailTemplate(id, data, userData['join']),
                        "to": email,
                        "subject": "[Webitel] Please verify your email '" + email + "'",
                        "from": From,
                        "attachments": [{
                            filename: 'logo1024.png',
                            path: __appRoot + '/public/static/logo1024.png',
                            cid: 'logo1024.png'
                        }]
                    };

                    mailService._send(mail, SmtpOption, function (err) {
                        if (err) {
                            log.error(err);
                            return cb(err);
                        };
                        log.trace('Send mail %s', email);
                        return cb(null, data);
                    });
                } else {
                    Service.createLogin(option['login'], password, function (err, res) {
                       if (err) {
                           // TODO - delete DB record.
                           log.error(err);
                           return cb(err)
                       };
                        return cb(null, data);
                    });
                };
            });
        });
    },

    /**
     *
     * @param login
     * @param password
     * @param cb
     * @returns {*|Suite|Object|number}
     */
    createLogin: function (login, password, cb) {
        var option = {
            "login": login,
            "password": password,
            "domain": CONFERENCE_DOMAIN,
            "role": 'user'
        };
        return accountService.create(Service.CALLER, option, cb);
    },

    /**
     *
     * @param id
     * @param cb
     */
    validate: function (id, cb) {
        let dbConference = application.DB._query.conference;
        dbConference.setConfirmed(id, true, function (err, data) {
            if (err) {
                log.error(err);
                return cb(err);
            };
            if (!data) {
                return cb(new Error("Bad query response"));
            };

            var login = data['login'];
            var password = data['password'];
            return Service.createLogin(login, password, cb);
        });
    },

    _runAutoDeleteUser: function (app) {
        if (ENABLE_CONFERENCE.toString() == 'true') {
            let _time = SCHEDULE_USER_HOUR * 60 * 60 * 1000;
            app.Schedule(_time, function () {
                log.debug('Schedule conference delete user.');
                let dbConference = application.DB._query.conference,
                    date = Date.now() - (ALIVE_USER_HOUR * 60 * 60 * 1000)
                ;

                dbConference._getDeleteAgents(date, function (err, result) {
                    if (err)
                        return log.error(err);
                    if (result && result.length > 0) {
                        result.forEach(function (item) {
                            var _o = {
                                "name": item['login'],
                                "domain": item['domain']
                            }
                            accountService.remove(Service.CALLER, _o, function(){});
                        });

                        dbConference._deleteExpiresAgents(date, function (err) {
                            if (err)
                                return log.error(err)
                            ;
                        });
                    };
                });
            });
        };
    },

    // TODO create systemRole
    CALLER: {
        "username": "root",
        "role": 2,
        "roleName": "root"
    }
};

module.exports = Service;

function _generate (count) {
    return new Date().getTime().toString();
};