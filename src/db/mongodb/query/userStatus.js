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
                (err, result) => {
                    if (err)
                        return cb(err);

                    if ((!data['status'] || !data['state']) && result && result['status'] && result['state']) {
                        data['status'] = result['status'];
                        data['state'] = result['state'];
                        data['description'] = result['description'] || "";
                    };

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
                    "account": userId,
                    "domain": domain
                }, function (err, res) {
                    return cb(err, res && res.result);
                });

            return 1;
        }
    }
};