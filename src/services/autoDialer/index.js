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
    END_CAUSE = require('./const').END_CAUSE,
    Collection = require(__appRoot + '/lib/collection'),
    VoiceDialer = require('./voice'),
    ProgressiveDialer = require('./progressive'),
    PredictiveDialer = require('./predictive'),
    eventsService = require(__appRoot + '/services/events')
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


        app.Broker.on('ccEvent', this.onAgentStatusChange.bind(this));
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

                this.dbDialer._updateDialer(d._objectId, d.state, d.cause, true, null, (err) => {
                    if (err)
                        log.error(err);
                    this.sendEvent(d, true, 'ready');

                });
            });

            dialer.once('end', (d) => {
                log.debug(`End dialer ${d.nameDialer} - ${d._id} - ${d.cause}`);

                if (dialer._agents instanceof Array && (dialer.type === DIALER_TYPES.ProgressiveDialer || dialer.type === DIALER_TYPES.PredictiveDialer))
                    this.agentManager.removeDialerInAgents(dialer._agents, dialer._id);

                let sleepTime = (d.state === DIALER_STATES.Sleep) ? (new Date(Date.now() + dialer._calendar.sleepTime)) : null;
                this.dbDialer._updateDialer(d._objectId, d.state, d.cause, d.state === DIALER_STATES.Sleep, sleepTime, (err) => {
                    if (err)
                        log.error(err);

                    this.activeDialer.remove(dialer._id);
                });
                if (sleepTime)
                    this.addTask(d._id, d._domain, dialer._calendar.sleepTime);

                this.sendEvent(d, d.state === DIALER_STATES.Sleep, 'end');
            });

            // // TODO Remove event;
            // dialer.on('sleep', (d) => {
            //
            //     this.dbDialer._updateDialer(d._objectId, d.cause, true, new Date(Date.now() + dialer._sleepNextTry), (err) => {
            //         if (err)
            //             log.error(err);
            //     });
            //     this.addTask(d._id, d._domain, dialer._sleepNextTry);
            // });

            dialer.on('error', (d) => {
                log.warn(`remove dialer ${d.nameDialer}`);
                this.activeDialer.remove(d._id);
            });

            if (dialer.type === DIALER_TYPES.PredictiveDialer || dialer.type === DIALER_TYPES.ProgressiveDialer) {

                this.agentManager.initAgents(dialer, (err, res) => {
                    if (err)
                        return log.error(err);

                    this.agentManager.addDialerInAgents(dialer._agents, dialer._id);

                    dialer.setReady();
                });

            } else {
                dialer.setReady();
            }
        });

        this.activeDialer.on('removed', (dialer) => {
            if (dialer.type === DIALER_TYPES.ProgressiveDialer && dialer._agents instanceof Array)
                this.agentManager.removeDialerInAgents(dialer._agents, dialer._id);

            log.info(`Remove active dialer ${dialer.nameDialer} : ${dialer._id} - ${dialer.cause}`);
            this.sendEvent(dialer, dialer.state === DIALER_STATES.Sleep, 'removed');
        });
    }

    onAgentStatusChange (e) {
        let agent = this.agentManager.getAgentById(e['CC-Agent']);
        if (!agent) return;
        if (e['CC-Action'] === 'agent-status-change') {
            agent.setStatus(e['CC-Agent-Status']);
        } else if (e['CC-Action'] === 'agent-state-change') {
            agent.setState(e['CC-Agent-State']);
        }
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

    sendAgentToDialer (agent) {
        for (let key of agent.dialers) {
            let d = this.activeDialer.get(key);
            if (d && d.setAgent(agent)) {
                break;
            } else  {
                log.debug('skip')
            }
        }
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
        return this.connectDb === true && this.connectFs === true && this.connectWConsole === true;
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
                    this.recoveryCrashDialer(dialer);
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

        let ad = this.activeDialer.get(id);
        if (ad)
            return cb && cb(null, {active: true});

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
                return new ProgressiveDialer(dialerDb, calendarDb);
            case DIALER_TYPES.VoiceBroadcasting:
                return new VoiceDialer(dialerDb, calendarDb);
            case  DIALER_TYPES.PredictiveDialer:
                dialerDb.agentManager = this.agentManager;
                return new PredictiveDialer(dialerDb, calendarDb);
        }
    }

    recoveryCrashDialer (dialer) {
        this.dbDialer._updateLockedMembers(
            dialer._id.toString(),
            this.id,
            END_CAUSE.PROCESS_CRASH,
            (err) => {
                if (err)
                    return log.error(err);
                this.addDialerFromDb(dialer);
            }
        );
    }
}

module.exports = AutoDialer;