/**
 * Created by igor on 24.05.16.
 */

'use strict';


let generateUuid = require('node-uuid'),
    log = require(__appRoot + '/lib/log')(module),
    EventEmitter2 = require('eventemitter2').EventEmitter2,
    dynamicSort = require(__appRoot + '/utils/sort').dynamicSort,
    MEMBER_STATE = require('./const').MEMBER_STATE,
    END_CAUSE = require('./const').END_CAUSE
    ;

module.exports = class Member extends EventEmitter2 {

    constructor (config, option) {
        super();

        this.tryCount = option.maxTryCount;
        this.nextTrySec = option.intervalTryCount;
        this.minCallDuration = option.minCallDuration;
        this.callSuccessful = false;

        this.score = option._score;
        this.queueName = option.queueName ;
        this._queueId = option.queueId;
        this._domain = option.domain;
        this.queueNumber = option.queueNumber || option.queueName;

        let lockedGws = option.lockedGateways;

        this.causesOK = option.causesOK;
        this.causesRetry = option.causesRetry;
        this.causesError = option.causesError;
        this.causesMinus = option.causesMinus;

        this._id = config._id;
        this.expire = config.expire;
        this.channelsCount = 0;

        this.sessionId = generateUuid.v4();
        this._log = {
            session: this.sessionId,
            callTime: Date.now(),
            callState: 0,
            callPriority: 0,
            callNumber: null, //+
            cause: null, //+
            callAttempt: null, // +
            callUUID: null,
            recordSessionSec: 0,
            steps: []
        };
        this.currentProbe = (config._probeCount || 0) + 1;
        this.endCause = null;
        // this.bigData = new Array(1e4).join('XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX\n');
        this.variables = {};

        this._data = config;
        this.name = (config.name || "_undef_").replace(/'/g, '_');


        for (let key in config.variables) {
            this.setVariable(key, config.variables[key]);
        }

        this.log(`create probe ${this.currentProbe}`);
        this.setCurrentAttempt(this.currentProbe);

        this.number = "";
        this._currentNumber = null;
        this._countActiveNumbers = 0;
        this._communications = config.communications;
        if (config.communications instanceof Array) {
            let n = config.communications.filter( (communication, position) => {
                let isOk = communication && communication.state === MEMBER_STATE.Idle;
                if (isOk)
                    this._countActiveNumbers++;

                if (isOk && (!lockedGws || ~lockedGws.indexOf(communication.gatewayPositionMap))) {
                    if (!communication._probe)
                        communication._probe = 0;
                    if (!communication.priority)
                        communication.priority = 0;
                    communication._score = communication.priority - (communication._probe + 1);
                    communication._id = position;
                    return true;
                }
                return false;
            });
            this._currentNumber = n.sort(dynamicSort('-_score'))[0];
            this.setCurrentNumber(this._currentNumber);

            if (this._currentNumber) {
                this._currentNumber._probe++;
                this.number = this._currentNumber.number.replace(/\D/g, '');
                this.log(`set number: ${this.number}`);
            } else {
                console.log('ERROR', this);
            }

        }
    }

    setCurrentAttempt (attempt) {
        this._log.callAttempt = attempt;
    }

    setCurrentNumber (communication) {
        if (!communication)
            return log.warn(`No communication in ${this._id}`);

        this._log.callNumber = communication.number;
        this._log.callPriority = communication.priority || 0;
    }

    checkExpire () {
        return this.expire && Date.now() >= this.expire;
    }

    minusProbe () {
        if (this._currentNumber)
            this._currentNumber._probe--;
        this.currentProbe--;
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
            "variable_domain_name": this._domain,
            "dialerId": this._queueId,
            "dialerName": this.queueName,
            "id": this._id.toString(),
            "name": this.name,
            "currentNumber": this._currentNumber && this._currentNumber.number,
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

            this.setCallUUID(e.getHeader('variable_uuid'))
        }
        
        if (~this.causesMinus.indexOf(endCause)) {
            this.minusProbe();
            this.log(`end cause ${endCause}`);
            this.nextTime = Date.now() + (this.nextTrySec * 1000);
            this._setStateCurrentNumber(MEMBER_STATE.Idle);
            this.emit('end', this);
            return;
        }

        if (~this.causesOK.indexOf(endCause)) {
            if (billSec >= this.minCallDuration) {
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

        if (~this.causesRetry.indexOf(endCause) || skipOk) {
            if (this.currentProbe >= this.tryCount) {
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

        if (~this.causesError.indexOf(endCause)) {
            this.log(`fatal: ${endCause}`);
            this._setStateCurrentNumber(MEMBER_STATE.End);
        }


        if (this.currentProbe >= this.tryCount) {
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