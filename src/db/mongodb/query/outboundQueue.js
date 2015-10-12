/**
 * Created by Igor Navrotskyj on 15.09.2015.
 * status:
 *  idle - ожидание;
 *  distribution - распределение;
 *  active - ;
 *  end -
 */

'use strict';

var conf = require(__appRoot + '/conf'),
    oQueryCollectionName = conf.get('mongodb:collectionOutboundQueue'),
    cdrCollectionName = conf.get('mongodb:collectionCDR'),
    dbUtils = require('./utils'),
    ObjectId = require('mongodb').ObjectId,
    STATUS = require(__appRoot + '/const/outboundQueue').STATUS;

module.exports = {
    addQuery: addQuery
};

function addQuery (db) {
    let resultInterface =  {
        getCallee: function (domain, option, cb) {
            let query = {
                "domain": domain,
                "status": STATUS.IDLE
            };
            if (option['tags']) {
                query['tags'] = option['tags'];
            };
            let limit = option['limit'] || 1;
            let sort = {
                "countOriginate": 1
            };
            let update = {
                "$set": {
                    "status": STATUS.DISTRIBUTION
                }
            };
            if (option['userId']) {
                update['$set']['userId'] = option['userId']
            };

            return db
                .collection(oQueryCollectionName)
                .findAndModify(
                    query,
                    sort,
                    update,
                    {limit: limit, multi: true},
                    cb
                );
        },

        getAvgBillSecUser: function (userId, cb) {
            return db
                .collection(cdrCollectionName)
                .aggregate(
                [
                    {
                        "$match": {
                            "variables.presence_id": userId
                        }
                    },
                    {
                        "$group": {
                            "_id": "$variables.presence_id",
                            "avg": {
                                "$avg": "$variables.billsec"
                            }
                        }
                    }
                ],
                function (err, res) {
                    if (err)
                        return cb(err);

                    if (!res || res.length < 1)
                        return cb(null, 0)

                    return cb(null, res[0].avg)
                }
            )
        },
        
        setState: function () {
            
        },
        
        updateItem: function (id, data, cb) {
            return db
                .collection(oQueryCollectionName)
                .update(
                {
                    "_id": id
                },
                data,
                cb
            );
        },
        
        insert: function (data, cb) {
            db
                .collection(oQueryCollectionName)
                .insert(
                    data,
                    cb
                );

            return 1;
        }
    };
    return resultInterface;
};