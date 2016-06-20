/**
 * Created by igor on 25.05.16.
 */

let Dialer = require('./dialer'),
    DialString = require('./dialString'),
    log = require(__appRoot + '/lib/log')(module),

    DIALER_TYPES = require('./const').DIALER_TYPES;

module.exports = class VoiceBroadcast extends Dialer {

    constructor (config, calendarConf) {
        super(DIALER_TYPES.VoiceBroadcasting, config, calendarConf);

        [this._originateTimeout = 60] = [config.parameters && config.parameters.originateTimeout];

        this._variables = Object.assign(this._variables, {
            'originate_timeout': this._originateTimeout
        });

        this._gw = new DialString(this._variables);

        let handleHangupEvent = (e) => {
            let member = this.members.get(e.getHeader('variable_dlr_member_id'));
            if (member) {
                member.channelsCount--;
                member.end(e.getHeader('variable_hangup_cause'), e);
            }
        };

        application.Esl.on(`esl::event::CHANNEL_HANGUP_COMPLETE::*`, handleHangupEvent);

        this.once('end', () => {
            application.Esl.off(`esl::event::CHANNEL_HANGUP_COMPLETE::*`, handleHangupEvent);
        });

        // console.log(this);
    }

    dialMember (member) {
        log.trace(`try call ${member.sessionId}`);

        let ds = this._gw.get(member);
        member.log(`dialString: ${ds}`);

        log.trace(`Call ${ds}`);

        application.Esl.bgapi(ds, (res) => {
            if (/^-ERR/.test(res.body)) {
                member.offEslEvent();
                let error =  res.body.replace(/-ERR\s(.*)\n/, '$1');
                member.end(error);
            }
            member.channelsCount++;
        });
    }
};