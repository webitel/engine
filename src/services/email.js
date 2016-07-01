/**
 * Created by Igor Navrotskyj on 26.08.2015.
 */

'use strict';

var log = require(__appRoot + '/lib/log')(module),
    nodemailer = require('nodemailer'),
    smtpPool = require('nodemailer-smtp-pool'),
    conf = require(__appRoot + '/conf'),
    CodeError = require(__appRoot + '/lib/error');

var Provider = {
    "smtp": smtpPool
};

var emailService = {
    get: function (domain, cb) {
        if (!domain) {
            return cb(new CodeError(400, 'Domain is required.'));
        };
        var dbEmail = application.DB._query.email;
        dbEmail.get({
            domain: domain
        }, cb);
    },

    set: function (settings, cb) {
        if (!settings || !settings['provider'] || !settings['options'] || !settings['domain']) {
            return cb(new CodeError(400, 'Bad parameters.'));
        };

        var dbEmail = application.DB._query.email;
        dbEmail.set(settings, cb);
    },

    update: function (settings, cb) {
        if (!settings || !settings['provider'] || !settings['options'] || !settings['domain']) {
            return cb(new CodeError(400, 'Bad parameters.'));
        };
        var dbEmail = application.DB._query.email;
        dbEmail.update(settings, cb);
    },

    remove: function (domain, cb) {
        if (!domain) {
            return cb(new CodeError(400, 'Bad parameters.'));
        };
        var dbEmail = application.DB._query.email;
        dbEmail.remove({
            domain: domain
        }, cb);
    },

    send: function (mailOption, domain, cb) {
        try {
            emailService.get(domain, function (err, res) {
                if (err) {
                    return cb(err);
                };

                if (!res || !res['options']) {
                    return cb(new CodeError(404, "Not settings EMail provider from domain " + domain));
                };

                if (typeof Provider[res.provider] != 'function') {
                    return cb(new CodeError(500, 'Bad provider name.'));
                };
                mailOption['from'] = mailOption['from'] || res['from'];
                try {
                    var transporter = nodemailer.createTransport(Provider[res.provider](res['options']));
                    transporter.sendMail(
                        mailOption,
                        cb
                    );
                } catch (e) {
                    return cb(e);
                };
            });
            return 1;
        } catch (e) {
            return cb(e);
        }
    },

    _removeByDomain: function (domain, cb) {
        if (!domain) {
            return cb(new CodeError(400, 'Domain is required.'));
        };
        var dbEmail = application.DB._query.email;
        dbEmail.removeByDomain(domain, cb);
    },
    
    _report: function (err, cb) {
        let option = conf.get('application:mailReport');
        if (!option || !option['smtp']) {
            return cb(new Error('Not config.'));
        };

        try {
            let smtpOption = option['smtp'],
                mail = {
                    "html": '<h1>WTF!!!  (O_o) Oops!  Server crashed.</h1>' +
                    '<h2>Server: ' + conf.get('server:host') + '</h2>' +
                    '<h2 style="color: red">Message: ' + err.message + '</h2>' +
                    '<h2 style="color: red">Stack: </h2>' +
                    '<h3 style="color: #333333">' + err.stack + '</h3>',
                    "to": option['to'],
                    "subject": "Webitel crashed :(",
                    "from": option['from']
                }
                ;

            let transporter = nodemailer.createTransport(smtpPool(smtpOption));
            transporter.sendMail(
                mail,
                cb
            )
        } catch (e) {
            return cb(e);
        }
    },

    _send: function (mail, smtpOption, cb) {
        if (!mail || !smtpOption) {
            return cb(new Error("mail or smtpOption undefined."));
        };
        try {
            let transporter = nodemailer.createTransport(smtpPool(smtpOption));
            transporter.sendMail(
                mail,
                cb
            )
        } catch (e) {
            log.error(e);
            return cb(e);
        };
    }
};

module.exports = emailService;