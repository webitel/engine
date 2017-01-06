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

        setStateToken: (domainName, uuid, state, cb) => {
            let update = {
                $set:  {"tokens.$.enabled": state}
            };

            db
                .collection(domainCollectionName)
                .update({
                    "name": domainName,
                    "tokens.uuid": uuid
                }, update, cb);
        },

        removeToken: (domainName, uuid, cb) => {
            let update = {
                "$pull": {
                    tokens: {
                        uuid
                    }
                }
            };

            db
                .collection(domainCollectionName)
                .update({
                    "name": domainName
                }, update, cb);
        },

        getTokenByKey: (domain, uuid, cb) => {
            db
                .collection(domainCollectionName)
                .findOne({
                    name: domain,
                    tokens: {$elemMatch: {uuid: uuid, enabled: true}}
                }, {"tokens.$": 1, _id: 0}, cb)
        },

        addToken: (domainName, data, cb) => {
            let update = {
                "$push": {
                    tokens: data
                }
            };

            db
                .collection(domainCollectionName)
                .update({
                    "name": domainName
                }, update, {upsert: true}, cb);
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