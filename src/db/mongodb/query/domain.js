/**
 * Created by igor on 07.09.16.
 */

"use strict";

var conf = require(__appRoot + '/conf'),
    domainCollectionName = conf.get('mongodb:collectionDomain')
    ;

module.exports = {
    addQuery: addQuery
};

function addQuery(db) {
    return {
        getByName: (name, cb) => {

            db
                .collection(domainCollectionName)
                .findOne({
                    "name": name
                }, cb);
        },

        updateOrInserParams: (domainName, params, cb) => {
            let update = {
                "$set": params
            };

            db
                .collection(domainCollectionName)
                .update({
                    "name": domainName
                }, update, {upsert: true}, cb);
        }
    }
}