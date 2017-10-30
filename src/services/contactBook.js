/**
 * Created by Igor Navrotskyj on 04.09.2015.
 */

'use strict';

const CodeError = require(__appRoot + '/lib/error'),
    validateCallerParameters = require(__appRoot + '/utils/validateCallerParameters'),
    checkPermissions = require(__appRoot + '/middleware/checkPermissions');

const Service = {

    list: function (caller, options, cb) {
        checkPermissions(caller, 'book', 'r', function (err) {
            if (err)
                return cb(err);

            let domain = validateCallerParameters(caller, options['domain']);
            if (!domain) {
                return cb(new CodeError(400, 'Bad request: domain is required.'));
            }
            options.domain = domain;
            application.PG.getQuery('contacts').list(options, cb);
        });
    },

    getById: function (caller, options, cb) {
        checkPermissions(caller, 'book', 'r', function (err) {
            if (err)
                return cb(err);

            let domain = validateCallerParameters(caller, options.domain);
            if (!domain) {
                return cb(new CodeError(400, 'Bad request: domain is required.'));
            }

            if (!options.id)
                return cb(new CodeError(400, 'Bad request: id is required.'));

            application.PG.getQuery('contacts').findById(options.id, domain, cb);
        });
    },

    create: function (caller, data, cb) {
        checkPermissions(caller, 'book', 'c', function (err) {
            if (err)
                return cb(err);

            const domain = validateCallerParameters(caller, data.domain);

            application.PG.getQuery('contacts').create(data, domain, cb);
        });
    },

    
    updateItem: function (caller, contact = {}, cb) {
        checkPermissions(caller, 'book', 'u', function (err) {
            if (err)
                return cb(err);

            const domain = validateCallerParameters(caller, contact.domain);
            if (!domain)
                return cb(new CodeError(400, "Domain is required."));

            if (!contact.id)
                return cb(new CodeError(400, "Id is required."));

            application.PG.getQuery('contacts').update(contact, contact.id, domain, cb);
        });
    },
    
    removeItem: function (caller, options, cb) {
        checkPermissions(caller, 'book', 'd', function (err) {
            if (err)
                return cb(err);

            const domain = validateCallerParameters(caller, options.domain);
            if (!domain) {
                return cb(new CodeError(400, "Domain is required."));
            }

            if (!options.id)
                return cb(new CodeError(400, "Id is required."));

            application.PG.getQuery('contacts').deleteById(options.id, domain, cb);
        });
    },
    
    _removeByDomain: function (domain, cb) {
        if (!domain) {
            return cb(new CodeError(400, "Domain is required."));
        }

        application.PG.getQuery('contacts').deleteByDomain(domain, cb);
    },
    
    types: {
        list: function (caller, options, cb) {
            checkPermissions(caller, 'book', 'r', function (err) {
                if (err)
                    return cb(err);

                let domain = validateCallerParameters(caller, options['domain']);
                if (!domain) {
                    return cb(new CodeError(400, 'Bad request: domain is required.'));
                }
                options.domain = domain;
                application.PG.getQuery('contacts').types.list(options, cb);
            });
        },
        
        create: function (caller, options, cb) {
            checkPermissions(caller, 'book', 'c', function (err) {
                if (err)
                    return cb(err);

                const domain = validateCallerParameters(caller, options['domain']);
                if (!domain)
                    return cb(new CodeError(400, 'Bad request: domain is required.'));


                if (!options.name)
                    return cb(new CodeError(400, 'Bad request: name is required.'));

                application.PG.getQuery('contacts').types.create(options.name, domain, cb);
            });
        }
    }
};

module.exports = Service;