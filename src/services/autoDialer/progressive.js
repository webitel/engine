/**
 * Created by igor on 25.05.16.
 */

let Dialer = require('./dialer'),
    log = require(__appRoot + '/lib/log')(module),
    async = require('async'),
    END_CAUSE = require('./const').END_CAUSE,
    VAR_SEPARATOR = require('./const').VAR_SEPARATOR,
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
                dialerManager.agentManager.huntingAgent(this, member, (err, agent) => {
                    if (err)
                        log.error(err);

                    if (agent) {
                        this.dialMember(member, agent);
                    } else {
                        member.log(`No found agent!!!`);
                        member.minusProbe();
                        member.end();
                    }
                });
            }

            engine();
        });

        this.members.on('removed', (m) => {
            this.rollback(m, m.getDestination(), {
                bridgedCall: m.bridgedCall,
                connectedCall: m.getConnectedFlag()
            }, e => {
                if (!e)
                    engine();
            });
        });

        this.on('availableAgent', a => {
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

                    if (this._active < this._limit && res.agents > 0 && res.members > 0 && this._active < res.allLogged) {
                        this.huntingMember();
                    } else if (this.members.length() === 0) {
                        this.tryStop();
                    } else {
                        tryNow()
                    }
                }
            );
        };

        this.on('wakeUp', () => {
            engine()
        });

        application.Esl.subscribe(['CHANNEL_DESTROY']);

        this.getDialString = (member, agent) => {
            let vars = [
                `dlr_member_id='${member._id.toString()}'`,
                `dlr_session_id='${member.getSessionId()}'`,
                `dlr_current_attempt=${member.currentProbe}`,
                `dlr_id=${member.getQueueId()}`,
                `presence_data='${member.getDomain()}'`,
                `cc_queue='${member.getQueueName()}'`
            ];

            const webitelData = {
                dlr_member_id: member._id.toString(),
                dlr_id: member.getQueueId(),
                ...this.getDescriptionMapping()
            };

            for (let key in this._variables) {
                if (this._variables.hasOwnProperty(key)) {
                    vars.push(`${key}='${this._variables[key]}'`);
                    webitelData[key] = this._variables[key];
                }
            }

            if (member._currentNumber && member._currentNumber.description) {
                vars.push(`dlr_member_number_description='${member.getCurrentNumberDescription()}'`);
            }

            for (let key of member.getVariableKeys()) {
                webitelData[key] = (member.getVariable(key) || "").replace(/\\'/g, '');
                vars.push(`'${key}'='${member.getVariable(key)}'`);
            }

            vars.push("webitel_data=\\'" + JSON.stringify(webitelData).replace(/\s/g, '\\s') + "\\'");

            const dest = member.getDestination();

            const apps = [];
            if (this._recordSession) {
                vars.push(
                    `RECORD_MIN_SEC=2`,
                    `RECORD_STEREO=${this.recordStereo()}`,
                    `RECORD_BRIDGE_REQ=true`,
                    `recording_follow_transfer=true`
                );
            }

            const gw = dest.gwProto === 'sip' && dest.gwName ? `sofia/gateway/${dest.gwName}/${dest.dialString}` : dest.dialString;
            let dialString = member.number.replace(dest._regexp, gw);

            apps.push(`bridge:\\'{bridge_early_media=true,cc_side=member,originate_timeout=${this._originateTimeout},ignore_display_updates=true,origination_callee_id_number=${member.number},origination_caller_id_number='${member.getCallerIdNumber()}'}${dialString}\\'`);

            vars.push(
                `originate_timeout=${this.getAgentOriginateTimeout(agent)}`,
                'webitel_direction=outbound',
                `cc_agent=${agent.name}`,
                `cc_side=agent`
            );
            return `originate {^^${VAR_SEPARATOR}${vars.join(VAR_SEPARATOR)}}user/${agent.name} 'set_user:${agent.name},${apps.join(',')}' inline default '${member.name}' '${member.number}'`;
        }
    }

    dialMember (member, agent) {
        log.trace(`try call ${member.sessionId} to ${agent.name}`);
        member.setAgent(agent);

        const ds = this.getDialString(member, agent);

        member.log(`dialString: ${ds}`);
        log.trace(`Call ${ds}`);

        const start = Date.now();
        let channelUuid = null;

        const onChannelDestroy = (e) => {
            if (+e.getHeader("variable_bridge_uepoch")) {
                member.bridgedCall = true;
                member.setBridgedTime(Math.round(+e.getHeader("variable_bridge_uepoch") / 1000));
            }

            member.end(e.getHeader('variable_hangup_cause'), e);

            this._am.setAgentStats(agent, this._objectId, {
                call: true,
                bridged: true,
                callTimeSec: +e.getHeader('variable_billsec') || 0,
                wrapTime: this.getAgentParam('wrap_up_time', agent),
                processing: member._processingSeconds > 0 ? member._processingSeconds : null,
                lastStatus: `end -> ${member._id}`,
                process: null
            }, (e) => {
                if (e)
                    return log.error(e);
            });
        };

        member.once('end', () => {
            if (channelUuid)
                application.Esl.off(`esl::event::CHANNEL_DESTROY::${channelUuid}`, onChannelDestroy);
        });

        member.channelsCount++;
        application.Esl.bgapi(ds, (res) => {
            log.trace(`fs response: ${res && res.body}`);
            const date = Date.now();

            const bgOkData = res.body.match(/^\+OK\s(.*)\n$/);

            if (bgOkData) {
                member.setConnectedFlag(true);
                member.setConnectToAgent();

                channelUuid = bgOkData[1];
                member.setCallUUID(channelUuid);
                application.Esl.on(`esl::event::CHANNEL_DESTROY::${channelUuid}`, onChannelDestroy);

                if (this._recordSession) {
                    application.Esl.bgapi(`uuid_record ${channelUuid} start http_cache://${application._storageUri}` +
                        encodeURI(`/sys/formLoadFile?domain=${member.getDomain()}&id=${channelUuid}&type=mp3&email=none&name=recordSession&.mp3`)
                        , res => {
                            log.trace(`Response uuid_record ${channelUuid} : ${res.body}`);
                        });
                }

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
            } else if (/^-ERR|^-USAGE/.test(res.body)) {
                let error =  res.body.replace(/-ERR\s(.*)\n/, '$1');
                member.log(`agent error: ${error}`);


                member.minusProbe();
                member.nextTrySec = 1;
                member.end(error === END_CAUSE.MANAGER_REQUEST ? END_CAUSE.MANAGER_REQUEST : undefined);

                if (error === 'NO_ANSWER') {
                    if (this.getAgentParam('max_no_answer', agent) > 0 && this.getAgentParam('max_no_answer', agent) <= (this.getAgentParam('no_answer_count', agent) + 1)) {
                        return this._am.setNoAnswerAgent(agent, e => {
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
                    }

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
            } else {
                log.error(res.body);
            }
        });

    }
};

function timeToSec(current, start) {
    return Math.round( (current - start) / 1000 )
}
