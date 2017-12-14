/**
 * Created by Igor Navrotskyj on 27.09.2015.
 */

'use strict';
var CodeError = require(__appRoot + '/lib/error'),
    validateCallerParameters = require(__appRoot + '/utils/validateCallerParameters'),
    authService = require(__appRoot + '/services/auth'),
    plainTableToJSON = require(__appRoot + '/utils/parse').plainTableToJSON,
    plainCollectionToJSON = require(__appRoot + '/utils/parse').plainCollectionToJSON,
    log = require(__appRoot + '/lib/log')(module),
    checkPermissions = require(__appRoot + '/middleware/checkPermissions')
    ;


function createData(domain) {
    application.PG.getQuery('contacts').types.create("Email", domain, err => {
        if (err)
            return log.error(err);
        log.trace(`Create default communication type Email successful`);
    });
    application.PG.getQuery('contacts').types.create("Phone", domain, err => {
        if (err)
            return log.error(err);

        log.trace(`Create default communication type Phone successful`);
    });
}

var Service = {
    /**
     *
     * @param caller
     * @param option
     * @param cb
     */
    create: (caller, option, cb) => {
        checkPermissions(caller, 'domain', 'c', function (err) {
            if (err)
                return cb(err);

            if (!option || !option['name'] || !option['customerId']) {
                return cb(new CodeError(400, 'Bad request.'));
            };

            //if (!/(?=^.{4,253}$)(^((?!-)[a-zA-Z0-9-]{0,62}[a-zA-Z0-9]\.)+[a-zA-Z]{2,63}$)/.test(option.name))
            // TODO
            if (/\//.test(option.name))
                return cb(new CodeError(400, "Bad domain name."));

            application.WConsole.domainCreate(caller, option['name'], option['customerId'], option, function (err, res) {
                if (err)
                    return cb(err);
                createData(option.name);
                return cb(null, res);
            });
        });
    },

    /**
     *
     * @param caller
     * @param option
     * @param cb
     */
    list: (caller, option, cb) => {
        checkPermissions(caller, 'domain', 'r', function (err) {
            if (err)
                return cb(err);

            if (!option) {
                return cb(new CodeError(400, 'Bad request.'));
            };

            application.WConsole.domainList(caller, option['customerId'], function (err, res) {
                if (err)
                    return cb(err);
                if (option['type'] == 'plain') {
                    return cb(null, res);
                }
                return plainTableToJSON(res, null, cb, 0);
            });
        });
    },

    /**
     *
     * @param caller
     * @param options
     * @param cb
     */
    item: (caller = {}, options, cb) => {
        const perm = caller.domain ? 'ro' : 'r';
        checkPermissions(caller, 'domain', perm, function (err) {
            if (err)
                return cb(err);

            if (!options || !options['name']) {
                return cb(new CodeError(400, 'Bad request.'));
            };

            var name = validateCallerParameters(caller, options['name']);

            application.WConsole.domainItem(caller, name, function (err, res) {
                if (err)
                    return cb(err);

                return plainCollectionToJSON(res, cb);
            });
        });
    },

    /**
     * 
     * @param caller
     * @param options
     * @param cb
     */
    update: (caller = {}, options, cb) => {
        const perm = caller.domain ? 'uo' : 'u';
        checkPermissions(caller, 'domain', perm, function (err) {
            if (err)
                return cb(err);

            if (!options || !options['name'] || !options['type'] || !(options['params'] instanceof Array)) {
                return cb(new CodeError(400, 'Bad request.'));
            };

            var name = validateCallerParameters(caller, options['name']);

            application.WConsole.updateDomain(caller, name, options, function (err, res) {
                if (err)
                    return cb(err);

                return plainCollectionToJSON(res, cb);
            });
        });
    },

    /**
     *
     * @param caller
     * @param options
     * @param cb
     */
    remove: (caller, options, cb) => {
        checkPermissions(caller, 'domain', 'd', function (err) {
            if (err)
                return cb(err);

            if (!options || !options['name']) {
                return cb(new CodeError(400, 'Bad request.'));
            };

            application.WConsole.domainRemove(caller, options['name'], function (err, res) {
                if (err)
                    return cb(err);

                return cb(null, res)
            });
        });
    },

    settings: {
        get: (caller, options, cb) => {
            const perm = caller.domain ? 'ro' : 'r';
            checkPermissions(caller, 'domain', perm, (err) => {
                if (err)
                    return cb(err);

                let domain = validateCallerParameters(caller, options && options.name);
                if (!domain)
                    return cb(400, 'Domain is required.');

                application.DB._query.domain.getByName(domain, cb);
            })
        },

        getTokenList: (caller, options = {}, cb) => {
            const perm = caller.domain ? 'ro' : 'r';
            checkPermissions(caller, 'domain', perm, (err) => {
                if (err)
                    return cb(err);

                const domain = validateCallerParameters(caller, options.domain);

                application.DB._query.domain.listToken(domain, options.filter, options.columns, cb);
            })
        },

        genToken: (caller, options = {}, cb) => {
            const perm = caller.domain ? 'uo' : 'u';
            checkPermissions(caller, 'domain', perm, (err) => {
                if (err)
                    return cb(err);

                let domain = validateCallerParameters(caller, options && options.name);
                if (!domain)
                    return cb(new CodeError(400, 'Domain is required.'));

                if (!options.expire || options.expire <= Date.now())
                    return cb(new CodeError(400, 'Bad expire date.'));

                if (!options.role)
                    return cb(new CodeError(400, 'Bad role name.'));

                const {data, token} = authService.genDomainToken(caller.id, domain, {exp: options.expire, roleName: options.role});

                application.DB._query.domain.addToken(domain, data, (err) => {
                    if (err)
                        return cb(err);

                    return cb(null, {data,token});
                });
            })
        },

        removeToken: (caller, options = {}, cb) => {
            const perm = caller.domain ? 'uo' : 'u';
            checkPermissions(caller, 'domain', perm, (err) => {
                if (err)
                    return cb(err);

                let domain = validateCallerParameters(caller, options && options.name);
                if (!domain)
                    return cb(400, 'Domain is required.');

                if (!options.uuid)
                    return cb(400, 'Token id is required.');

                application.DB._query.domain.removeToken(domain, options.uuid, cb);
            })
        },

        setStateToken: (caller, options = {}, cb) => {
            const perm = caller.domain ? 'uo' : 'u';
            checkPermissions(caller, 'domain', perm, (err) => {
                if (err)
                    return cb(err);

                let domain = validateCallerParameters(caller, options && options.name);
                if (!domain)
                    return cb(400, 'Domain is required.');

                if (!options.uuid)
                    return cb(400, 'Token id is required.');

                if (typeof options.state !== 'boolean')
                    return cb(400, 'Bad state token.');

                application.DB._query.domain.setStateToken(domain, options.uuid, options.state, cb);
            })
        },

        updateOrInsert: (caller, options, cb) => {
            const perm = caller.domain ? 'uo' : 'u';
            checkPermissions(caller, 'domain', perm, (err) => {
                if (err)
                    return cb(err);

                let domain = validateCallerParameters(caller, options && options.name);
                if (!domain)
                    return cb(400, 'Domain is required.');

                delete options.domain;
                application.DB._query.domain.updateOrInserParams(domain, options, cb);
            })
        },

        _remove: (domainName, cb) => {
            application.DB._query.domain.remove(domainName, cb);
        }
    }
};

module.exports = Service;