/**
 * Created by igor on 05.07.17.
 */


'use strict';

const conf = require(__appRoot + '/conf'),
    CALLBACK_COLLECTION = conf.get('mongodb:collectionCallback'),
    MEMBERS_COLLECTION = conf.get('mongodb:collectionCallbackMembers'),
    utils = require('./utils'),
    CodeError = require(__appRoot + '/lib/error'),
    log = require(__appRoot + '/lib/log')(module),
    ObjectID = require("mongodb").ObjectID
;

module.exports = {
    addQuery: addQuery
};

function addQuery(db) {


    function removeMemberFromQueue(queue) {
        return db
            .collection(MEMBERS_COLLECTION)
            .removeOne({queue}, {multi: true}, e => {
                if (e)
                    log.error(e)
            });
    }

    function existsQueue(queue, domain, cb) {
        if (!ObjectID.isValid(queue))
            return cb(new CodeError(400, 'Bad queueId.'));

        return db
            .collection(CALLBACK_COLLECTION)
            .find({_id: new ObjectID(queue), domain}, {_id: 1})
            .count(cb);
    }


    return {
        create: (doc, cb) => {
            return db
                .collection(CALLBACK_COLLECTION)
                .insert(doc, cb);
        },

        search: (options, cb) => {
            return utils.searchInCollection(db, CALLBACK_COLLECTION, options, cb);
        },

        findById: function (_id, domainName, cb) {
            if (!ObjectID.isValid(_id))
                return cb(new CodeError(400, 'Bad objectId.'));

            return db
                .collection(CALLBACK_COLLECTION)
                .findOne({_id: new ObjectID(_id), domain: domainName}, cb);
        },

        update: function (_id, domainName, doc = {}, cb) {
            if (!ObjectID.isValid(_id))
                return cb(new CodeError(400, 'Bad objectId.'));

            let data = {
                $set: {}
            };
            for (let key in doc) {
                if (doc.hasOwnProperty(key) && key !== '_id' && key !== 'domain' ) {
                    data.$set[key] = doc[key];
                }
            }
            return db
                .collection(CALLBACK_COLLECTION)
                .updateOne({_id: new ObjectID(_id), domain: domainName}, data, cb);

        },

        remove: function (_id, domainName, cb) {
            if (!ObjectID.isValid(_id))
                return cb(new CodeError(400, 'Bad objectId.'));

            return db
                .collection(CALLBACK_COLLECTION)
                .removeOne({_id: new ObjectID(_id), domain: domainName}, (err, res) => {
                    if (err)
                        return cb(err);
                    removeMemberFromQueue(_id);
                    return cb(null, res);
                });
        },

        members: {
            create: (doc, cb) => {

                existsQueue(doc.queue, doc.domain, (err, count) => {
                    if (!count) {
                        return cb(new CodeError(404, `Not found ${doc.queue}`))
                    }

                    return db
                        .collection(MEMBERS_COLLECTION)
                        .insert(doc, cb);

                });
            },

            search: (options, cb) => {
                return utils.searchInCollection(db, MEMBERS_COLLECTION, options, cb);
            },

            findById: function (_id, queue,  domainName, cb) {
                if (!ObjectID.isValid(_id))
                    return cb(new CodeError(400, 'Bad objectId.'));

                return db
                    .collection(MEMBERS_COLLECTION)
                    .findOne({_id: new ObjectID(_id), queue, domain: domainName}, cb);
            },

            update: function (_id, queue, domainName, doc = {}, cb) {
                if (!ObjectID.isValid(_id))
                    return cb(new CodeError(400, 'Bad objectId.'));

                let data = {
                    $set: {}
                };
                for (let key in doc) {
                    if (doc.hasOwnProperty(key) && key !== '_id' && key !== 'domain' && key !== 'queue' && key !== 'comments') {
                        data.$set[key] = doc[key];
                    }
                }
                return db
                    .collection(MEMBERS_COLLECTION)
                    .updateOne({_id: new ObjectID(_id), queue, domain: domainName, done: {$ne: true}}, data, cb);

            },

            remove: function (_id, queue, domainName, cb) {
                if (!ObjectID.isValid(_id))
                    return cb(new CodeError(400, 'Bad objectId.'));

                return db
                    .collection(MEMBERS_COLLECTION)
                    .removeOne({_id: new ObjectID(_id), queue, domain: domainName}, cb);
            },

            addComment: function (_id, queue, domainName, doc = {}, cb) {
                if (!ObjectID.isValid(_id))
                    return cb(new CodeError(400, 'Bad objectId.'));

                doc._id = new ObjectID();
                let data = {
                    $push: {
                        comments :doc
                    }
                };

                return db
                    .collection(MEMBERS_COLLECTION)
                    .updateOne({_id: new ObjectID(_id), queue, domain: domainName}, data, e => {
                        if (e)
                            return cb(e)

                        return cb(null, doc)
                    });
            },

            removeComment: function (_id, queue, domainName, commentId, cb) {
                if (!ObjectID.isValid(_id))
                    return cb(new CodeError(400, 'Bad objectId.'));

                if (!ObjectID.isValid(commentId))
                    return cb(new CodeError(400, 'Bad commentId.'));

                let data = {
                    $pull: { comments: {  _id:  new ObjectID(commentId) }  }
                };

                return db
                    .collection(MEMBERS_COLLECTION)
                    .updateOne({_id: new ObjectID(_id), queue, domain: domainName}, data, cb);
            },

            updateComment: function (_id, queue, domainName, commentId, text, cb) {
                if (!ObjectID.isValid(_id))
                    return cb(new CodeError(400, 'Bad objectId.'));

                if (!ObjectID.isValid(commentId))
                    return cb(new CodeError(400, 'Bad commentId.'));

                let data = {
                    $set: { "comments.$.comment": text }
                };

                return db
                    .collection(MEMBERS_COLLECTION)
                    .updateOne(
                        {
                            _id: new ObjectID(_id),
                            queue,
                            domain: domainName,
                            comments: {$elemMatch: {_id: new ObjectID(commentId)}}
                        },
                        data,
                        cb
                    );
            }
        }
    }
}