/**
 * Created by Igor Navrotskyj on 27.08.2015.
 */

'use strict';

var conf = require(__appRoot + '/conf'),
    CodeError = require(__appRoot + '/lib/error'),
    buildFilterQuery = require('./utils').buildFilterQuery,
    blacklistCollectionName = conf.get('mongodb:collectionBlackList');

module.exports = {
    addQuery: addQuery
};

function addQuery(db) {
    return {

        /**
         *
         * @param data
         * @param cb
         * @returns {number}
         */
        createOrUpdate: function (data, cb) {
            db
                .collection(blacklistCollectionName)
                .update({
                        "domain": data['domain'],
                        "name": data['name'],
                        "number": data['number']
                    },
                    data,
                    {upsert: true},
                    cb
            );

            return 1;
        },

        /**
         *
         * @param domain
         * @param cb
         * @returns {number}
         */
        getNames: function (domain, cb) {
            db
                .collection(blacklistCollectionName)
                .aggregate([
                    {"$match": {"domain": domain}},
                    {"$project": {"name": 1}},
                    {"$group": {"_id": {"name": "$name"}}},
                    {"$project": {"name": "$_id.name", "_id": 0}}
                ], cb);

            return 1;
        },

        /**
         * 
         * @param domain
         * @param option
         * @param cb
         * @returns {number}
         */
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
                    .collection(blacklistCollectionName)
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

        /**
         *
         * @param domain
         * @param option
         * @param cb
         * @returns {*}
         */
        remove: function (domain, option, cb) {
            var filter = {
                "domain": domain
            };

            if (option['name']) {
                filter['name'] = option['name'];
            };

            if (option['number']) {
                filter['number'] = option['number'];
            };

            if (Object.keys(filter).length == 1) {
                return cb(new CodeError(400, 'Bad request'));
            };

            db
                .collection(blacklistCollectionName)
                .remove(filter, cb);

            return 1;
        },
        
        removeByDomain: function (domain, cb) {
            db
                .collection(blacklistCollectionName)
                .remove({
                    "domain": domain
                }, function (err, res) {
                    return cb(err, res && res.result);
                });

            return 1;
        }
    };
};