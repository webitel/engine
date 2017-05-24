/**
 * Created by Igor Navrotskyj on 27.08.2015.
 */

'use strict';

var ObjectId = require('mongodb').ObjectId;

let Utils = module.exports = {
    buildFilterQuery: function (filter) {
        var filterArray = [];
        if (filter) {
            for (var key in filter) {
                if (!filter[key]) {
                    delete filter[key];
                    continue;
                };
                if (key == '_id' && ObjectId.isValid(filter[key])) {
                    filter[key] = ObjectId(filter[key]);
                    continue;
                }
                for (var item in filter[key]) {
                    if (filter[key][item] == '_id' && ObjectId.isValid(filter[key])) {
                        filter[key][item] = ObjectId(filter[key]);
                    }
                }
            }
            filterArray.push(filter)
        };

        return {
            "$and": filterArray
        };
    },

    validateUuid: function (id) {
        if (ObjectID.isValid(id)) {
            return true;
        };
        return false;
    },

    searchInCollection: function (db, collectionName, options, cb) {
        let filter = options['filter'],
            columns = options['columns'] || {},
            limit = parseInt(options['limit'], 10) || 40,
            sort = options['sort'] || {},
            pageNumber = parseInt(options['pageNumber'], 10) || 0,
            domain = options.domain;

        let query = Utils.buildFilterQuery(filter);

        if (domain) {
            query['$and'].push({
                "domain": domain
            });
        }
        try {

            db
                .collection(collectionName)
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

    countInCollection: function (db, collectionName, options, cb) {
        let filter = options['filter'],
            domain = options.domain;

        let query = Utils.buildFilterQuery(filter);

        if (domain) {
            query['$and'].push({
                "domain": domain
            });
        }
        try {

            db
                .collection(collectionName)
                .find(query['$and'].length == 0 ? {} : query)
                .count(cb);

            return 1;
        } catch (e) {
            cb(e);
        }
    }
};