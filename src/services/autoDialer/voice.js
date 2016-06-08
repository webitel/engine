/**
 * Created by igor on 25.05.16.
 */

let Dialer = require('./dialer'),
    Router = require('./router'),
    log = require(__appRoot + '/lib/log')(module),

    DIALER_TYPES = require('./const').DIALER_TYPES;

module.exports = class VoiceBroadcast extends Dialer {

    constructor (config, calendarConf) {
        super(DIALER_TYPES.VoiceBroadcasting, config, calendarConf);

        [this._originateTimeout = 60] = [config.parameters && config.parameters.originateTimeout];

        this._variables = Object.assign(this._variables, {
            'originate_timeout': this._originateTimeout
        });
        this._router = new Router(config.resources, this._variables);

        if (this._limit > this._router._limit) {
            log.warn(`skip dialer limit, max resources ${this._router._limit}`);
            this._limit = this._router._limit;
        }



        // console.log(this);
    }

    dialMember (member) {

        log.trace(`try call ${member.sessionId}`);

        let gw = this._router.getDialStringFromMember(member);

        if (gw.found) {
            if (gw.dialString) {
                let ds = gw.dialString();
                member.log(`dialString: ${ds}`);

                member.once('end', () => {
                    this._router.freeGateway(gw);
                });

                let onChannelHangup = (e) => {
                    member.channelsCount--;
                    member.end(e.getHeader('variable_hangup_cause'), e);
                };

                member.offEslEvent = function () {
                    application.Esl.off(`esl::event::CHANNEL_HANGUP_COMPLETE::${member.sessionId}`, onChannelHangup);
                };

                application.Esl.once(`esl::event::CHANNEL_HANGUP_COMPLETE::${member.sessionId}`, onChannelHangup);

                log.trace(`Call ${ds}`);

                application.Esl.bgapi(ds, (res) => {

                    if (/^-ERR/.test(res.body)) {
                        member.offEslEvent();
                        let error =  res.body.replace(/-ERR\s(.*)\n/, '$1');
                        member.end(error);
                    }
                    member.channelsCount++;
                });
            } else {
                member.minusProbe();
                this.nextTrySec = 0;
                member.end();
            }

        } else {
            member.end(gw.cause);
        }
    }
};