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
    dialerHistoryCollectionName = conf.get('mongodb:collectionDialerHistory'),
    calendarCollectionName = conf.get('mongodb:collectionCalendar'),
    memberCollectionName = conf.get('mongodb:collectionDialerMembers'),
    agentsCollectionName = conf.get('mongodb:collectionDialerAgents'),
    AGENT_STATUS = require(__appRoot + '/services/autoDialer/const').AGENT_STATUS,
    AGENT_STATE = require(__appRoot + '/services/autoDialer/const').AGENT_STATE,
    log = require(__appRoot + '/lib/log')(module),
    utils = require('./utils'),
    getDomainFromStr = require(__appRoot + '/utils/parse').getDomainFromStr
;

module.exports = {
    addQuery: addQuery
};

function addQuery (db) {

    const fnFilterDialerCommunications = `function(communications, codes, ranges, allCodes) {
        return communications.filter( (i, key) => {
                    if (i.state !== 0)
                        return false;
                    
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


    function removeDialerHistory(dialerId, cb) {
        if (!ObjectID.isValid(dialerId))
            return cb(new Error(`Bad dialer object id: ${dialerId}`));

        return db
            .collection(dialerHistoryCollectionName)
            .remove({dialer: dialerId}, {multi: true}, cb);
    }

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

            const oid = new ObjectID(_id);

            removeDialerHistory(oid, e => {
                if (e)
                    log.error(e);
            });

            return db
                .collection(dialerCollectionName)
                .removeOne({_id: oid, domain: domainName}, cb);
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
                if (doc.hasOwnProperty(key) && key !== '_id' && key !== 'domain' && key !== 'stats' && key !== 'active' && key !== 'state') {
                    data.$set[key] = doc[key];
                }
            }
            return db
                .collection(dialerCollectionName)
                .updateOne({_id: new ObjectID(_id), domain: domainName}, data, cb);

        },

        resetProcessStatistic: function (options, domainName, cb) {
            const dialerId = options.id;
            if (!ObjectID.isValid(dialerId))
                return cb(new CodeError(400, 'Bad objectId.'));

            const _id = new ObjectID(dialerId);

            const $set = {};

            if (options.resetProcess) {
                $set['stats.active'] = 0;
                $set['stats.resource'] = {};
            }

            if (options.resetStats) {
                $set['stats.callCount'] = 0;
                $set['stats.errorCall'] = 0;
                $set['stats.successCall'] = 0;
                $set['stats.predictAbandoned'] = 0;
                $set['stats.queueLimit'] = 0;
                $set['stats.predictAdjust'] = 0;
                $set['stats.bridgedCall'] = 0;
                $set['stats.connectedCall'] = 0;
                $set['stats.waitSec'] = 0;
                $set['stats.amd'] = {};
            }

            return db
                .collection(dialerCollectionName)
                .updateOne(
                    options.skipActive ? {_id, domain: domainName} : {_id, domain: domainName, active: {$ne: true}},
                    {
                        $set
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
            try {
                return db
                    .collection(memberCollectionName)
                    .insertMany(doc, cb);
            } catch (e) {
                return cb(e)
            }
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

        _removeByDomain: function (domain, cb) {
            const dialers = [];
            db
                .collection(dialerCollectionName)
                .find({domain}, {_id: 1})
                .toArray((err, res) => {
                    if (err)
                        return cb(err);

                    if (!(res instanceof Array)) {
                        return cb();
                    }

                    for (let dialer of res) {
                        dialers.push(dialer._id);
                    }

                    db
                        .collection(dialerHistoryCollectionName)
                        .remove({dialer: {$in: dialers}}, {multi: true}, e => {
                            if (e)
                                return log.error(e);
                        });


                    db
                        .collection(memberCollectionName)
                        .remove({dialer: {$in: dialers.map( i => i.toString())}}, {multi: true}, e => {
                            if (e)
                                return log.error(e);
                        });

                    return db
                        .collection(dialerCollectionName)
                        .remove({domain}, {multi: true}, cb);

                });
        },

        removeMemberByDialerId: function (dialerId, cb) {
            return db
                .collection(memberCollectionName)
                .remove({dialer: dialerId}, cb);
        },

        updateMember: function (_id, dialerId, doc, cb) {
            if (!ObjectID.isValid(_id))
                return cb(new CodeError(400, 'Bad objectId.'));

            db
                .collection(memberCollectionName)
                .findOne({_id: new ObjectID(_id), dialer: dialerId}, (err, res) => {
                    if (err)
                        return cb(err);

                    if (!res)
                        return cb(new CodeError(404, `Member ${_id} not found`));

                    if (res._lock)
                        return cb(new CodeError(406, `Member ${_id} locked`));

                    let data = {
                        $set: {}
                    };

                    for (let key in doc) {
                        if (key === 'communications') {
                            if (doc[key] instanceof Array) {
                                doc[key] = doc[key].map( (comm, idx) => {
                                    let storageComm = findCommunications(res.communications, comm.number);
                                    if (storageComm) {
                                        PROTECTED_FIELDS_COMMUNICATION.forEach(colName => {
                                            if (storageComm.hasOwnProperty(colName)) {
                                                comm[colName] = storageComm[colName]
                                            }
                                        })
                                    }

                                    return comm;
                                });
                            }

                            data.$set[key] = doc[key];
                        } else if (doc.hasOwnProperty(key) && key != '_id' && key != 'dialer') {
                            data.$set[key] = doc[key];
                        }
                    }

                    return db
                        .collection(memberCollectionName)
                        .updateOne({_id: new ObjectID(_id), dialer: dialerId}, data, cb);
                });
        },

        terminateMember: function(memberDb, cause, cb) {
            const $set = {_endCause: cause, callSuccessful: true};

            if (memberDb.communications instanceof Array) {
                for (let i = 0; i < memberDb.communications.length; i++) {
                    $set[`communications.${i}.state`] = 2
                }
            }

            db
                .collection(memberCollectionName)
                .findOneAndUpdate({
                    _id: memberDb._id,
                    dialer: memberDb.dialer,
                    _lock: null,
                    _endCause: null,
                    _waitingForResultStatusCb: {$ne: 1}
                }, {$set}, {projection: {_id: 1}}, (err, res) => {
                    if (err)
                        return cb(err);

                    if (!res || !res.value)
                        return cb(new CodeError(404, `Member ${memberDb._id} not found`));

                    return cb(null, res.value);
                });
        },

        aggregateMembers: function (dialerId, aggregateQuery, cb) {
            let query = [
                {$match:{dialer: dialerId}}
            ];
            query = query.concat(aggregateQuery);
            return db
                .collection(memberCollectionName)
                .aggregate(query, {allowDiskUse:true}, cb);
        },

        _updateDialer: function (_id, state, cause, active, nextTick, cb) {
            if (!ObjectID.isValid(_id))
                return cb(new CodeError(400, 'Bad objectId.'));

            if (typeof _id === 'string') {
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

        _setNumbersStateCrashMembers: (cause, state) => {
            var fn = () => {
                db
                    .collection(memberCollectionName)
                    .update(
                        {"_endCause":"PROCESS_CRASH", "communications": {$elemMatch: {state: {$ne: state}}}},
                        { $set: { "communications.$.state": state}}, {multi: true},
                        (err, res) => {
                            if (err) {
                                return log.error(err)
                            }

                            if (res.result.n !== 0) {
                                fn();
                                return
                            }
                            log.debug(`end set numbers end state`)
                        }
                    )
            };

            fn()
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
                    {sort, projection: {_log: 0}, returnOriginal: false},
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

        _setActiveAgents: function (dialerId, active, cb) {
            return db
                .collection(agentsCollectionName)
                .update(
                    {
                        dialer: {$elemMatch: {_id: dialerId}}
                    }, {
                        $set: {"dialer.$.active": active}
                    },
                    {multi: true},
                    cb
                );
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

        _initAgent: function (agentId, domain, params = {}, skills, cb) {
            const update = {
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
                    // randomPoint: [Math.random(), 0]
                    randomValue: Math.random()
                },
                $max: {
                    noAnswerCount: +params.no_answer_count
                },
                // $addToSet: {setAvailableTime: null},
                $currentDate: { lastModified: true }
            };

            const time = Date.now();
            if (params.status !== AGENT_STATUS.LoggedOut) {
                update.$min = {
                    lastLoggedInTime: time,
                    lastStateChange: time
                };
                update.$unset = {loggedOutTime : 1};
            } else {
                update.$min = {
                    loggedOutTime: time,
                    lastStateChange: time
                };
                update.$unset = {lastLoggedInTime : 1};
            }
            return db
                .collection(agentsCollectionName)
                .update({agentId: agentId, domain}, update, {upsert: true}, cb)
        },

        _setAgentState: function (agentId, state, cb) {
            return db
                .collection(agentsCollectionName)
                .findAndModify(
                    {agentId: agentId, domain: getDomainFromStr(agentId)},
                    {},
                    {
                        $set: {
                            state,
                            randomValue: Math.random(),
                            lastStateChange: Date.now()
                        },
                        $currentDate: { lastModified: true }
                    },
                    {upsert: true, new: true},
                    cb
                )
        },

        _setAgentStatus: function (agentId, status, cb) {
            let filter = {agentId: agentId, domain: getDomainFromStr(agentId)};

            const time = Date.now();
            const update = {
                $set : {
                    status,
                    randomValue: Math.random(),
                    lastStatusChange: time

                },
                $currentDate: { lastModified: true }
            };

            return db
                .collection(agentsCollectionName)
                .findOne(
                    filter,
                    {lastStatusChange: 1, statusInfo: 1, status: 1, dialer: 1, lastStateChange: 1, state: 1},
                    (err, res) => {
                        if (err)
                            return cb(err);

                        if (res) {
                            if (res.status && res.lastStatusChange) {
                                const sec = Math.round( (time - res.lastStatusChange) / 1000 );
                                update.$inc = {
                                    [`statusInfo.${res.status}`]: sec
                                };

                                let idle = false;
                                let lastChangeStatusState = null;
                                if (res.state === AGENT_STATE.Waiting && (res.status === AGENT_STATUS.Available || res.status === AGENT_STATUS.AvailableOnDemand)) {
                                    idle = true;
                                    lastChangeStatusState = Math.max(res.lastStateChange, res.lastStatusChange);
                                }

                                if (res.dialer instanceof Array) {
                                    for (let i = 0; i < res.dialer.length; i++) {
                                        if (res.dialer[i].active > 0) {
                                            if (res.lastStatusChange > res.dialer[i].active) {
                                                update.$inc[`dialer.${i}.${res.status}`] = sec;
                                            } else {
                                                update.$inc[`dialer.${i}.${res.status}`] =  Math.round( (time - res.dialer[i].active) / 1000 );
                                            }

                                            if (idle) {
                                                if (lastChangeStatusState > res.dialer[i].active) {
                                                    update.$inc[`dialer.${i}.idleSec`] = Math.round( (time - lastChangeStatusState) / 1000 );
                                                } else {
                                                    update.$inc[`dialer.${i}.idleSec`] = Math.round( (time - res.dialer[i].active) / 1000 );
                                                }

                                                update.$inc[`stats.${res.dialer[i]._id}.idleSec`] = update.$inc[`dialer.${i}.idleSec`];
                                            }

                                        }
                                    }
                                }
                            }
                        }

                        if (status !== AGENT_STATUS.LoggedOut) {
                            update.$set.lastLoggedInTime = time;
                            update.$unset = {loggedOutTime : 1};
                        } else {
                            update.$set.loggedOutTime = time;
                            update.$unset = {lastLoggedInTime : 1};
                        }
                        return db
                            .collection(agentsCollectionName)
                            .findAndModify(
                                filter,
                                {},
                                update,
                                {upsert: true, new: true},
                                cb
                            );
                    }
                );
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
            console.dir(filter,{depth: 100, colors: true});
            console.dir(sort,{depth: 100, colors: true});
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
                            "lastStateChange": 1,
                            "maxNoAnswer" : 1,
                            "noAnswerDelayTime" : 1,
                            "rejectDelayTime" : 1,
                            "wrapUpTime" : 1,
                            "callTimeout" : 1,
                            "skills" : 1,
                            "randomValue" : 1,
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

        _resetMembers: (dialerId, resetLog = false, fromDate, callerId, cb) => {

            let bulk = db.collection(memberCollectionName).initializeOrderedBulkOp(),
                count = 0;

            const time = Date.now();

            const cursor = db
                .collection(memberCollectionName)
                .find({
                    dialer: dialerId,
                    _endCause: {$ne: null} ,
                    createdOn: {$gte: fromDate},
                    callSuccessful: {$ne: true},
                    communications: {$elemMatch: {stopCommunication: {$ne: true}}}
                }, {communications: 1, _id: 1});

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
                            //_lastNumberId: 1,
                            _lastMinusProbe: 1,
                            _nextTryTime: 1,
                            lastCall: 1
                        };

                        const $set = {};
                        for (let i = 0, len = doc.communications.length; i < len; i++) {
                            if (doc.communications[i].stopCommunication === true)
                                continue;

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
        },

        _getCursor: (filter, projection) => {
            return db
                .collection(memberCollectionName)
                .find(filter, projection);
        },

        _updateOneMember: (filter, update, cb) => {
            return db
                .collection(memberCollectionName)
                .updateOne(filter, update, {}, cb);
        },

        _setExpireOnDay: (filter, update, cb) => {
            return db
                .collection(memberCollectionName)
                .findOneAndUpdate(filter, update, {
                    projection: {
                        domain: 1,
                        dialer: 1,
                        id: 1,
                        name: 1,
                        variables: 1,
                        _probeCount: 1,
                        _lastNumberId: 1,
                        communications: 1,
                    }
                }, cb);
        },

        insertDialerHistory: (dialerId, data = {}, cb) => {
            data.createdOn = Date.now();

            if (typeof dialerId === 'string' && ObjectID.isValid(dialerId)) {
                dialerId = new ObjectID(dialerId);
            }

            data.dialer = dialerId;

            return db
                .collection(dialerHistoryCollectionName)
                .insert(data, cb);
        },

        listHistory: (options = {}, cb) => {
            if (options.filter && ObjectID.isValid(options.filter.dialer)) {
                options.filter.dialer = new ObjectID(options.filter.dialer);
            }
            options.domain = null;
            return utils.searchInCollection(db, dialerHistoryCollectionName, options, cb);
        },

        removeDialerHistory: removeDialerHistory,

        getTimezoneFromDialer: (dialerId, cb) => {
            if (!ObjectID.isValid(dialerId))
                return cb(new CodeError(400, 'Bad objectId.'));

            return db
                .collection(dialerCollectionName)
                .aggregate([
                    {"$match": {_id: new ObjectID(dialerId)}},
                    {
                        $lookup:
                            {
                                from: calendarCollectionName,
                                localField: "_id.toString()",
                                foreignField: "calendar.id",
                                as: "calendarDoc"
                            }
                    },
                    {
                        "$unwind": "$calendarDoc"
                    },
                    {
                        "$project": {
                            "timezone": "$calendarDoc.timeZone.id"
                        }
                    }
                ], (err, res) => {
                    if (err)
                        return cb(err);

                    if (res && res[0] && res[0].timezone) {
                        return cb(null, res[0].timezone)
                    }
                    return cb(null, null)
                });
        },
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

const PROTECTED_FIELDS_COMMUNICATION = ["lastCall", "rangeAttempts", "rangeId", "_score", "_probe", "_id", "state", "status", "stopCommunication"];

function findCommunications(arr = [], number) {
    for (let i = 0; i < arr.length; i++)
        if (arr[i].number === number)
            return arr[i]
}
