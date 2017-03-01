/**
 * Created by igor on 25.05.16.
 */

const DIALER_STATES = require('./const').DIALER_STATES,
    DIALER_CAUSE = require('./const').DIALER_CAUSE,
    MEMBER_STATE = require('./const').MEMBER_STATE,

    CODE_RESPONSE_ERRORS = require('./const').CODE_RESPONSE_ERRORS,
    CODE_RESPONSE_RETRY = require('./const').CODE_RESPONSE_RETRY,
    CODE_RESPONSE_OK = require('./const').CODE_RESPONSE_OK,
    CODE_RESPONSE_MINUS_PROBE = require('./const').CODE_RESPONSE_MINUS_PROBE,

    NUMBER_STRATEGY = require('./const').NUMBER_STRATEGY,

    Member = require('./member'),
    Calendar = require('./calendar'),
    Collection = require(__appRoot + '/lib/collection'),
    log = require(__appRoot + '/lib/log')(module),
    EventEmitter2 = require('eventemitter2').EventEmitter2,
    dialerService = require(__appRoot + '/services/dialer'),
    async = require('async')
    ;

module.exports = class Dialer extends EventEmitter2 {

    constructor (type, config, calendarConfig, dialerManager) {
        super();
        this.type = type;
        this._id = config._id.toString();
        this._objectId = config._id;
        this._instanceId = application._instanceId;
        this._active = 0;
        this._agents = [];
        this._recources = {};

        this.consumerTag = null;
        this.queueName = `engine.dialer.${this._id}`;
        this._lastModified = 0;

        this._dbDialer = application.DB.collection('dialer');

        this._dbDialer.update({_id: this._objectId, "stats.active": null}, {
            $set: {"stats.active": 0}
        });

        // this.bigData = new Array(1e6).join('XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX\n');

        this.nameDialer = config.name;
        this.number = config.number || this.nameDialer;

        this._domain = config.domain;
        this.state = DIALER_STATES.Idle;
        this.cause = DIALER_CAUSE.Init;

        this._calendar = new Calendar(calendarConfig, config.communications);

        this.once('end', () => {
            this._calendar.stop();
            this.closeChannel();
        });

        this._setConfig(config);

        this.countMembers = 0;

        log.debug(`Init dialer: ${this.nameDialer}@${this._domain}`);

        this.members = new Collection('id');

        this.members.on('added', (member) => {
            log.trace(`Members length ${this.members.length()}`);

            member.once('end', (m) => {
                const $set = {_lastSession: m.sessionId, variables: m.variables, callSuccessful: m.callSuccessful},
                    $max = {
                        _nextTryTime: m.nextTime
                    };

                const update = {
                    $push: {_log: m._log},
                    $set,
                    $max
                };

                if (m._currentNumber) {
                    let communications = m._communications;
                    if (communications instanceof Array) {
                        for (let i = 0, len = communications.length; i < len; i++) {
                            if (i === m._currentNumber._id) {
                                $max[`communications.${i}.state`] = m._currentNumber.state;

                                $set[`communications.${i}._id`] = m._currentNumber._id;
                                $set[`communications.${i}._probe`] = m._currentNumber._probe;
                                $set[`communications.${i}._score`] = m._currentNumber._score;
                                $set[`communications.${i}.rangeId`] = m._currentNumber.rangeId;
                                $set[`communications.${i}.rangeAttempts`] = m._currentNumber.rangeAttempts;

                                $set[`communications.${i}.lastCall`] = m._minusProbe ? 0 : Date.now();

                                if (this._waitingForResultStatus) {
                                    if (m._minusProbe) {
                                        $set._waitingForResultStatusCb = null;
                                        $set._waitingForResultStatus = null;
                                        $set[`communications.${i}.checkResult`] = null;
                                    } else {
                                        update.$min = {
                                            _waitingForResultStatusCb: 1
                                        };
                                        update.$min[`communications.${i}.checkResult`] = 1;

                                        $max._waitingForResultStatus =  Date.now() + (this._wrapUpTime * 1000);
                                    }
                                }

                            } else {
                                // TODO option strategy X2 = set false
                                // Separate attempts for numbers with the same type

                                if (m._currentNumber.type === communications[i].type) {
                                    $set[`communications.${i}.rangeId`] = m._currentNumber.rangeId;
                                    $set[`communications.${i}.rangeAttempts`] = m._currentNumber.rangeAttempts;
                                }

                                if (m.endCause) {
                                    $set[`communications.${i}.state`] = MEMBER_STATE.End;
                                }
                            }
                        }
                    }
                    $set._lastNumberId = m._currentNumber._id;
                }

                if (m.endCause && !this._waitingForResultStatus) {
                    $set._endCause = m.endCause;
                }


                $set._lastMinusProbe = m._minusProbe;
                $set._lock = null;

                if (m._minusProbe) {
                    update.$inc = {_probeCount: -1}
                }

                // console.log(update);
                dialerService.members._updateByIdFix(
                    m._id,
                    update,
                    (err, res) => {
                        if (err)
                            log.error(err);

                    }
                );

                log.trace(`removed ${m.sessionId}`);
                if (!this.members.remove(m._id))
                    log.error(new Error(m));

                if (m.endCause) {
                    m.broadcast();
                }
            });

        });

        this.members.on('removed', (m) => {
            log.trace(`Members length ${this.members.length()}`);

            this.countMembers--;
            this.checkSleep();
            if (!this.isReady() || this.members.length() === 0)
                return this.tryStop();
        });
    }

    initChannel (cb) {
        this.channel = application.Broker.channel;
        this.channel.assertQueue(this.queueName, {autoDelete: true, durable: true, exclusive: false}, (err, qok) => {
            if (err)
                throw err; // TODO set log

            this.channel.consume(qok.queue, (msg) => {
                try {
                    this._huntingMember();
                } catch (e) {
                    log.error(e);
                }
            }, {noAck: true}, (e, res) => {
                if (e)
                    throw e; //TODO set log
                this.consumerTag = res.consumerTag;
                return cb(e)
            });
        });
    }

    closeChannel () {
        if (this.consumerTag) {
            this.channel.cancel(this.consumerTag)
        }
    }

    _setConfig (config) {

        this._stats = config.stats || {};
        if (this._stats.resources) {
            this._recources = this._stats.resources;
        }


        if (config.lastModified && config.lastModified.equals(this._lastModified)) {
            return;
        }

        this._lastModified = config.lastModified;

        this.resources = [];
        if (config.resources instanceof Array) {
            for (let res of config.resources) {
                const regexp = strToRegExp(res.dialedNumber);
                if (regexp)
                    this.resources.push({
                        dialedNumber: res.dialedNumber,
                        regexp: regexp,
                        destinations: res.destinations
                    });
            }
        }

        this.updateResources(this.resources);

        this._memberErrorCauses = config.causesError instanceof Array ? config.causesError : CODE_RESPONSE_ERRORS;
        this._memberMinusCauses = config.causesMinus instanceof Array ? config.causesMinus : CODE_RESPONSE_MINUS_PROBE;
        this._memberOKCauses = config.causesOK instanceof Array ? config.causesOK : CODE_RESPONSE_OK;
        this._memberRetryCauses = config.causesRetry instanceof Array ? config.causesRetry : CODE_RESPONSE_RETRY;

        this.communications = (config.communications && config.communications.types) instanceof Array
            ? config.communications.types
            : [];

        let parameters = (config && config.parameters) || {};
        [
            this._limit = 999,
            this._maxTryCount = 5,
            this._intervalTryCount = 5,
            this._minBillSec = 0,
            this._waitingForResultStatus = null,
            this._wrapUpTime = 60,
            this._originateTimeout = 60,
            this.lockId = `my best lock`,
            this._skills = [],
            this._recordSession = true,
            this._amd = {
                enabled: false
            },
            this._predictAdjust = 150,
            this._targetPredictiveSilentCalls = 2.5,
            this._maxPredictiveSilentCalls = 3
        ] = [
            parameters.limit,
            parameters.maxTryCount,
            parameters.intervalTryCount,
            parameters.minBillSec,
            parameters.waitingForResultStatus,
            parameters.wrapUpTime,
            parameters.originateTimeout,
            config.lockId,
            config.skills,
            parameters.recordSession,
            config.amd,
            parameters.predictAdjust,
            parameters.targetPredictiveSilentCalls,
            parameters.maxPredictiveSilentCalls
        ];

        if (this._amd.enabled) {
            const amdParams = [];
            if (this._amd.hasOwnProperty('silenceThreshold')) {
                amdParams.push(`silence_threshold=${this._amd.silenceThreshold}`);
            }
            if (this._amd.hasOwnProperty('maximumWordLength')) {
                amdParams.push(`maximum_word_length=${this._amd.maximumWordLength}`);
            }
            if (this._amd.hasOwnProperty('maximumNumberOfWords')) {
                amdParams.push(`maximum_number_of_words=${this._amd.maximumNumberOfWords}`);
            }
            if (this._amd.hasOwnProperty('betweenWordsSilence')) {
                amdParams.push(`between_words_silence=${this._amd.betweenWordsSilence}`);
            }
            if (this._amd.hasOwnProperty('minWordLength')) {
                amdParams.push(`min_word_length=${this._amd.minWordLength}`);
            }
            if (this._amd.hasOwnProperty('totalAnalysisTime')) {
                amdParams.push(`total_analysis_time=${this._amd.totalAnalysisTime}`);
            }
            if (this._amd.hasOwnProperty('afterGreetingSilence')) {
                amdParams.push(`after_greeting_silence=${this._amd.afterGreetingSilence}`);
            }
            if (this._amd.hasOwnProperty('greeting')) {
                amdParams.push(`greeting=${this._amd.greeting}`);
            }
            if (this._amd.hasOwnProperty('initialSilence')) {
                amdParams.push(`initial_silence=${this._amd.initialSilence}`);
            }

            this._amd._string = amdParams.join(' ');
        }
        this._broadcastPlaybackUri = config.playbackFile && config.playbackFile.uri;

        this.numberStrategy = config.numberStrategy || NUMBER_STRATEGY.BY_PRIORITY; // byPriority

        this.agentStrategy = config.agentStrategy;
        this.defaultAgentParams = config.agentParams || {};
        if (config.agents instanceof Array)
            this._agents = config.agents.map((i)=> `${i}@${this._domain}`);

        this._variables = config.variables || {};
        this._variables.domain_name = this._domain;
    }

    updateResources () {
        for (let resource of this.resources) {
            if (resource.destinations instanceof Array) {
                for (let dest of resource.destinations) {

                    if (dest.enabled !== true) continue;

                    let res = this.getResourceStat(dest.uuid);
                    if (!res) {
                        dest.active = 0;
                        dest.gwActive = 0;
                        dest.regexp = resource.dialedNumber;
                        this._recources[dest.uuid] = res = dest;
                    }
                }
            }
        }
    }

    getResourceStat (uuid) {
        return this._recources[uuid];
    }

    getFreeResourceRoutes () {
        const res = [];
        for (let key in this._recources) {
            if (this._recources[key].active < this._recources[key].limit && !~res.indexOf(this._recources[key].regexp))
                res.push(this._recources[key].regexp)
        }
        return res;
    }

    getResourceRoutes (cb) {
        const gws = new Map(),
            $or = [];



        gws.forEach( (g, name) => {
            $or.push({
                name: name
            })
        });

        application.DB.collection('gateway').find({$or}).toArray((e, res) => {
            if (e)
                return cb(e);

            for (let dbGw of res) {
                if (gws.has(dbGw.name))
                    gws.get(dbGw.name).resources.forEach( i => {
                        this.getResourceStat(i).gwActive = res.stats && res.stats.active
                    });
            }

            return cb(null);
        });

        //return res;
    }


    getAgentParam (paramName, agent = {}) {
        if (this.defaultAgentParams[paramName])
            return this.defaultAgentParams[paramName];

        return agent[paramName]
    }

    rollback (member, dest, stats, cb) {
        let $inc = {"stats.active": -1, "stats.callCount": 1};

        if (member) {
            if (member.callSuccessful) {
                $inc["stats.successCall"] = 1;
            } else {
                $inc["stats.errorCall"] = 1;
            }
        }

        if (dest && dest.uuid) {
            $inc[`stats.resource.${dest.uuid}`] = -1;
        }

        const update = {
        };


        if (stats instanceof Object) {
            update.$set = {};
            for (let key in stats) {
                if (key === 'predictAbandoned' || key === 'bridgedCall') {
                    if (stats[key] === true)
                        $inc[`stats.${key}`] = 1;
                } else {
                    update.$set[`stats.${key}`] = stats[key]
                }
            }
        }

        update.$inc = $inc;

        this._dbDialer.findAndModify(
            {_id: this._objectId, "stats.active": {$gt: 0}},
            {},
            update,
            {},
            (e, r) => {
                if (e)
                    log.error(e);
                if (r.lastErrorObject.n !== 1 || r.lastErrorObject.updatedExisting !== true)
                    throw r;
                this._active--;
                return cb && cb(e)
            }
        );
    }

    huntingMember () {
        if (!this.isReady())
            return;

        this.channel.sendToQueue(this.queueName, new Buffer(JSON.stringify({action: "call"})));
    }

    countAvailableMembers (limit = 1, cb) {
        // console.dir(this.getFilterAvailableMembers(), {depth: 10, colors: true});
        dialerService.members._aggregate([
                {$match: this.getFilterAvailableMembers()},
                {$limit: limit},
                {$count: "availableMembers"}
            ],
            (err, res) => {
                if (err)
                    return cb(err);

                return cb(err, (res[0] && res[0].availableMembers) || 0);
            }
        );
    }

    checkMember (cb) {

        const {codes, ranges, allCodes} = this.getCommunicationCodes();

        const skipCode = allCodes.filter( i => ~codes.indexOf(i) ? false : true );

        const patterns = [".*"];
        const regex = [];

        const $or = [];

        for (let p of patterns) {
            regex.push(new RegExp(p));
            $or.push({
                "communications.number": {$regex: p},
                "_p": p
            })
        }

        let codeFilter = [
            {
                "communications.type": null
            }
        ];

        for (let type of ranges) {
            codeFilter.push({
                $or: [
                    {
                        "communications.type": type.code,
                        $or: [
                            {
                                "communications.rangeAttempts": {
                                    "$lt": type.attempts || 0
                                }
                            },
                            {
                                "communications.rangeId": {
                                    "$ne": type.rangeId
                                }
                            }
                        ]
                    },
                    {
                        "communications.type": type.code,
                        "communications.rangeId": null
                    }
                ]
            })
        }

        const agg = [
            {
                $match: {
                    dialer: this._id,
                    _waitingForResultStatusCb: null,
                    _endCause: null,
                    _lock: null,
                    "communications.state": 0,
                    $and: [
                        {'$or': [ { _nextTryTime: { '$lte': Date.now()} }, { _nextTryTime: null }]},
                        {'$or': codeFilter}
                    ]
                }
            },

            {
                $sort: { "_nextTryTime": -1, "priority": -1, _id: -1}
            },

            {$limit: 1000},

            {$unwind: {path: "$communications",  includeArrayIndex: "_idx" }},

            {
                $match: {
                    "communications.state": 0,
                    "communications.type": {$nin: skipCode},
                    $or: codeFilter
                }
            },

            {
                $project: {
                    _nextTryTime: 1,
                    priority: 1,
                    communications: 1,
                    "isType": {
                        $switch: {
                            "branches": [
                                {
                                    case: {$ifNull: ["$communications.type", null]},
                                    then: 1
                                }
                            ],
                            default: 0
                        }
                    },
                    "lastCall": { $ifNull: [ "$communications.lastCall", -1 ] },
                    _idx: 1,
                    index: { $indexOfArray: [ [null].concat(codes), "$communications.type" ] }
                }
            },

            // {$sort: {"isType": -1, "communications.lastCall": 1, index: 1, "communications.priority": -1, "communications._probe": 1}}, //TOD-DOWN
            {$sort: {"isType": -1, index: -1, "communications.lastCall": 1,  "communications.priority": -1, "communications._probe": 1}}, //PRIORITY

            {
                $group: {
                    _id: "$_id",
                    communications: {$first: "$communications"},
                    _nextTryTime: {$first: "$_nextTryTime"},
                    priority: {$first: "$priority"},
                    _idx: {$first: "$_idx"}
                }
            },

            {$limit: 10},

            {$addFields: { "_p": patterns }},

            {$unwind: "$_p"},

            {$match: {$or: $or}},

            {
                $group: {
                    _id: "$_id",
                    communications: {$first: "$communications"},
                    _nextTryTime: {$first: "$_nextTryTime"},
                    priority: {$first: "$priority"},
                    _p: {$first: "$_p"},
                    _idx: {$first: "$_idx"}
                }
            },

            {$sort: { "_nextTryTime": -1, "priority": -1, "_id": -1}}
        ];

        console.dir(agg, {depth: 100});

        dialerService.members._aggregate(agg, (e, res = []) => {
            if (e)
                throw e;

            console.dir(res[0] && res[0].communications, {depth: 10, colors: true});

            return cb(e, res);
        })
    }

    getLimit () {
        return this._limit
    }

    _huntingMember () {
        if (!this.isReady())
            return;

        log.trace(`hunting on member ${this.nameDialer} - members queue: ${this._active} state: ${this.state}`);

        this._dbDialer.findAndModify(
            {_id: this._objectId, "stats.active": {$lt: this.getLimit()}},
            {},
            {
                $inc: {"stats.active": 1}
            },
            {new: true},
            (err, res) => {
                if (err)
                    return log.error(err);

                if (res.value) {
                    this._setConfig(res.value);
                    this._active = res.value.stats.active;

                    this.reserveMember((err, member, number, destination) => {
                        if (err) {
                            log.error(err);
                        }

                        if (!member) {
                            if (this.members.length() === 0)
                                this.tryStop();
                            this.rollback();
                            return log.debug (`Not found available members in ${this.nameDialer}`);
                        }

                        if (!this.isReady() || !destination || !number) {
                            this.rollback(null, destination);
                            return this.unReserveMember(member._id, (err) => {
                                if (err)
                                    return log.error(err);
                            });
                        }

                        let m = new Member(member, number, destination, this);
                        this.members.add(m._id, m);
                    });


                }
            }
        );
    }

    getCommunicationCodes () {
        const minOfDay = this._calendar.calcCurrentTimeOfDay(),
            codes = [],
            allCodes = [],
            ranges = [],
            date = new Date().getDate()
            ;

        for (let comm of this.communications) {
            allCodes.push(comm.code);
            for (let range of comm.ranges) {
                if (range.startTime <= minOfDay && range.endTime > minOfDay) {
                    codes.push(comm.code);
                    range.code = comm.code;
                    range.rangeId = `${date}_${range.startTime}_${range.endTime}`;
                    ranges.push(range);
                    break;
                }
            }
        }
        return {codes, ranges, allCodes};
    }

    getFilterAvailableMembers () {

        const {codes, ranges} = this.getCommunicationCodes();

        const communications = {
            $elemMatch: {
                state: {
                    $in: [MEMBER_STATE.Idle, null]
                },
                type: {
                    $in: [].concat(codes, null)
                }
            }
        };

        let codeFilter = [
            {
                type: {$nin: codes}
            }
        ];

        for (let type of ranges) {
            codeFilter.push({
                $or: [
                    {
                        "type": type.code,
                        $or: [
                            {
                                "rangeAttempts": {
                                    "$lt": type.attempts || 0
                                }
                            },
                            {
                                "rangeId": {
                                    "$ne": type.rangeId
                                }
                            }
                        ]
                    },
                    {
                        "type": type.code,
                        "rangeId": null
                    }
                ]
            })
        }

        if (codeFilter.length > 0) {
            communications.$elemMatch.$or = codeFilter;
        }

        if (this._lockedGateways && this._lockedGateways.length > 0)
            communications.$elemMatch.gatewayPositionMap = {
                $nin: this._lockedGateways
            };

        return {
            dialer: this._id,
            _waitingForResultStatusCb: null,
            _endCause: null,
            _lock: null,
            communications,
            $or: [{_nextTryTime: {$lte: Date.now()}}, {_nextTryTime: null}]
        };
    }

    getSortAvailableMembers () {
        return {
                "_nextTryTime": -1,
                "priority": -1,
                "_id": -1
        }
    }

    getMemberNumber (member, codes, ranges, allCodes) {
        if (member.communications instanceof Array) {

            const communicationsMap = member.communications.filter((i, key) => {
                if (i.state !== 0)
                    return false;

                const idx = codes.indexOf(i.type);

                if (~idx) {
                    i.isTypeFound = 1;
                    const rangeProperty = ranges[idx];
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

                i._id = key;
                if (!i._probe)
                    i._probe = 0;

                if (!i.lastCall)
                    i.lastCall = 0;

                return true;
            });

            let sort = {};

            if (this.numberStrategy === NUMBER_STRATEGY.TOP_DOWN) {
                sort = {
                    isTypeFound: "desc", //-
                    lastCall: "asc",
                    rangePriority: "desc",
                    priority: "desc",
                    _probe: "asc"
                };
            } else {
                sort = {
                    isTypeFound: "desc", //-
                    rangePriority: "desc",
                    lastCall: "asc",
                    priority: "desc",
                    _probe: "asc"
                };
            }

            return keySort(communicationsMap, sort)[0]
        }
    }

    trySetCallDestination (uuid, limit, cb) {
        const filterNull = {};
        filterNull[`stats.resource.${uuid}`] = null;
        const filterLimit = {};
        filterLimit[`stats.resource.${uuid}`] = {$lt: limit};

        const $inc = {};
        $inc[`stats.resource.${uuid}`] = 1;

        if(!this.isReady())
            return cb();
        this._dbDialer.findAndModify(
            {_id: this._objectId, $or: [filterNull, filterLimit], active: true},
            {},
            {$inc},
            {new: false}, //TODO
            cb
        )
    }

    detectNumberInDestinations (destinations, cb) {
        async.detectSeries(
            destinations,
            (dest, callback) => {
                const res = this.getResourceStat(dest.uuid);
                if (res && res.active < dest.limit) {
                    this.trySetCallDestination(res.uuid, dest.limit, (err, res) => {
                        if (err)
                            return callback(err);

                        if (!res || !res.value) {
                            callback(null);
                            return;
                        }

                        callback(null, true);
                    });

                    /* TODO check Max call in gw ?
                    if (dest.gwProto === 'sip') {
                        application.DB.collection('gateway').findOne({name: dest.gwName}, (err, gw) => {
                            if (err)
                                return callback(err);


                        });
                        this.trySetCallDestination(res.uuid, dest.limit, (err, res) => {
                            if (err)
                                return callback(err);

                            if (!res.value)
                                return callback();

                            return callback(null, dest);
                        });

                    } else {
                        this.trySetCallDestination(res.uuid, dest.limit, (err, res) => {
                            if (err)
                                return callback(err);

                            if (!res.value)
                                return callback();

                            return callback(null, dest);
                        });
                    }
                    */

                } else {
                    callback(null)
                }
            },
            cb
        )
    }

    detectNumberInRoutes (numberConfig = {}, cb) {
        let dest = null; //todo;
        async.detectSeries(
            this.resources,
            (resource, callback) => {
                if (resource.regexp.test(numberConfig.number)) {
                    this.detectNumberInDestinations(resource.destinations, (err, res) => {
                        if (err)
                            return callback(err);

                        if (!res)
                            return callback(null);

                        dest = res;
                        dest._regexp = resource.regexp;
                        callback(null, true)
                    });
                } else {
                    callback(null)
                }
            },
            (err, res) => {
                return cb(err, dest);
            }
        );
    }

    reserveMember (cb) {

        const regexp = this.getFreeResourceRoutes();
        if (regexp.length === 0)
            return cb(null, null);

        const $set = {
            _lock: this._instanceId
        };

        if (this._waitingForResultStatus) {
            $set._waitingForResultStatus = Date.now() + (this._wrapUpTime * 1000); //ERROR
            $set._waitingForResultStatusCb = 1;
            $set._maxTryCount = this._maxTryCount;
        }

        //console.dir(this.getFilterAvailableMembers(), {depth: 10, colors: true});

        let {codes, ranges, allCodes} = this.getCommunicationCodes();

        const filter = this.getFilterAvailableMembers();


        if (this.numberStrategy === NUMBER_STRATEGY.TOP_DOWN) {
            filter['$where'] = `function () {
            
                var homes = fnFilterDialerCommunications(
                    this.communications, ${JSON.stringify(codes)},
                    ${JSON.stringify(ranges)},
                    ${JSON.stringify(allCodes)}
                );
                var a = fnKeySort(homes, {
                    isTypeFound: "desc",
                    lastCall: "asc",
                    rangePriority: "desc",
                    priority: "desc",
                    _probe: "asc"
                })[0];
            
                printjson(a);
                return true
            }`;
        } else {
            filter['$where'] = `function () {
            
                var homes = fnFilterDialerCommunications(
                    this.communications, ${JSON.stringify(codes)},
                    ${JSON.stringify(ranges)},
                    ${JSON.stringify(allCodes)}
                );
                var a = fnKeySort(homes, {
                    isTypeFound: "desc",
                    rangePriority: "desc",
                    lastCall: "asc",
                    priority: "desc",
                    _probe: "asc"
                })[0];
            
                printjson(a);
                return true
            }`;
        }

        // console.dir(filter, {depth: 10, colors: true});

        dialerService.members._updateMember(
            filter,
            {
                $set,
                $inc: {_probeCount: 1},
                $currentDate: {lastModified: true}
            },
            this.getSortAvailableMembers(),
            (err, res) => {
                if (err)
                    return cb(err);

                if (!res && !res.value)
                    return cb(null, null);

                const number = this.getMemberNumber(res.value, codes, ranges, allCodes);
                if (!number)
                    return cb(null, res.value);

                this.detectNumberInRoutes(number, (err, destination) => {
                    if (err)
                        return cb(err);

                    return cb(null, res.value, number, destination);
                });
            }
        );
    }

    unReserveMember (id, cb) {
        dialerService.members._updateById(
            id,
            {$set: {_lock: null}},
            cb
        )
    }

    getCountAndNextTryTime (cb) {
        dialerService.members._aggregate(
            [
                {$match: {dialer: this._id, _endCause: null, "communications.state": 0}},
                {
                    $group: {
                        _id: '',
                        nextTryTime: {
                            $min: "$_nextTryTime"
                        },
                        count: {
                            $sum: 1
                        }
                    }
                }
            ],
            (err, res) => {
                if (err)
                        return cb(err);

                return cb(null, (res && res[0]) || {});
            }
        )
    }

    reCalcCalendar () {
        this._calendar.reCalc();
        let calendar = this._calendar;
        
        if (calendar.expire) {
            this.state = DIALER_STATES.End;
            this.cause = DIALER_CAUSE.ProcessExpire;
        } else if (calendar.sleepTime > 0) {
            this.state = DIALER_STATES.Sleep;
            this.cause = DIALER_CAUSE.ProcessSleep;
        } else if (calendar.deadLineTime > 0) {
            this.state = DIALER_STATES.Work;
            this.cause = DIALER_CAUSE.ProcessReady;
        }
    }

    checkSleep () {
        if (this._calendar.expire) {
            this.cause = DIALER_CAUSE.ProcessExpire;
            this.emit('end', this);
            return;
        }
        if (Date.now() >= this._calendar.deadLineTime) {
            if (this.state !== DIALER_STATES.Sleep) {
                this.reCalcCalendar();
                this.cause = DIALER_CAUSE.ProcessSleep;
                this.setState(DIALER_STATES.Sleep)
            }
        }
        if (this.state === DIALER_STATES.Sleep) {
            this.closeNoChannelsMembers(DIALER_STATES.Sleep);
            if (this.members.length() === 0) {
                this.emit('sleep', this);
                this.emit('end', this);
            }
            return true;
        }
        return false;
    }

    isReady () {
        return this.state === DIALER_STATES.Work;
    }

    isError () {
        return this.state === DIALER_STATES.Error;
    }

    toJson () {
        return {
            "members": this.members.length(),
            "state": this.state
        }
    }

    setState (state) {
        this.state = state;

        if (this.isError()) {
            let ms = this.members.getKeys();
            ms.forEach((key) => {
                let m = this.members.get(key);
                //TODO
                m.removeAllListeners();
                this.members.remove(key)
            });

            this.emit('error', this);
            return;
        }

        if (state === DIALER_STATES.ProcessStop) {
            if (this.members.length() === 0) {
                this.cause = DIALER_CAUSE.ProcessStop;
                this.emit('end', this)
            } else {
                this.closeNoChannelsMembers(DIALER_STATES.ProcessStop);
            }
        }
    }

    closeNoChannelsMembers (cause) {
        let mKeys = this.members.getKeys();
        for (let key of mKeys) {
            let m = this.members.get(key);
            // TODO error...
            if (m && m.channelsCount === 0) {
                if (m.currentProbe > 0) {
                    m.minusProbe();
                }
                m.log(`Stop dialer cause ${cause || 'empty'}`);
                m.end();
            }
        }
    }

    tryStop () {
        console.log('state', this.state, this.members.length());

        if (this.isError()) {
            log.warn(`Force stop process.`);
            return;
        }

        if (this.state === DIALER_STATES.ProcessStop) {
            if (this.members.length() != 0)
                return;

            log.info('Stop dialer');

            this.cause = DIALER_CAUSE.ProcessStop;
            this.active = false;
            this.emit('end', this);
            return
        }

        if (this.state === DIALER_STATES.Sleep) {
            return
        }

        if (this._processTryStop)
            return;

        this._processTryStop = true;
        console.log('Try END -------------');

        this.getCountAndNextTryTime((err, res) => {
            if (err)
                return log.error(err);

            if (!res && this.members.length() === 0) {
                this.cause = DIALER_CAUSE.ProcessNotFoundMember;
                this.setState(DIALER_STATES.End);
                this.emit('end', this);
                return log.info(`STOP DIALER ${this.name}`);
            }

            if (!res)
                return;

            log.trace(`Status ${this.nameDialer} : state - ${this.state}; count - ${res.count || 0}; nextTryTime - ${res.nextTryTime}`);

            if (!res.count || res.count === 0) {
                this.cause = DIALER_CAUSE.ProcessComplete;
                this.setState(DIALER_STATES.End);
                this.emit('end', this);
                return log.info(`STOP DIALER ${this.name}`);
            }
            this.countMembers = res.count;

            this._processTryStop = false;
            if (!res.nextTryTime) {
                res.nextTryTime = Date.now() + 1000;
            } else if (res.nextTryTime < Date.now()) {
                return this.emit('wakeUp')
            }

            if (res.nextTryTime > 0) {
                let nextTime = res.nextTryTime - Date.now();
                if (nextTime < 1)
                    nextTime = 1000;

                if (nextTime > 2147483647)
                    nextTime = 2147483647;

                console.log(nextTime);

                this._timerId = setTimeout(() => {
                    clearTimeout(this._timerId);
                    this.emit('wakeUp')
                }, nextTime);
            }

        });
    }
    
    setReady () {
        this.getCountAndNextTryTime((err, {count = 0, nextTryTime = 0}) => {
            if (err) {
                log.error(err);
                this.cause = `${err.message}`;
                this.emit('end', this);
                return;
            }

            if (count === 0) {
                this.cause = DIALER_CAUSE.ProcessNotFoundMember;
                this.state = DIALER_STATES.End;
                this.emit('end', this);
                return;
            }

            this.reCalcCalendar();
            this.checkSleep();

            if (this.state === DIALER_STATES.Work) {
                this.countMembers = count;
                this.initChannel((e, res) => {
                    if (e) {
                        this.cause = `${err.message}`;
                        return this.emit('end', this);
                    }
                    this.emit('ready', this);
                    log.trace(`found in ${this.nameDialer} ${count} members. run hunting...`);
                });
            }
        });
    }
};




const keySort = function(arr = [], keys) {

    keys = keys || {};

    const sortFn = function(a, b) {
        let sorted = 0, ix = 0;

        while (sorted === 0 && ix < KL) {
            let k = obIx(keys, ix);
            if (k) {
                let dir = keys[k];
                sorted = _keySort(a[k], b[k], dir);
                ix++;
            }
        }
        return sorted;
    };

    const obIx = function(obj, ix){
        return Object.keys(obj)[ix];
    };

    const _keySort = function(a, b, d) {
        d = d !== null ? d : 1;
        // a = a.toLowerCase(); // this breaks numbers
        // b = b.toLowerCase();
        if (a == b)
            return 0;
        return a > b ? 1 * d : -1 * d;
    };

    const KL = Object.keys(keys).length;

    if (!KL)
        return arr.sort(sortFn);

    for ( let k in keys) {
        // asc unless desc or skip
        keys[k] =
            keys[k] == 'desc' || keys[k] == -1  ? -1
                : (keys[k] == 'skip' || keys[k] === 0 ? 0
                : 1);
    }
    arr = arr.sort(sortFn);
    return arr;
};

function strToRegExp(str = "") {
    try {
        return new RegExp(str);
    } catch (e) {
        return null;
    }
}