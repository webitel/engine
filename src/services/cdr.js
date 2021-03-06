/**
 * Created by Igor Navrotskyj on 07.09.2015.
 */

'use strict';

var CodeError = require(__appRoot + '/lib/error'),
    authService = require('./auth'),
    conf = require(__appRoot + '/conf'),
    CDR_SERVER_URL = `${conf.get('cdrServer:useProxy')}` === 'true' ? conf.get('server:baseUrl') : conf.get("cdrServer:host"),
    CDR_GET_FILE_API = '/api/v2/files/',
    checkPermissions = require(__appRoot + '/middleware/checkPermissions')
    ;


if (CDR_SERVER_URL) {
    CDR_SERVER_URL = CDR_SERVER_URL.replace(/\/$/g, '')
}

var Service = {
    getRecordFile: function (caller, uuid, cb) {

        if (!uuid)
            return cb(new CodeError(403, "UUID is required."));

        let callback = (err) => {
            if (err)
                return cb(err);

            existsRecordFile(uuid, function (err, exists) {
                if (err)
                    return cb(err);

                if (exists) {
                    var diff = 24 * 60 * 60 * 1000; // + day
                    authService.getTokenMaxExpires(caller, diff, function (err, result) {
                        if (err)
                            return cb(err);

                        var url = CDR_SERVER_URL + CDR_GET_FILE_API + uuid + '?x_key=' + result['key'] +
                            '&access_token=' + result['token'];

                        return cb(null, {
                            "body": url
                        })
                    });
                } else {
                    return cb(null, {
                        "body": "+OK: Not found."
                    });
                }
            });
        };

        checkPermissions(caller, 'cdr', 'r', function (err) {
            if (err) {
                return checkPermissions(caller, 'cdr', 'ro', callback);
            }

            return callback(null);
        });
    }
};

function existsRecordFile(uuid, cb) {
    var dbCdr = application.DB._query.cdr;
    return dbCdr.existsRecordFile(uuid, cb);
}

module.exports = Service;