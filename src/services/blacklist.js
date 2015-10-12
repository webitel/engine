/**
 * Created by Igor Navrotskyj on 27.08.2015.
 */

'use strict';

var CodeError = require(__appRoot + '/lib/error'),
    checkPermissions = require(__appRoot + '/middleware/checkPermissions'),
    validateCallerParameters = require(__appRoot + '/utils/validateCallerParameters');

var Service = {
    create: function (caller, option, cb) {
        checkPermissions(caller, 'blacklist', 'c', function (err) {
            if (err)
                return cb(err);

            option = option || {};
            var domain = validateCallerParameters(caller, option['domain']);
            option['domain'] = domain;


            if (!domain || !option['name'] || !option['number']) {
                return cb(new CodeError(400, 'Bad request: domain, name or number is required.'));
            };
            option['number'] = option['number'].toString();
            var dbBlacklist = application.DB._query.blacklist;
            return dbBlacklist.createOrUpdate(option, cb);
        });
    },
    
    getNames: function (caller, domain, cb) {
        checkPermissions(caller, 'blacklist', 'r', function (err) {
            if (err)
                return cb(err);

            domain = validateCallerParameters(caller, domain);

            if (!domain) {
                return cb(new CodeError(400, 'Domain is required.'));
            }
            ;

            var dbBlacklist = application.DB._query.blacklist;
            return dbBlacklist.getNames(domain, cb);
        });
    },

    search: function (caller, domain, option, cb) {
        checkPermissions(caller, 'blacklist', 'r', function (err) {
            if (err)
                return cb(err);

            domain = validateCallerParameters(caller, domain);
            var dbBlacklist = application.DB._query.blacklist;

            return dbBlacklist.search(domain, option, cb);
        });
    },

    getFromName: function (caller, name, domain, query, cb) {
        checkPermissions(caller, 'blacklist', 'r', function (err) {
            if (err)
                return cb(err);
            if (!name) {
                return cb(new CodeError(400, "Name is required."));
            };
            query = query || {};
            var pageNumber = query['page'];
            var limit = query['limit'];
            var order = query['order'];
            var orderValue = query['orderValue'];
            var option = {
                "filter": {
                    "name": name
                }
            };

            if (order) {
                option['sort'] = {};
                option.sort[order] = orderValue == 1 ? 1 : -1;
            }
            ;

            option['limit'] = parseInt(limit);
            option['pageNumber'] = pageNumber;

            return Service.search(caller, domain, option, cb);
        });
    },

    getNumberFromName: function (caller, option, cb) {
        checkPermissions(caller, 'blacklist', 'r', function (err) {
            if (err)
                return cb(err);

            option = option || {};
            var query = {
                "filter": {
                    "name": option['name'],
                    "number": option['number']
                }
            };
            return Service.search(caller, option['domain'], query, cb);
        });
    },
    
    remove: function (caller, domain, option, cb) {
        checkPermissions(caller, 'blacklist', 'd', function (err) {
            if (err)
                return cb(err);

            if (!option) {
                return cb(new CodeError(400, 'Bad request.'));
            }
            ;
            domain = validateCallerParameters(caller, domain);
            if (!domain) {
                return cb(new CodeError(400, 'Domain is required.'));
            }
            ;

            var dbBlacklist = application.DB._query.blacklist;
            return dbBlacklist.remove(domain, option, cb);
        });
    },

    _removeByDomain: function (domain, cb) {
        if (!domain) {
            return cb(new CodeError(400, "Domain is required."));
        };
        var dbBlacklist = application.DB._query.blacklist;
        return dbBlacklist.removeByDomain(domain, cb);
    }
};

module.exports = Service;