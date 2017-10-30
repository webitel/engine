/**
 * Created by I. Navrotskyj on 30.10.17.
 */

"use strict";

const log = require(__appRoot + '/lib/log')(module),
    checkPermissions = require(__appRoot + '/middleware/checkPermissions'),
    validateCallerParameters = require(__appRoot + '/utils/validateCallerParameters'),
    CodeError = require(__appRoot + '/lib/error');

const Service = {
    item: (caller, options = {}, cb) => {
        checkPermissions(caller, 'metadata', 'r', function (err) {
            if (err)
                return cb(err);


            let domain = validateCallerParameters(caller, options.domain);
            if (!domain) {
                return cb(new CodeError(400, 'Bad request: domain is required.'));
            }

            application.PG.getQuery('metadata').item(domain, options.object_name, cb);
        });
    },

    createOrReplace: (caller, options = {}, cb) => {
        checkPermissions(caller, 'metadata', 'u', function (err) {
            if (err)
                return cb(err);


            let domain = validateCallerParameters(caller, options.domain);
            if (!domain) {
                return cb(new CodeError(400, 'Bad request: domain is required.'));
            }

            application.PG.getQuery('metadata').createOrReplace(domain, options.object_name, options.data, cb);
        });
    },

    remove: (caller, options = {}, cb) => {
        checkPermissions(caller, 'metadata', 'd', function (err) {
            if (err)
                return cb(err);


            let domain = validateCallerParameters(caller, options.domain);
            if (!domain) {
                return cb(new CodeError(400, 'Bad request: domain is required.'));
            }

            application.PG.getQuery('metadata').remove(domain, options.object_name, cb);
        });
    }
};

module.exports = Service;