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
    ObjectID = require('mongodb').ObjectID,
    dialerService = require(__appRoot + '/services/dialer'),
    encodeRK = require(__appRoot + '/utils/helper').encodeRK
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
        this.agentManager = new AgentManager(this);

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

        // app.Broker.on('ccEvent', this.onAgentStatusChange.bind(this));
        // app.Broker.on('webitelEvent', (e) => {
        //
        //     // TODO replace skills;
        //     return void 0;
        //     if (e['Event-Name'] == 'USER_MANAGED') {
        //         let agentId = `${e['Event-Account']}@${e['User-Domain']}`;
        //         let agent = this.agentManager.getAgentById(agentId);
        //         if (agent) {
        //
        //         }
        //     }
        // });

        this.activeDialer.on('added', (dialer) => {

            dialer.on('ready', (d) => {
                log.debug(`Ready dialer ${d.nameDialer} - ${d._id}`);

                this.dbDialer._dialerCollection.findOneAndUpdate(
                    {_id: d._objectId},
                    {
                        $set: {state: d.state, _cause: d.cause, active: true, nextTick: null},
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

                let sleepTime = (d.state === DIALER_STATES.Sleep) ? (new Date(Date.now() + dialer._calendar.sleepTime)) : null;

                this.dbDialer._dialerCollection.findOneAndUpdate(
                    {_id: d._objectId},
                    {
                        $set: {state: d.state, _cause: d.cause, active: d.state === DIALER_STATES.Sleep, nextTick: sleepTime},
                        $pull: {"stats.process": this._app._instanceId}
                    },
                    e => {
                        if (e)
                            log.error(e);
                        this.activeDialer.remove(dialer._id);
                    }
                );

                if (sleepTime)
                    this.addTask(d._id, d._domain, dialer._calendar.sleepTime);

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
        this.connectBroker = true;
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
            channel.bindQueue(qok.queue, application.Broker.Exchange.ENGINE, "*.dialer.systemm");
            log.debug(`Init queue - successful`);
            this.emit('changeConnection');
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
            this.dbDialer._setAgentStatus(e['CC-Agent'], e['CC-Agent-Status'], (err, res) => {
                if (err)
                    return log.error(err);

                if (res.value) {
                    this.sendAgentToDialer(res.value);
                }
            });
        } else if (e['CC-Action'] === 'agent-state-change') {
            this.dbDialer._setAgentState(e['CC-Agent'], e['CC-Agent-State'], (err, res) => {
                if (err)
                    return log.error(err);

                if (res.value) {
                    this.sendAgentToDialer(res.value);
                }
            })
        }
    }

    sendAgentToDialer (agent = {}) {
        if (agent.state === AGENT_STATE.Waiting &&
            (agent.status === AGENT_STATUS.Available || agent.status === AGENT_STATUS.AvailableOnDemand) &&
            agent.dialer instanceof Array) {

            for (let agentDialer of agent.dialer) {
                let dialer = this.activeDialer.get('' + agentDialer._id);
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
        application.Broker.publish(application.Broker.Exchange.ENGINE, `${encodeRK(application._instanceId)}.dialer.systemm`, data, e => {
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
        this.dbDialer._getActiveDialer((err, res) => {
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
                ad.huntingMember();

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
            dialerDb.lockId = this.id;
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
                            $currentDate: {lastModified: true},
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
                "communications.checkResult": 1,
                _lock: null
            },
            {
                $set : {
                    _waitingForResultStatus: null,
                    _waitingForResultStatusCb: null,
                    "communications.$.checkResult": null,
                    "communications.$.lastCall": -1
                },
                $inc: {_probeCount: -1, "communications.$._probe": -1, "communications.$.rangeAttempts": -1},
                $push: {
                    _log: {
                        time: Date.now(),
                        text: `Schedule no response result status`
                    }
                },
                $currentDate: {lastModified: true}
            },
            cb
        );
    }
}

module.exports = AutoDialer;