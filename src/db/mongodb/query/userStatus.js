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
            var collection = db
                .collection(statusCollectionName);

            collection
                .findAndModify(
                {"account": data['account'], "domain": data['domain']},
                {"date": -1},
                {"$set": {"endDate": data['date']}},
                {limit: 1},
                (err) => {
                    if (err)
                        return cb(err);

                    collection
                        .insert(data, (err) => {
                            if (err)
                                return cb(err);

                            return cb(null);
                        });
                }
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