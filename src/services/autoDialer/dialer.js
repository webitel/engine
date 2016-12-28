/**
 * Created by igor on 25.05.16.
 */

let DIALER_STATES = require('./const').DIALER_STATES,
    DIALER_CAUSE = require('./const').DIALER_CAUSE,
    MEMBER_STATE = require('./const').MEMBER_STATE,
    END_CAUSE = require('./const').END_CAUSE,

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

    checkSkill (skills) {
        return this._skillsReg.test(skills)
    }

    constructor (type, config, calendarConfig) {
        super();
        this.type = type;
        this._id = config._id.toString();
        this._objectId = config._id;

        // this.bigData = new Array(1e6).join('XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX\n');

        this.nameDialer = config.name;
        this.number = config.number || this.nameDialer;

        this._domain = config.domain;
        this.state = DIALER_STATES.Idle;
        this.cause = DIALER_CAUSE.Init;

        this._calendar = new Calendar(calendarConfig, config.communications);

        this.once('end', () => {
            this._calendar.stop();
        });

        this._memberErrorCauses = config.causesError instanceof Array ? config.causesError : CODE_RESPONSE_ERRORS;
        this._memberMinusCauses = config.causesMinus instanceof Array ? config.causesMinus : CODE_RESPONSE_MINUS_PROBE;
        this._memberOKCauses = config.causesOK instanceof Array ? config.causesOK : CODE_RESPONSE_OK;
        this._memberRetryCauses = config.causesRetry instanceof Array ? config.causesRetry : CODE_RESPONSE_RETRY;

        this.countMembers = 0;
        this._countRequestHunting = 0;

        let parameters = (config && config.parameters) || {};
        [
            this._limit = 999,
            this._maxTryCount = 5,
            this._intervalTryCount = 5,
            this._minBillSec = 0,
            this._waitingForResultStatus = false,
            this._wrapUpTime = 60,
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
        
        this._skillsReg = this._skills.length > 0 ? new RegExp('\\b' + this._skills.join('\\b|\\b') + '\\b', 'i') : /.*/;


        this._variables = config.variables || {};
        this._variables.domain_name = this._domain;

        log.debug(`
            Init dialer: ${this.nameDialer}@${this._domain}
            Config:
                type: ${this.type},
                limit: ${this._limit},
                minBillSec: ${this._minBillSec},
                maxTryCount: ${this._maxTryCount},
                intervalTryCount: ${this._intervalTryCount},
                deadLine: ${new Date(this._calendar.deadLineTime)}
        `);

        this.members = new Collection('id');

        // this.membersQueue = async.queue((member, cb) => {
        //     member.once('end', cb);
        //     this.dialMember(member);
        // }, this._limit);
        //
        // this.membersQueue.drain = (e) => {
        //     console.log('drain', e);
        // };

        this.members.on('added', (member) => {
            log.trace(`Members length ${this.members.length()}`);
            // this.membersQueue.push(member, () => {
            //     let $set = {_nextTryTime: member.nextTime, _lastSession: member.sessionId, _endCause: member.endCause, variables: member.variables, _probeCount: member.currentProbe};
            //     if (member._currentNumber)
            //         $set[`communications.${member._currentNumber._id}`] = member._currentNumber;
            //
            //     dialerService.members._updateById(
            //         member._id,
            //         {
            //             $push: {_log: member._log},
            //             $set,
            //             $unset: {_lock: 1}//, $inc: {_probeCount: 1}
            //         },
            //         (err) => {
            //             if (err)
            //                 return log.error(err);
            //
            //             log.trace(`removed ${member.sessionId}`);
            //             if (!this.members.remove(member._id))
            //                 log.error(new Error(member));
            //         }
            //     );
            // });
            // Close member session
            member.once('end', (m) => {
                let $set = {_nextTryTime: m.nextTime, _lastSession: m.sessionId, variables: m.variables, callSuccessful: m.callSuccessful},
                    $max = {
                        _probeCount: m.currentProbe
                    };

                if (m._currentNumber) {
                    let communications = m._communications;
                    if (communications instanceof Array) {
                        for (let i = 0, len = communications.length; i < len; i++) {
                            if (i == m._currentNumber._id) {
                                $max[`communications.${i}.state`] = m._currentNumber.state;
                                $set[`communications.${i}._id`] = m._currentNumber._id;
                                $set[`communications.${i}._probe`] = m._currentNumber._probe;
                                $set[`communications.${i}._score`] = m._currentNumber._score;
                                $set[`communications.${i}._range`] = m._currentNumber._range;
                            } else {
                                if (m.endCause) {
                                    $set[`communications.${i}.state`] = MEMBER_STATE.End;
                                }
                            }
                        }
                    }
                    $set._lastNumberId = m._currentNumber._id;
                }

                if (m.endCause) {
                    $set._endCause = m.endCause;
                }

                if (m._minusProbe && this._waitingForResultStatus)
                    $set._waitingForResultStatus = false;

                $set._lastMinusProbe = m._minusProbe;
                $set._lock = null;

                dialerService.members._updateByIdFix(
                    m._id,
                    {
                        $push: {_log: m._log},
                        $set,
                        $max
                        // $unset: {_lock: 1}//, $inc: {_probeCount: 1}
                    },
                    (err) => {
                        if (err)
                            log.error(err);

                        log.trace(`removed ${m.sessionId}`);
                        if (!this.members.remove(m._id))
                            log.error(new Error(m));
                    }
                );

                if (m.endCause) {
                    m.broadcast();
                }
            });

            if (member.checkExpire()) {
                member.endCause = END_CAUSE.MEMBER_EXPIRED;
                member.end(END_CAUSE.MEMBER_EXPIRED)
            } else {
                if (member._currentNumber) {
                    this.dialMember(member)
                } else {
                    member.end();
                }
            }
        });

        this.members.on('removed', () => {
            log.trace(`Members length ${this.members.length()}`);
            this.countMembers--;
            this.checkSleep();
            if (!this.isReady() || this.members.length() === 0)
                return this.tryStop();

            if (!this.checkLimit())
                this.huntingMember();
        });
    }

    addMemberCallbackQueue (m, e, wrapTime) {
        log.trace(`End channels ${m.sessionId}`);

        if (this._waitingForResultStatus) {
            m.log('check callback');
            m.nextTrySec += (this._wrapUpTime || 0);
            m.end(null, e);
        } else {
            m.end(e.getHeader('variable_hangup_cause'), e);
        }
    }

    huntingMember () {

        if (this.checkLimit())
            return;

        if (!this.isReady())
            return;

        log.trace(`hunting on member ${this.nameDialer} - members queue: ${this.members.length()} state: ${this.state}`);

        this._countRequestHunting++;
        this.reserveMember((err, res) => {
            this._countRequestHunting--;

            if (err)
                return log.error(err);

            if (!res || !res.value) {
                if (this.members.length() === 0)
                    this.tryStop();
                return log.debug (`Not found members in ${this.nameDialer}`);
            }

            if (!this.isReady()) {
                return this.unReserveMember(res.value._id, (err) => {
                    if (err)
                        return log.error(err);
                });
            }
            
            if (this.members.existsKey(res.value._id)) {
                return log.warn(`Member in queue ${this.nameDialer} : ${res.value._id}`);
            }

            let option = {
                maxTryCount: this._maxTryCount,
                intervalTryCount: this._intervalTryCount,
                lockedGateways: this._lockedGateways,
                queueId: this._id,
                queueName: this.nameDialer,
                queueNumber: this.number,
                minCallDuration: this._minBillSec,
                domain: this._domain,
                _waitingForResultStatus: this._waitingForResultStatus,

                causesOK: this._memberOKCauses,
                causesRetry: this._memberRetryCauses,
                causesError: this._memberErrorCauses,
                causesMinus: this._memberMinusCauses
            };
            // TODO remove options
            let m = new Member(res.value, option, this);
            this.members.add(m._id, m);

            if (!this.checkLimit()) {
                this.huntingMember();
            }
        });
    }

    reserveMember (cb) {
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
                                "_range.attempts": {
                                    "$lt": type.range.attempts
                                }
                            },
                            {
                                "_range.rangeId": {
                                    "$ne": type.rangeId
                                }
                            }
                        ]
                    },
                    {
                        "type": type.code,
                        "_range": null
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

        const filter = {
            dialer: this._id,
            _endCause: null,
            _lock: null,
            communications,
            $or: [{_nextTryTime: {$lte: Date.now()}}, {_nextTryTime: null}]
        };

        const $set = {
            _lock: this.lockId
        };

        if (this._waitingForResultStatus) {
            $set._waitingForResultStatus = true;
            $set._maxTryCount = this._maxTryCount;
        }

        //console.dir(filter, {depth: 10, colors: true});

        dialerService.members._updateMember(
            filter,
            {
                $set
            },
            {sort: [["_nextTryTime", -1],["priority", -1], ["_id", -1]]},
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
                {$match: {dialer: this._id, _lock: null, _endCause: null, "communications.state": 0}},
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

    checkLimit () {
        return (this._countRequestHunting + this.members.length() >= this._limit || this.members.length()  >= this._limit);
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
                m.removeAllListeners();
                if (typeof m.offEslEvent == 'function') {
                    m.offEslEvent();
                }
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

        if (this._processTryStop || this.checkLimit())
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
                    this.huntingMember()
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
                this.emit('ready', this);
                log.trace(`found in ${this.nameDialer} ${count} members. run hunting...`);
                this.huntingMember();
            }
        });
    }
};