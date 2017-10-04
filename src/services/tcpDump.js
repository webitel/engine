/**
 * Created by I. Navrotskyj on 04.10.17.
 */

"use strict";

const log = require(__appRoot + '/lib/log')(module),
    checkPermissions = require(__appRoot + '/middleware/checkPermissions'),
    checkEslError = require(__appRoot + '/middleware/checkEslError'),
    CodeError = require(__appRoot + '/lib/error');

const Service = {
    list: (caller, option = {}, cb) => {
        if (!validateRoot(caller)) {
            return cb(new CodeError(403, 'Permission denied!'))
        }

        if (!option)
            return cb(new CodeError(400, "Bad request options"));

        application.PG.getQuery('tcpDump').list(option, cb);
    },

    create: (caller, option = {}, cb) => {
        if (!validateRoot(caller)) {
            return cb(new CodeError(403, 'Permission denied!'))
        }

        if (!option)
            return cb(new CodeError(400, "Bad request options"));

        if ( !(option.duration < 3601))
            return cb(new CodeError(400, "Bad request duration"));


        application.PG.getQuery('tcpDump').create(option, (err, res) => {
            if (err)
                return cb(err);

            application.Esl.bgapi(
                `luarun DumpUpload.lua ${res} ${option.duration} '${option.filter}'`,
                function (resFs) {
                    const err = checkEslError(resFs);
                    if (err) {
                        application.PG.getQuery('tcpDump').remove(res, e => {
                            if (e)
                                log.error(e)
                        });
                        return cb(err, res);
                    }

                    return cb(null, res);
                }
            )
        })
    },

    get: (caller, option = {}, cb) => {
        if (!validateRoot(caller)) {
            return cb(new CodeError(403, 'Permission denied!'))
        }

        if (!option.id)
            return cb(new CodeError(400, "Bad request id is required"));

        application.PG.getQuery('tcpDump').get(option.id, cb);
    },

    remove: (caller, option = {}, cb) => {
        if (!validateRoot(caller)) {
            return cb(new CodeError(403, 'Permission denied!'))
        }

        if (!option.id)
            return cb(new CodeError(400, "Bad request id is required"));

        application.PG.getQuery('tcpDump').remove(option.id, cb);
    },

    update: (caller, option = {}, cb) => {
        if (!validateRoot(caller)) {
            return cb(new CodeError(403, 'Permission denied!'))
        }

        if (!option.id)
            return cb(new CodeError(400, "Bad request id is required"));

        application.PG.getQuery('tcpDump').update(option.id, option, cb);
    }
};

function validateRoot(caller) {
    return caller && !caller.domain;
}

module.exports = Service;