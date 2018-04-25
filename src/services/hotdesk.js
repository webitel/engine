/**
 * Created by Igor Navrotskyj on 21.09.2018.
 */

'use strict';

const log = require(__appRoot + '/lib/log')(module),
    checkPermissions = require(__appRoot + '/middleware/checkPermissions'),
    Collection = require(__appRoot + '/lib/collection'),
    validateCallerParameters = require(__appRoot + '/utils/validateCallerParameters'),
    CodeError = require(__appRoot + '/lib/error');

let loggedOutHotdesk = null;

const Service = {
    /**
     *
     * @param app
     */
    init: (app) => {
        if (loggedOutHotdesk) {
            return;
        }
        log.info('Init hotdesk');
        loggedOutHotdesk = new Collection('id');

        app.Schedule(60 * 1000, function () {
            log.debug('Schedule logout hotdesk.');
            if (loggedOutHotdesk.length() > 0) {
                const currentTime = Date.now();
                let hotdesk;
                for (let key of loggedOutHotdesk.getKeys()) {
                    hotdesk = loggedOutHotdesk.get(key);
                    if (hotdesk.time < currentTime) {
                        loggedOutHotdesk.remove(key);
                        application.WConsole.hotdeskSignOutById(`${hotdesk.id}@${hotdesk.domain}`, e => {
                            if (e) {
                                log.error(e);
                            }
                        });
                    }
                }
            }
        });

    },

    /**
     *
     * @param key
     * @param id
     * @param domain
     * @returns {boolean}
     */
    addTask: (key, id, domain) => {
        if (!loggedOutHotdesk) {
            return false;
        }
        loggedOutHotdesk.add(key, {id, domain, time: Date.now() + 60000});
    },

    /**
     *
     * @param caller
     * @param options
     * @param cb
     */
    signIn: (caller, options = {}, cb) => {
        checkPermissions(caller, 'hotdesk', 'u', (err) => {
            if (err)
                return cb(err);

            if (!options.address) {
                return cb(new CodeError(400, "Address is required"))
            }
            loggedOutHotdesk.remove(caller.id);
            application.WConsole.hotdeskSignIn(caller, options, cb);
        });
    },

    /**
     *
     * @param caller
     * @param options
     * @param cb
     */
    signOut: (caller, options, cb) => {
        checkPermissions(caller, 'hotdesk', 'u', (err) => {
            if (err)
                return cb(err);

            if (!options.address) {
                return cb(new CodeError(400, "Address is required"))
            }
            application.WConsole.hotdeskSignOut(caller, options, cb);
        });
    },

    /**
     *
     * @param caller
     * @param options
     * @param cb
     */
    list: (caller, options, cb) => {
        checkPermissions(caller, 'hotdesk', 'r', (err) => {
            if (err)
                return cb(err);

            options.domain = validateCallerParameters(caller, options.domain);
            if (!options.domain) {
                return cb(new CodeError(400, 'Domain is required'))
            }

            application.WConsole.hotdeskList(caller, options, cb);
        });
    },

    /**
     *
     * @param caller
     * @param options
     * @param cb
     */
    item: (caller, options, cb) => {
        checkPermissions(caller, 'hotdesk', 'r', (err) => {
            if (err)
                return cb(err);

            const domain = validateCallerParameters(caller, options.domain);
            if (!domain) {
                return cb(new CodeError(400, 'Domain is required'))
            }
            if (!options.id) {
                return cb(new CodeError(400, 'Id is required'))
            }

            application.WConsole.hotdeskList(caller, {domain, filter: {id: options.id}}, (err, res = []) => {
                if (err)
                    return cb(err);
                //TODO
                return cb(null, res[0]);
            });
        });
    },

    /**
     *
     * @param caller
     * @param options
     * @param cb
     */
    create: (caller, options = {}, cb) => {
        checkPermissions(caller, 'hotdesk', 'c', (err) => {
            if (err)
                return cb(err);

            options.domain = validateCallerParameters(caller, options.domain);
            if (!options.domain) {
                return cb(new CodeError(400, 'Domain is required'))
            }
            if (!options.id) {
                return cb(new CodeError(400, 'Id is required'))
            }

            application.WConsole.hotdeskCreate(caller, options, cb);
        });
    },

    /**
     *
     * @param caller
     * @param options
     * @param cb
     */
    remove: (caller, options, cb) => {
        checkPermissions(caller, 'hotdesk', 'd', (err) => {
            if (err)
                return cb(err);

            options.domain = validateCallerParameters(caller, options.domain);
            if (!options.domain) {
                return cb(new CodeError(400, 'Domain is required'))
            }
            if (!options.id) {
                return cb(new CodeError(400, 'Id is required'))
            }

            application.WConsole.hotdeskRemove(caller, options, cb);
        });
    },

    /**
     *
     * @param caller
     * @param options
     * @param cb
     */
    update: (caller, options, cb) => {
        checkPermissions(caller, 'hotdesk', 'u', (err) => {
            if (err)
                return cb(err);

            options.domain = validateCallerParameters(caller, options.domain);
            if (!options.domain) {
                return cb(new CodeError(400, 'Domain is required'))
            }
            if (!options.id) {
                return cb(new CodeError(400, 'Id is required'))
            }

            application.WConsole.hotdeskUpdate(caller, options, cb);
        });
    }
};

module.exports = Service;