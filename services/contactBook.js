/**
 * Created by Igor Navrotskyj on 04.09.2015.
 */

'use strict';

var CodeError = require(__appRoot + '/lib/error'),
    validateCallerParameters = require(__appRoot + '/utils/validateCallerParameters'),
    checkPermissions = require(__appRoot + '/middleware/checkPermissions');

var Service = {
    create: function (caller, domain, data, cb) {
        checkPermissions(caller, 'book', 'c', function (err) {
            try {
                if (err)
                    return cb(err);

                domain = validateCallerParameters(caller, domain);
                if (!domain || !data || typeof data['name'] != 'string' || !(data['phones'] instanceof Array)) {
                    return cb(new CodeError(400, 'Bad request.'));
                };
                data['domain'] = domain;
                var dbBook = application.DB._query.book;
                dbBook.create(data, cb);
            } catch (e) {
                return cb(e);
            };
        });
    },

    search: function (caller, domain, option, cb) {
        checkPermissions(caller, 'book', 'r', function (err) {
            if (err)
                return cb(err);

            domain = validateCallerParameters(caller, domain);
            var dbBook = application.DB._query.book;
            return dbBook.search(domain, option, cb);
        });
    },
    
    list: function (caller, domain, option, cb) {
        checkPermissions(caller, 'book', 'r', function (err) {
            if (err)
                return cb(err);

            option = option || {};
            var query = {
                "filter": {
                    "name": option['name'],
                    "phones": option['phone'],
                    "tag": option['tag']
                },
                "limit": option['limit']
            };
            return Service.search(caller, domain, query, cb);
        });
    },
    
    getById: function (caller, domain, id, cb) {
        checkPermissions(caller, 'book', 'r', function (err) {
            if (err)
                return cb(err);

            var query = {
                "filter": {
                    "_id": id
                }
            };
            return Service.search(caller, domain, query, function (err, res) {
                if (err) {
                    return cb(err);
                };

                if (res && res.length == 0) {
                    return cb(new CodeError(404, "Not found."))
                };

                return cb(null, res && res[0])
            });
        });
    },
    
    updateItem: function (caller, domain, id, data, cb) {
        checkPermissions(caller, 'book', 'u', function (err) {
            if (err)
                return cb(err);

            domain = validateCallerParameters(caller, domain);
            if (!domain) {
                return cb(new CodeError(400, "Domain is required."));
            };

            if (!data) {
                return cb(new CodeError(400, "Bad request."))
            };

            if (!data['phones'] || !data['name']) {
                return cb(new CodeError(400, "Phones or name is required."))
            };

            var dbBook = application.DB._query.book;
            return dbBook.updateById(domain, id, data, cb);
        });
    },
    
    removeItem: function (caller, domain, id, cb) {
        checkPermissions(caller, 'book', 'd', function (err) {
            if (err)
                return cb(err);

            domain = validateCallerParameters(caller, domain);
            if (!domain) {
                return cb(new CodeError(400, "Domain is required."));
            };

            var dbBook = application.DB._query.book;
            return dbBook.removeById(domain, id, function (err, res) {
                if (err)
                    return cb(err);
                var result = res && res['result'];
                if (!result || result.n != 1) {
                    return cb(new CodeError(404, "Not found"));
                }
                return cb(null, result);
            });
        });
    },
    
    _removeByDomain: function (domain, cb) {
        if (!domain) {
            return cb(new CodeError(400, "Domain is required."));
        };
        var dbBook = application.DB._query.book;
        return dbBook.removeByDomain(domain, cb);
    }
};

module.exports = Service;