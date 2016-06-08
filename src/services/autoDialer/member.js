/**
 * Created by igor on 24.05.16.
 */

'use strict';


let generateUuid = require('node-uuid'),
    log = require(__appRoot + '/lib/log')(module),
    EventEmitter2 = require('eventemitter2').EventEmitter2,
    dynamicSort = require(__appRoot + '/utils/sort').dynamicSort,

    CODE_RESPONSE_OK = require('./const').CODE_RESPONSE_OK,
    CODE_RESPONSE_RETRY = require('./const').CODE_RESPONSE_RETRY,
    CODE_RESPONSE_ERRORS = require('./const').CODE_RESPONSE_ERRORS,
    MEMBER_STATE = require('./const').MEMBER_STATE,
    END_CAUSE = require('./const').END_CAUSE
    ;

module.exports = class Member extends EventEmitter2 {

    constructor (config, option) {
        super();
        if (config._lock)
            throw config;

        this.tryCount = option.maxTryCount;
        this.nextTrySec = option.intervalTryCount;
        this.minCallDuration = option.minCallDuration;

        this.score = option._score;
        this.queueName = option.queueName ;
        this._queueId = option.queueId;
        this._domain = option.domain;
        this.queueNumber = option.queueNumber || option.queueName;

        let lockedGws = option.lockedGateways;

        this._id = config._id;
        this.channelsCount = 0;

        this.sessionId = generateUuid.v4();
        this._log = {
            session: this.sessionId,
            callUUid: null,
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

        this.number = "";
        this._currentNumber = null;
        this._countActiveNumbers = 0;
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

            if (this._currentNumber) {
                this._currentNumber._probe++;
                this.number = this._currentNumber.number.replace(/\D/g, '');
                this.log(`set number: ${this.number}`);
            } else {
                console.log('ERROR', this);
            }

        }
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

    setCallUUID (uuid) {
        this.log(`set uuid ${uuid}`);
        this._log.callUUID = uuid;
    }

    _setStateCurrentNumber (state) {
        if (!this._currentNumber)
            return;
        this._currentNumber.state = state;
    }

    end (endCause, e) {
        
        if (this.processEnd) return;
        this.processEnd = true;

        log.trace(`end member ${this._id} cause: ${this.endCause || endCause || ''}`) ;

        let skipOk = false,
            billSec = e && +e.getHeader('variable_billsec');

        if (e) {
            let recordSec = +e.getHeader('variable_record_seconds');
            if (recordSec)
                this.setRecordSession(recordSec);

            this.setCallUUID(e.getHeader('variable_uuid'))
        }

        if (~CODE_RESPONSE_OK.indexOf(endCause)) {
            if (billSec >= this.minCallDuration) {
                this.endCause = endCause;
                this.log(`OK: ${endCause}`);
                this._setStateCurrentNumber(MEMBER_STATE.End);
                this.emit('end', this);
                return;
            } else {
                skipOk = true;
                this.log(`skip ok: bill sec ${billSec}`)
            }

        }

        if (~CODE_RESPONSE_RETRY.indexOf(endCause) || skipOk) {
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

        if (~CODE_RESPONSE_ERRORS.indexOf(endCause)) {
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