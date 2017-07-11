/**
 * Created by igor on 05.07.17.
 */

'use strict';

const conf = require(__appRoot + '/conf'),
    WIDGET_COLLECTION = conf.get('mongodb:collectionWidget'),
    utils = require('./utils'),
    CodeError = require(__appRoot + '/lib/error'),
    ObjectID = require("mongodb").ObjectID
;

module.exports = {
    addQuery: addQuery
};

function addQuery(db) {
    return {
        newId: () => new ObjectID(),
        create: (doc, cb) => {
            return db
                .collection(WIDGET_COLLECTION)
                .insert(doc, cb);
        },

        search: (options, cb) => {
            return utils.searchInCollection(db, WIDGET_COLLECTION, options, cb);
        },

        findById: function (_id, domainName, options, cb) {
            if (!ObjectID.isValid(_id))
                return cb(new CodeError(400, 'Bad objectId.'));

            return db
                .collection(WIDGET_COLLECTION)
                .findOne({_id: new ObjectID(_id), domain: domainName}, options, cb);
        },

        update: function (_id, domainName, doc = {}, cb) {
            if (!ObjectID.isValid(_id))
                return cb(new CodeError(400, 'Bad objectId.'));

            let data = {
                $set: {}
            };
            for (let key in doc) {
                if (doc.hasOwnProperty(key) && key !== '_id' && key !== 'domain' ) {
                    data.$set[key] = doc[key];
                }
            }
            return db
                .collection(WIDGET_COLLECTION)
                .updateOne({_id: new ObjectID(_id), domain: domainName}, data, cb);

        },

        remove: function (_id, domainName, cb) {
            if (!ObjectID.isValid(_id))
                return cb(new CodeError(400, 'Bad objectId.'));

            return db
                .collection(WIDGET_COLLECTION)
                .findOneAndDelete({_id: new ObjectID(_id), domain: domainName}, {projection: {_id:1, _filePath: 1} }, cb);
        },
    }
}
