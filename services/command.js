/**
 * Created by Igor Navrotskyj on 07.08.2015.
 */

'use strict';

var Service = {
    bgApi: function (execString, cb) {
        application.Esl.bgapi(
            execString,
            function (res) {
                return cb(null, res);
            }
        );
    },
    broadcast: function (options, cb) {
        Service.bgApi(
            ('uuid_broadcast ' + options['uuid'] + ' ' + options['application']),
            cb
        );
    }
};

module.exports = Service;