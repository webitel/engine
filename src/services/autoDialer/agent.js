/**
 * Created by igor on 24.05.16.
 */

'use strict';

let log = require(__appRoot + '/lib/log')(module),
    AGENT_STATE = require('./const').AGENT_STATE,
    AGENT_STATUS = require('./const').AGENT_STATUS
    ;

class Agent {

    constructor (key, params, skills, position) {
        this.id = key;
        this.number = key.split('@')[0];
        this.state = params.state;
        this.status = params.status;
        this.skills = params.skills;
        // this.bigData = new Array(1e6).join('XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX\n');
        this.maxNoAnswer = +(params.max_no_answer || 0);
        this.noAnswerDelayTime = +(params.no_answer_delay_time || 0);
        this.wrapUpTime = +(params.wrap_up_time || 10);
        this.rejectDelayTime = +(params.reject_delay_time || 0);
        let cTimeout = params.contact && /originate_timeout=([^,|}]*)/.exec(params.contact);
        this.callTimeout = +(cTimeout && cTimeout[1]) || 10;
        [this.skills = ''] = [skills];
        this.dialers = [];
        this.lockTime = 0;
        this.unIdleTime = 0;
        this.availableTime = Infinity;
        this.lock = false;
        this.timerId = null;
        this.callCount = 0;
        this.lastBridgeCallTimeStart = null;
        this.lastBridgeCallTimeEnd = null;
        this.callTimeMs = 0;

        this.position = position;

        this._noAnswerCallCount = 0;
    }

    getSkills () {}
    setSkills () {}

    setState (state = '') {
        this.state = state;
        if (this.state === AGENT_STATE.Waiting && (this.status === AGENT_STATUS.Available || this.status === AGENT_STATUS.AvailableOnDemand)) {
            log.info(`new time agent`);
            if (!this.lock)
                this.availableTime = Date.now()
            ;
        }

        log.trace(`Change agent state ${this.id} ${this.state} ${this.status} ${this.lock}`);
    }

    setStatus (status = '') {
        this.status = status;
        if (this.state === AGENT_STATE.Waiting && (this.status === AGENT_STATUS.Available || this.status === AGENT_STATUS.AvailableOnDemand)) {
            log.info(`new time agent`);
            if (!this.lock)
                this.availableTime = Date.now()
            ;
        }
        log.trace(`Change agent status ${this.id} ${this.state} ${this.status} ${this.lock}`);
    }

    addDialer (dialerId) {
        let id = this.dialers.indexOf(dialerId);
        if (!~id)
            this.dialers.push(dialerId);
    }

    removeDialer (dialerId) {
        let id = this.dialers.indexOf(dialerId);
        if (~id)
            this.dialers.splice(id, 1);
    }
}

module.exports = Agent;