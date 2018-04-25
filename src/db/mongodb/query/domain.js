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

        listToken: (domain, filter, columns, cb) => {
            const _f = filter instanceof Object ? filter : {};
            if (domain !== '*') {
                _f.name = domain;
            }

            db
                .collection(domainCollectionName)
                .find(_f, columns)
                .limit(40)
                .toArray(cb);
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
        },

        remove: (domainName, cb) => {
            db
                .collection(domainCollectionName)
                .remove({
                    "name": domainName
                }, (err, res) => {
                    return cb(err, res && res.result)
                });
        },

        getAuthSettings: (name, cb) => {
            db
                .collection(domainCollectionName)
                .findOne({name, "auth.enable": true}, {fields: {"auth": 1}}, (err, res) => {
                    return cb(err, res && res.auth)
                });
        }
    }
}