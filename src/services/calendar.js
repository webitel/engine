/**
 * Created by igor on 12.04.16.
 */

'use strict';

var CodeError = require(__appRoot + '/lib/error'),
    checkPermissions = require(__appRoot + '/middleware/checkPermissions'),
    checkEslError = require(__appRoot + '/middleware/checkEslError'),
    parsePlainTableToJSONArray = require(__appRoot + '/utils/parse').plainTableToJSONArray,
    validateCallerParameters = require(__appRoot + '/utils/validateCallerParameters');

let Service = {

    /**
     *
     * @param caller
     * @param option
     * @param cb
     */
    list: function (caller, option, cb) {
        checkPermissions(caller, 'calendar', 'r', function (err) {
            if (err)
                return cb(err);

            option = option || {};
            let domain = validateCallerParameters(caller, option['domain']);
            option['domain'] = domain;

            let dbCalendar = application.DB._query.calendar;

            return dbCalendar.search(domain, option, cb);

        });
    },

    /**
     *
     * @param caller
     * @param option
     * @param cb
     */
    item: function (caller, option, cb) {
        checkPermissions(caller, 'calendar', 'r', function (err) {
            if (err)
                return cb(err);

            if (!option || !option.id)
                return cb(new CodeError(400, "Bad calendar id"));

            let domain = validateCallerParameters(caller, option['domain']);
            let dbCalendar = application.DB._query.calendar;

            return dbCalendar.findById(domain, option.id, cb);
        });
    },

    /**
     *
     * @param caller
     * @param option
     * @param cb
     */
    create: function (caller, option, cb) {
        checkPermissions(caller, 'calendar', 'c', function (err) {
            if (err)
                return cb(err);

            if (!option)
                return cb(new CodeError(400, "Bad request"));

            if (!option.name)
                return cb(new CodeError(400, "Name is required."));

            let domain = validateCallerParameters(caller, option['domain']);

            if (!domain)
                return cb(new CodeError(400, "Domain is required."));

            if (!(option.accept instanceof Array))
                return cb(new CodeError(400, "Accept is required."));

            let dbCalendar = application.DB._query.calendar;

            return dbCalendar.insert(option, cb);
        });
    },

    /**
     * 
     * @param caller
     * @param option
     * @param cb
     */
    update: function (caller, option, cb) {
        checkPermissions(caller, 'calendar', 'u', function (err) {
            if (err)
                return cb(err);

            if (!option)
                return cb(new CodeError(400, "Bad request"));

            if (!option.id)
                return cb(new CodeError(400, "Id is required."));

            if (!option.data)
                return cb(new CodeError(400, "Data is required."));

            let calendar = option.data;

            let domain = validateCallerParameters(caller, calendar['domain']);

            if (!domain)
                return cb(new CodeError(400, "Domain is required."));

            if (!(calendar.accept instanceof Array))
                return cb(new CodeError(400, "Accept is required."));

            let dbCalendar = application.DB._query.calendar;

            return dbCalendar.updateById(domain, option.id, calendar, cb);
        });
    },
    
    remove: function (caller, option, cb) {
        checkPermissions(caller, 'calendar', 'd', function (err) {
            if (err)
                return cb(err);

            if (!option)
                return cb(new CodeError(400, "Bad request"));

            if (!option.id)
                return cb(new CodeError(400, "Id is required."));

            let domain = validateCallerParameters(caller, option['domain']);

            if (!domain)
                return cb(new CodeError(400, "Domain is required."));

            let dbCalendar = application.DB._query.calendar;

            return dbCalendar.removeById(domain, option.id, cb);
        });
    },
    
    _removeByDomain: function (domainName, cb) {
        let dbCalendar = application.DB._query.calendar;

        return dbCalendar.removeByDomain(domainName, cb);
    }
};

module.exports = Service;