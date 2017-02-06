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
            this.rollback({
                callSuccessful: m.callSuccessful
            }, e => {
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
            vars.push(
                // `origination_uuid=${member.sessionId}`,
                `origination_caller_id_number='${member.getQueueNumber()}'`,
                `origination_caller_id_name='${member.getQueueName()}'`,
                `origination_callee_id_number='${member.number}'`,
                `origination_callee_id_name='${member.name}'`,
                `loopback_bowout_on_execute=true`
            );
            return `originate {${vars}}loopback/${member.number}/default 'set:dlr_member_id=${member._id.toString()},set:dlr_queue=${member.getQueueId()},socket:` + '$${acr_srv}' + `' inline`;
        };

        const handleHangupEvent = (e) => {
            let member = this.members.get(e.getHeader('variable_dlr_member_id'));
            if (member) {
                console.log(e.getHeader('variable_uuid'));
                member.channelsCount--;
                member.end(e.getHeader('variable_hangup_cause'), e);
            }
        };

        application.Esl.on(`esl::event::CHANNEL_HANGUP_COMPLETE::*`, handleHangupEvent);

        application.Esl.subscribe(['CHANNEL_HANGUP_COMPLETE']);
        this.once('end', () => {
            application.Esl.off(`esl::event::CHANNEL_HANGUP_COMPLETE::*`, handleHangupEvent);
        });
    }

    dialMember (member) {
        log.trace(`try call ${member.sessionId}`);
        let ds = this.getDialString(member);
        member.log(`dialString: ${ds}`);

        log.trace(`Call ${ds}`);

        application.Esl.bgapi(ds, (res) => {
            if (/^-ERR/.test(res.body)) {
                member.end(res.body.replace(/-ERR\s(.*)\n/, '$1'));
            }
            member.channelsCount++;
        });
    }
};