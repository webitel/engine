/**
 * Created by Igor Navrotskyj on 07.09.2015.
 */

'use strict';


var conf = require(__appRoot + '/conf'),
    cdrCollectionName = conf.get('mongodb:collectionFile');

module.exports = {
    addQuery: addQuery
};

function addQuery(db) {
    return {
        existsRecordFile: function (uuid, cb) {
            return db
                .collection(cdrCollectionName)
                .findOne({
                    "uuid": uuid
                }, function (err, res) {
                    if (err) {
                        return cb(err);
                    };

                    return cb(null, !!res);
                });
        }
    };
};