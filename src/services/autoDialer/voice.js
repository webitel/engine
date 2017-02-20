/**
 * Created by igor on 25.05.16.
 */

let Dialer = require('./dialer'),
    log = require(__appRoot + '/lib/log')(module),
    END_CAUSE = require('./const').END_CAUSE,
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

        this.on('ready', () => {
            engine()
        });

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
            this.rollback(m, m.getDestination(), e => {
                if (!e)
                    this.huntingMember();
            });
        });
        
        this.getDialString = (member) => {
            const vars = [`presence_data='${member.getDomain()}'`, `cc_queue='${member.getQueueName()}'`, `originate_timeout=${this._originateTimeout}`];

            for (let key in this._variables) {
                if (this._variables.hasOwnProperty(key)) {
                    vars.push(`${key}='${this._variables[key]}'`);
                }
            }

            for (let key of member.getVariableKeys()) {
                vars.push(`${key}='${member.getVariable(key)}'`);
            }

            const dest = member.getDestination();

            vars.push(
                `origination_uuid=${member.sessionId}`,
                `dlr_member_id=${member._id.toString()}`,
                `origination_caller_id_number='${dest.callerIdNumber}'`,
                `origination_caller_id_name='${member.getQueueName()}'`,

                `origination_callee_id_number='${member.number}'`,
                `origination_callee_id_name='${member.name}'`,
                `destination_number='${member.number}'`,

                `dlr_queue=${member.getQueueId()}`
            );

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

            apps.push(`socket:` + '$${acr_srv}');

            return `originate {${vars}}${dialString} '${apps.join(',')}' inline`;
        };

        const handleHangupEvent = (e) => {
            let member = this.members.get(e.getHeader('variable_dlr_member_id'));
            if (member) {
                if (--member.channelsCount !== 0)
                    return;

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

            if (/^-ERR/.test(res.body)) {
                return member.end(res.body.replace(/-ERR\s(.*)\n/, '$1'));
            }
        });
    }
};