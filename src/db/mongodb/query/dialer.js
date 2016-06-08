/**
 * Created by igor on 26.04.16.
 */

'use strict';


var conf = require(__appRoot + '/conf'),
    CodeError = require(__appRoot + '/lib/error'),
    dialerCollectionName = conf.get('mongodb:collectionDialer'),
    memberCollectionName = conf.get('mongodb:collectionDialerMembers'),
    ObjectID = require('mongodb').ObjectID,
    utils = require('./utils')
    ;

module.exports = {
    addQuery: addQuery
};

function addQuery (db) {
    return {
        search: function (options, cb) {
            return utils.searchInCollection(db, dialerCollectionName, options, cb);
        },
        
        findById: function (_id, domainName, cb) {
            if (!ObjectID.isValid(_id))
                return cb(new CodeError(400, 'Bad objectId.'));

            return db
                .collection(dialerCollectionName)
                .findOne({_id: new ObjectID(_id), domain: domainName}, cb);
        },
        
        removeById: function (_id, domainName, cb) {
            if (!ObjectID.isValid(_id))
                return cb(new CodeError(400, 'Bad objectId.'));

            return db
                .collection(dialerCollectionName)
                .removeOne({_id: new ObjectID(_id), domain: domainName}, cb);
        },

        create: function (doc, cb) {

            return db
                .collection(dialerCollectionName)
                .insert(doc, cb);
        },
        
        update: function (_id, domainName, doc, cb) {
            if (!ObjectID.isValid(_id))
                return cb(new CodeError(400, 'Bad objectId.'));

            let data = {
                $set: {}
            };

            for (let key in doc) {
                if (doc.hasOwnProperty(key) && key != '_id' && key != 'domain') {
                    data.$set[key] = doc[key];
                }
            };
            return db
                .collection(dialerCollectionName)
                .updateOne({_id: new ObjectID(_id), domain: domainName}, data, cb);

        },
        
        memberList: function (options, cb) {
            return utils.searchInCollection(db, memberCollectionName, options, cb);
        },
        
        memberCount: function (options, cb) {
            return utils.countInCollection(db, memberCollectionName, options, cb);
        },
        
        memberById: function (_id, dialerName, cb) {
            if (!ObjectID.isValid(_id))
                return cb(new CodeError(400, 'Bad objectId.'));

            return db
                .collection(memberCollectionName)
                .findOne({_id: new ObjectID(_id), dialer: dialerName}, cb);
        },
        
        createMember: function (doc, cb) {
            return db
                .collection(memberCollectionName)
                .insert(doc, cb);
        },
        
        removeMemberById: function (_id, dialerId, cb) {
            if (!ObjectID.isValid(_id))
                return cb(new CodeError(400, 'Bad objectId.'));

            return db
                .collection(memberCollectionName)
                .removeOne({_id: new ObjectID(_id), dialer: dialerId}, cb);
        },

        removeMemberByDialerId: function (dialerId, cb) {
            return db
                .collection(memberCollectionName)
                .remove({dialer: dialerId}, cb);
        },

        updateMember: function (_id, dialerId, doc, cb) {
            if (!ObjectID.isValid(_id))
                return cb(new CodeError(400, 'Bad objectId.'));

            let data = {
                $set: {}
            };

            for (let key in doc) {
                if (doc.hasOwnProperty(key) && key != '_id' && key != 'dialer') {
                    data.$set[key] = doc[key];
                }
            };

            return db
                .collection(memberCollectionName)
                .updateOne({_id: new ObjectID(_id), dialer: dialerId}, data, cb);

        },
        
        aggregateMembers: function (dialerId, aggregateQuery, cb) {
            let query = [
                {$match:{dialer: dialerId}}
            ];
            query = query.concat(aggregateQuery);
            return db
                .collection(memberCollectionName)
                .aggregate(query, cb);
        },
        
        _updateDialer: function (_id, state, cause, active, nextTick, cb) {
            if (!ObjectID.isValid(_id))
                return cb(new CodeError(400, 'Bad objectId.'));

            if (typeof _id == 'string') {
                _id = new ObjectID(_id);
            }
            return db
                .collection(dialerCollectionName)
                .findOneAndUpdate(
                {_id: _id},
                {$set: {state: state, _cause: cause, active: active === true, nextTick: nextTick}},
                cb
            );
        },

        _getActiveDialer: function (cb) {
            return db
                .collection(dialerCollectionName)
                .find({
                    active: true
                })
                .toArray(cb)
        },
        
        _getDialerById: function (id, domain, cb) {
            if (typeof id === 'string' && ObjectID.isValid(id)) {
                id = new ObjectID(id);
            }

            return db
                .collection(dialerCollectionName)
                .findOne({_id: new ObjectID(id), domain: domain}, cb);
        },
        
        _updateLockedMembers: function (id, lockId, cause, cb) {
            return db
                .collection(memberCollectionName)
                .update(
                    {dialer: id, _lock: lockId},
                    {$set: {_endCause: cause}, $unset: {_lock: null}}, {multi: true},
                    cb
                )
        },
        
        _updateMember: function (filter, doc, sort, cb) {
            return db
                .collection(memberCollectionName)
                .findOneAndUpdate(
                    filter,
                    doc,
                    sort,
                    cb
                )
        },
        
        _aggregateMembers: function (agg, cb) {
            return db
                .collection(memberCollectionName)
                .aggregate(
                    agg,
                    cb
                )
        }
    }
}