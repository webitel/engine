/**
 * Created by igor on 31.05.16.
 */

const Dialer = require('./dialer'),
    log = require(__appRoot + '/lib/log')(module),
    async = require('async'),
    DIALER_TYPES = require('./const').DIALER_TYPES,
    AGENT_STATUS = require('./const').AGENT_STATUS,
    END_CAUSE = require('./const').END_CAUSE,
    CANCEL_CAUSE = 'ORIGINATOR_CANCEL'
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
                    bridgedCall: m.bridgedCall,
                    connectedCall: m.getConnectedFlag(),
                    predictAdjust: this._stats.predictAdjust,
                    amd: m.getAmdResult()
                },
                e => {
                    if (!e)
                        engine();
                }
            );
        });

        this.getLimit = () => {
            if (this._stats.queueLimit && this._stats.queueLimit < this._limit) {
                return this._stats.queueLimit;
            }
            return this._limit;
        };

        this.on('availableAgent', () => {
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

                    if (!this.isReady() || res.members < 1) {
                        if (this.members.length() === 0) {
                            this.tryStop();
                        }
                        return;
                    }

                    log.trace('engine response: ', res);

                    if (res.agents === 0 )
                        return;

                    if (!this._stats.predictAbandoned)
                        this._stats.predictAbandoned = 0;

                    if (!this._stats.bridgedCall)
                        this._stats.bridgedCall = 0;

                    if (!this._stats.callCount)
                        this._stats.callCount = 0;

                    if (!this._stats.successCall)
                        this._stats.successCall = 0;

                    if (!this._stats.predictAdjust)
                        this._stats.predictAdjust = this._predictAdjust;

                    console.log(`all agent ${res.allLogged} active ${this._active}`);

                    if (this._stats.bridgedCall < 10) {
                        this._stats.queueLimit = res.allLogged;
                        if (this._active < res.allLogged ) {
                            this.huntingMember();
                        } else {
                            if (this.members.length() === 0) {
                                this.tryStop();
                            }
                        }
                        return;
                    }


                    if (this._stats.callCount > 200 ) {
                        const silentCalls = this._targetPredictiveSilentCalls - ((this._stats.predictAbandoned  * 100) / this._stats.callCount);

                        this._stats.predictAdjust = Math.round( (Math.pow(2,silentCalls) / Math.pow(0.05,silentCalls)) * 35) ;
                        // Math.round( ((silentCalls - this._targetPredictiveSilentCalls) * 2 ) / 0.05 );

                        if (this._stats.predictAdjust > 1000) {
                            this._stats.predictAdjust = 1000;
                        } else if (this._stats.predictAdjust < 0) {
                            this._stats.predictAdjust = 0
                        }

                        console.log('>>>>>>>>>>>>',this._stats.predictAdjust, ">>>", silentCalls);
                    }

                    if (this._stats.predictAdjust === 0) {
                        this._stats.queueLimit = res.allLogged;
                    } else {
                        const connectRate = this._stats.callCount / this._stats.bridgedCall;
                        const overDial = Math.abs((res.agents / connectRate) - res.agents);

                        const call = Math.ceil(res.agents + (overDial * this._stats.predictAdjust) / 100 );

                        this._stats.queueLimit =  call + (res.allLogged - res.agents) ;
                    }

                    console.log(`CALL ->> -->> ${this._stats.queueLimit} all agents: ${res.allLogged}`);

                    if ((res.agents <= 2 && this._active < res.allLogged) || (res.agents >= 2  && this._active < this._stats.queueLimit) ) {
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
                `cc_side=member`,
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

    _originateAgent (member, agent) {
        member.setAgent(agent);

        let agentVars = [
            `cc_side=agent`,
            `cc_agent='${agent.agentId}'`,
            `cc_queue='${this.nameDialer}'`,
            `dlr_session='${member.sessionId}'`,
            `origination_callee_id_number='${agent.agentId}'`,
            `origination_callee_id_name='${agent.agentId}'`,
            `origination_caller_id_number='${member.number}'`,
            `origination_caller_id_name='${member.name}'`,
            `destination_number='${member.number}'`,
            `effective_caller_id_number='${agent.agentId}'`,
            `effective_callee_id_number='${member.number}'`,
            'webitel_direction=inbound'
        ];

        for (let key in this._variables) {
            if (this._variables.hasOwnProperty(key)) {
                agentVars.push(`${key}='${this._variables[key]}'`);
            }
        }

        application.Esl.bgapi(`uuid_setvar ${member.sessionId} cc_agent ${agent.agentId}`);

        const start = Date.now();

        member._predAgentOriginate = true;

        const agentDs = `originate {${agentVars}}user/${agent.agentId} &park()`;
        member.log(`Agent ds: ${agentDs}`);
        application.Esl.bgapi(agentDs, (res) => {
            member.log(`agent ${agent.agentId} fs res -> ${res.body}`);
            const bgOkData = res.body.match(/^\+OK\s(.*)\n$/);
            const date = Date.now();

            if (member.processEnd) {
                member.agent = null;

                if (bgOkData)
                    application.Esl.bgapi(`uuid_kill ${bgOkData[1]} ${CANCEL_CAUSE}`);

                this._am.setAgentStats(agent.agentId, this._objectId, {
                    lastStatus: `error bridge member end`,
                    setAvailableTime: agent.status === AGENT_STATUS.AvailableOnDemand ? null : Date.now() + (this.getAgentParam('wrapUpTime', agent) * 1000),
                    process: null,
                    idleSec: this._am.getIdleTimeSecAgent(agent)
                }, (e) => {
                    if (e)
                        log.error(e);

                });
                return
            }

            if (bgOkData) {
                application.Esl.bgapi(`uuid_bridge ${member.sessionId} ${bgOkData[1]}`, (bridge) => {
                    member._predAgentOriginate = false;

                    member.log(`fs response bridge agent: ${bridge.body}`);

                    if (/^-ERR|^-USAGE/.test(bridge.body)) {
                        this._am.setAgentStats(agent.agentId, this._objectId, {
                            lastStatus: `error bridge -> ${bridge.body}`,
                            setAvailableTime: agent.status === AGENT_STATUS.AvailableOnDemand ? null : Date.now() + (this.getAgentParam('wrapUpTime', agent) * 1000),
                            process: null,
                            idleSec: this._am.getIdleTimeSecAgent(agent)
                        }, (e) => {
                            if (e)
                                log.error(e);

                            member.agent = null;
                            this._joinAgent(member);
                        });
                    } else {
                        member.predictAbandoned = false;
                        member.bridgedCall = true;
                        this._am.setAgentStats(agent.agentId, this._objectId, {
                            lastBridgeCallTimeStart: date,
                            connectedTimeSec: timeToSec(date, start),
                            lastStatus: `active -> ${member._id}`,
                            idleSec: this._am.getIdleTimeSecAgent(agent)
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
            } else if (/^-ERR|^-USAGE/.test(res.body)) {
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
                                process: null,
                                idleSec: this._am.getIdleTimeSecAgent(agent),
                                missedCall: true
                            }, (e) => {
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
                        process: null,
                        idleSec: this._am.getIdleTimeSecAgent(agent),
                        missedCall: true
                    }, (e) => {
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
                        process: null,
                        idleSec: this._am.getIdleTimeSecAgent(agent),
                        missedCall: true
                    }, (e) => {
                        if (e)
                            return log.error(e);
                    });
                }

                this._joinAgent(member);
            } else {
                log.error(res.body);
            }
        });
    }

    _joinAgent (member) {
        if (member.getBridgedTime() + (1000 * this._maxLocateAgentSec) <= Date.now() ) {
            application.Esl.bgapi(`uuid_kill ${member.sessionId} LOSE_RACE`); // todo add conf hangup cause
            return;
        }

        if (member._predTimer) {
            clearTimeout(member._predTimer);
            member._predTimer = null;
        }

        this._am.huntingAgent(this._objectId, this._agents, this._skills, this.agentStrategy, this._readyTime,  member, (err, agent) => {
            if (err)
                log.error(err);


            if (!agent) {
                member.log(`no found agent, try now`);
                member._predTimer = setTimeout(() => {
                    this._joinAgent(member);
                }, 500); // TODO add config
                return;
            } else if (member.processEnd) {
                member.log(`set agent ${agent.agentId} rollback`);
                this._am.setAgentStats(agent.agentId, this._objectId, {
                    lastStatus: `rollback -> ${member._id}`,
                    setAvailableTime: agent.status === AGENT_STATUS.AvailableOnDemand ? null : Date.now() + (this.getAgentParam('wrapUpTime', agent) * 1000),
                    process: null,
                    idleSec: this._am.getIdleTimeSecAgent(agent)
                }, (e) => {
                    if (e)
                        log.error(e);
                });
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

        const  onChannelPark = (e) => {
            member.setConnectedFlag(true);
            
            if (this._amd.enabled === true) {
                let amdResult = e.getHeader('variable_amd_result');
                member.log(`amd_result=${amdResult}`);
                if ( !(amdResult === 'HUMAN' || (this._amd.allowNotSure && amdResult === 'NOTSURE')) ) {
                    application.Esl.bgapi(`uuid_kill ${member.sessionId} NORMAL_UNSPECIFIED`);
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

            member.channelsCount--;

            if (member._predAgentOriginate === true) {
                log.trace(`hangup agent channel for dlr_session ${member.sessionId}`);
                application.Esl.bgapi(`hupall ${CANCEL_CAUSE} dlr_session ${member.sessionId}`);
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
                    process: null,
                }, (e) => {
                    if (e)
                        log.error(e);
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
        member.channelsCount++;
        application.Esl.bgapi(ds, (res) => {
            log.trace(`fs response: ${res && res.body}`);
            member.log(`fs response: ${res.body}`);
            if (/^-ERR|^-USAGE/.test(res.body)) {
                member.channelsCount--;
                member.end(res.body.replace(/-ERR|-USAGE\s(.*)\n/, '$1'));
            }
        });

    }

};


function timeToSec(current, start) {
    return Math.round( (current - start) / 1000 )
}