/**
 * Created by i.navrotskyj on 15.03.2016.
 */
'use strict';

var conf = require(__appRoot + '/conf'),
    buildFilterQuery = require('./utils').buildFilterQuery,
    ObjectID = require('mongodb').ObjectID,
    collectionHookName = conf.get('mongodb:collectionHook');

module.exports = {
    addQuery: addQuery
};

function addQuery(db) {
    return {
        list: function (filter, cb) {
            let _f = (filter instanceof Object) ? filter : {};
            return db
                .collection(collectionHookName)
                .find(_f)
                .toArray(cb);
        },

        item: function (id, domain, col, cb) {
            if (!ObjectID.isValid(id))
                return cb(new Error('Bad id'));

            return db
                .collection(collectionHookName)
                .findOne({"_id": new ObjectID(id), "domain": domain}, col || {}, cb)
        },
        
        update: function (id, domain, doc, cb) {
            if (!ObjectID.isValid(id))
                return cb(new Error('Bad id'));

            return db
                .collection(collectionHookName)
                .updateOne({"_id": new ObjectID(id), "domain": domain}, doc, cb)
        },
        
        count: function (filter, cb) {
            return db
                .collection(collectionHookName)
                .count(filter, cb)
        },
        
        create: function (doc, cb) {
            return db
                .collection(collectionHookName)
                .insert(doc, cb);
        },

        remove: function (id, domain, cb) {
            if (!ObjectID.isValid(id))
                return cb(new Error('Bad id'));

            return db
                .collection(collectionHookName)
                .removeOne({"_id": new ObjectID(id), "domain": domain}, cb)
        },

        search: function (domain, option, cb) {
            var filter = option['filter'];
            var columns = option['columns'] || {};
            var limit = parseInt(option['limit'], 10) || 40;
            var sort = option['sort'] || {};
            var pageNumber = parseInt(option['pageNumber'], 10);

            var query = buildFilterQuery(filter);
            if (domain) {
                query['$and'].push({
                    "domain": domain
                });
            };
            try {

                db
                    .collection(collectionHookName)
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
    };
};