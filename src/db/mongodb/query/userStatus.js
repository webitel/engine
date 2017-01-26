/**
 * Created by Igor Navrotskyj on 16.09.2015.
 */

'use strict';

var conf = require(__appRoot + '/conf'),
    log = require(__appRoot + '/lib/log')(module),
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
                {"account": data['account'], "domain": data['domain'], "endDate": null},
                {"date": -1},
                {"$set": {"endDate": data['date']}},
                {limit: 1, new: true},
                (err, result) => {
                    if (err)
                        return cb(err);

                    const dbStatus = result && result.value;

                    if ((!data['status'] || !data['state']) && dbStatus && dbStatus['status'] && dbStatus['state']) {
                        data['status'] = dbStatus['status'];
                        data['state'] = dbStatus['state'];
                        data['description'] = dbStatus['description'] || "";
                    }

                    if (dbStatus) {
                        dbStatus.timeSec = Math.round((dbStatus.endDate - dbStatus.date) / 1000);
                        application.Broker.publish(application.Broker.Exchange.STORAGE_COMMANDS, 'storage.commands.inbound',
                            {'exec-api': 'userStatus.saveToElastic', 'exec-args': result.value}, e => {
                                if (e)
                                    return log.error(e)
                            }
                        );
                    }
                    
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