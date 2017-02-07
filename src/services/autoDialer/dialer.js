/**
 * Created by igor on 25.05.16.
 */

let DIALER_STATES = require('./const').DIALER_STATES,
    DIALER_CAUSE = require('./const').DIALER_CAUSE,
    MEMBER_STATE = require('./const').MEMBER_STATE,

    CODE_RESPONSE_ERRORS = require('./const').CODE_RESPONSE_ERRORS,
    CODE_RESPONSE_RETRY = require('./const').CODE_RESPONSE_RETRY,
    CODE_RESPONSE_OK = require('./const').CODE_RESPONSE_OK,
    CODE_RESPONSE_MINUS_PROBE = require('./const').CODE_RESPONSE_MINUS_PROBE,

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

        this.consumerTag = null;
        this.queueName = `engine.dialer.${this._id}`;

        this._dbDialer = application.DB.collection('dialer');

        this._dbDialer.update({_id: this._objectId, "stats.active": null}, {
            $currentDate: {lastModified: true},
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

                                if (this._waitingForResultStatus) {
                                    if (m._minusProbe) {
                                        $set._waitingForResultStatusCb = null;
                                        $set._waitingForResultStatus = null;
                                    } else {
                                        update.$min = {
                                            _waitingForResultStatusCb: 1
                                        };
                                        $max._waitingForResultStatus =  Date.now() + (this._wrapUpTime * 1000);
                                        $set[`communications.${i}.checkResult`] = 1
                                    }
                                }

                            } else {
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
        this._memberErrorCauses = config.causesError instanceof Array ? config.causesError : CODE_RESPONSE_ERRORS;
        this._memberMinusCauses = config.causesMinus instanceof Array ? config.causesMinus : CODE_RESPONSE_MINUS_PROBE;
        this._memberOKCauses = config.causesOK instanceof Array ? config.causesOK : CODE_RESPONSE_OK;
        this._memberRetryCauses = config.causesRetry instanceof Array ? config.causesRetry : CODE_RESPONSE_RETRY;

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
            }
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
            config.amd
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

        this.agentStrategy = config.agentStrategy;
        this.defaultAgentParams = config.agentParams || {};
        if (config.agents instanceof Array)
            this._agents = config.agents.map( (i)=> `${i}@${this._domain}`);

        this._variables = config.variables || {};
        this._variables.domain_name = this._domain;
    }

    getAgentParam (paramName, agent = {}) {
        if (this.defaultAgentParams[paramName])
            return this.defaultAgentParams[paramName];

        return agent[paramName]
    }

    rollback (params = {}, cb) {
        let $inc = {"stats.active": -1, "stats.callCount": 1};

        if (params.callSuccessful) {
            $inc["stats.successCall"] = 1;
        } else {
            $inc["stats.errorCall"] = 1;
        }

        this._dbDialer.findAndModify(
            {_id: this._objectId, "stats.active": {$gt: 0}},
            {},
            {
                $currentDate: {lastModified: true},
                $inc
            },
            {},
            (e, r) => {
                if (e)
                    log.error(e);
                if (r.lastErrorObject.n !== 1 || r.lastErrorObject.updatedExisting !== true)
                    throw r;
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

    _huntingMember () {
        if (!this.isReady())
            return;

        log.trace(`hunting on member ${this.nameDialer} - members queue: ${this._active} state: ${this.state}`);

        this._dbDialer.findAndModify(
            {_id: this._objectId, "stats.active": {$lt: this._limit}},
            {},
            {
                $currentDate: {lastModified: true},
                $inc: {"stats.active": 1}
            },
            {new: true},
            (err, res) => {
                if (err)
                    return log.error(err);

                if (res.value) {
                    this._setConfig(res.value);
                    this._active = res.value.stats.active;

                    this.reserveMember((err, member) => {
                        if (err) {
                            log.error(err);
                            return this.rollback();
                        }

                        if (!member || !member.value) {
                            if (this.members.length() === 0)
                                this.tryStop();
                            this.rollback();
                            return log.debug (`Not found members in ${this.nameDialer}`);
                        }

                        if (!this.isReady()) {
                            this.rollback();
                            return this.unReserveMember(member.value._id, (err) => {
                                if (err)
                                    return log.error(err);
                            });
                        }
                        // this.unReserveMember(member.value._id, (err) => {
                        //     if (err)
                        //         return log.error(err);
                        // });
                        // return this.rollback({}, () => this.huntingMember());

                        let m = new Member(member.value, this);
                        this.members.add(m._id, m);
                    });
                }
            }
        );
    }

    getFilterAvailableMembers () {
        const communications = {
            $elemMatch: {
                state: {
                    $in: [MEMBER_STATE.Idle, null]
                },
                type: {
                    $in: [].concat(this._calendar.getCommunicationCodes(), null)
                }
            }
        };

        let codeFilter = [
            {
                type: {$nin: this._calendar.getCommunicationCodes()}
            }
        ];

        for (let type of this._calendar._currentCommunicationsRanges) {
            codeFilter.push({
                $or: [
                    {
                        "type": type.code,
                        $or: [
                            {
                                "rangeAttempts": {
                                    "$lt": type.range.attempts
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
        return [
                ["_nextTryTime", -1],
                ["priority", -1],
                ["_id", -1]
            ]
    }

    reserveMember (cb) {
        const $set = {
            _lock: this._instanceId
        };

        if (this._waitingForResultStatus) {
            $set._waitingForResultStatus = Date.now() + (this._wrapUpTime * 1000);
            $set._waitingForResultStatusCb = 1;
            $set._maxTryCount = this._maxTryCount;
        }

        // console.dir(this.getFilterAvailableMembers(), {depth: 10, colors: true});

        dialerService.members._updateMember(
            this.getFilterAvailableMembers(),
            {
                $set,
                $inc: {_probeCount: 1},
                $currentDate: {lastModified: true}
            },
            {sort: this.getSortAvailableMembers()},
            cb
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
            if (!res.nextTryTime) res.nextTryTime = Date.now() + 1000;
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