/**
 * Created by Igor Navrotskyj on 26.08.2015.
 */

'use strict';

var conf = require(__appRoot + '/conf'),
    authCollectionName = conf.get('mongodb:collectionAuth');

module.exports = {
    addQuery: addQuery
};

function addQuery(db) {
    return {
        add: function (option, cb) {
            return db
                .collection(authCollectionName)
                .findAndModify({"key": option['key']}, [], option, {"upsert": true}, cb);
        },

        getByKey: function (key, cb) {
            return db
                .collection(authCollectionName)
                .findOne({"key": key}, cb);
        },
        
        getByUserName: function (username, expires, cb) {
            return db
                .collection(authCollectionName)
                .find({
                    "username": username,
                    "expires": {
                        "$gt": expires
                    }
                })
                .sort({"expires": -1})
                .limit(1)
                .toArray(cb);
        },

        removeUserTokens: function (username, domain, cb) {
            return db
                .collection(authCollectionName)
                .remove({
                    "username": username,
                    "domain": domain
                }, cb);
        },
        
        removeDomainTokens: function (domainName, cb) {
            return db
                .collection(authCollectionName)
                .remove({
                    "domain": domainName
                }, cb);
        },
        
        remove: function (key, cb) {
            try {
                var query = {
                    "$or": [{
                        "expires": {
                            "$lt": new Date().getTime()
                        }
                    },
                        {
                            "key": key
                        }
                    ]
                };
                return db
                    .collection(authCollectionName)
                    .remove(query, cb);
            } catch (e) {
                return cb(e);
            }
        }
    };
};