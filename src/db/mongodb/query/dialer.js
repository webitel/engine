/**
 * Created by igor on 26.04.16.
 */

'use strict';


var conf = require(__appRoot + '/conf'),
    CodeError = require(__appRoot + '/lib/error'),
    dialerCollectionName = conf.get('mongodb:collectionDialer'),
    memberCollectionName = conf.get('mongodb:collectionDialerMembers'),
    agentsCollectionName = conf.get('mongodb:collectionDialerAgents'),
    ObjectID = require('mongodb').ObjectID,
    utils = require('./utils')
    ;

module.exports = {
    addQuery: addQuery
};

function addQuery (db) {
    return {
        //TODO del collection
        _dialerCollection: db.collection(dialerCollectionName),

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
        
        memberById: function (_id, dialerName, cb, addDialer) {
            if (!ObjectID.isValid(_id) || !ObjectID.isValid(dialerName))
                return cb(new CodeError(400, 'Bad objectId.'));
            // TODO...

            return db
                .collection(memberCollectionName)
                .findOne({_id: new ObjectID(_id), dialer: dialerName}, (err, res) => {
                    if (err)
                        return cb(err);

                    if (addDialer) {
                        db
                            .collection(dialerCollectionName)
                            .findOne({_id: new ObjectID(dialerName)}, (err, resDialer) => {
                                return cb(err, res, resDialer);
                            });
                    } else {
                        return cb(err, res)
                    }

                });
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

        removeMemberByFilter: function (dialerId, filter, cb) {
            let _f = filter || {};
            _f.dialer = dialerId;
            return db
                .collection(memberCollectionName)
                .remove(_f, {multi: true}, cb);
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
        
        _updateLockedMembers: (id, lockId, cause, cb) => {
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

        _updateMemberFix: (id, data, cb) => {
            return db
                .collection(memberCollectionName)
                .update(
                    {_id: id},
                    data,
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
        },

        _lockCount: function (dialerId, cb) {
            return db
                .collection(memberCollectionName)
                .find({dialer: dialerId, _lock: true})
                .count(cb)
                
        },

        _initAgentInDialer: function (id, dialerId, cb) {
            return db
                .collection(agentsCollectionName)
                .update({agentId: id, "dialer._id": {$ne: dialerId}}, {
                    $push: {
                        dialer: {
                            _id: dialerId,
                            callCount: 0,
                            gotCallCount: 0,
                            callTimeSec: 0,
                            lastBridgeCallTimeStart: 0,
                            lastBridgeCallTimeEnd: 0,
                            connectedTimeSec: 0,
                            process: null,
                            setAvailableTime: null,
                            lastStatus: ""
                        }
                    },
                    $currentDate: { lastModified: true }
                }, {upsert: false}, cb)
        },

        _initAgent: function (agentId, params = {}, skills, cb) {
            return db
                .collection(agentsCollectionName)
                .update({agentId: agentId}, {
                    $set: {
                        state: params.state,
                        status: params.status,
                        busyDelayTime: +params.busy_delay_time,
                        lastStatusChange: +params.last_status_change * 1000,
                        maxNoAnswer: +params.max_no_answer,
                        noAnswerDelayTime: +params.no_answer_delay_time,
                        rejectDelayTime: +params.reject_delay_time,
                        wrapUpTime: +params.reject_delay_time,
                        skills: skills,
                        randomPoint: [Math.random(), 0]
                    },
                    $max: {
                        noAnswerCount: +params.no_answer_count
                    },
                    $currentDate: { lastModified: true }
                }, {upsert: true}, cb)
        },

        _setAgentState: function (agentId, state, cb) {
            return db
                .collection(agentsCollectionName)
                .findAndModify(
                    {agentId: agentId},
                    {},
                    {
                        $set: {
                            state,
                            randomPoint: [Math.random(), 0],
                            lastStatusChange: Date.now()
                        },
                        $currentDate: { lastModified: true }
                    },
                    {upsert: true, new: true},
                    cb
                )
        },

        _setAgentStatus: function (agentId, status, cb) {
            return db
                .collection(agentsCollectionName)
                .findAndModify(
                    {agentId: agentId},
                    {},
                    {
                        $set: {
                            status,
                            randomPoint: [Math.random(), 0],
                            lastStatusChange: Date.now()
                        },
                        $currentDate: { lastModified: true }
                    },
                    {upsert: true, new: true},
                    cb
                )
        },

        _getAgentCount: (filter = {}, cb) => {
            return db
                .collection(agentsCollectionName)
                .find(filter)
                .count(cb);
        },

        _findAndModifyAgent: (filter, sort, update, cb) => {
            return db
                .collection(agentsCollectionName)
                .findAndModify(filter, sort, update, {new: true}, cb)
        }
    }
}