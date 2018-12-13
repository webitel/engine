/**
 * Created by Igor Navrotskyj on 01.09.2015.
 */

'use strict';

var log = require(__appRoot + '/lib/log')(module),
    CodeError = require(__appRoot + '/lib/error');

module.exports = function (caller, resource, action, cb) {
    try {
        if (!caller || !caller.roleName) {
            log.error('Bad caller.');
            return cb(new CodeError(403, 'Permission denied!'), false);
        }
        application.acl.areAnyRolesAllowed(caller.roleName, resource, action, function (err, res) {
            if (err) {
                return cb(new CodeError(500, err.message));
            }

            if (!res) {
                return cb(new CodeError(403, 'Permission denied!'))
            }

            cb(null, res);
        });
    } catch (e) {
        log.error(e);
        cb(e, false);
    }
};

module.exports.ROOT = {
    roleName: "root"
};