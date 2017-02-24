/**
 * Created by igor on 25.05.16.
 */

let Dialer = require('./dialer'),
    log = require(__appRoot + '/lib/log')(module),
    async = require('async'),
    END_CAUSE = require('./const').END_CAUSE,
    AGENT_STATUS = require('./const').AGENT_STATUS,
    DIALER_TYPES = require('./const').DIALER_TYPES;

module.exports = class Progressive extends Dialer {
    constructor (config, calendarConf, dialerManager) {
        super(DIALER_TYPES.ProgressiveDialer, config, calendarConf, dialerManager);

        this._am = config.agentManager;

        this.members.on('added', (member) => {
            if (member.checkExpire()) {
                member.endCause = END_CAUSE.MEMBER_EXPIRED;
                member.end(END_CAUSE.MEMBER_EXPIRED);
            } else {
                dialerManager.agentManager.huntingAgent(config._id, this._agents, this._skills, this.agentStrategy, (err, agent) => {
                    if (err)
                        throw err;

                    if (agent) {
                        this.dialMember(member, agent);
                    } else {
                        member.end(); //TODO
                    }
                });
            }

            engine();
        });

        this.members.on('removed', (m) => {
            this.rollback(m, m.getDestination(), null, e => {
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
                    } else if (this.members.length() === 0) {
                        this.tryStop();
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
                m.end(e.getHeader('variable_hangup_cause'), e);
                const agent = m.getAgent();

                this._am.setAgentStats(agent.agentId, this._objectId, {
                    call: true,
                    gotCall: true, //TODO
                    clearNoAnswer: true,
                    lastBridgeCallTimeEnd: Date.now(),
                    callTimeSec: +e.getHeader('variable_billsec') || 0,
                    lastStatus: `end -> ${m._id}`,
                    setAvailableTime:
                        agent.status === AGENT_STATUS.AvailableOnDemand ? null : Date.now() + (this.getAgentParam('wrapUpTime', agent) * 1000),
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

            const webitelData = {
                dlr_member_id: member._id.toString(),
                dlr_id: member.getQueueId()
            };

            for (let key of member.getVariableKeys()) {
                webitelData[key] = member.getVariable(key);
                vars.push(`${key}='${member.getVariable(key)}'`);
            }

            vars.push("webitel_data=\\'" + JSON.stringify(webitelData).replace(/\s/g, '\\s') + "\\'");

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
            let dialString = member.number.replace(dest._regexp, gw);

            apps.push(`bridge:\\'{cc_side=member,origination_caller_id_number='${dest.callerIdNumber}'}${dialString}\\'`);

            vars.push(
                `origination_callee_id_number='${agent.agentId}'`,
                `origination_callee_id_name='${agent.agentId}'`,
                `origination_caller_id_number='${member.number}'`,
                `origination_caller_id_name='${member.name}'`,
                `destination_number='${member.number}'`,
                `originate_timeout=${this.getAgentParam('callTimeout', agent)}`,
                'webitel_direction=outbound',
                `cc_side=agent`
            );
            return `originate {${vars}}user/${agent.agentId} 'set_user:${agent.agentId},${apps.join(',')}' inline`;
        }
    }

    dialMember (member, agent) {
        log.trace(`try call ${member.sessionId} to ${agent.agentId}`);
        member.setAgent(agent);

        const ds = this.getDialString(member, agent);

        member.log(`dialString: ${ds}`);
        log.trace(`Call ${ds}`);

        const start = Date.now();

        application.Esl.bgapi(ds, (res) => {
            log.trace(`fs response: ${res && res.body}`);
            const date = Date.now();

            if (/^-ERR|^-USAGE/.test(res.body)) {
                let error =  res.body.replace(/-ERR\s(.*)\n/, '$1');
                member.log(`agent error: ${error}`);
                
                
                member.minusProbe();
                member.nextTrySec = 1;
                member.end();

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
            } else {
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
};

function timeToSec(current, start) {
    return Math.round( (current - start) / 1000 )
}