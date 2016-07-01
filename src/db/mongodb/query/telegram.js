/**
 * Created by igor on 28.06.16.
 */


'use strict';

var conf = require(__appRoot + '/conf'),
    CodeError = require(__appRoot + '/lib/error'),
    ObjectID = require('mongodb').ObjectID,
    telegramCollectionName = conf.get('mongodb:collectionTelegram')
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
                .collection(telegramCollectionName)
                .insert(data, function (err, res) {
                    var result = res && res['ops'];
                    if (result instanceof Array) {
                        result = result[0];
                    }
                    cb(err, result);
                });

            return 1;
        },

        /**
         * 
         * @param userId
         * @param cb
         * @returns {Promise}
         */
        getByUserId: function (userId, cb) {
            return db
                .collection(telegramCollectionName)
                .findOne({
                    "user": userId
                }, {chatId: 1}, cb);
        },

        /**
         *
         * @param chatId
         * @param cb
         * @returns {Promise}
         */
        getSession: function (chatId, cb) {
            return db
                .collection(telegramCollectionName)
                .findOne({
                    "chatId": chatId
                }, cb);
        },

        /**
         *
         * @param chatId
         * @param username
         * @param cb
         * @returns {Promise}
         */
        removeSession:function (chatId, username, cb) {
            return db
                .collection(telegramCollectionName)
                .remove({
                    "chatId": chatId,
                    "user": username
                }, cb);
        }
    }
}