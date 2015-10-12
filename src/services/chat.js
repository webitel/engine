/**
 * Created by Igor Navrotskyj on 07.08.2015.
 */

'use strict';

var CodeError = require(__appRoot + '/lib/error'),
    log = require(__appRoot + '/lib/log')(module)
    ;

var Service = {
    bgApi: function (execString, cb) {
        log.debug(execString);
        application.Esl.bgapi(
            execString,
            function (res) {
                return cb(null, res);
            }
        );
    },
    
    send: function (caller, options, cb) {
        try {
            var profile = options['profile'] || 'verto';
            var from = options['from'];
            var to = options['to'];
            var message = options['message'];

            if (!from || !to || !message) {
                return cb(new Error("-ERR Bad request"));
            }
            ;

            var data = [].concat(profile, from, to, message).join('|');
            Service.bgApi(
                'chat ' + data,
                cb
            );
        } catch (e) {
            cb(e);
        };
    },
    
    sendInternalWSMessage: function (caller, option, cb) {
        try {
            // TODO
            if (!caller) {
                return cb(new CodeError(401, "Bad caller."));
            };
            if (!option || !option['to'] || !option['body']) {
                return cb(new CodeError(400, 'Bad request'));
            };
            var toId = option['to'].toString();

            if (caller['domain']) {
                toId = toId.replace(/@.*/, '') + '@' + caller['domain'];
            };

            var _to = application.Users.get(toId);
            if (!_to) {
                return cb(new CodeError(404, "User not found"));
            };

            var msg = {
                to: _to['id'],
                from: caller['id'],
                body: option['body'],
                'webitel-event-name': 'WEBITEL-CUSTOM',
                'Event-Name': 'WEBITEL-CUSTOM-MESSAGE'
            };

            _to.sendObject(msg);
            return cb(null, 'Message send.');

        } catch (e) {
            log.error(e);
        }
    }
};

module.exports = Service;