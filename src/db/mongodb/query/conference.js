/**
 * Created by i.navrotskyj on 08.10.2015.
 */
'use strict';


var conf = require(__appRoot + '/conf'),
    CodeError = require(__appRoot + '/lib/error'),
    ObjectID = require('mongodb').ObjectID,
    conferenceCollectionName = conf.get('mongodb:collectionConference')
    ;

const TYPE = {
    USER: "user",
    CONFERENCE: "conference"
};

module.exports = {
    addQuery: addQuery
};

function addQuery(db) {
    return {
        /**
         * 
         * @param email
         * @param cb
         * @returns {Promise}
         */
        getByEMail: function (email, cb) {
            return db
                .collection(conferenceCollectionName)
                .findOne({
                    "email": email
                }, cb);
        },

        /**
         *
         * @param email
         * @param password
         * @param data
         * @param cb
         * @returns {Promise}
         */
        insert: function (email, password, data, cb) {
            data['email'] = email;
            data['password'] = password;
            data['_createdOn'] = new Date().getTime();
            return db
                .collection(conferenceCollectionName)
                .insert(data, {fullResult: true}, function (err, r) {
                    if (err)
                        return cb(err);

                    try {
                        return cb(null, r && r['ops'] && r.ops[0]);
                    } catch (e) {
                        return cb(e);
                    };
                });
        },
        
        existsEmail: function (email, cb) {
            return db
                .collection(conferenceCollectionName)
                .findOne({
                    "email": email
                },
                function (err, res) {
                    if (err)
                        return cb(err);

                    return cb(
                        null,
                        res ? true : false
                    );
                }
            );
        },
        
        setConfirmed: function (_id, confirmed, cb) {
            if (!ObjectID.isValid(_id)) {
                return cb(new Error('Bad id.'));
            };
            var collection = db
                .collection(conferenceCollectionName)
            ;
            // TODO FindAndModify!!!
            return collection
                .update({
                    "_id": new ObjectID(_id),
                    "confirmed": !confirmed
                },
                {
                    "$set": {
                        "confirmed": confirmed
                    }
                },
                function (err, res) {
                    if (err)
                        return cb(err);
                    var _nMod = res && res.result && res.result['nModified'] == 1;
                    if (!_nMod) {
                        return cb(new Error('Not found.'));
                    };
                    return collection
                        .findOne({"_id": new ObjectID(_id)}, cb)
                }
            );
        },
        
        _getDeleteAgents: function (date, cb) {
            return db
                .collection(conferenceCollectionName)
                .find({
                    "_createdOn": {
                        "$lte": date
                    }
                })
                .toArray(cb);
        },
        
        _deleteExpiresAgents: function (date, cb) {
            return db
                .collection(conferenceCollectionName)
                .remove({
                    "_createdOn": {
                        "$lte": date
                    }
                }, cb);
        }
    }
}