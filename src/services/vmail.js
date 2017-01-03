/**
 * Created by igor on 30.12.16.
 */

"use strict";
    
const bgapi = require('./channel').bgApi,
    log = require(__appRoot + '/lib/log')(module),
    validateCallerParameters = require(__appRoot + '/utils/validateCallerParameters'),
    checkPermissions = require(__appRoot + '/middleware/checkPermissions'),
    CodeError = require(__appRoot + '/lib/error');

const Service = {
    list: (caller, option = {}, cb) => {

        const domain = validateCallerParameters(caller, option.domain);
        let perm = caller.id !== option.id + '@' + domain ? 'r' : 'ro';

        checkPermissions(caller, 'vmail', perm, (e) => {
            if (e)
                return cb(e);

            if (!domain)
                return cb(new CodeError(400, 'Domain is required.'));

            if (!option.id)
                return cb(new CodeError(400, 'Account id is required.'));

            bgapi(`vm_list ${option.id}@${domain}`, (err, res) => {
                if (err)
                    return cb(err);

                const lines = (res.body || '')
                    .match(/[^\r\n]+/g);

                if (!lines)
                    return cb(null, []);

                try {
                    return cb(
                        null,
                        lines
                            .map(i => {
                                const message = i.split(/:/g);
                                return {
                                    createdOn: +message[0],
                                    readOn: +message[1],
                                    userName: message[2],
                                    domain: message[3],
                                    folder: message[4],
                                    path: message[5],
                                    uuid: message[6],
                                    cidName: message[7],
                                    cidNumber: message[8],
                                    messageLen: +message[9]
                                }
                            })
                    )
                } catch (e) {
                    return cb(e)
                }
            });
        });
    },

    setState: (caller, option = {}, cb) => {
        const domain = validateCallerParameters(caller, option.domain);

        if (!domain)
            return cb(new CodeError(400, 'Domain is required.'));

        if (!option.id)
            return cb(new CodeError(400, 'Account id is required.'));

        const perm = caller.id !== option.id + '@' + domain ? 'u' : 'uo';

        checkPermissions(caller, 'vmail', perm, (e) => {
            if (e)
                return cb(e);

            if (!option.uuid)
                return cb(new CodeError(400, 'Message id is required.'));

            if (option.state !== 'read' && option.state !== 'unread')
                return cb(new CodeError(400, `Bad state: ${option.state || '_undef_'}`));

            bgapi(`vm_read ${option.id}@${domain} ${option.state} ${option.uuid}`, (err, res) => {
                if (err)
                    return cb(err);

                return cb(null, res && res.body);
            });
        });

    },

    remove: (caller, option = {}, cb) => {
        const domain = validateCallerParameters(caller, option.domain);

        if (!domain)
            return cb(new CodeError(400, 'Domain is required.'));

        if (!option.id)
            return cb(new CodeError(400, 'Account id is required.'));

        if (!option.uuid)
            return cb(new CodeError(400, 'Message id is required.'));

        const perm = caller.id !== option.id + '@' + domain ? 'r' : 'ro';

        checkPermissions(caller, 'vmail', perm, (e) => {
            if (e)
                return cb(e);

            bgapi(`vm_delete ${option.id}@${domain} ${option.uuid}`, (err, res) => {
                if (err)
                    return cb(err);

                return cb(null, res && res.body);
            })
        });

    }
};

module.exports = Service;

