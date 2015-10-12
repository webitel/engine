/**
 * Created by Igor Navrotskyj on 04.09.2015.
 */

'use strict';

var conf = require(__appRoot + '/conf'),
    CodeError = require(__appRoot + '/lib/error'),
    buildFilterQuery = require('./utils').buildFilterQuery,
    ObjectID = require('mongodb').ObjectID,
    contactBookCollectionName = conf.get('mongodb:collectionContactBook')
    ;

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
        create: function (data, cb) {
            db
                .collection(contactBookCollectionName)
                .insert(data, function (err, res) {
                    var result = res && res['ops'];
                    if (result instanceof Array) {
                        result = result[0];
                    }
                    ;
                    cb(err, result);
                });

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
            var limit = parseInt(option['limit']) || 40;

            var sort = option['sort'] || {};
            var pageNumber = option['pageNumber'];

            var query = buildFilterQuery(filter);
            if (domain) {
                query['$and'].push({
                    "domain": domain
                });
            };

            db
                .collection(contactBookCollectionName)
                .find(query['$and'].length == 0 ? {} : query, columns)
                .sort(sort)
                .skip(pageNumber > 0 ? ((pageNumber - 1) * limit) : 0)
                .limit(limit)
                .toArray(cb);

            return 1;
        },

        /**
         *
         * @param domain
         * @param _id
         * @param data
         * @param cb
         * @returns {*}
         */
        updateById: function (domain, _id, data, cb) {
            if (!ObjectID.isValid(_id)) {
                return cb(new CodeError(400, "Bad id."))
            };
            data['domain'] = domain;

            db
                .collection(contactBookCollectionName)
                .findAndModify(
                {"_id": new ObjectID(_id), "domain": domain},
                [],
                data,
                function (err, result) {
                    if (err) return cb(err);

                    if (result && !result['value'])
                        return cb(new CodeError(404, 'Not found'));

                    if (result)
                        return cb(null, result.value);
                });

            return 1;
        },

        /**
         *
         * @param domain
         * @param _id
         * @param cb
         * @returns {*}
         */
        removeById: function (domain, _id, cb) {
            if (!ObjectID.isValid(_id)) {
                return cb(new CodeError(400, "Bad id."))
            };

            db
                .collection(contactBookCollectionName)
                .remove({"_id": new ObjectID(_id), "domain": domain}, cb);
        },
        
        removeByDomain: function (domain, cb) {
            db
                .collection(contactBookCollectionName)
                .remove({
                    "domain": domain
                }, function (err, res) {
                    return cb(err, res && res.result);
                });

            return 1;
        }
    }
}