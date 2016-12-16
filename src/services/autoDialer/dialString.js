/**
 * Created by igor on 17.06.16.
 */

module.exports = class DialString {
    constructor (variables) {
        this._vars = [];
        if (variables) {
            for (let key in variables)
                this._vars.push(`${key}=${variables[key]}`);
        }
    }

    get (member, agent) {
        let vars = [`presence_data='${member._domain}'`, `cc_queue='${member.queueName}'`].concat(this._vars);

        for (let key of member.getVariableKeys()) {
            vars.push(`${key}='${member.getVariable(key)}'`);
        }
        vars.push(
            // `origination_uuid=${member.sessionId}`,
            `origination_caller_id_number='${member.queueNumber}'`,
            `origination_caller_id_name='${member.queueName}'`,
            `origination_callee_id_number='${member.number}'`,
            `origination_callee_id_name='${member.name}'`,
            `loopback_bowout_on_execute=true`
        );
        return `originate {${vars}}loopback/${member.number}/default 'set:dlr_member_id=${member._id.toString()},set:dlr_queue=${member._queueId},socket:` + '$${acr_srv}' + `' inline`;
    }
};