/**
 * Created by igor on 24.05.16.
 */

'use strict';


const generateUuid = require('node-uuid'),
    log = require(__appRoot + '/lib/log')(module),
    EventEmitter2 = require('eventemitter2').EventEmitter2,
    MEMBER_STATE = require('./const').MEMBER_STATE,
    MEMBER_VARIABLE_CALLER_NUMBER = require('./const').MEMBER_VARIABLE_CALLER_NUMBER,
    END_CAUSE = require('./const').END_CAUSE,
    getHangupCode = require('./const').getHangupCode,
    DIALER_TYPES = require('./const').DIALER_TYPES,
    channelService = require(__appRoot + '/services/channel'),
    VARIABLES = require('./const').VARIABLES
;

const SPAM_VARIABLE = "cc_min_count";

module.exports = class Member extends EventEmitter2 {

    constructor (config, currentNumber, destination, dialer = {}) {
        super();

        this._id = config._id;
        this.expire = config.expire;
        this.timezone = config.timezone; //TODO
        this.sessionId = config._lastSession;
        this.currentProbe = (config._probeCount || 0);
        this.callSuccessful = false;
        this.connectedCall = false;
        this.connectedAgent = false;
        this.successfulCount = (config.successfulCount || 0);

        this.terminateOn = null;

        this.name = ('' + (config.name || "_undef_")).replace(/'/g, '_');
        this.variables = {};

        for (let key in config.variables) {
            this.setVariable(key, config.variables[key]);
        }

        this.getMaxAttemptsCount = () => dialer._maxTryCount;
        this.getIntervalAttemptSec = () => dialer._intervalTryCount;
        this.getMinBillSec = () => dialer._minBillSec;

        this.nextTrySec = dialer._intervalTryCount;

        this.getQueueName = () => dialer.nameDialer;
        this.getQueueNumber = () => dialer.number || this.getQueueName();
        this.getQueueId = () => dialer._id;
        this.getDomain = () => dialer._domain;

        this.getCausesSuccessful = (cause) => dialer._memberOKCauses.indexOf(cause);
        this.getCausesRetry = (cause) => dialer._memberRetryCauses.indexOf(cause);
        this.getCausesError = (cause) => dialer._memberErrorCauses.indexOf(cause);
        this.getCausesMinus = (cause) => dialer._memberMinusCauses.indexOf(cause);

        this.getNoRetryAbandoned = () => !dialer.retryAbandoned;

        this.getDialerType = () => dialer.type;

        this.getDestination = () => destination;

        this.getCallerIdNumber = () => {
            if (this.variables.hasOwnProperty(MEMBER_VARIABLE_CALLER_NUMBER)) {
                return `${this.getVariable(MEMBER_VARIABLE_CALLER_NUMBER)}`
            }
            if (destination._callerIdNumbersArr && destination._callerIdNumbersArr.length > 0) {
                return destination._callerIdNumbersArr[randomInteger(0, destination._callerIdNumbersArr.length - 1)]
            } else {
                return destination.callerIdNumber
            }
        };

        this.getDestinationUuid = () => destination;
        this.channelsCount = 0;
        this._minusProbe = false;
        this.agent = null;

        this._log = {
            session: this.sessionId,
            destinationId: destination.uuid,
            callDescription: "",
            callTime: Date.now(),
            talkSec: 0,
            waitSec: null,
            callSuccessful: false,
            bridgedTime: null,
            connectedTime: null,
            callState: 0,
            callPriority: 0,
            callNumber: null, //+
            callTypeCode: "",
            callTypeName: "",
            callPositionIndex: 0, //+
            cause: null, //+
            causeQ850: null,
            callAttempt: null, // +
            callUUID: null,
            recordSessionSec: 0,
            agentId: null,
            steps: []
        };

        this._waitingForResultStatus = dialer._waitingForResultStatus;

        this.endCause = null;
        // this.bigData = new Array(1e5).join('XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX\n');

        this.log(`create probe ${this.currentProbe} - max ${dialer._maxTryCount}`);
        this.setCurrentAttempt(this.currentProbe);

        this.number = "";
        this._countActiveNumbers = 0;
        this._communications = config.communications;

        this._currentNumber = currentNumber;
        //this._currentNumber.checkResult = this._waitingForResultStatus ? 1 : null;

        this.setCurrentNumber(this._currentNumber, config.communications, dialer);

        if (this._currentNumber) {
            this._currentNumber._probe++;

            if (this._currentNumber.rangeId)
                this._currentNumber.rangeAttempts++;

            this.number = (this._currentNumber.number + '').replace(/\D/g, '');
            this.log(`set number: ${this.number}`);
        } else {
            console.log('ERROR', this);
        }
    }

    getSessionId() {
        return this.sessionId;
    }

    setConnectToAgent() {
        this.connectedAgent = true;
    }

    getConnectToAgent() {
        return this.connectedAgent;
    }

    setCurrentAttempt (attempt) {
        this._log.callAttempt = attempt;
    }

    setConnectedFlag (val = true) {
        this.connectedCall = val;
        if (val) {
            this.setConnectedTime()
        }
    }

    getConnectedFlag () {
        return this.connectedCall
    }

    setTalkSec (e, useAmd) {
        this._log.talkSec = 0;

        const answerTime = getIntValueFromEventHeader(e, 'variable_answer_epoch');
        if (answerTime === 0)
            return this._log.talkSec;

        if (this.getDialerType() === DIALER_TYPES.VoiceBroadcasting) {
            this._log.talkSec = getIntValueFromEventHeader(e, 'variable_end_epoch') - answerTime;
            if (useAmd && this._log.talkSec > 0) {
                this._log.talkSec -= getIntValueFromEventHeader(e, 'variable_amd_result_epoch') - answerTime;
            }
        } else {
            const bridgeEpoch = getIntValueFromEventHeader(e, 'variable_bridge_epoch');
            if (bridgeEpoch > 0)
                this._log.talkSec = getIntValueFromEventHeader(e, 'variable_end_epoch') - getIntValueFromEventHeader(e, 'variable_bridge_epoch');
        }
        if (this.getDialerType() !== DIALER_TYPES.ProgressiveDialer) {
            if (this._log.bridgedTime) {
                this._log.waitSec = Math.round((this._log.bridgedTime - this._log.connectedTime) / 1000);
            } else if (this._log.connectedTime && this._log.amdResult !== "MACHINE") {
                this._log.waitSec = Math.round((Date.now() - this._log.connectedTime) / 1000);
            }
        }
        return this._log.talkSec;
    }

    getTalkSec () {
        return this._log.talkSec;
    }

    getWaitSec () {
        return this._log.waitSec;
    }

    setAmdResult (result, cause) {
        this.log(`Set amd ${result} - ${cause}`);
        this._log.amdResult = result;
        this._log.amdCause = cause;
    }

    getAmdResult () {
        if (this._log.amdResult) {
            return {
                result: this._log.amdResult,
                cause: this._log.amdCause
            }
        }
        return null;
    }

    setConnectedTime () {
        this._log.connectedTime = Date.now();
    }

    getConnectedTime() {
        return this._log.connectedTime;
    }

    setBridgedTime (time) {
        this._log.bridgedTime = time || Date.now()
    }

    getBridgedTime () {
        return this._log.bridgedTime
    }

    startCall () {
        this._log.callTime = Date.now()
    }

    setAgent (agent = {}) {
        this.log(`set agent: ${agent.name}`);
        this._log.agentId = agent.name;
        //todo
        this.agent = agent;
    }

    getAgent () {
        return this.agent;
    }

    setCurrentNumber (communication, all, queue) {
        if (!communication)
            return log.warn(`No communication in ${this._id}`);

        this._log.callNumber = communication.number;
        this._log.callPriority = communication.priority || 0;
        this._log.callDescription = communication.description || "";
        if (communication.type) {
            this._log.callTypeCode = communication.type;
            this._log.callTypeName = queue.getCommunicationTypeName(communication.type);

            this.setVariable(`${VARIABLES.COMMUNICATION_TYPE_NAME}`, this._log.callTypeName);
            this.setVariable(`${VARIABLES.COMMUNICATION_TYPE_CODE}`, this._log.callTypeCode);
        }

        // const $set = {};
        // $set[`communications.${communication._id}`] = communication;

        // dialerService.members._updateByIdFix(
        //     this._id,
        //     {$set},
        //     (err, res) => {
        //         if (err)
        //             log.error(err);
        //
        //     }
        // );

        if (all instanceof Array) {
            for (let num of all) {
                if (num.state === MEMBER_STATE.Idle) {
                    this._countActiveNumbers++
                }
            }
            this._log.callPositionIndex = all.indexOf(communication);
        }
    }

    checkExpire () {
        return this.expire && Date.now() >= this.expire;
    }

    minusProbe () {
        if (this._currentNumber) {
            this._currentNumber._probe--;
            if (this._currentNumber.rangeId && isFinite(this._currentNumber.rangeAttempts)) {
                this._currentNumber.rangeAttempts--;
            }
        }

        this.currentProbe--;
        this._minusProbe = true;
        this.log(`minus probe: ${this.currentProbe}`);
    }

    setVariable (varName, value) {
        this.variables[varName] = value;
        return true
    }

    getVariable (varName) {
        return this.variables[varName] && (this.variables[varName] + '').replace(/\'|\r|\n|\t|\\/g, ``);
    }

    getCurrentNumberDescription () {
        return this._currentNumber.description && (this._currentNumber.description + '').replace(/\'|\r|\n|\t|\\/g, ``);
    }

    getVariableKeys () {
        return Object.keys(this.variables);
    }

    log (str) {
        log.trace(this._id + ' -> ' + str);
        this._log.steps.push({
            time: Date.now(),
            data: str
        });
    }

    setRecordSession (sec) {
        this._log.recordSessionSec = sec;
    }

    setProbeEndCause (cause) {
        this._log.cause = cause;
        this._log.causeQ850 = getHangupCode(cause)
    }

    setProbeQ850Code (code) {
        if (+code)
            this._log.causeQ850 = +code;
    }

    setCallUUID (uuid) {
        this.log(`set uuid ${uuid}`);
        this._log.callUUID = uuid;
    }

    _setStateCurrentNumber (state) {
        if (!this._currentNumber)
            return;
        this._currentNumber.state = state;
        this._log.callState = state;
    }

    toJSON () {
        let e = {
            "Event-Name": "CUSTOM",
            "Event-Subclass": "engine::dialer_member_end",
            "variable_domain_name": this.getDomain(),
            "dialerId": this.getQueueId(),
            "dialerName": this.getQueueName(),
            "id": this._id.toString(),
            "name": this.name,
            "currentNumber": this._currentNumber && this._currentNumber.number,
            "dlr_member_number_description": this._currentNumber && this._currentNumber.description,
            "currentProbe": this.currentProbe,
            "session": this.sessionId,
            "endCause": this._endCause || this.endCause
        };
        if (this.expire)
            e.expire = this.expire;

        if (this.agent) {
            e.agentId = this.agent.name;
        }

        for (let key in this.variables) {
            if (this.variables.hasOwnProperty(key))
                e[`variable_${key}`] = this.variables[key]
        }
        return e;
    }

    broadcast () {
        // console.log(this.toJSON());
        application.Broker.publish(application.Broker.Exchange.FS_EVENT, `.CUSTOM.engine%3A%3Adialer_member_end..`, this.toJSON());
    }

    setCallSuccessful (val) {
        this.callSuccessful = val; // TODO delete
        this._log.callSuccessful = val;
    }

    setTerminate(caller, cb) {
        this.terminateOn = caller;
        channelService.hupByVariable('dlr_member_id',this._id.toString(), END_CAUSE.MANAGER_REQUEST, (err) => {
            if (err) {
                return cb(err)
            }

            return cb(null, {
                _id: this._id,
                waitHangupCause: END_CAUSE.MANAGER_REQUEST
            })
        });
    }

    end (endCause, e) {

        if (this.processEnd) return;

        if (endCause)
            endCause = endCause.trim();

        this.processEnd = true;

        let spamCount = 0;

        log.trace(`end member ${this._id} cause: ${this.endCause || endCause || ''}`) ;

        if (e) {
            const recordSec = +e.getHeader('variable_record_seconds');
            if (recordSec)
                this.setRecordSession(recordSec);

            if (e.getHeader('variable_amd_result')) {
                this.setAmdResult(e.getHeader('variable_amd_result'), e.getHeader('variable_amd_cause'));
                this.setTalkSec(e, true);
            } else {
                this.setTalkSec(e, false);
            }

            //this.setProbeQ850Code(e.getHeader('variable_hangup_cause_q850'));
            this.setCallUUID(e.getHeader('variable_uuid'));
            spamCount = (+e.getHeader(`variable_${SPAM_VARIABLE}`))
        }

        if (endCause === END_CAUSE.MANAGER_REQUEST && this.terminateOn) {
            this.log(`User ${this.terminateOn} terminate call`);
            this._setStateCurrentNumber(MEMBER_STATE.End);
            this.endCause = END_CAUSE.MANAGER_REQUEST;
            this.predictAbandoned = true;
            this.emit('end', this);
            return
        }

        if (this.predictAbandoned) {

            this.setProbeEndCause(END_CAUSE.ABANDONED);

            this.callSuccessful = false;
            if (this.getNoRetryAbandoned()) {
                this.log(`Abandoned`);
                this._setStateCurrentNumber(MEMBER_STATE.End);
                this.endCause = END_CAUSE.ABANDONED;
                this.emit('end', this);
                return;
            }
        } else {
            this.setProbeEndCause(endCause);
        }

        if (this._waitingForResultStatus && endCause !== END_CAUSE.MEMBER_EXPIRED && this.bridgedCall === true) {
            this.nextTime = Date.now() + (this.nextTrySec * 1000);
            this.log(`Check callback`);
            this.emit('end', this);
            return;
        }

        let skipOk = false;

        if (~this.getCausesMinus(endCause)) {
            this.minusProbe();
            this.log(`end cause ${endCause}`);
            this.nextTime = Date.now() + (this.nextTrySec * 1000);
            this._setStateCurrentNumber(MEMBER_STATE.Idle);
            this.emit('end', this);
            return;
        }

        //TODO
        if (~this.getCausesSuccessful(endCause) && (this.bridgedCall || this.getDialerType() === DIALER_TYPES.VoiceBroadcasting)) {
            if (this.getTalkSec() >= this.getMinBillSec()) {
                this.successfulCount++;

                if (spamCount > 0 && this.successfulCount < spamCount) {
                    this.log(`check spam count ${this.successfulCount}`);
                    skipOk = true;
                } else {
                    this.endCause = endCause;
                    this.log(`OK: ${endCause}`);
                    this.setCallSuccessful(true);
                    this._setStateCurrentNumber(MEMBER_STATE.End);
                    this.emit('end', this);
                    return
                }
            } else {
                skipOk = true;
                this.log(`skip ok: talk sec ${this.getTalkSec()}`)
            }

        }

        if (~this.getCausesRetry(endCause) || skipOk || this.predictAbandoned) {
            if (this.currentProbe >= this.getMaxAttemptsCount()) {
                this.log(`max try count`);
                this.endCause = END_CAUSE.MAX_TRY;
                this._setStateCurrentNumber(MEMBER_STATE.End);
            } else {
                this.nextTime = Date.now() + (this.nextTrySec * 1000);
                this.log(`min next time: ${this.nextTime}`);
                this.log(`Retry: ${endCause}`);
                this._setStateCurrentNumber(MEMBER_STATE.Idle);
            }

            this.emit('end', this);
            return;
        }

        if (~this.getCausesError(endCause)) {
            this.log(`fatal: ${endCause}`);
            this._setStateCurrentNumber(MEMBER_STATE.End);
            if (this._countActiveNumbers === 1 && endCause)
                this.endCause = endCause;
        }


        if (this.currentProbe >= this.getMaxAttemptsCount()) {
            this.log(`max try count`);
            this.endCause = END_CAUSE.MAX_TRY;
            this._setStateCurrentNumber(MEMBER_STATE.End)
        } else {
            // if (this._countActiveNumbers === 1 && endCause)
            //     this.endCause = endCause;

            if (!this._minusProbe)
                this.nextTime = Date.now() + (this.nextTrySec * 1000);
        }
        this.log(`end cause: ${endCause || ''}`);
        this.emit('end', this);
    }
};


const keySort = function(arr = [], keys) {

    keys = keys || {};

    const sortFn = function(a, b) {
        let sorted = 0, ix = 0;

        while (sorted === 0 && ix < KL) {
            let k = obIx(keys, ix);
            if (k) {
                let dir = keys[k];
                sorted = _keySort(a[k], b[k], dir);
                ix++;
            }
        }
        return sorted;
    };

    const obIx = function(obj, ix){
        return Object.keys(obj)[ix];
    };

    const _keySort = function(a, b, d) {
        d = d !== null ? d : 1;
        // a = a.toLowerCase(); // this breaks numbers
        // b = b.toLowerCase();
        if (a == b)
            return 0;
        return a > b ? 1 * d : -1 * d;
    };

    const KL = Object.keys(keys).length;

    if (!KL)
        return arr.sort(sortFn);

    for ( let k in keys) {
        // asc unless desc or skip
        keys[k] =
            keys[k] == 'desc' || keys[k] == -1  ? -1
                : (keys[k] == 'skip' || keys[k] === 0 ? 0
                : 1);
    }
    arr = arr.sort(sortFn);
    return arr;
};


function getIntValueFromEventHeader(e, name) {
    return +e.getHeader(name) || 0;
}

function randomInteger(min, max) {
    return Math.round( min - 0.5 + Math.random() * (max - min + 1) );
}
