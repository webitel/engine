/**
 * Created by Igor Navrotskyj on 16.09.2015.
 */

'use strict';

const conf = require(__appRoot + '/conf'),
    log = require(__appRoot + '/lib/log')(module),
    statusCollectionName = conf.get('mongodb:collectionAgentStatus')
;

module.exports = {
    addQuery: addQuery
};


function addQuery(db) {

    return {
        create: function (data, cb) {

        },

        _removeByUserId: function (domain, userId, cb) {
            db
                .collection(statusCollectionName)
                .remove({
                    "account": userId,
                    "domain": domain
                }, function (err, res) {
                    return cb(err, res && res.result);
                });

            return 1;
        }
    }
}

