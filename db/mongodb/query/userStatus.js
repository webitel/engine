/**
 * Created by Igor Navrotskyj on 16.09.2015.
 */

'use strict';

var conf = require(__appRoot + '/conf'),
    statusCollectionName = conf.get('mongodb:collectionAgentStatus')
    ;

module.exports = {
    addQuery: addQuery
};

function addQuery(db) {
    return {
        create: function (data, cb) {
            db
                .collection(statusCollectionName)
                .insert(data, function (err, res) {
                    var result = res && res['ops'];
                    if (result instanceof Array) {
                        result = result[0];
                    }
                    ;
                    cb(err, result);
                });

            return 1;
        },
        
        setDuration: function (userId, duration, cb) {
            db
                .collection(statusCollectionName)
                .findAndModify(
                    {"userId": userId},
                    {"date": -1},
                    {"$set": {"duration": duration}},
                    cb
                );

            return 1;
        },
        
        _removeByUserId: function (domain, userId, cb) {
            db
                .collection(statusCollectionName)
                .remove({
                    "userId": userId,
                    "domain": domain
                }, function (err, res) {
                    return cb(err, res && res.result);
                });

            return 1;
        }
    }
};