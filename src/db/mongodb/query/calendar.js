/**
 * Created by igor on 12.04.16.
 */

'use strict';

var conf = require(__appRoot + '/conf'),
    CodeError = require(__appRoot + '/lib/error'),
    buildFilterQuery = require('./utils').buildFilterQuery,
    ObjectID = require('mongodb').ObjectID,
    calendarCollectionName = conf.get('mongodb:collectionCalendar');

module.exports = {
    addQuery: addQuery
};

function addQuery (db) {
    return {
        search: function (domain, option, cb) {
            var filter = option['filter'];
            var columns = option['columns'] || {};
            var limit = option['limit'] || 40;
            var sort = option['sort'] || {};
            var pageNumber = option['pageNumber'];

            var query = buildFilterQuery(filter);
            if (domain) {
                query['$and'].push({
                    "domain": domain
                });
            };
            try {

                db
                    .collection(calendarCollectionName)
                    .find(query['$and'].length == 0 ? {} : query, columns)
                    .sort(sort)
                    .skip(pageNumber > 0 ? ((pageNumber - 1) * limit) : 0)
                    .limit(limit)
                    .toArray(cb);

                return 1;
            } catch (e) {
                cb(e);
            }
        },

        findById: function (domain, id, cb) {
            if (!ObjectID.isValid(id))
                return cb(new CodeError(400, "Bad request calendar id."));

            db
                .collection(calendarCollectionName)
                .findOne({
                    "_id": new ObjectID(id),
                    "domain": domain
                }, cb);

            return 1;
        },
        
        insert: function (doc, cb) {
            return db
                .collection(calendarCollectionName)
                .insert(doc, cb);
        },

        updateById: function (domain, id, doc, cb) {
            if (!ObjectID.isValid(id))
                return cb(new CodeError(400, "Bad request calendar id."));

            return db
                .collection(calendarCollectionName)
                .updateOne(
                    {_id: new ObjectID(id), domain: domain},
                    doc,
                    cb
            );
        },
        
        removeById: function (domain, id, cb) {
            if (!ObjectID.isValid(id))
                return cb(new CodeError(400, "Bad request calendar id."));

            return db
                .collection(calendarCollectionName)
                .remove(
                {_id: new ObjectID(id), domain: domain},
                cb
            );
        } 
    }
}