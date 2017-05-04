/**
 * Created by igor on 24.05.16.
 */

'use strict';

let EventEmitter2 = require('eventemitter2').EventEmitter2,
    log = require(__appRoot + '/lib/log')(module),
    AgentManager = require('./agentManager'),
    DIALER_TYPES = require('./const').DIALER_TYPES,
    DIALER_STATES = require('./const').DIALER_STATES,
    DIALER_CAUSE = require('./const').DIALER_CAUSE,
    AGENT_STATE = require('./const').AGENT_STATE,
    AGENT_STATUS = require('./const').AGENT_STATUS,
    END_CAUSE = require('./const').END_CAUSE,
    Collection = require(__appRoot + '/lib/collection'),
    VoiceDialer = require('./voice'),
    ProgressiveDialer = require('./progressive'),
    PredictiveDialer = require('./predictive'),
    eventsService = require(__appRoot + '/services/events'),
    dialerService = require(__appRoot + '/services/dialer'),
    encodeRK = require(__appRoot + '/utils/helper').encodeRK,
    async = require('async'),
    calendarManager = require('./calendarManager'),
    Scheduler = require(__appRoot + '/lib/scheduler'),
    conf = require(__appRoot + '/conf'),
    dialerCbMinusAttempt = `${conf.get('application:dialerCbMinusAttempt')}` === 'true'
    ;

const EVENT_CHANGE_STATE = `DC::CHANGE_STATE`;
eventsService.registered(EVENT_CHANGE_STATE);

class AutoDialer extends EventEmitter2 {

    constructor (app) {
        super();
        this._app = app;
        this.id = 'lock id';
        this.connectDb = false;
        this.connectWConsole = false;
        this.connectFs = false;
        this.connectBroker = false;

        this.activeDialer = new Collection('id');
        this.agentManager = new AgentManager();

        this.agentManager.on('unReserveHookAgent', this.sendAgentToDialer.bind(this));

        log.debug('Init AutoDialer');

        this.on(`changeConnection`, (e) => {
            if (this.isReady()) {
                this.loadCampaign();
            } else {
                this.forceStop();
            }
        });

        app.on('sys::connectDb', this.onConnectDb.bind(this));
        app.on('sys::reconnectDb', this.onConnectDb.bind(this));

        app.on('sys::wConsoleConnect', this.onConnectWConsole.bind(this));
        app.on('sys::wConsoleConnectError', this.onConnectWConsoleError.bind(this));

        app.on('sys::connectDbError', this.onConnectDbError.bind(this));
        app.on('sys::closeDb', this.onConnectDbError.bind(this));

        app.on('sys::connectFsApi', this.onConnectFs.bind(this));
        app.on('sys::errorConnectFsApi', this.onConnectFsError.bind(this));

        app.Broker.on('error:close', this.onConnectBrokerError.bind(this));
        app.Broker.on('init:broker', this.onConnectBrokerSuccessful.bind(this));

        this.activeDialer.on('added', (dialer) => {

            dialer.on('ready', (d) => {
                log.debug(`Ready dialer ${d.nameDialer} - ${d._id}`);
                d.state = DIALER_STATES.Work;
                this.dbDialer._dialerCollection.findOneAndUpdate(
                    {_id: d._objectId},
                    {
                        $set: {state: d.state, _cause: d.cause, active: true, nextTick: null, "stats.readyOn": Date.now(), "stats.stopOn": null},
                        $addToSet: {"stats.process": this._app._instanceId}
                    },
                    e => {
                        if (e)
                            log.error(e);
                        this.sendEvent(d, true, 'ready');
                    }
                );
            });

            dialer.once('end', (d) => {
                log.debug(`End dialer ${d.nameDialer} - ${d._id} - ${d.cause}`);

                this.dbDialer._dialerCollection.findOneAndUpdate(
                    {_id: d._objectId},
                    {
                        $set: {state: d.state, _cause: d.cause, active: d.state === DIALER_STATES.Sleep, "stats.readyOn": null, "stats.stopOn": Date.now()},
                        $pull: {"stats.process": this._app._instanceId}
                    },
                    e => {
                        if (e)
                            log.error(e);
                        this.activeDialer.remove(dialer._id);
                    }
                );

                this.sendEvent(d, d.state === DIALER_STATES.Sleep, 'end');
            });

            dialer.on('error', (d) => {
                log.warn(`remove dialer ${d.nameDialer}`);
                this.activeDialer.remove(d._id);
            });

            if (dialer.type === DIALER_TYPES.PredictiveDialer || dialer.type === DIALER_TYPES.ProgressiveDialer) {

                this.agentManager.initAgents(dialer, (err, res) => {
                    if (err)
                        return log.error(err);

                    dialer.setReady();
                });

            } else {
                dialer.setReady();
            }
        });

        this.activeDialer.on('removed', (dialer) => {

            log.info(`Remove active dialer ${dialer.nameDialer} : ${dialer._id} - ${dialer.cause}`);
            this.sendEvent(dialer, dialer.state === DIALER_STATES.Sleep, 'removed');
        });

        this.on(`changeDialerState`, (dialer, calendar, currentTime) => {

            if (currentTime.expire) {
                log.debug(`Set dialer ${dialer._id} expire`);
                const d = this.activeDialer.get(dialer._id.toString());
                if (d) {
                    d.setState(DIALER_STATES.End);
                    d.cause = DIALER_CAUSE.ProcessExpire;
                }
                this.dbDialer._dialerCollection.findOneAndUpdate(
                    {_id: dialer._id},
                    {
                        $set: {
                            active: false,
                            state: DIALER_STATES.End,
                            _cause: DIALER_CAUSE.ProcessExpire,
                            "stats.minuteOfDay": null
                        }
                    },
                    e => {
                        if (e)
                            log.error(e);
                    }
                );
            } else if ((dialer.state === DIALER_STATES.Work || dialer.state === DIALER_STATES.Idle) && currentTime.currentTimeOfDay === null) {
                log.debug(`Set dialer ${dialer._id} sleep`);
                const d = this.activeDialer.get(dialer._id.toString());
                if (d) {
                    d.setState(DIALER_STATES.Sleep);
                    d.cause = DIALER_CAUSE.ProcessSleep;
                }
                this.dbDialer._dialerCollection.findOneAndUpdate(
                    {_id: dialer._id},
                    {
                        $set: {
                            active: true,
                            state: DIALER_STATES.Sleep,
                            _cause: DIALER_CAUSE.ProcessSleep,
                            "stats.minuteOfDay": currentTime.currentTimeOfDay,
                            "stats.weekOfDay": currentTime.currentWeek
                        }
                    },
                    e => {
                        if (e)
                            log.error(e);
                    }
                );

            } else if (dialer.state === DIALER_STATES.Sleep && currentTime.currentTimeOfDay !== null) {
                log.debug(`Set dialer ${dialer._id} ready`);

                this.dbDialer._dialerCollection.findOneAndUpdate(
                    {_id: dialer._id},
                    {
                        $set: {
                            state: DIALER_STATES.Idle,
                            _cause: DIALER_CAUSE.Init,
                            "stats.minuteOfDay": currentTime.currentTimeOfDay,
                            "stats.weekOfDay": currentTime.currentWeek
                        }
                    },
                    e => {
                        if (e)
                            log.error(e);

                        this.runDialerById(dialer._id, dialer.domain, (err) => {
                            if (err)
                                log.error(err);
                        });
                    }
                );
            } else if (currentTime.currentTimeOfDay !== null) {
                log.debug(`Set dialer ${dialer._id} time of day ${currentTime.currentTimeOfDay}`);

                this.dbDialer._dialerCollection.findOneAndUpdate(
                    {_id: dialer._id},
                    {
                        $set: {
                            "stats.minuteOfDay": currentTime.currentTimeOfDay,
                            "stats.weekOfDay": currentTime.currentWeek
                        }
                    },
                    e => {
                        if (e)
                            log.error(e);
                    }
                );
            }

            if (dialer.stats && dialer.stats.weekOfDay !== currentTime.currentWeek) {
                this.dbDialer._dialerCollection.findOneAndUpdate(
                    {_id: dialer._id},
                    {
                        $set: {
                            "stats.weekOfDay": currentTime.currentWeek
                        }
                    },
                    e => {
                        if (e)
                            log.error(e);

                        this.agentManager.resetAgentsStats(dialer._id, e => {
                            if (e)
                                log.error(e);
                        })
                    }
                );
            }

        });

        const checkSetAvailableTimeSec = (cb) => {
            if (!this.isReady()) {
                return cb();
            }

            this.agentManager.checkSetAvailableTime(e => {
                if (e)
                    log.error(e);

                return cb();
            });
        };

        const clearAttemptOnDeadlineResultSec = dialerCbMinusAttempt
            ? (cb) => {
                if (!this.isReady()) {
                    return cb();
                }

                this.clearAttemptOnDeadlineResultStatus(e => {
                    if (e)
                        log.error(e);

                    return cb();
                });
              }
            : (cb) => {
                if (!this.isReady()) {
                    return cb();
                }

                this.clearAttemptOnDeadlineResultStatusAddAttempts(e => {
                    if (e)
                        log.error(e);

                    return cb();
                });
              };

        // let currentDay = das;

        const fnScheduleMin = (cb) => {

            if (!this.isReady()) {
                return cb();
            }

            this.calendarAgent((err) => {
                if (err)
                    log.error(err);
                return cb();
            });
        };


        new Scheduler('1-59/1 * * * * *', checkSetAvailableTimeSec, {log: false});
        new Scheduler('1-59/1 * * * * *', clearAttemptOnDeadlineResultSec, {log: false});
        new Scheduler('*/1 * * * *', fnScheduleMin, {log: false});
    }

    sendEvent (d, active, callingName) {
        let e = {
            "Event-Name": EVENT_CHANGE_STATE,
            "Dialer-Id": d._id,
            "Dialer-Name": d.nameDialer,
            "Active": active,
            "Cause": d.cause,
            "State": d.state,
            "Event-Calling-Function": callingName,
            "Type": d.type,
            "Members-Count": d.countMembers,
            "Event-Domain": d._domain
        };
        log.trace(`fire event ${EVENT_CHANGE_STATE} ${d._domain} ${d._id} ${callingName}`);
        eventsService.fire(EVENT_CHANGE_STATE, d._domain, e);
        eventsService.fire(EVENT_CHANGE_STATE, 'root', e);
    }

    onConnectFs (esl) {
        log.debug(`On init esl`);
        this.connectFs = true;
        esl.subscribe('CHANNEL_HANGUP_COMPLETE');

        this.emit('changeConnection');
    }

    onConnectWConsole () {
        this.connectWConsole = true;
        this.emit('changeConnection');
    }
    onConnectWConsoleError () {
        this.connectWConsole = false;
        this.emit('changeConnection');
    }

    onConnectBrokerSuccessful () {
        const channel = application.Broker.channel;

        channel.assertQueue('', {autoDelete: true, durable: true, exclusive: true}, (err, qok) => {
            if (err)
                return log.error(err);

            channel.consume(qok.queue, (msg) => {
                try {
                    this.onBrokerMessage(msg);
                } catch (e) {
                    log.error(e);
                }
            }, {noAck: true});

            /*
            {
                "action": stop | start | sleep ?
                "args": {}
            }
             */
            channel.bindQueue(qok.queue, application.Broker.Exchange.ENGINE, "*.dialer.system", {}, (e) => {
                log.debug(`Init queue - successful`);
                this.connectBroker = true;
                this.emit('changeConnection');
            });
        });

        channel.assertQueue('engine.agents', {autoDelete: true, durable: true, exclusive: false}, (err, qok) => {
            if (err)
                return log.error(err);

            channel.consume(qok.queue, (msg) => {
                try {
                    this.onAgentStatusChange(msg);
                } catch (e) {
                    log.error(e);
                }
            }, {noAck: true});
            //#FreeSWITCH-Hostname,Event-Subclass,CC-Action,CC-Queue,Unique-ID
            channel.bindQueue(qok.queue, application.Broker.Exchange.FS_CC_EVENT, "*.callcenter%3A%3Ainfo.agent-status-change.*.*");
            channel.bindQueue(qok.queue, application.Broker.Exchange.FS_CC_EVENT, "*.callcenter%3A%3Ainfo.agent-state-change.*.*");
            log.debug(`Init queue agents - successful`);
        });
    }

    onAgentStatusChange (msg) {
        if (!msg)
            return;

        const e = JSON.parse(msg.content.toString());

        if (e['CC-Action'] === 'agent-status-change') {
            log.trace(`try set status ${e['CC-Agent']} -> ${e['CC-Agent-Status']}`);
            this.dbDialer._setAgentStatus(e['CC-Agent'], e['CC-Agent-Status'], (err, res) => {
                if (err)
                    return log.error(err);

                if (res.value) {
                    log.debug(`OK set status ${e['CC-Agent']} -> ${e['CC-Agent-Status']}`);
                    this.sendAgentToDialer(res.value);
                } else {
                    log.error(`Agent ${e['CC-Agent']} set status no response db: `, e, res);
                }
            });
        } else if (e['CC-Action'] === 'agent-state-change') {
            log.trace(`try set state ${e['CC-Agent']} -> ${e['CC-Agent-State']}`);
            this.dbDialer._setAgentState(e['CC-Agent'], e['CC-Agent-State'], (err, res) => {
                if (err)
                    return log.error(err);

                if (res.value) {
                    log.trace(`OK set state ${e['CC-Agent']} -> ${e['CC-Agent-State']}`);
                    this.sendAgentToDialer(res.value);
                } else {
                    log.error(`Agent ${e['CC-Agent']} set state no response db: `, e, res);
                }
            })
        }
    }

    getMemberFromActiveDialer (dialerId, memberId) {
        const dialer = this.activeDialer.get(dialerId);
        if (!dialer)
            return null;

        return dialer.members.get(memberId);
    }

    sendAgentToDialer (agent = {}) {
        if (agent.state === AGENT_STATE.Waiting &&
            (agent.status === AGENT_STATUS.Available || agent.status === AGENT_STATUS.AvailableOnDemand) &&
            agent.dialer instanceof Array) {

            for (let agentDialer of agent.dialer) {
                let dialer = this.activeDialer.get(agentDialer._id.toString());
                if (dialer) {
                    if (dialer.emit('availableAgent', agent))
                        return;
                }
            }
        }
    }

    onBrokerMessage (msg) {
        const {action, args = {}} = JSON.parse(msg.content.toString());
        switch (action) {
            case "start":
                this._runDialerById(args.id, args.domain, (err, res) => {
                    if (err)
                        log.error(err);
                });
                break;
            case "stop":
                this._stopDialerById(args.id, args.domain, (err, res) => {
                    if (err)
                        log.error(err);
                });
                break;

            default:
                return log.warn(`bad action: `, content);
        }
        // console.dir(content);
    }

    sendToBroker (data = {}, cb) {
        application.Broker.publish(application.Broker.Exchange.ENGINE, `${encodeRK(application._instanceId)}.dialer.system`, data, e => {
            if (e)
                log.error(e);
            return cb && cb(e);
        })
    }

    onConnectBrokerError () {
        this.connectBroker = false;
        this.emit('changeConnection');
    }

    onConnectFsError (e) {
        this.connectFs = false;
        this.emit('changeConnection', false);
    }

    onConnectDb (db) {
        log.debug(`On init db`);
        this.connectDb = true;
        this.dbDialer = db._query.dialer;
        this.dbMember = db._query.dialer;
        this.dbCalendar = db._query.calendar;

        this.emit('changeConnection');
    }

    onConnectDbError (e) {
        log.warn('Db error');
        this.connectDb = false;
        this.emit('changeConnection', false);
    }

    isReady () {
        return this.connectDb === true && this.connectFs === true && this.connectWConsole === true && this.connectBroker === true;
    }

    addTask (dialerId, domain, time) {
        if (!time)
            time = 1000;
        log.info(`Dialer ${dialerId}@${domain} next try ${new Date(Date.now() + time)}`);

        setTimeout(() => {
            if (!this.isReady()) {
                // sleep recovery min
                return this.addTask(dialerId, domain, 60 * 1000);
            }

            this.runDialerById(dialerId, domain, () => {})
        }, time);
    }

    loadCampaign () {
        this.dbDialer._getActiveDialer({}, (err, res) => {
            if (err)
                return log.error(err);

            if (res instanceof Array) {
                log.info(`Found ${res.length} dialer`);
                res.forEach((dialer) => {
                    if (dialer.stats && dialer.stats.process instanceof Array && ~dialer.stats.process.indexOf(this._app._instanceId)) {
                        log.warn(`recovery members by lock id ${this._app._instanceId}`);
                        this.recoveryCrashDialer(dialer, (e) => {
                            if (e)
                                return log.error(e);

                            this.runDialerById(dialer._id, dialer.domain);
                        })
                    } else {
                        this.runDialerById(dialer._id, dialer.domain);
                    }
                })
            } else {
                log.debug('Not found dialer');
            }
        });
    }

    forceStop () {
        let keys = this.activeDialer.getKeys();
        for (let key of keys) {
            this.activeDialer.get(key).setState(DIALER_STATES.Error);
        }
    }

    stopDialerById (id, domain, cb) {
        this.sendToBroker({
            action: "stop",
            args: {
                id,
                domain
            }
        }, cb);
    }
    _stopDialerById (id, domain, cb) {
        let dialer = this.activeDialer.get(id);
        if (!dialer) {
            this.dbDialer._updateDialer(
                id,
                DIALER_STATES.ProcessStop,
                DIALER_CAUSE.ProcessStop,
                false,
                null,
                (err, c) => {
                    if (err)
                        return cb(err);

                    return cb(null,  {state: DIALER_STATES.ProcessStop, members: 0})
                }
            );
        } else {
            dialer.setState(DIALER_STATES.ProcessStop);
            return cb(null, dialer.toJson());
        }
    }

    runDialerById(id, domain, cb) {

        this.sendToBroker({
            action: "start",
            args: {
                id,
                domain
            }
        }, cb);
    }
    _runDialerById(id, domain, cb) {

        let ad = this.activeDialer.get(id);
        if (ad) {
            if (ad.state === DIALER_STATES.Work)
                ad.emit('wakeUp');

            return cb && cb(null, {active: true});
        }

        this.dbDialer._getDialerById(id, domain, (err, res) => {
            if (err)
                return cb(err);

            if (!res)
                return cb(`Not found dialer ${id}@${domain}`);

            let error = this.addDialerFromDb(res);
            if (error)
                return cb(error);
            return cb(null, {active: true});
        });
    }

    addDialerFromDb (dialerDb) {

        if (dialerDb.active) {
            log.debug(`Dialer ${dialerDb.name} - ${dialerDb._id} is active.`);
            //return new Error("Dialer is active...");
        }

        let calendarId = dialerDb && dialerDb.calendar && dialerDb.calendar.id;


        this.dbCalendar.findById(dialerDb.domain, calendarId, (err, res) => {
            if (err)
                return log.error(err);
            // todo

            if (!res)
                return log.error('Not found calendar');

            dialerDb.lockId = this.id;
            dialerDb.state = DIALER_STATES.Idle;

            const currentTime = calendarManager.getCurrentTimeOfDay(res);
            dialerDb._currentMinuteOfDay = currentTime.currentTimeOfDay;
            dialerDb._currentWeek = currentTime.currentWeek;
            if (currentTime.expire || !currentTime.currentTimeOfDay) {
                this.emit('changeDialerState', dialerDb, res, currentTime);
                return;
            }

            let dialer = this.newInstanceDialer(dialerDb, res, this.id, this.agentManager);
            if (!dialer)
                return new Error('Bad dialer type');
            
            this.activeDialer.add(dialer._id, dialer);
        });
    }

    newInstanceDialer (dialerDb, calendarDb) {
        switch (dialerDb.type) {
            case DIALER_TYPES.ProgressiveDialer:
                dialerDb.agentManager = this.agentManager;
                return new ProgressiveDialer(dialerDb, calendarDb, this);
            case DIALER_TYPES.VoiceBroadcasting:
                return new VoiceDialer(dialerDb, calendarDb, this);
            case  DIALER_TYPES.PredictiveDialer:
                dialerDb.agentManager = this.agentManager;
                return new PredictiveDialer(dialerDb, calendarDb, this);
        }
    }

    recoveryCrashDialer (dialer, cb) {
        this.dbDialer._updateLockedMembers(
            dialer._id.toString(),
            this._app._instanceId,
            END_CAUSE.PROCESS_CRASH,
            (err, res) => {
                if (err)
                    return log.error(err);

                if (res.result.nModified) {
                    log.info(`Minus active call ${res.result.nModified}`);
                    this.dbDialer._dialerCollection.findAndModify(
                        {_id: dialer._id, "stats.active": {$gt: 0}},
                        {},
                        {
                            $currentDate: {lastModified: {$type: "timestamp" }},
                            $inc: {"stats.active": 0 - res.result.nModified}
                        },
                        {},
                        cb
                    );
                } else {
                    log.warn(`No found my lock`);
                    return cb && cb();
                }
            }
        );
    }

    clearAttemptOnDeadlineResultStatus (cb) {
        dialerService.members._updateMultiMembers(
            {
                _waitingForResultStatus: {$lte: Date.now()},
                _waitingForResultStatusCb: 1,
                "communications": {$elemMatch: {checkResult: 1}},
                _lock: null
            },
            {
                $set : {
                    _waitingForResultStatus: null,
                    _waitingForResultStatusCb: null,
                    // "communications.$.checkResult": 0,
                    "communications.$.lastCall": -1
                },
                $unset: {"checkResult": 1},
                $inc: {_probeCount: -1, "communications.$._probe": -1, "communications.$.rangeAttempts": -1},
                $push: {
                    _callback: {
                        time: Date.now(),
                        from: "system",
                        data: {
                            msg: "System schedule: no response result status, minus attempts"
                        }
                    }
                },
                $currentDate: {lastModified: true}
            },
            cb
        );
    }

    clearAttemptOnDeadlineResultStatusAddAttempts (cb) {
        const cursor = dialerService.members._getCursor({
            _waitingForResultStatus: {$lte: Date.now()},
            _waitingForResultStatusCb: 1,
            communications: {$elemMatch: {checkResult: 1}},
            _lock: null
        }, {_probeCount: 1, _maxTryCount: 1, _id: 1, variables: 1, name: 1, dialer: 1, domain: 1, communications: 1, _lastNumberId: 1});

        cursor.count((e, count) => {
            if (e)
                return cb(e);

            if (count < 1)
                return cb();


            const _exec = (e, data) => {
                if (e) {
                    return cb(e)
                }

                if (!data) {
                    return cb()
                }

                const lastAttempts = data._maxTryCount <= data._probeCount;
                dialerService.members._updateOneMember({
                    _id: data._id,
                    _waitingForResultStatusCb: 1,
                    communications: {$elemMatch: {checkResult: 1}},
                    _lock: null
                }, _getUpdateMember(lastAttempts, data.communications.length), (e, res) => {
                    if (e) {
                        return log.error(e);
                    }

                    if (res && res.result.nModified === 1 && lastAttempts) {
                        _broadcastMemberEnd(data)
                    }
                });

                cursor.nextObject(_exec);
            };

            cursor.nextObject(_exec);
        });
    }

    calendarAgent (cb) {
        calendarManager.checkDialerDeadline(this, this.dbDialer, this.dbCalendar, cb);
        //
        // this.dbDialer._getActiveDialer({calendar: 1, domain: 1}, (err, res) => {
        //     if (err)
        //         return log.error(err);
        //
        //     if (res instanceof Array) {
        //         async.forEachOf(res, (dialer, key, callback) => {
        //             const calendarId = dialer.calendar && dialer.calendar.id;
        //             if (!calendarId)
        //                 return callback();
        //
        //             this.dbCalendar.findById(dialer.domain, calendarId, (err, res) => {
        //                 if (err) {
        //                     log.error(err);
        //                     return cb(err);
        //                 }
        //
        //                 console.log(dialer, res);
        //                 callback();
        //             });
        //         }, () => {});
        //     }
        // })
    }
}


function _getUpdateMember(end, communicationsLength) {
    const $set = {
        _waitingForResultStatus: null,
        _waitingForResultStatusCb: null
    };

    if (end) {
        $set._endCause = END_CAUSE.MAX_TRY;
        for (let i = 0; i < communicationsLength; i++)
            $set[`communications.${i}.state`] = 2;
    }

    return {
        $set,
        $unset: {"checkResult": 1},
        $push: {
            _callback: {
                time: Date.now(),
                from: "system",
                data: {
                    msg: "System schedule: no response result status, save attempts"
                }
            }
        },
        $currentDate: {lastModified: true}
    }
}


function _broadcastMemberEnd(member) {
    const event = {
        "Event-Name": "CUSTOM",
        "Event-Subclass": "engine::dialer_member_end",
        // TODO
        "variable_domain_name": member.domain,
        "dialerId": member.dialer,
        "id": member._id.toString(),
        "name": member.name,
        "currentProbe": member._probeCount,
        "endCause": END_CAUSE.MAX_TRY,
        "reason": "callback",
        "callback_user_id": "system"
    };

    const lastNumber = isFinite(member._lastNumberId) && member.communications[member._lastNumberId]
            ? member.communications[member._lastNumberId]
            : null
        ;


    if (lastNumber) {
        event.currentNumber = lastNumber.number;
        event.dlr_member_number_description = lastNumber.description || ''
    }

    for (let key in member.variables) {
        if (member.variables.hasOwnProperty(key))
            event[`variable_${key}`] = member.variables[key]
    }
    console.log(event);
    application.Broker.publish(application.Broker.Exchange.FS_EVENT, `.CUSTOM.engine%3A%3Adialer_member_end..`, event);
}


module.exports = AutoDialer;