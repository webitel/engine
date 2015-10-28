/**
 * Created by Igor Navrotskyj on 16.09.2015.
 */

'use strict';

var CodeError = require(__appRoot + '/lib/error'),
    log = require(__appRoot + '/lib/log')(module)
    ;

var Service = {
    insert: function (option) {
        let status = option['status'];
        let state = option['state'];
        let userId = option['account'];

        if (!state || !status || !userId) {
            return log.error('Caller %s status or state undefined.', userId);
        };

        let data = option;
        data['date'] = Date.now();
        let dbUserStatus = application.DB._query.userStatus;

        dbUserStatus.create(data, (err) => {
            if (err)
                log.error(err);
        });
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