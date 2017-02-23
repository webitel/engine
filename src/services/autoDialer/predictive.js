/**
 * Created by igor on 31.05.16.
 */

const Dialer = require('./dialer'),
    generateUuid = require('node-uuid'),
    log = require(__appRoot + '/lib/log')(module),
    async = require('async'),
    DIALER_TYPES = require('./const').DIALER_TYPES,
    AGENT_STATUS = require('./const').AGENT_STATUS,
    END_CAUSE = require('./const').END_CAUSE,

    ControllerIntegralGain = 0.05,
    ControllerProportionalGain = 2
    ;

module.exports = class Predictive extends Dialer {
    constructor (config, calendarConf, dialerManager) {
        super(DIALER_TYPES.PredictiveDialer, config, calendarConf, dialerManager);

        this._am = config.agentManager;

        application.Esl.subscribe([ 'CHANNEL_HANGUP_COMPLETE', 'CHANNEL_PARK']);

        this.members.on('added', (member) => {
            if (member.checkExpire()) {
                member.endCause = END_CAUSE.MEMBER_EXPIRED;
                member.end(END_CAUSE.MEMBER_EXPIRED);
            } else {
                this.dialMember(member)
            }

            engine();
        });

        this.members.on('removed', (m) => {
            this.rollback(
                m,
                m.getDestination(),
                {
                    queueLimit: this._stats.queueLimit || this._limit,
                    predictAbandoned: m.predictAbandoned,
                    predictAdjust: this._stats.predictAdjust
                },
                e => {
                    if (!e)
                        engine();
                }
            );
        });

        this.getLimit = () => {
            if (this._stats.queueLimit) {
                return this._stats.queueLimit;
            }
            return this._limit;
        };


        this.on('availableAgent', agent => {
            engine();
        });

        const engine = () => {
            async.parallel(
                {
                    agents: (cb) => {
                        dialerManager.agentManager.getAvailableCount(this._objectId, this._agents, this._skills, cb);
                    },
                    allLogged: (cb) => {
                        dialerManager.agentManager.getAllLoggedAgent(this._objectId, this._agents, this._skills, cb);
                    },
                    members: (cb) => {
                        this.countAvailableMembers(this._limit, cb);
                    }
                },
                (err, res) => {
                    if (err)
                        return log.error(err);

                    if (!this.isReady()) {
                        if (this.members.length() === 0) {
                            this.tryStop();
                        }
                        return;
                    }

                    if (res.agents === 0 )
                        return;

                    if (!this._stats.successCall || this._stats.successCall <= 10 || res.agents <=2) {
                        this._stats.queueLimit = this._active + res.agents - 1;
                        if ( (this._active - (res.allLogged - res.agents)) - res.agents < 0 ) {
                            this.huntingMember();
                        } else {
                            if (this.members.length() === 0) {
                                this.tryStop();
                            }
                        }
                        return;
                    }


                    // if ( (this._active - (res.allLogged - res.agents)) - res.agents < 0)
                    //     this.huntingMember();
                    //
                    // return;

                    // Init state

                    const silentCalls = (this._stats.predictAbandoned * 100) / this._stats.callCount;
                    this._stats.predictAdjust = this._stats.predictAdjust || this._predictAdjust;

                    // this._stats.predictAdjust += -( ((silentCalls - this._targetPredictiveSilentCalls) * ControllerProportionalGain) / ControllerIntegralGain  );
                    
                    this._stats.predictAdjust = ((100 - Math.abs(silentCalls) ) * ControllerProportionalGain * ControllerIntegralGain )  * 100;

                    console.log(silentCalls, this._stats.predictAdjust);

                    const connectRate = this._stats.callCount / this._stats.successCall;
                    const overDial = Math.abs((res.agents / connectRate) - res.agents);
                    const call = Math.ceil(res.agents + (overDial * this._stats.predictAdjust) / 100 );

                    this._stats.queueLimit = this._active + call - 1;
                    console.log(`CALL ->> +${call - res.agents} -->> ${this._stats.queueLimit}`);

                    if ( (this._active - (res.allLogged - call)) - call < 0 ) {
                        this.huntingMember();
                    } else {
                        if (this.members.length() === 0) {
                            this.tryStop();
                        }
                    }

                }
            );
        };

        this.on('ready', () => {
            engine();
        });

        this.on('wakeUp', () => {
            engine()
        });

        this.getDialString = (member) => {
            let vars = [
                `origination_uuid=${member.sessionId}`,
                `dlr_member_id=${member._id.toString()}`,
                `dlr_id=${member.getQueueId()}`,
                `presence_data='${member.getDomain()}'`,
                `cc_queue='${member.getQueueName()}'`
            ];

            for (let key in this._variables) {
                if (this._variables.hasOwnProperty(key)) {
                    vars.push(`${key}='${this._variables[key]}'`);
                }
            }

            if (member._currentNumber && member._currentNumber.description) {
                vars.push(`dlr_member_number_description='${member._currentNumber.description}'`);
            }

            for (let key of member.getVariableKeys()) {
                vars.push(`${key}='${member.getVariable(key)}'`);
            }

            const dest = member.getDestination();

            const apps = [];
            if (this._recordSession) {
                vars.push(
                    `RECORD_MIN_SEC=2`,
                    `RECORD_STEREO=true`,
                    `RECORD_BRIDGE_REQ=false`,
                    `recording_follow_transfer=true`
                );

                let sessionUri = 'http_cache://$${cdr_url}' +
                    encodeURI(`/sys/formLoadFile?domain=${member.getDomain()}&id=${member.sessionId}&type=mp3&email=none&name=recordSession&.mp3`);

                apps.push(`record_session:${sessionUri}`)
            }

            const gw = dest.gwProto === 'sip' && dest.gwName ? `sofia/gateway/${dest.gwName}/${dest.dialString}` : dest.dialString;
            const dialString = member.number.replace(dest._regexp, gw);

            vars.push(
                `origination_callee_id_number='${member.number}'`,
                `origination_callee_id_name='${member.name}'`,

                `origination_caller_id_number='${dest.callerIdNumber}'`,
                `origination_caller_id_name='${member.getQueueName()}'`,

                `destination_number='${member.number}'`,

                `originate_timeout=${this._originateTimeout}`,
                'webitel_direction=outbound'
            );

            if (this._amd && this._amd.enabled) {
                apps.push(`amd:${this._amd._string}`);

            }

            vars.push('ignore_early_media=true'); //TODO move config

            apps.push(`park:`);

            return `originate {${vars}}${dialString} '${apps.join(',')}' inline`;
        }
    }

    calcLimit (agent) {
        console.log(`call request ${this._callRequestCount}`);
        if (!this.isReady()) {
            this._queueCall.length = 0;
            return;
        }

        let aC = 0;
        let cc = 0;
        if (agent && this._gotCallCount > 10) {
            aC = 1;
            this._skipAgents.push(agent);
        } else {
            if (this._callRequestCount != 0) return;
            this._skipAgents = this._am.getFreeAgents(this._agents, this.agentStrategy);
            aC = this._skipAgents.length;
        }

        if (aC == 0)
                return;

        if (this._predictAdjust != 0 && this._gotCallCount > 10 && aC > 0) {

            if (this.__dumpLastRecalc < this._gotCallCount) {
                let avgBad = (this._badCallCount * 100) / this._allCallCount;
                if (avgBad > 2) {
                    this._predictAdjust -= this._predictAdjust * (avgBad / 100);
                    if (this._predictAdjust <= 0)
                        this._predictAdjust = 1;

                } else if (avgBad > 0 && avgBad < 2 && this._predictAdjust < 1000) {
                    this._predictAdjust += this._predictAdjust * ((100 - avgBad) / 100);
                    if (this._predictAdjust > 1000)
                        this._predictAdjust = 1000;
                } else  if (avgBad === 0) {
                    this._predictAdjust += this._predictAdjust * 0.055;
                }
                this.__dumpLastRecalc = this._gotCallCount + 10;
            }

            let connectRate = this._allCallCount / this._gotCallCount;
            let overDial = Math.abs((aC / connectRate) - aC);
            console.log(`connectRate: ${connectRate} overDial: ${overDial}`);
            cc =  Math.ceil(aC + (overDial * this._predictAdjust) / 100 );
        } else {
            cc = aC;
        }

        console.log(`concurrency: ${this._activeCallCount + cc}; cc: ${cc}; aC: ${aC}; calls: ${this._activeCallCount}; adjust: ${this._predictAdjust}; all: ${this._allCallCount}; got: ${this._gotCallCount}; bad: ${this._badCallCount}`);

        if (this._queueCall.length > 0) {
            for (let i = 0; i < cc; i++) {
                let m = this._queueCall.shift();
                if (!m) {
                    break;
                }
                this.__dial(m, this.calcLimit.bind(this));
            }
        } else {
            this._skipAgents.length = 0;
        }
    }

    _originateAgent (member, agent) {
        member.setAgent(agent);

        member._predAgentOriginateUuid = generateUuid.v4();

        let agentVars = [
            `origination_uuid=${member._predAgentOriginateUuid}`,
            `origination_callee_id_number='${agent.agentId}'`,
            `origination_callee_id_name='${agent.agentId}'`,
            `origination_caller_id_number='${member.number}'`,
            `origination_caller_id_name='${member.name}'`,
            `destination_number='${member.number}'`,
            `effective_caller_id_number='${agent.agentId}'`,
            `effective_callee_id_number='${member.number}'`
        ];

        for (let key in this._variables) {
            if (this._variables.hasOwnProperty(key)) {
                agentVars.push(`${key}='${this._variables[key]}'`);
            }
        }

        application.Esl.bgapi(`uuid_setvar ${member.sessionId} cc_agent ${agent.agentId}`);

        const start = Date.now();
        application.Esl.bgapi(`originate {${agentVars}}user/${agent.agentId} &eval('` + '${uuid_bridge(' + member.sessionId + ' ${uuid}' +  `)}')`, (res) => {
            member.log(`agent fs res -> ${res.body}`);
            member._predAgentOriginateUuid = null;
            if (member.processEnd)
                return;
            const date = Date.now();

            if (/^-ERR|^-USAGE/.test(res.body)) {
                let error =  res.body.replace(/-ERR\s(.*)\n/, '$1');
                member.log(`agent error: ${error}`);

                if (error === 'NO_ANSWER') {
                    if (this.getAgentParam('maxNoAnswer', agent) <= (this.getAgentParam('noAnswerCount', agent) + 1)) {
                        return this._am.setNoAnswerAgent(agent, e => {
                            if (e)
                                log.error(e);

                            this._am.setAgentStats(agent.agentId, this._objectId, {
                                call: true,
                                gotCall: false,
                                clearNoAnswer: true,
                                setAvailableTime: null,
                                connectedTimeSec: timeToSec(date, start),
                                lastStatus: `NO_ANSWER -> ${member._id} -> MAX`,
                                process: null
                            }, (e, res) => {
                                if (e)
                                    return log.error(e);
                            });
                        });
                    }

                    this._am.setAgentStats(agent.agentId, this._objectId, {
                        call: true,
                        gotCall: false,
                        noAnswer: true,
                        connectedTimeSec: timeToSec(date, start),
                        lastStatus: `NO_ANSWER -> ${member._id}`,
                        setAvailableTime: date + (this.getAgentParam('noAnswerDelayTime', agent) * 1000),
                        process: "checkState"
                    }, (e, res) => {
                        if (e)
                            return log.error(e);
                    });
                } else {
                    this._am.setAgentStats(agent.agentId, this._objectId, {
                        call: true,
                        gotCall: false,
                        connectedTimeSec: timeToSec(date, start),
                        lastStatus: `REJECT -> ${member._id} -> ${error}`,
                        setAvailableTime: date + (this.getAgentParam('rejectDelayTime', agent) * 1000),
                        process: "checkState"
                    }, (e, res) => {
                        if (e)
                            return log.error(e);
                    });
                }

                this._joinAgent(member);
            } else {
                member.predictAbandoned = false;
                this._am.setAgentStats(agent.agentId, this._objectId, {
                    lastBridgeCallTimeStart: date,
                    connectedTimeSec: timeToSec(date, start),
                    lastStatus: `active -> ${member._id}`
                }, (e, res) => {
                    if (e)
                        return log.error(e);

                    if (res && res.value) {
                        //member._agent = res.value
                        //TODO
                    }
                });
            }
        });
    }

    _joinAgent (member) {
        if (member.getBridgedTime() + (1000 * 10) <= Date.now() ) { //todo move conf 10 sec found
            application.Esl.bgapi(`uuid_kill ${member.sessionId} LOSE_RACE`); // todo add conf hangup cause
            return;
        }

        this._am.huntingAgent(this._objectId, this._agents, this._skills, this.agentStrategy, (err, agent) => {
            if (err)
                throw err;

            if (!agent) {
                member.log(`no found agent, try now`);
                member._predTimer = setTimeout(() => {
                    this._joinAgent(member);
                }, 500); // TODO add config
                return;
            }

            this._originateAgent(member, agent);
        });
    }

    dialMember (member) {
        log.trace(`try call ${member.sessionId}`);
        const ds = this.getDialString(member);
        member.log(`dialString: ${ds}`);
        log.trace(`Call ${ds}`);

        application.Esl.bgapi(ds, (res) => {
            log.trace(`fs response: ${res && res.body}`);
            member.channelsCount++;

            if (/^-ERR|^-USAGE/.test(res.body)) {
                let error =  res.body.replace(/-ERR\s(.*)\n/, '$1');

                member.minusProbe();
                member.nextTrySec = 1;
                member.end();
            }
        });

        const  onChannelPark = (e) => {
            if (this._amd.enabled === true) {
                let amdResult = e.getHeader('variable_amd_result');
                member.log(`amd_result=${amdResult}`);
                if (amdResult !== 'HUMAN') {
                    application.Esl.bgapi(`uuid_kill ${member.sessionId} USER_BUSY`);
                    return;
                }
            }

            if (this._broadcastPlaybackUri) {
                log.trace(`broadcast ${member.sessionId} playback::${this._broadcastPlaybackUri} aleg`);
                application.Esl.bgapi(`uuid_broadcast ${member.sessionId} playback::${this._broadcastPlaybackUri} aleg`);
            }
            member.predictAbandoned = true;
            member.setBridgedTime();
            member.log(`answer`);

            this._joinAgent(member);
        };

        const onChannelHangup = (e) => {
            if (member._predTimer) {
                clearTimeout(member._predTimer);
            }

            if (member._predAgentOriginateUuid) {
                application.Esl.bgapi(`uuid_kill ${member._predAgentOriginateUuid} ORIGINATOR_CANCEL`);
            }

            const agent = member.getAgent();
            if (agent) {
                member.log(`set agent ${agent.agentId} check status`);
                this._am.setAgentStats(agent.agentId, this._objectId, {
                    call: true,
                    gotCall: true, //TODO
                    clearNoAnswer: true,
                    lastBridgeCallTimeEnd: Date.now(),
                    callTimeSec: +e.getHeader('variable_billsec') || 0,
                    lastStatus: `end -> ${member._id}`,
                    setAvailableTime: agent.status === AGENT_STATUS.AvailableOnDemand ? null : Date.now() + (this.getAgentParam('wrapUpTime', agent) * 1000),
                    process: null
                }, (e, res) => {
                    if (e)
                        return log.error(e);
                });
            }

            member.end(e.getHeader('variable_hangup_cause'), e);
        };

        member.once('end', (m) => {
            application.Esl.off(`esl::event::CHANNEL_PARK::${member.sessionId}`, onChannelPark);
            application.Esl.off(`esl::event::CHANNEL_HANGUP_COMPLETE::${member.sessionId}`, onChannelHangup);
        });

        application.Esl.once(`esl::event::CHANNEL_PARK::${member.sessionId}`, onChannelPark);
        application.Esl.once(`esl::event::CHANNEL_HANGUP_COMPLETE::${member.sessionId}`, onChannelHangup);

    }

};


function timeToSec(current, start) {
    return Math.round( (current - start) / 1000 )
}