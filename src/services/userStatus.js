/**
 * Created by Igor Navrotskyj on 16.09.2015.
 */

'use strict';

var CodeError = require(__appRoot + '/lib/error'),
    log = require(__appRoot + '/lib/log')(module)
    ;

var Service = {
    insert: function (option, cb) {
        let status = option['status'];
        let state = option['state'];
        let userId = option['userId'];

        if (!state && !status) {
            return cb(new CodeError(500, 'Caller %s status or state undefined.', option['userId']));
        };

        let data = option;
        data['date'] = new Date().getTime();

        // TODO agent, logged, count session...

        let dbUserStatus = application.DB._query.userStatus;
        dbUserStatus.setDuration(
            userId,
            option['prevDate'],
            function (err) {
                if (err)
                    log.error(err);

                return dbUserStatus.create(data, cb);
            }
        );
    },
    
    _removeByUserId: function (userId, domain, cb) {
        if (!domain || !userId) {
            return cb(new CodeError(400, "Domain is required."));
        };

        var dbUserStatus = application.DB._query.userStatus;
        return dbUserStatus._removeByUserId(domain, userId, cb);
    }
};

module.exports = Service;