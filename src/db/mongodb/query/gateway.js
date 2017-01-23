/**
 * Created by igor on 19.01.17.
 */

"use strict";

const CodeError = require(__appRoot + '/lib/error'),
    conf = require(__appRoot + '/conf'),
    gatewayCollectionName = conf.get('mongodb:collectionGateway')
    ;

module.exports = {
    addQuery: addQuery
};

function addQuery(db) {
    return {
        item: (name, cb) => {
            db
                .collection(gatewayCollectionName)
                .findOne({name: name}, cb)
        },

        insertOrUpdate: (name, data, cb) => {
            db
                .collection(gatewayCollectionName)
                .update({name: name}, data, {upsert: true}, cb)
        },

        removeByName: (name, cb) => {
            db
                .collection(gatewayCollectionName)
                .remove({name: name}, cb)
        },

        removeNotNames: (names, cb) => {
            db
                .collection(gatewayCollectionName)
                .remove({name: {$nin: names}}, {multi: true}, cb)
        },

        incrementLineByName: (name, inc, direction, cb) => {
            const $inc = {
                "stats.active": inc
            };

            if (inc === -1) {
                $inc[`stats.${direction === 'inbound' ? 'callsIn' : 'callsOut'}`] = 1
            }

            db
                .collection(gatewayCollectionName)
                .findAndModify({name: name}, {}, {$inc}, {upsert: true, new: true}, cb)
        },


        incrementLineByRealm: (realm, inc, direction, cb) => {
            const $inc = {
                "stats.active": inc
            };

            if (inc === -1) {
                $inc[`stats.${direction === 'inbound' ? 'callsIn' : 'callsOut'}`] = 1
            }

            db
                .collection(gatewayCollectionName)
                .findAndModify({"params.realm": realm}, {}, {$inc}, {upsert: false, new: true}, cb)
        },

        cleanActiveChannels: (cb) => {
            db
                .collection(gatewayCollectionName)
                .update({}, {$set: {"stats.active": 0}}, {multi: true}, cb)
        }
    }
}