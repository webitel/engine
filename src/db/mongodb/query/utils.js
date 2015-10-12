/**
 * Created by Igor Navrotskyj on 27.08.2015.
 */

'use strict';

var ObjectId = require('mongodb').ObjectId;

module.exports = {
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
    }
};