/**
 * Created by igor on 26.04.16.
 */

'use strict';


const conf = require(__appRoot + '/conf'),
    CodeError = require(__appRoot + '/lib/error'),
    Mongo = require("mongodb"),
    ObjectID = Mongo.ObjectID,
    generateUuid = require('node-uuid'),
    dialerCollectionName = conf.get('mongodb:collectionDialer'),
    memberCollectionName = conf.get('mongodb:collectionDialerMembers'),
    agentsCollectionName = conf.get('mongodb:collectionDialerAgents'),
    AGENT_STATUS = require(__appRoot + '/services/autoDialer/const').AGENT_STATUS,
    log = require(__appRoot + '/lib/log')(module),
    utils = require('./utils')
    ;

module.exports = {
    addQuery: addQuery
};

function addQuery (db) {

    const fnFilterDialerCommunications = `function(communications, codes, ranges, allCodes) {
        return communications.filter( (i, key) => {
                    if (i.state !== 0)
                        return false;
                    print(codes[0]);
                    var idx = codes.indexOf(i.type);
    
                    if (~idx) {
                        i.isTypeFound = 1;
                        var rangeProperty = ranges[idx];
                        if (i.rangeId && i.rangeId === rangeProperty.rangeId) {
                            if (i.rangeAttempts >= rangeProperty.attempts) {
                                return false;
                            }
                        } else {
                            i.rangeId = rangeProperty.rangeId;
                            i.rangeAttempts = 0;
                        }
                        i.rangePriority = rangeProperty.priority || 0;
                        
                    } else if (~allCodes.indexOf(i.type)) {
                        return false
                    } else {
                        if (!i.rangeAttempts) i.rangeAttempts = 0;
                        i.isTypeFound = 0;
                        i.rangePriority = -1;
                    }
                    
                    if (!i.lastCall)
                        i.lastCall = 0;
    
                    return true;
                })
    }`;

    const fnKeySort = `function(arr, keys) {

        keys = keys || {};
    
        var sortFn = function(a, b) {
            var sorted = 0, ix = 0;
    
            while (sorted === 0 && ix < KL) {
                var k = obIx(keys, ix);
                if (k) {
                    var dir = keys[k];
                    sorted = _keySort(a[k], b[k], dir);
                    ix++;
                }
            }
            return sorted;
        };
    
        var obIx = function(obj, ix){
            return Object.keys(obj)[ix];
        };
    
        var _keySort = function(a, b, d) {
            d = d !== null ? d : 1;
            // a = a.toLowerCase(); // this breaks numbers
            // b = b.toLowerCase();
            if (a == b)
                return 0;
            return a > b ? 1 * d : -1 * d;
        };
    
        var KL = Object.keys(keys).length;
    
        if (!KL)
            return arr.sort(sortFn);
    
        for ( var k in keys) {
            // asc unless desc or skip
            keys[k] =
                keys[k] == 'desc' || keys[k] == -1  ? -1
                    : (keys[k] == 'skip' || keys[k] === 0 ? 0
                    : 1);
        }
        arr = arr.sort(sortFn);
        return arr;
    };`;

    db.collection("system.js")
        .update({_id: "fnFilterDialerCommunications"}, {$set: {
            value: new Mongo.Code(fnFilterDialerCommunications)
        }}, {upsert: true}, e => {
            if (e)
                throw e;
        });

    db.collection("system.js")
        .update({_id: "fnKeySort"}, {$set: {
            value: new Mongo.Code(fnKeySort)
        }}, {upsert: true}, e => {
            if (e)
                throw e;
        });

    const _resetAgents = (dialerId) => {
        return db
            .collection(agentsCollectionName)
            .update(
                {"dialer": {$elemMatch: {_id: dialerId, process: {$ne: null}}}},
                {$set: {"dialer.$.process": null}},
                {multi: true}
            );
    };

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

        create: function (doc = {}, cb) {
            setDefUuidDestination(doc.resources);
            return db
                .collection(dialerCollectionName)
                .insert(doc, cb);
        },
        
        update: function (_id, domainName, doc = {}, cb) {
            if (!ObjectID.isValid(_id))
                return cb(new CodeError(400, 'Bad objectId.'));

            let data = {
                $set: {},
                $currentDate: {
                    lastModified: {$type: "timestamp" }
                }
            };
            setDefUuidDestination(doc.resources);
            for (let key in doc) {
                if (doc.hasOwnProperty(key) && key != '_id' && key != 'domain' && key !== 'stats') {
                    data.$set[key] = doc[key];
                }
            };
            return db
                .collection(dialerCollectionName)
                .updateOne({_id: new ObjectID(_id), domain: domainName}, data, cb);

        },

        resetProcessStatistic: function (dialerId, domainName, cb) {
            if (!ObjectID.isValid(dialerId))
                return cb(new CodeError(400, 'Bad objectId.'));

            const _id = new ObjectID(dialerId);

            _resetAgents(_id);

            return db
                .collection(dialerCollectionName)
                .updateOne(
                    {_id, domain: domainName, active: {$ne: true}},
                    {
                        $set: {
                            "stats.active": 0,
                            "stats.resource": {}
                        }
                    },
                    cb
                );
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
                .removeOne({_id: new ObjectID(_id), dialer: dialerId, _lock: null}, (e, res) => {
                    if (e)
                        return cb(e);

                    if (res && res.result && res.result.n === 0)
                        return cb(new CodeError(406, `Not Acceptable`));

                    return cb(null, res);
                });
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

        _getActiveDialer: function (project, cb) {
            return db
                .collection(dialerCollectionName)
                .find({
                    active: true
                }, project)
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

        _updateMultiMembers: (filter, update, cb) => {
            return db
                .collection(memberCollectionName)
                .update(
                    filter,
                    update,
                    {multi: true},
                    cb
                )
        },
        
        _updateMember: function (filter, doc, sort, cb) {
            return db
                .collection(memberCollectionName)
                .findOneAndUpdate(
                    filter,
                    doc,
                    {sort, projection: {_log: 0}},
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
                        callTimeout: 10, // TODO
                        skills: skills,
                        randomPoint: [Math.random(), 0]
                    },
                    $max: {
                        noAnswerCount: +params.no_answer_count
                    },
                    // $addToSet: {setAvailableTime: null},
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
                            lastStatusChange: Date.now() // todo ?
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
        },

        _findAndModifyAgentByHunting: (dialerId, filter, sort, update, cb) => {

            return db
                .collection(agentsCollectionName)
                .findAndModify(
                    filter,
                    sort,
                    update,
                    {
                        new: true,
                        fields: {
                            "_id" : 1,
                            "agentId" : 1,
                            "state" : 1,
                            "status" : 1,
                            "busyDelayTime" : 1,
                            "lastStatusChange" : 1,
                            "maxNoAnswer" : 1,
                            "noAnswerDelayTime" : 1,
                            "rejectDelayTime" : 1,
                            "wrapUpTime" : 1,
                            "callTimeout" : 1,
                            "skills" : 1,
                            "randomPoint" : 1,
                            "noAnswerCount": 1,
                            "lastModified": 1,
                            "dialer": {$elemMatch: {_id: dialerId}}
                        }
                    },
                    cb
                )
        },

        _updateAgentMulti: (filter, update, cb) => {
            return db
                .collection(agentsCollectionName)
                .update(filter, update, {multi: true}, cb)
        },

        _resetMembers: (dialerId, resetLog = false, callerId, cb) => {

            let bulk = db.collection(memberCollectionName).initializeOrderedBulkOp(),
                count = 0;

            const time = Date.now();

            const cursor = db
                .collection(memberCollectionName)
                .find({dialer: dialerId, _endCause: {$ne: null}}, {communications: 1, _id: 1});

            const respBulk = err => {
                if (err)
                    log.error(err);
            };

            let getUpdate;

            if (resetLog) {
                getUpdate = ($unset, $set) => {
                    $set._log = [{
                        steps: [{
                            "time" : time,
                            "data" : `Reset by ${callerId}`
                        }]
                    }];
                    return {
                        $unset,
                        $set
                    }
                }
            } else {
                getUpdate = ($unset, $set) => {
                    return {
                        $unset,
                        $set,
                        $push: {
                            _log: {
                                steps: [{
                                    "time" : time,
                                    "data" : `Reset by ${callerId}`
                                }]
                            }
                        }
                    }
                }
            }

            cursor.each((err, doc) => {
                if (err)
                    return cb(err);

                if (doc) {
                    if (doc.communications instanceof Array) {

                        const $unset = {
                            _endCause: 1,
                            _probeCount: 1,
                            callSuccessful: 1,
                            _lastNumberId: 1,
                            _lastMinusProbe: 1,
                            _nextTryTime: 1
                        };

                        const $set = {};
                        for (let i = 0, len = doc.communications.length; i < len; i++) {
                            $set[`communications.${i}.state`] = 0;
                            $unset[`communications.${i}._id`] = 1;
                            $unset[`communications.${i}._probe`] = 1;
                            $unset[`communications.${i}._score`] = 1;
                            $unset[`communications.${i}.rangeId`] = 1;
                            $unset[`communications.${i}.rangeAttempts`] = 1;
                            $unset[`communications.${i}.lastCall`] = 1;
                        }

                        bulk.find({_id: doc._id}).updateOne(getUpdate($unset, $set));
                        count++;

                        if ( count % 1000 == 0 ) {
                            log.debug(`Exec bulk member reset: ${count}`);
                            bulk.execute(respBulk);
                            bulk = db.collection(memberCollectionName).initializeOrderedBulkOp();
                        }
                    }

                } else {
                    // Execute any pending operations
                    if ( count % 1000 != 0 ) {
                        bulk.execute(respBulk);
                        log.debug(`Exec bulk member reset: ${count}`);
                    }

                    return cb(null, count);
                }
            });
        }
    }
}

function setDefUuidDestination(resources) {
    if (resources instanceof Array) {
        for (let res of resources) {
            if (res.destinations instanceof Array) {
                for (let dest of res.destinations) {
                    if (!dest.uuid)
                        dest.uuid = generateUuid.v4();
                }
            }
        }
    }
}