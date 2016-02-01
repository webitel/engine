/**
 * Created by Igor Navrotskyj on 27.09.2015.
 */

'use strict';
var CodeError = require(__appRoot + '/lib/error'),
    validateCallerParameters = require(__appRoot + '/utils/validateCallerParameters'),
    plainTableToJSON = require(__appRoot + '/utils/parse').plainTableToJSON,
    plainCollectionToJSON = require(__appRoot + '/utils/parse').plainCollectionToJSON,
    checkPermissions = require(__appRoot + '/middleware/checkPermissions')
    ;

var Service = {
    /**
     *
     * @param caller
     * @param option
     * @param cb
     */
    create: function (caller, option, cb) {
        checkPermissions(caller, 'domain', 'c', function (err) {
            if (err)
                return cb(err);

            if (!option || !option['name'] || !option['customerId']) {
                return cb(new CodeError(400, 'Bad request.'));
            };

            if (!/(?=^.{4,253}$)(^((?!-)[a-zA-Z0-9-]{0,62}[a-zA-Z0-9]\.)+[a-zA-Z]{2,63}$)/.test(option.name))
                return cb(new CodeError(400, "Bad domain name."));

            application.WConsole.domainCreate(caller, option['name'], option['customerId'], option, function (err, res) {
                if (err)
                    return cb(err);

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
    list: function (caller, option, cb) {
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
                return plainTableToJSON(res, null, cb);
            });
        });
    },

    /**
     *
     * @param caller
     * @param options
     * @param cb
     */
    item: function (caller, options, cb) {
        checkPermissions(caller, 'domain/item', 'r', function (err) {
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
    update: function (caller, options, cb) {
        checkPermissions(caller, 'domain/item', 'u', function (err) {
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
    remove: function (caller, options, cb) {
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
    }
};

module.exports = Service;