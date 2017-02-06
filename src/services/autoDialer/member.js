/**
 * Created by igor on 24.05.16.
 */

'use strict';


let generateUuid = require('node-uuid'),
    log = require(__appRoot + '/lib/log')(module),
    EventEmitter2 = require('eventemitter2').EventEmitter2,
    MEMBER_STATE = require('./const').MEMBER_STATE,
    END_CAUSE = require('./const').END_CAUSE
    ;

module.exports = class Member extends EventEmitter2 {

    constructor (config, dialer = {}) {
        super();

        this._id = config._id;
        this.expire = config.expire;
        this.timezone = config.timezone; //TODO
        this.sessionId = generateUuid.v4();
        this.currentProbe = (config._probeCount || 0) + 1;
        this.callSuccessful = false;

        this.name = (config.name || "_undef_").replace(/'/g, '_');
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

        // let lockedGws = option.lockedGateways; // TODO

        this.getCausesSuccessful = (cause) => dialer._memberOKCauses.indexOf(cause);
        this.getCausesRetry = (cause) => dialer._memberRetryCauses.indexOf(cause);
        this.getCausesError = (cause) => dialer._memberErrorCauses.indexOf(cause);
        this.getCausesMinus = (cause) => dialer._memberMinusCauses.indexOf(cause);

        this.channelsCount = 0;
        this._minusProbe = false;

        this._log = {
            session: this.sessionId,
            callTime: Date.now(),
            callSuccessful: false,
            callState: 0,
            callPriority: 0,
            callNumber: null, //+
            callPositionIndex: 0, //+
            cause: null, //+
            callAttempt: null, // +
            callUUID: null,
            recordSessionSec: 0,
            steps: []
        };

        this._waitingForResultStatus = dialer._waitingForResultStatus;

        // if (config._waitingForResultStatus && !config._lastMinusProbe) {
        //     // minus prev probe (no callback)
        //     this.log(`No prev call result status`);
        //     this.currentProbe--;
        // }

        this.endCause = null;
        // this.bigData = new Array(1e5).join('XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX\n');

        this.log(`create probe ${this.currentProbe}`);
        this.setCurrentAttempt(this.currentProbe);

        this.number = "";
        this._currentNumber = null;
        this._countActiveNumbers = 0;
        this._communications = config.communications;
        if (config.communications instanceof Array) {

            let typeCodes = dialer._calendar.getCommunicationCodes();

            const communicationsMap = config.communications.map( (i, key) => {
                i._id = key;
                if (!i._probe)
                    i._probe = 0;

                if (!i._score)
                    i._score = 0;

                i._codeFound = dialer._calendar.checkCommunicationInAllCode(i.type) ? true : false;
                let idx = typeCodes.indexOf(i.type);

                if (~idx) {
                    i._codeWorkPrior = idx;
                    let rangeProperty = dialer._calendar.getCommunicationByPosition(i._codeWorkPrior);
                    if (rangeProperty) {
                        if (i._range && i._range.rangeId === rangeProperty.rangeId) {

                            if (i._range.attempts >= rangeProperty.range.attempts) {
                                i._skipByAttempt = 1;
                                i._codeWorkPrior = 1002;
                            }

                        } else {
                            i._range = {
                                attempts: 0,
                                rangeId: rangeProperty.rangeId || null
                            }
                        }
                    }
                } else {
                    i._codeWorkPrior = i._codeFound ? 1001 : 1000;
                }
                i._score = i._codeWorkPrior + i._probe;
                return i;
            });

            const communications = keySort(communicationsMap, {
                state: "asc",
                // _skipByAttempt: "desc",
                _codeWorkPrior: "asc",
                _score: "asc",
                priority: "desc"
            });

            let currentNumber = communications[0];

            if (currentNumber.state !== MEMBER_STATE.Idle) {
                // end call... ?? Error: must no find
                console.error(`TODO currentNumber.state !== MEMBER_STATE.Idle`);
                this.nextTrySec = 5;
                return;
            } else if (currentNumber._codeFound === true && !currentNumber._range || currentNumber._skipByAttempt) {
                // no current range: sleep to next range
                console.error(`TODO sleep number`);
                this.nextTrySec = 5;
                return;
            }

            this._currentNumber = currentNumber;

           // console.dir(communications, {depth: 10, colors: true});

            this.setCurrentNumber(this._currentNumber, config.communications);

            if (this._currentNumber) {
                this._currentNumber._probe++;

                if (this._currentNumber._range)
                    this._currentNumber._range.attempts++;

                this.number = (this._currentNumber.number + '').replace(/\D/g, '');
                this.log(`set number: ${this.number}`);
            } else {
                console.log('ERROR', this);
            }

        }
    }

    setCurrentAttempt (attempt) {
        this._log.callAttempt = attempt;
    }

    setAmdResult (result, cause) {
        console.log(`set amd ${result} - ${cause}`);
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

    setAgent (agent = {}) {

    }

    setCurrentNumber (communication, all) {
        if (!communication)
            return log.warn(`No communication in ${this._id}`);

        this._log.callNumber = communication.number;
        this._log.callPriority = communication.priority || 0;
        if (all instanceof Array)
            this._log.callPositionIndex = all.indexOf(communication);
    }

    checkExpire () {
        return this.expire && Date.now() >= this.expire;
    }

    minusProbe () {
        if (this._currentNumber) {
            this._currentNumber._probe--;
            if (this._currentNumber._range && isFinite(this._currentNumber._range && this._currentNumber._range.attempts)) {
                this._currentNumber._range.attempts--;
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
        return this.variables[varName];
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
            "dialerId": this._queueId,
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

        if (this._agent) {
            e.agentId = this._agent.id;
        }

        for (let key in this.variables) {
            if (this.variables.hasOwnProperty(key))
                e[`variable_${key}`] = this.variables[key]
        }
        return e;
    }

    broadcast () {
        application.Broker.publish(application.Broker.Exchange.FS_EVENT, `.CUSTOM.engine%3A%3Adialer_member_end..`, this.toJSON());
    }

    end (endCause, e) {
        
        if (this.processEnd) return;
        this.processEnd = true;
        this.setProbeEndCause(endCause);

        log.trace(`end member ${this._id} cause: ${this.endCause || endCause || ''}`) ;

        let skipOk = false,
            billSec = e && +e.getHeader('variable_billsec');

        if (e) {
            let recordSec = +e.getHeader('variable_record_seconds');
            if (recordSec)
                this.setRecordSession(recordSec);

            if (e.getHeader('variable_amd_result'))
                this.setAmdResult(e.getHeader('variable_amd_result'), e.getHeader('variable_amd_cause'));

            this.setCallUUID(e.getHeader('variable_uuid'))
        }
        
        if (~this.getCausesMinus(endCause)) {
            this.minusProbe();
            this.log(`end cause ${endCause}`);
            this.nextTime = Date.now() + (this.nextTrySec * 1000);
            this._setStateCurrentNumber(MEMBER_STATE.Idle);
            this.emit('end', this);
            return;
        }

        if (~this.getCausesSuccessful(endCause)) {
            if (billSec >= this.getMinBillSec()) {
                this.endCause = endCause;
                this.log(`OK: ${endCause}`);
                this.callSuccessful = true;
                this._setStateCurrentNumber(MEMBER_STATE.End);
                this.emit('end', this);
                return;
            } else {
                skipOk = true;
                this.log(`skip ok: bill sec ${billSec}`)
            }

        }

        if (~this.getCausesRetry(endCause) || skipOk) {
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
        }


        if (this.currentProbe >= this.getMaxAttemptsCount() && !this._waitingForResultStatus) {
            this.log(`max try count`);
            this.endCause = endCause || END_CAUSE.MAX_TRY;
            this._setStateCurrentNumber(MEMBER_STATE.End)
        } else {
            if (this._countActiveNumbers == 1 && endCause)
                this.endCause = endCause;
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
