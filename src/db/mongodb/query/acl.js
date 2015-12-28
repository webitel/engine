/**
 * Created by i.navrotskyj on 22.12.2015.
 */
'use strict';
var conf = require(__appRoot + '/conf'),
    collectionAclPermissions = conf.get('mongodb:collectionAclPermissions');

module.exports = {
    addQuery: addQuery
};

function addQuery (db) {
    return {
        getRoles: function (fields, cb) {
            if (typeof fields == 'function') {
                cb = fields;
                fields = {};
            }

            return db
                .collection(collectionAclPermissions)
                .find({}, fields)
                .toArray(cb);
        },

        insert: function (data, cb) {
            return db
                .collection(collectionAclPermissions)
                .insert(data, cb);
        },

        update: function (query, data, cb) {
            return db
                .collection(collectionAclPermissions)
                .update(query, data, cb);
        },
        
        removeByName: function (name, cb) {
            return db
                .collection(collectionAclPermissions)
                .remove({"roles": name}, cb);
        },

        removeById: function (id, cb) {
            return db
                .collection(collectionAclPermissions)
                .remove({
                    "_id": id
                }, cb);
        }
    }
}