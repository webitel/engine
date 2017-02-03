/**
 * Created by igor on 25.05.16.
 */

let Dialer = require('./dialer'),
    log = require(__appRoot + '/lib/log')(module),
    async = require('async'),
    END_CAUSE = require('./const').END_CAUSE,
    DIALER_TYPES = require('./const').DIALER_TYPES;

module.exports = class Progressive extends Dialer {
    constructor (config, calendarConf, dialerManager) {
        super(DIALER_TYPES.ProgressiveDialer, config, calendarConf, dialerManager);

        this._am = config.agentManager;

        this.members.on('added', (member) => {
            if (member.checkExpire()) {
                member.endCause = END_CAUSE.MEMBER_EXPIRED;
                member.end(END_CAUSE.MEMBER_EXPIRED);
                return;
            }

            dialerManager.agentManager.huntingAgent(config._id, this._agents, this._skills, this.agentStrategy, (err, agent) => {
                if (err)
                    throw err;

                if (agent) {
                    this.dialMember(member, agent);
                } else {
                    member.end(); //TODO
                }
            });

            engine();
        });

        this.members.on('removed', (m) => {
            this.rollback({
                callSuccessful: m.callSuccessful
            }, e => {
                if (!e)
                    engine();
            });
        });

        this.on('availableAgent', a => {
            this.huntingMember();
        });

        const engine = () => {
            async.parallel(
                {
                    agents: (cb) => {
                        dialerManager.agentManager.getAvailableCount(this._objectId, this._agents, this._skills, cb);
                    },
                    members: (cb) => {
                        this.countAvailableMembers(this._limit, cb);
                    }
                },
                (err, res) => {
                    if (err)
                        return log.error(err);

                    if (this._active < this._limit && res.agents > 0 && res.members > 0) {
                        this.huntingMember();
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


        let getMembersFromEvent = (e) => {
            return this.members.get(e.getHeader('variable_dlr_member_id'))
        };

        let onChannelCreate = (e) => {
            let m = getMembersFromEvent(e);
            if (m) {
                m.channelsCount++;
                m.setCallUUID(e.getHeader('variable_uuid'));
            }
        };
        let onChannelDestroy = (e) => {
            let m = getMembersFromEvent(e);
            if (m && --m.channelsCount === 0) {
                m.end(e);

                this._am.setAgentStats(m.agentId, this._objectId, {
                    call: true,
                    gotCall: true, //TODO
                    clearNoAnswer: true,
                    lastBridgeCallTimeEnd: Date.now(),
                    callTimeSec: Math.round(( Date.now() - m.startCall) / 1000),
                    lastStatus: `end -> ${m._id}`,
                    setAvailableTime: Date.now() + (m.wrapUpTime * 1000),
                    process: null
                }, (e, res) => {
                    if (e)
                        return log.error(e);
                });

            }
        };

        this.once('end', () => {
            log.debug('Off channel events');
            application.Esl.off('esl::event::CHANNEL_DESTROY::*', onChannelDestroy);
            application.Esl.off('esl::event::CHANNEL_CREATE::*', onChannelCreate);
        });
        application.Esl.subscribe(['CHANNEL_CREATE', 'CHANNEL_DESTROY']);

        application.Esl.on('esl::event::CHANNEL_DESTROY::*', onChannelDestroy);
        application.Esl.on('esl::event::CHANNEL_CREATE::*', onChannelCreate);

        this.getDialString = (member, agent) => {
            let vars = [
                `dlr_member_id=${member._id.toString()}`,
                `dlr_id=${member._queueId}`,
                `presence_data='${member._domain}'`,
                `cc_queue='${member.queueName}'`
            ];

            for (let key in this._variables) {
                if (this._variables.hasOwnProperty(key)) {
                    vars.push(`${key}='${this._variables[key]}'`);
                }
            }

            if (member._currentNumber && member._currentNumber.description) {
                vars.push(`dlr_member_number_description='${member._currentNumber.description}'`);
            }

            const webitelData = {
                dlr_member_id: member._id.toString(),
                dlr_id: member._queueId
            };

            for (let key of member.getVariableKeys()) {
                webitelData[key] = member.getVariable(key);
                vars.push(`${key}='${member.getVariable(key)}'`);
            }

            vars.push("webitel_data=\\'" + JSON.stringify(webitelData).replace(/\s/g, '\\s') + "\\'");

            vars.push(
                `origination_callee_id_number='${agent.agentId}'`,
                `origination_callee_id_name='${agent.agentId}'`,
                `origination_caller_id_number='${member.number}'`,
                `origination_caller_id_name='${member.name}'`,
                `destination_number='${member.number}'`,
                `originate_timeout=10`, // TODO
                'webitel_direction=outbound'
            );
            return `originate {${vars}}user/${agent.agentId} 'set_user:${agent.agentId},transfer:${member.number}' inline`;
        }

    }

    dialMember (member, agent) {
        log.trace(`try call ${member.sessionId} to ${agent.agentId}`);
        member.log(`set agent: ${agent.agentId}`);
       // member._agent = agent;
        member.agentId = agent.agentId;

        const ds = this.getDialString(member, agent);

        member.log(`dialString: ${ds}`);
        log.trace(`Call ${ds}`);

        // connectedTime = Date
        const start = member.startCall = Date.now();
        member.wrapUpTime = agent.wrapUpTime;

        application.Esl.bgapi(ds, (res) => {
            log.trace(`fs response: ${res && res.body}`);
            const date = Date.now();

            if (/^-ERR|^-USAGE/.test(res.body)) {
                let error =  res.body.replace(/-ERR\s(.*)\n/, '$1');
                member.log(`agent: ${error}`);
                member.minusProbe();
                member.nextTrySec = 1;
                member.end();

                if (error === 'NO_ANSWER') {
                    if (agent.maxNoAnswer <= ++agent.noAnswerCount) {
                        return this._am.setNoAnswerAgent(agent, (e) => {
                            this._am.setAgentStats(agent.agentId, this._objectId, {
                                clearNoAnswer: true,
                                setAvailableTime: null,
                                connectedTimeSec: Math.round( (date - start) / 1000),
                                lastStatus: `NO_ANSWER -> ${member._id} -> MAX`,
                                process: null
                            }, (e, res) => {
                                if (e)
                                    return log.error(e);
                            });
                        });
                    }

                    this._am.setAgentStats(agent.agentId, this._objectId, {
                        noAnswer: true,
                        connectedTimeSec: Math.round( (date - start) / 1000),
                        lastStatus: `NO_ANSWER -> ${member._id}`,
                        setAvailableTime: date + (agent.rejectDelayTime * 1000),
                        process: "checkState"
                    }, (e, res) => {
                        if (e)
                            return log.error(e);
                    });
                } else {
                    this._am.setAgentStats(agent.agentId, this._objectId, {
                        connectedTimeSec: Math.round( (date - start) / 1000),
                        lastStatus: `REJECT -> ${member._id} -> ${error}`,
                        setAvailableTime: date + (agent.rejectDelayTime * 1000),
                        process: "checkState"
                    }, (e, res) => {
                        if (e)
                            return log.error(e);
                    });
                }
            } else {
                this._am.setAgentStats(agent.agentId, this._objectId, {
                    lastBridgeCallTimeStart: date,
                    connectedTimeSec: Math.round( (date - start) / 1000),
                    lastStatus: `active -> ${member._id}`
                }, (e, res) => {
                    if (e)
                        return log.error(e);

                    if (res && res.value) {
                        //member._agent = res.value
                    }
                });
            }
        });

        return;
        let gw = this._gw.fnDialString(member);

        this.findAvailAgents( (agent) => {
            member.log(`set agent: ${agent.id}`);
            let ds = gw(agent, null, null, this.defaultAgentParams);
            member._gw = gw;
            member._agent = agent;

            member.log(`dialString: ${ds}`);
            log.trace(`Call ${ds}`);
            member.inCall = true;
            application.Esl.bgapi(ds, (res) => {
                log.trace(`fs response: ${res && res.body}`);
                if (/^-ERR|^-USAGE/.test(res.body)) {
                    let error =  res.body.replace(/-ERR\s(.*)\n/, '$1');
                    member.log(`agent: ${error}`);
                    member.minusProbe();
                    member.nextTrySec = 1;
                    // TODO ??
                    this.nextTrySec = 0;
                    let delayTime = agent.getTime('rejectDelayTime', this.defaultAgentParams);
                    if (error == 'NO_ANSWER') {
                        agent._noAnswerCallCount++;
                        member._agentNoAnswer = true;
                        delayTime = agent.getTime('noAnswerDelayTime', this.defaultAgentParams);
                    }
                    member.end();
                    this._am.taskUnReserveAgent(agent, delayTime, false, this.defaultAgentParams);
                } else {
                    agent.lastBridgeCallTimeStart = Date.now();
                }
            });
        });

    }
};