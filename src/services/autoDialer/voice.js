/**
 * Created by igor on 25.05.16.
 */

let Dialer = require('./dialer'),
    log = require(__appRoot + '/lib/log')(module),
    END_CAUSE = require('./const').END_CAUSE,
    VAR_SEPARATOR = require('./const').VAR_SEPARATOR,
    DIALER_TYPES = require('./const').DIALER_TYPES;

module.exports = class VoiceBroadcast extends Dialer {

    constructor (config, calendarConf, dialerManager) {
        super(DIALER_TYPES.VoiceBroadcasting, config, calendarConf, dialerManager);

        const engine = () => {
            if (this._active < this._limit) {
                this.countAvailableMembers(this._limit, (e, count) => {
                    if (e) {
                        log.error(e);

                    }
                    return this.huntingMember();

                    // let i = count - this._active;
                    //
                    // if (i < 1)
                    //     return;
                    //
                    // while (i--) {
                    //     this.huntingMember();
                    // }
                })
            }
        };

        this.on('wakeUp', () => {
            engine()
        });

        this.members.on('added', (member) => {
            if (member.checkExpire()) {
                member.endCause = END_CAUSE.MEMBER_EXPIRED;
                member.end(END_CAUSE.MEMBER_EXPIRED)
            } else {
                if (member._currentNumber) {
                    this.dialMember(member)
                } else {
                    member.end();
                }
            }

            engine();
        });

        this.members.on('removed', (m) => {
            this.rollback(
                m,
                m.getDestination(),
                {
                    amd: m.getAmdResult(),
                    bridgedCall: m.bridgedCall,
                    connectedCall: m.getConnectedFlag(),
                    waitSec: m.getWaitSec()
                },
                e => {
                    if (!e)
                        this.huntingMember();
                }
            );
        });

        this.getDialString = (member) => {
            const vars = [`presence_data='${member.getDomain()}'`, `cc_queue='${member.getQueueName()}'`, 'ignore_early_media=true', `originate_timeout=${this._originateTimeout}`];

            for (let key in this._variables) {
                if (this._variables.hasOwnProperty(key)) {
                    vars.push(`${key}='${this._variables[key]}'`);
                }
            }

            for (let key of member.getVariableKeys()) {
                vars.push(`${key}='${member.getVariable(key)}'`);
            }

            const dest = member.getDestination();

            const callerIdNumber = member.getCallerIdNumber();

            vars.push(
                `origination_uuid=${member.sessionId}`,
                `dlr_member_id=${member._id.toString()}`,
                `dlr_current_attempt=${member.currentProbe}`,
                `origination_caller_id_number='${callerIdNumber}'`,
                `origination_caller_id_name='${member.getQueueName()}'`,

                `origination_callee_id_number='${member.number}'`,
                `origination_callee_id_name='${member.name}'`,
                `destination_number='${member.number}'`,
                `webitel_direction=dialer`,
                `dlr_queue=${member.getQueueId()}`,
                // `cc_member_session_uuid=${member.sessionId}`,
                `cc_side=member`
            );

            if (member._currentNumber && member._currentNumber.description) {
                vars.push(`dlr_member_number_description='${member.getCurrentNumberDescription()}'`);
            }

            const apps = [];
            if (this._recordSession) {
                vars.push(
                    `RECORD_MIN_SEC=2`,
                    `RECORD_STEREO=false`,
                    `RECORD_BRIDGE_REQ=false`,
                    `recording_follow_transfer=true`
                );
                // ${direction|uuid}
                ///TODO records
                let sessionUri = 'http_cache://$${cdr_url}' +
                    encodeURI(`/sys/formLoadFile?domain=${member.getDomain()}&id=${member.sessionId}&type=mp3&email=none&name=recordSession&.mp3`);

                apps.push(`record_session:${sessionUri}`)
            }

            const gw = dest.gwProto === 'sip' && dest.gwName ? `sofia/gateway/${dest.gwName}/${dest.dialString}` : dest.dialString;
            let dialString = member.number.replace(dest._regexp, gw);

            if (this._amd && this._amd.enabled) {
                vars.push("hangup_after_bridge=true");

                vars.push(`amd_on_human='transfer::${member.getQueueId()} XML dialer ${callerIdNumber} ${member.getQueueName()}'`);
                // vars.push(`amd_on_human='transfer::dialer'`);
                vars.push(`amd_on_machine=hangup::NORMAL_UNSPECIFIED`);
                vars.push(`amd_on_notsure=${this._amd.allowNotSure ? `'transfer::${member.getQueueId()} XML dialer ${callerIdNumber} ${member.getQueueName()}'` : 'hangup::NORMAL_UNSPECIFIED'}`);

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
                apps.push(`transfer:${member.getQueueId()} XML dialer ${callerIdNumber} ${member.getQueueName()}`);
            }

            vars.push(`cc_member_attempt_count=${member.currentProbe}`);
            vars.push(`cc_member_successful_count=${member.successfulCount}`);
            vars.push(`export_vars=cc_member_attempt_count,cc_member_successful_count`);

            return `originate {^^${VAR_SEPARATOR}${vars.join(VAR_SEPARATOR)}}${dialString} '${apps.join(',')}' inline`;
        };

        const handleHangupEvent = (e) => {
            let member = this.members.get(e.getHeader('variable_dlr_member_id'));
            if (member) {
                member.channelsCount--;

                if (member.channelsCount > 0)
                    return;

                if (member.channelsCount < 0) {
                    log.warn(`Member no handle channel_create`);
                    log.warn(member.toJSON());
                }

                if (member.getConnectedTime() > 0 && e.getHeader("Caller-Context") === "dialer") {
                    member.bridgedCall = true;
                    member.setBridgedTime(member.getConnectedTime());
                }

                member.end(e.getHeader('variable_hangup_cause'), e);
            }
        };

        const cr = (e) => {
            let member = this.members.get(e.getHeader('variable_dlr_member_id'));
            if (member) {
                member.channelsCount++;
            }
        };

        application.Esl.on(`esl::event::CHANNEL_HANGUP_COMPLETE::*`, handleHangupEvent);
        application.Esl.on(`esl::event::CHANNEL_CREATE::*`, cr);

        application.Esl.subscribe(['CHANNEL_HANGUP_COMPLETE']);
        application.Esl.subscribe(['CHANNEL_CREATE']);
        this.once('end', () => {
            application.Esl.off(`esl::event::CHANNEL_HANGUP_COMPLETE::*`, handleHangupEvent);
            application.Esl.off(`esl::event::CHANNEL_CREATE::*`, cr);
        });
    }

    dialMember (member) {
        log.trace(`try call ${member.sessionId}`);
        let ds = this.getDialString(member);
        member.log(`dialString: ${ds}`);

        log.trace(`Call ${ds}`);

        application.Esl.bgapi(ds, (res) => {
            member.log(`fs response ${res.body}`);

            if (/^-ERR|^-USAGE/.test(res.body)) {
                member.bridgedCall = false;
                return member.end(res.body.replace(/-ERR\s(.*)\n/, '$1'));
            }
            member.setConnectedFlag(true);
        });
    }
};
