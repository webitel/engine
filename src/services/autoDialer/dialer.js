/**
 * Created by igor on 25.05.16.
 */

let DIALER_STATES = require('./const').DIALER_STATES,
    DIALER_CAUSE = require('./const').DIALER_CAUSE,
    MEMBER_STATE = require('./const').MEMBER_STATE,
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

        this._calendar = new Calendar(calendarConfig);

        this.countMembers = 0;
        this._countRequestHunting = 0;

        let parameters = (config && config.parameters) || {};
        [
            this._limit = 999,
            this._maxTryCount = 5,
            this._intervalTryCount = 5,
            this._minBillSec = 0,
            this.lockId = `my best lock`,
            this._skills = []
        ] = [
            parameters.limit,
            parameters.maxTryCount,
            parameters.intervalTryCount,
            parameters.minBillSec,
            config.lockId,
            config.skills
        ];
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
                let $set = {_nextTryTime: m.nextTime, _lastSession: m.sessionId, _endCause: m.endCause, variables: m.variables, _probeCount: m.currentProbe};
                if (m._currentNumber)
                    $set[`communications.${m._currentNumber._id}`] = m._currentNumber;

                dialerService.members._updateById(
                    m._id,
                    {
                        $push: {_log: m._log},
                        $set,
                        $unset: {_lock: 1}//, $inc: {_probeCount: 1}
                    },
                    (err) => {
                        if (err)
                            return log.error(err);

                        log.trace(`removed ${m.sessionId}`);
                        if (!this.members.remove(m._id))
                            log.error(new Error(m));
                    }
                )
            });
            this.dialMember(member);
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
                domain: this._domain
            };
            let m = new Member(res.value, option);
            this.members.add(m._id, m);

            if (!this.checkLimit()) {
                this.huntingMember();
            }
        });
    }

    reserveMember (cb) {
        let communications = {
            $elemMatch: {
                $or: [{state: MEMBER_STATE.Idle}, {state: null}]
            }
        };

        if (this._lockedGateways && this._lockedGateways.length > 0)
            communications.$elemMatch.gatewayPositionMap = {
                $nin: this._lockedGateways
            };

        let filter = {
            dialer: this._id,
            _endCause: null,
            _lock: null,
            communications,
            $or: [{_nextTryTime: {$lte: Date.now()}}, {_nextTryTime: null}]
        };

        dialerService.members._updateMember(
            filter,
            {
                $set: {_lock: this.lockId}
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
            this.closeNoChannelsMembers();
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
                this.closeNoChannelsMembers();
            }
        }
    }

    closeNoChannelsMembers () {
        let mKeys = this.members.getKeys();
        for (let key of mKeys) {
            let m = this.members.get(key);
            // TODO error...
            if (m && m.channelsCount === 0) {
                if (m.currentProbe > 0)
                    m.minusProbe();
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