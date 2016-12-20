/**
 * Created by igor on 25.05.16.
 */

let log = require(__appRoot + '/lib/log')(module);

class Gw {
    constructor (conf = {}, regex, variables) {
        this.activeLine = 0;
        // TODO link regex...
        this.regex = regex;
        this.maxLines = conf.limit || Infinity;
        this.gwName = conf.gwName;

        this._vars = [];

        if (variables) {
            for (let key in variables)
                this._vars.push(`${key}=${variables[key]}`);
        }
        if (regex)
            this._callerIdNumber = conf.callerIdNumber;

        this.dialString = conf.gwProto == 'sip' && conf.gwName ? `sofia/gateway/${conf.gwName}/${conf.dialString}` : conf.dialString;
    }

    dialAgent (agent) {
        return `originate user/${agent.id} &park()`;
        //originate {dlr_member_id=58187332446277017bc8227b,dlr_id=573dd6bb578ee6e832b47fd6,presence_data='10.10.10.144',cc_queue='Igor',domain_name=10.10.10.144,gatewayPositionMap='0>0',origination_uuid=a185f1b1-2d1e-4f0f-8549-3d18a4a62501,origination_caller_id_name='Igor',origination_callee_id_number='84908031329',origination_callee_id_name='LE VAN GIAL'}sofia/nonreg/sip:AutoDialerTest@pre.webitel.com:5080 &park()
    }

    fnDialString (member) {
        return (agent, sysVars, park, agentParams = {}, dialer = {}) => {
            let vars = [`dlr_member_id=${member._id.toString()}`, `dlr_id=${member._queueId}`, `presence_data='${member._domain}'`, `cc_queue='${member.queueName}'`].concat(this._vars);

            if (sysVars instanceof Array) {
                vars = vars.concat(sysVars);
            }

            if (member._currentNumber && member._currentNumber.description) {
                vars.push(`dlr_member_number_description='${member._currentNumber.description}'`);
            }

            var webitelData = {
                dlr_member_id: member._id.toString(),
                dlr_id: member._queueId
            };
            
            for (let key of member.getVariableKeys()) {
                webitelData[key] = member.getVariable(key);
                vars.push(`${key}='${member.getVariable(key)}'`);
            }

            vars.push("webitel_data=\\'" + JSON.stringify(webitelData).replace(/\s/g, '\\s') + "\\'");

            if (agent) {
                vars.push(
                    `origination_callee_id_number='${agent.id}'`,
                    `origination_callee_id_name='${agent.id}'`,
                    `origination_caller_id_number='${member.number}'`,
                    `origination_caller_id_name='${member.name}'`,
                    `destination_number='${member.number}'`,
                    `originate_timeout=${isFinite(agentParams.callTimeout) ? agentParams.callTimeout :  agent.callTimeout}`,
                    'webitel_direction=outbound'
                );
                return `originate {${vars}}user/${agent.id} 'set_user:${agent.id},transfer:${member.number}' inline`;
            }

            let exportVars = [];
            vars.forEach( i => {
                exportVars.push(i.split('=')[0]);
            });

            vars.push(
                `export_vars='${exportVars}'`,
                `origination_uuid=${member.sessionId}`,
                // `origination_caller_id_number='${member.queueNumber}'`,
                `origination_callee_id_name='${member.queueName}'`
            );

            if (this._callerIdNumber) {
                vars.push(`origination_caller_id_number='${this._callerIdNumber}'`)
            } else {
                vars.push(
                    `origination_caller_id_number='${member.number}'`,
                    `origination_caller_id_name='${member.name}'`
                )
            }

            if (park) {
                const apps = [];

                if (dialer._recordSession) {
                    vars.push(
                        `RECORD_MIN_SEC=2`,
                        `RECORD_STEREO=true`,
                        `RECORD_BRIDGE_REQ=true`,
                        `recording_follow_transfer=true`
                    );

                    let sessionUri = 'http_cache://$${cdr_url}' +
                        encodeURI(`/sys/formLoadFile?domain=${member._domain}&id=${member.sessionId}&type=mp3&email=none&name=recordSession&.mp3`);

                    apps.push(`record_session:${sessionUri}`)
                }

                if (dialer._amd && dialer._amd.enabled) {
                    apps.push(`amd:${dialer._amd._string}`);
                }

                vars.push('ignore_early_media=true');

                apps.push(`park:`);

                let gwString = member.number.replace(this.regex, this.dialString);

                return `originate {${vars}}${gwString} '${apps.join(',')}' inline`;
            } else {
                return `originate {${vars}}loopback/${member.number}/default 'set:dlr_queue=${member._queueId},socket:` + '$${acr_srv}' + `' inline`;
            }

        };
    }

    tryLock (member) {
        if (this.activeLine >= this.maxLines)
            return false;

        this.activeLine++;
        log.debug(`Active line: ${this.dialString} ->> ${this.activeLine}`);

        return this.fnDialString(member)
    }

    unLock () {
        let unLocked = false;
        if (this.activeLine === this.maxLines && this.maxLines !== 0)
            unLocked = true;
        this.activeLine--;
        return unLocked;
    }
}

module.exports = Gw;
