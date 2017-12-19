/**
 * Created by igor on 31.05.16.
 */

const Dialer = require('./dialer'),
    log = require(__appRoot + '/lib/log')(module),
    async = require('async'),
    DIALER_TYPES = require('./const').DIALER_TYPES,
    AGENT_STATUS = require('./const').AGENT_STATUS,
    END_CAUSE = require('./const').END_CAUSE,
    VAR_SEPARATOR = require('./const').VAR_SEPARATOR,
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
                    waitSec: m.getWaitSec(),
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

        let t = null;
        const tryNow = () => {
            clearTimeout(t);
            t = setTimeout(() => {
                engine();
            }, 1000);
        };

        const engine = () => {
            async.parallel(
                {
                    agents: (cb) => {
                        dialerManager.agentManager.getAvailableCount(this._objectId, this._agents, this._skills, cb);
                        // cb(null, 2)
                    },
                    allLogged: (cb) => {
                        dialerManager.agentManager.getAllLoggedAgent(this._objectId, this._agents, this._skills, cb);
                        // cb(null, 50)
                    },
                    members: (cb) => {
                        this.countAvailableMembers(this._limit, cb);
                    }
                },
                (err, res) => {
                    if (err)
                        return log.error(err);

                    if (!this.isReady() || res.members < 1) {
                        this.tryStop();
                        return;
                    }

                    log.trace('engine response: ', res);

                    if (res.agents === 0 ) {
                        tryNow();
                        return;
                    }

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

                    if (this._stats.bridgedCall < this._predictStartBridgedCount) {
                        this._stats.queueLimit = res.allLogged;
                        if (this._active < res.allLogged ) {
                            this.huntingMember();
                        } else if (this.members.length() === 0) {
                            this.tryStop();
                        } else {
                            tryNow();
                        }
                        return;
                    }


                    if (this._stats.callCount > this._predictStartCallCount ) {
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
                    } else if (this.members.length() === 0) {
                        this.tryStop();
                    } else {
                        tryNow();
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
                vars.push(`'${key}'='${member.getVariable(key)}'`);
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

                `origination_caller_id_number='${member.getCallerIdNumber()}'`,
                `origination_caller_id_name='${member.getQueueName()}'`,

                `destination_number='${member.number}'`,

                `originate_timeout=${this._originateTimeout}`,
                'webitel_direction=outbound'
            );

            vars.push('ignore_early_media=true'); //TODO move config

            if (this._amd && this._amd.enabled) {
                vars.push("hangup_after_bridge=true");

                vars.push("amd_on_human=park");
                vars.push(`amd_on_machine=hangup::NORMAL_UNSPECIFIED`);
                vars.push(`amd_on_notsure=${this._amd.allowNotSure ? 'park' : 'hangup::NORMAL_UNSPECIFIED'}`);

                apps.push(`amd:${this._amd._string}`);

                if (this._amd.playbackFile) {

                    if (this._amd.beforePlaybackFileTime > 0)
                        apps.push(`sleep:${this._amd.beforePlaybackFileTime}`);

                    apps.push(`playback:${this._amd.playbackFile}`);

                    if (this._amd.totalAnalysisTime - this._amd.beforePlaybackFileTime > 0) {
                        apps.push(`sleep:${this._amd.totalAnalysisTime - this._amd.beforePlaybackFileTime + 100}`);
                    }
                } else {
                    apps.push(`sleep:${this._amd.totalAnalysisTime + 100}`);
                }

            } else {
                apps.push(`park:`);
            }

            return `originate {^^${VAR_SEPARATOR}${vars.join(VAR_SEPARATOR)}}${dialString} '${apps.join(',')}' inline`;
        }
    }

    _originateAgent (member, agent) {
        member.setAgent(agent);

        let agentVars = [
            `cc_side=agent`,
            `cc_agent='${agent.name}'`,
            `cc_queue='${this.nameDialer}'`,
            `originate_timeout=${this.getAgentOriginateTimeout(agent)}`,
            `dlr_session='${member.sessionId}'`,
            `origination_callee_id_number='${agent.name}'`,
            `origination_callee_id_name='${agent.name}'`,
            `origination_caller_id_number='${member.number}'`,
            `origination_caller_id_name='${member.name}'`,
            `destination_number='${member.number}'`,
            `effective_caller_id_number='${agent.name}'`,
            `effective_callee_id_number='${member.number}'`,
            'webitel_direction=inbound'
        ];

        const webitelData = {
            dlr_member_id: member._id.toString(),
            dlr_id: member.getQueueId()
        };

        for (let key in this._variables) {
            if (this._variables.hasOwnProperty(key)) {
                agentVars.push(`${key}='${this._variables[key]}'`);
            }
        }
        for (let key of member.getVariableKeys()) {
            webitelData[key] = member.getVariable(key);
            agentVars.push(`'${key}'='${member.getVariable(key)}'`);
        }

        agentVars.push("webitel_data=\\'" + JSON.stringify(webitelData).replace(/\s/g, '\\s') + "\\'");

        application.Esl.bgapi(`uuid_setvar ${member.sessionId} cc_agent ${agent.name}`);

        const start = Date.now();

        member._predAgentOriginate = true;

        const agentDs = `originate {^^${VAR_SEPARATOR}${agentVars.join(VAR_SEPARATOR)}}user/${agent.name} &park()`;
        member.log(`Agent ds: ${agentDs}`);
        application.Esl.bgapi(agentDs, (res) => {
            member.log(`agent ${agent.name} fs res -> ${res.body}`);
            const bgOkData = res.body.match(/^\+OK\s(.*)\n$/);
            const date = Date.now();

            if (member.processEnd) {
                member.agent = null;

                if (bgOkData)
                    application.Esl.bgapi(`uuid_kill ${bgOkData[1]} ${CANCEL_CAUSE}`);

                this._am.setAgentStats(agent, this._objectId, {
                    lastStatus: `error bridge member end`,
                    call: true,
                    process: null
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
                        this._am.setAgentStats(agent, this._objectId, {
                            lastStatus: `error bridge -> ${bridge.body}`,
                            call: true,
                            process: null
                        }, (e) => {
                            if (e)
                                log.error(e);

                            member.agent = null;
                            this._joinAgent(member);
                        });
                    } else {
                        member.predictAbandoned = false;
                        member.bridgedCall = true;
                        member.setBridgedTime();
                        this._am.setAgentStats(agent, this._objectId, {
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
            } else if (/^-ERR|^-USAGE/.test(res.body)) {
                let error =  res.body.replace(/-ERR\s(.*)\n/, '$1');
                member.log(`agent error: ${error}`);

                if (error === 'NO_ANSWER') {
                    if (this.getAgentParam('max_no_answer', agent) <= (this.getAgentParam('no_answer_count', agent) + 1)) {
                        this._am.setNoAnswerAgent(agent, e => {
                            if (e)
                                log.error(e);

                            this._am.setAgentStats(agent, this._objectId, {
                                call: true,
                                bridged: false,
                                lastStatus: `NO_ANSWER -> ${member._id} -> MAX`,
                                connectedTimeSec: timeToSec(date, start),
                                process: null,
                                noAnswer: true,
                                missedCall: true
                            }, (e) => {
                                if (e)
                                    return log.error(e);
                            });
                        });
                    } else {
                        this._am.setAgentStats(agent, this._objectId, {
                            call: true,
                            bridged: false,
                            noAnswer: true,
                            wrapTime: this.getAgentParam('no_answer_delay_time', agent),
                            connectedTimeSec: timeToSec(date, start),
                            lastStatus: `NO_ANSWER -> ${member._id}`,
                            process: null,
                            missedCall: true
                        }, (e) => {
                            if (e)
                                return log.error(e);
                        });
                    }

                } else {
                    this._am.setAgentStats(agent, this._objectId, {
                        call: true,
                        bridged: false,
                        noAnswer: true,
                        wrapTime: this.getAgentParam('reject_delay_time', agent),
                        lastStatus: `REJECT -> ${member._id} -> ${error}`,
                        connectedTimeSec: timeToSec(date, start),
                        process: null,
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
        member.agent = null;
        if (member.getConnectedTime() + (1000 * this._maxLocateAgentSec) <= Date.now() ) {
            application.Esl.bgapi(`uuid_kill ${member.sessionId} LOSE_RACE`); // todo add conf hangup cause
            return;
        }

        if (member._predTimer) {
            return;
            // clearTimeout(member._predTimer);
            // member._predTimer = null;
        }

        this._am.huntingAgent(this,  member, (err, agent) => {
            if (err)
                log.error(err);


            if (!agent) {
                member.log(`no found agent, try now`);
                member._predTimer = setTimeout(() => {
                    member._predTimer = null;
                    this._joinAgent(member);
                }, 1000); // TODO add config
                return;
            } else if (member.processEnd) {
                member.log(`set agent ${agent.name} rollback`);
                this._am.setAgentStats(agent, this._objectId, {
                    lastStatus: `rollback -> ${member._id}`,
                    call: true,
                    process: null
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

        const onChannelPark = (e) => {
            member.setConnectedFlag(true);

            if (this._amd.enabled === true) {
                let amdResult = e.getHeader('variable_amd_result');
                member.log(`amd_result=${amdResult}`);
                // if ( !(amdResult === 'HUMAN' || (this._amd.allowNotSure && amdResult === 'NOTSURE')) ) {
                //     application.Esl.bgapi(`uuid_kill ${member.sessionId} NORMAL_UNSPECIFIED`);
                //     return;
                // }
            }

            if (this._broadcastPlaybackUri) {
                log.trace(`broadcast ${member.sessionId} playback::${this._broadcastPlaybackUri} aleg`);
                application.Esl.bgapi(`uuid_broadcast ${member.sessionId} playback::${this._broadcastPlaybackUri} aleg`);
            }
            member.predictAbandoned = true;
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
                member.log(`set agent ${agent.name} check status`);
                this._am.setAgentStats(agent, this._objectId, {
                    call: true,
                    bridged: true,
                    callTimeSec: +e.getHeader('variable_billsec') || 0,
                    wrapTime: this.getAgentParam('wrap_up_time', agent),
                    lastStatus: `end -> ${member._id}`,
                    process: null
                }, (e) => {
                    if (e)
                        log.error(e);
                });
                member.agent = null;
            }

            //FAH-218
            if (this._amd.enabled && !e.getHeader('variable_amd_result') && +e.getHeader('variable_answer_epoch') > 0) {
                member.end('ORIGINATOR_CANCEL', e);
            } else {
                member.end(e.getHeader('variable_hangup_cause'), e);
            }
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