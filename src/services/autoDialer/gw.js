/**
 * Created by igor on 25.05.16.
 */

let log = require(__appRoot + '/lib/log')(module);

class Gw {
    constructor (conf, regex, variables) {
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

        this.dialString = conf.gwProto == 'sip' && conf.gwName ? `sofia/gateway/${conf.gwName}/${conf.dialString}` : conf.dialString;
    }

    dialAgent (agent) {
        return `originate user/${agent.id} &park()`;
    }

    fnDialString (member) {
        return (agent, sysVars, park) => {
            let vars = [`dlr_member_id=${member._id.toString()}`, `cc_queue='${member.queueName}'`].concat(this._vars);

            if (sysVars instanceof Array) {
                vars = vars.concat(sysVars);
            }

            var webitelData = {};
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
                    `originate_timeout=${agent.callTimeout}`,
                    'webitel_direction=outbound'
                );
                return `originate {${vars}}user/${agent.id} 'set_user:${agent.id},transfer:${member.number}' inline`;
            }

            vars.push(
                `origination_uuid=${member.sessionId}`,
                // `origination_caller_id_number='${member.queueNumber}'`,
                `origination_caller_id_name='${member.queueName}'`,
                `origination_callee_id_number='${member.number}'`,
                `origination_callee_id_name='${member.name}'`
            );

            let gwString = member.number.replace(this.regex, this.dialString);
            if (park) {
                return `originate {${vars}}${gwString} &park()`;
            } else {
                return `originate {${vars}}loopback/${member.number}/default 'set:dlr_queue=${member._queueId},socket:` + '$${acr_srv}' + `' inline`;
                // vars.push(`dlr_queue=${member._queueId}`);
                // return `originate {${vars}}${gwString} ` +  '&socket($${acr_srv})';
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
