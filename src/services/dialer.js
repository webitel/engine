/**
 * Created by igor on 26.04.16.
 */

'use strict';

const CodeError = require(__appRoot + '/lib/error'),
    request = require('request'),
    validateCallerParameters = require(__appRoot + '/utils/validateCallerParameters'),
    log = require(__appRoot + '/lib/log')(module),
    checkPermissions = require(__appRoot + '/middleware/checkPermissions'),
    END_CAUSE = require('./autoDialer/const').END_CAUSE,
    conf = require(__appRoot + '/conf'),
    generateUuid = require('node-uuid'),
    BASE_URI = conf.get('server:baseUrl').replace(/(\/+)$/, ''),
    Scheduler = require(__appRoot + '/lib/scheduler'),
    cronParser = require('cron-parser'),
    expVal = require(__appRoot + '/utils/validateExpression')
;


class CronJobs {
    constructor() {
        this.jobs = new Map();
    }

    cancel(id) {
        if (this.jobs.has(id)) {
            const job = this.jobs.get(id);
            job.cancel();
            this.jobs.delete(id);
            this.info();
            return true;
        }
        this.info();
        return false
    }

    add(id, cronFormat, data) {
        application.DB._query.dialer.getTimezoneFromDialer(data.dialerId, (err, timezone) => {
            if (err)
                log.error(err);

            const res = {};
            try {
                this.jobs.set(id, new Scheduler(cronFormat, function JobExecuteTemplate(cb) {
                    _startExecute(data, (err, res) => {
                        if (err)
                            log.error(err);

                        cb(null, res);
                    });
                },  {log: true, timezone}));

            } catch (e) {
                res.err = e;
            }
            this.info();
            return res;
        });
    }

    recreate(id, cronFormat, data) {
        this.cancel(id);
        this.add(id, cronFormat, data);
    }

    info() {
        log.debug(`Active job: ${this.jobs.size}`)
    }
}

const cronJobs = new CronJobs();

let Service = {

    /**
     *
     * @param caller
     * @param option
     * @param cb
     */
    list: function (caller, option, cb) {
        checkPermissions(caller, 'dialer', 'r', function (err) {
            if (err)
                return cb(err);

            if (!option)
                return cb(new CodeError(400, "Bad request options"));

            let domain = validateCallerParameters(caller, option['domain']);
            if (!domain) {
                return cb(new CodeError(400, 'Bad request: domain is required.'));
            };
            option.domain = domain;

            let db = application.DB._query.dialer;
            return db.search(option, cb);
        });
    },


    setState: function (caller, option, cb) {
        checkPermissions(caller, 'dialer', 'u', function (err) {
            if (err)
                return cb(err);

            if (!option)
                return cb(new CodeError(400, "Bad request options"));

            if (!option.id)
                return cb(new CodeError(400, "Bad request: id is required"));

            let domain = validateCallerParameters(caller, option['domain']);
            if (!domain) {
                return cb(new CodeError(400, 'Bad request: domain is required.'));
            }

            if (typeof option.state !== 'number') {
                return cb(new CodeError(400, 'Bad request: state is required.'));
            }

            option.domain = domain;
            if (option.state === 1) {
                return application.AutoDialer.runDialerById(option.id, domain, cb);
            } else if (option.state === 3) {
                return application.AutoDialer.stopDialerById(option.id, domain, cb);
            } else {
                return cb(new CodeError(400, 'Bad set state'));
            }

            //cb(null, {
            //    activeCall: 20,
            //    activeState: 2
            //});
        });
    },

    /**
     *
     * @param caller
     * @param option
     * @param cb
     */
    item: function (caller, option, cb) {
        checkPermissions(caller, 'dialer', 'r', function (err) {
            if (err)
                return cb(err);

            if (!option)
                return cb(new CodeError(400, "Bad request options"));

            let domain = validateCallerParameters(caller, option['domain']);
            if (!domain) {
                return cb(new CodeError(400, 'Bad request: domain is required.'));
            }

            if (!option.id)
                return cb(new CodeError(400, 'Bad request: id is required.'));

            let db = application.DB._query.dialer;

            return db.findById(option.id, domain, cb);
        });
    },

    /**
     *
     * @param caller
     * @param option
     * @param cb
     */
    create: function (caller, option, cb) {
        checkPermissions(caller, 'dialer', 'c', function (err) {
            if (err)
                return cb(err);

            if (!option)
                return cb(new CodeError(400, "Bad request options"));

            let domain = validateCallerParameters(caller, option['domain']);

            if (!domain) {
                return cb(new CodeError(400, 'Bad request: domain is required.'));
            }


            let dialer = option;
            dialer.domain = domain;

            if (!dialer.name)
                return cb(new CodeError(400, "Name is required"));

            if (!dialer.type)
                return cb(new CodeError(400, "Type is required"));

            if (!dialer.calendar || !dialer.calendar.id)
                return cb(new CodeError(400, "Calendar is required"));

            let db = application.DB._query.dialer;

            if (dialer._cf) {
                replaceExpression(dialer._cf)
            }

            if (!dialer.agents) {
                dialer.agents = [];
            }

            return db.create(dialer, cb);
        });
    },

    /**
     *
     * @param caller
     * @param option
     * @param cb
     */
    remove: function (caller, option, cb) {
        checkPermissions(caller, 'dialer', 'd', function (err) {
            if (err)
                return cb(err);

            if (!option)
                return cb(new CodeError(400, "Bad request options"));

            let domain = validateCallerParameters(caller, option['domain']);
            if (!domain) {
                return cb(new CodeError(400, 'Bad request: domain is required.'));
            }

            if (!option.id)
                return cb(new CodeError(400, 'Bad request: id is required.'));

            let db = application.DB._query.dialer,
                dialerId = option.id;

            return db.removeById(dialerId, domain, (err, res) => {
                if (!err) {
                    db.removeMemberByDialerId(dialerId, (err) => {
                        if (err)
                            log.error(err);
                    });
                }
                return cb && cb(err, res);
            });
        });
    },

    /**
     *
     * @param caller
     * @param option
     * @param cb
     */
    update: function (caller, option, cb) {
        checkPermissions(caller, 'dialer', 'u', function (err) {
            if (err)
                return cb(err);

            if (!option)
                return cb(new CodeError(400, "Bad request options"));


            if (!option.id)
                return cb(new CodeError(400, 'Bad request: id is required.'));

            if (!option.data)
                return cb(new CodeError(400, 'Bad request: data is required.'));

            let domain = validateCallerParameters(caller, option['domain']);
            if (!domain) {
                return cb(new CodeError(400, 'Bad request: domain is required.'));
            }

            if (option.data._cf) {
                replaceExpression(option.data._cf)
            }

            let db = application.DB._query.dialer;
            return db.update(option.id, domain, option.data, cb);

        });
    },

    /**
     *
     * @param caller
     * @param option
     * @param cb
     */
    resetProcessStatistic: function (caller, option, cb) {
        checkPermissions(caller, 'dialer', 'u', function (err) {
            if (err)
                return cb(err);

            if (!option)
                return cb(new CodeError(400, "Bad request options"));

            if (!option.id)
                return cb(new CodeError(400, 'Bad request: id is required.'));


            let domain = validateCallerParameters(caller, option['domain']);
            if (!domain) {
                return cb(new CodeError(400, 'Bad request: domain is required.'));
            }

            if (option.id) {
                const d = application.AutoDialer.activeDialer.get(option.id);
                if (d) {
                    log.warn(`Dialer ${d.nameDialer} in memory! members count ${d.members.length()}`);
                    let m;
                    for (let key of d.members.getKeys()) {
                        m = d.members.get(key);
                        if (m) {
                            log.warn(`Member ${m._id} clean session ${m.sessionId} and minus probe!`);
                            m.log(`Reset probe by reset process`);
                            m.minusProbe();
                            m.end();
                        }
                    }

                    if (!d.members.length()) {
                        d.emit('end', d);
                    }

                    log.warn(`Dialer ${d.nameDialer} in memory! members count ${d.members.length()}`);

                }
            }

            Service._resetProcessStatistic(option, domain, cb);
        });
    },

    _resetProcessStatistic: function (option, domain, cb) {
        if (option.resetAgents === true) {
            application.PG.getQuery('agents').reset(
                option.id,
                (err) => {
                    if (err)
                        return log.error(err);
                }
            )
        }
        let db = application.DB._query.dialer;

        return db.resetProcessStatistic(option, domain, cb);
    },


    _removeByDomain: function (domainName, cb) {
        const db = application.DB._query.dialer;
        db._removeByDomain(domainName, cb);
    },

    listHistory: function (caller, option, cb) {
        checkPermissions(caller, 'dialer', 'r', (err) => {
            if (err)
                return cb(err);

            if (!option.dialer)
                return cb(new CodeError(400, "Bad request dialer is required."));

            let domain = validateCallerParameters(caller, option['domain']);

            if (!domain) {
                return cb(new CodeError(400, 'Bad request: domain is required.'));
            }

            if (!option.filter)
                option.filter = {};

            option.filter["dialer"] = option.dialer;

            const db = application.DB._query.dialer;
            return db.listHistory(option, cb);
        })
    },

    members: {
        list: function (caller, option, cb) {
            checkPermissions(caller, 'dialer/members', 'r', function (err) {
                if (err)
                    return cb(err);

                if (!option)
                    return cb(new CodeError(400, "Bad request options."));

                if (!option.dialer)
                    return cb(new CodeError(400, "Bad request dialer is required."));

                let domain = validateCallerParameters(caller, option['domain']);

                if (!domain) {
                    return cb(new CodeError(400, 'Bad request: domain is required.'));
                }

                // TODO  before select dialer
                option.domain = null;

                if (!option.filter)
                    option.filter = {};

                option.filter["dialer"] = option.dialer;

                let db = application.DB._query.dialer;
                return db.memberList(option, cb);
            });
        },

        count: function (caller, option, cb) {
            checkPermissions(caller, 'dialer/members', 'r', function (err) {
                if (err)
                    return cb(err);

                if (!option)
                    return cb(new CodeError(400, "Bad request options."));

                if (!option.dialer)
                    return cb(new CodeError(400, "Bad request dialer is required."));

                let domain = validateCallerParameters(caller, option['domain']);

                if (!domain) {
                    return cb(new CodeError(400, 'Bad request: domain is required.'));
                }

                // TODO  before select dialer
                option.domain = null;

                if (!option.filter)
                    option.filter = {};

                option.filter["dialer"] = option.dialer;

                let db = application.DB._query.dialer;
                return db.memberCount(option, cb);
            });
        },
        
        item: function (caller, option, cb) {
            checkPermissions(caller, 'dialer/members', 'r', function (err) {
                if (err)
                    return cb(err);

                if (!option)
                    return cb(new CodeError(400, "Bad request options."));

                if (!option.id)
                    return cb(new CodeError(400, "Bad request id is required."));

                if (!option.dialer)
                    return cb(new CodeError(400, "Bad request dialer is required."));

                let domain = validateCallerParameters(caller, option['domain']);

                if (!domain) {
                    return cb(new CodeError(400, 'Bad request: domain is required.'));
                }

                // TODO  before select dialer
                option.domain = null;

                let db = application.DB._query.dialer;
                return db.memberById(option.id, option.dialer, cb);
            });
        },
        
        create: function (caller, option, cb) {
            checkPermissions(caller, 'dialer/members', 'c', function (err) {
                if (err)
                    return cb(err);

                if (!option)
                    return cb(new CodeError(400, "Bad request options"));

                if (!option.dialer)
                    return cb(new CodeError(400, 'Bad request: dialer id is required.'));

                if (!option.data)
                    return cb(new CodeError(400, 'Bad request: data is required.'));

                // TODO check dialer in domain
                let domain = validateCallerParameters(caller, option['domain']);
                if (!domain) {
                    return cb(new CodeError(400, 'Bad request: domain is required.'));
                }

                if (! (option.data instanceof Array)) {
                    option.data = [option.data];
                }

                let m;
                for (let i = 0; i < option.data.length; i++) {
                    m = option.data[i];
                    m.domain = domain;
                    m.dialer = option.dialer;
                    m.createdOn = Date.now();
                    m.randomValue = Math.random();
                    m._score = m.createdOn + (m.priority || 0);

                    if (!(m.communications instanceof Array) || m.communications.length === 0)
                        return cb(new CodeError(400, 'Bad communications (must array)'));

                    for (let comm of m.communications) {
                        if (!comm.number)
                            return cb(new CodeError(400, `Bad communication number`));
                        comm.state = 0;
                    }
                }

                let db = application.DB._query.dialer;
                return db.createMember(option.data, (err, res) => {
                    if (err)
                        return cb(err);

                    if (option.autoRun === 'true') {
                        try {
                            application.AutoDialer.runDialerById(option.dialer, domain, (e) => {
                                if (e)
                                    log.error(e);
                            });
                        } catch (e) {
                            log.error(e);
                        }
                    }

                    return cb(null, res);
                });

            });
        },

        update: function (caller, option, cb) {
            checkPermissions(caller, 'dialer/members', 'u', function (err) {
                if (err)
                    return cb(err);

                if (!option)
                    return cb(new CodeError(400, "Bad request options"));


                if (!option.id)
                    return cb(new CodeError(400, 'Bad request: id is required.'));

                if (!option.dialer)
                    return cb(new CodeError(400, 'Bad request: dialer id is required.'));

                if (!option.data)
                    return cb(new CodeError(400, 'Bad request: data is required.'));

                // TODO check dialer in domain
                let domain = validateCallerParameters(caller, option['domain']);
                if (!domain) {
                    return cb(new CodeError(400, 'Bad request: domain is required.'));
                };

                if (option.data.createdOn)
                    option.data._score = option.data.createdOn + (option.data.priority || 0);

                let db = application.DB._query.dialer;
                return db.updateMember(option.id, option.dialer, option.data, cb);

            });
        },

        remove: function (caller, option, cb) {
            checkPermissions(caller, 'dialer/members', 'd', function (err) {
                if (err)
                    return cb(err);

                if (!option)
                    return cb(new CodeError(400, "Bad request options"));


                if (!option.id)
                    return cb(new CodeError(400, 'Bad request: id is required.'));

                if (!option.dialer)
                    return cb(new CodeError(400, 'Bad request: dialer id is required.'));


                // TODO check dialer in domain
                let domain = validateCallerParameters(caller, option['domain']);
                if (!domain) {
                    return cb(new CodeError(400, 'Bad request: domain is required.'));
                }

                let db = application.DB._query.dialer;
                return db.removeMemberById(option.id, option.dialer, cb);

            });
        },

        removeByFilter: function (caller, option, cb) {
            checkPermissions(caller, 'dialer/members', 'd', function (err) {
                if (err)
                    return cb(err);

                if (!option)
                    return cb(new CodeError(400, "Bad request options"));


                if (!option.filter)
                    return cb(new CodeError(400, 'Bad request: filter is required.'));

                if (!option.dialer)
                    return cb(new CodeError(400, 'Bad request: dialer id is required.'));


                // TODO check dialer in domain
                let domain = validateCallerParameters(caller, option['domain']);
                if (!domain) {
                    return cb(new CodeError(400, 'Bad request: domain is required.'));
                }

                let db = application.DB._query.dialer;
                return db.removeMemberByFilter(option.dialer, option.filter, cb);

            });
        },
        
        aggregate: function (caller, option, cb) {
            checkPermissions(caller, 'dialer/members', 'r', function (err) {
                if (err)
                    return cb(err);

                if (!option)
                    return cb(new CodeError(400, "Bad request options"));


                if (!option.data || !(option.data instanceof Array))
                    return cb(new CodeError(400, 'Bad request: data is required.'));

                if (!option.dialer)
                    return cb(new CodeError(400, 'Bad request: dialer id is required.'));


                // TODO check dialer in domain
                let domain = validateCallerParameters(caller, option['domain']);
                if (!domain) {
                    return cb(new CodeError(400, 'Bad request: domain is required.'));
                }

                let db = application.DB._query.dialer;
                return db.aggregateMembers(option.dialer, option.data, cb);

            });
        },

        resetMembers: (caller, options = {}, cb) => {
            checkPermissions(caller, 'dialer/members', 'u', function (err) {
                if (err)
                    return cb(err);

                if (!options.dialer)
                    return cb(new CodeError(400, 'Bad request: dialer id is required.'));

                // TODO check dialer in domain
                // const domain = validateCallerParameters(caller, options['domain']);

                if (options.resetLog && caller.domain) {
                    return cb(new CodeError(403, 'Bad request: resetLog allow from root.'));
                }

                application.DB._query.dialer._resetMembers(options.dialer, options.resetLog, options.fromDate, caller.id, (err, count) => {
                    if (err) {
                        application.AutoDialer.addLogDialer(options.dialer, "RESET_MEMBERS", `Error: ${err.message}`);
                        return cb(err)
                    }
                    application.AutoDialer.addLogDialer(options.dialer, "RESET_MEMBERS", `Reset (${caller.id}) count ${count}`);
                    return cb(null, count)
                });
            })
        },

        setCallback: (caller, options = {}, cb) => {

            log.debug(`Set callback: %j, from: %s`, options, caller.id);

            if (!options.dialer)
                return cb(new CodeError(400, "Dialer id is required."));

            if (!options.member)
                return cb(new CodeError(400, "Member id is required."));

            if (!(options.callback instanceof Object))
                return cb(new CodeError(400, "Bad callback options"));

            let dbDialer = application.DB._query.dialer;

            dbDialer.memberById(options.member, options.dialer, (err, memberDb, dialerDb) => {
                if (err)
                    return cb(err);

                if (!memberDb)
                    return cb(new CodeError(404, `Not found ${options.member} in dialer ${options.dialer}`));

                if (memberDb._waitingForResultStatusCb !== 1) {
                    dbDialer._updateMember(
                        {_id: memberDb._id},
                        {$push: {
                            _callback: {
                                from: caller.id,
                                time: Date.now(),
                                data: {
                                    success: `timeout:${options.callback.success === true ? 'true' : 'false'}`,
                                    description: `Woow! Slow down! You ip: ${options.callerIp}!!1`,
                                    request: options.callback,
                                }
                            }
                        }},
                        {},
                        (e) => {
                            if (e) {
                                log.error(e)
                            }
                        }
                    );

                    return cb(new CodeError(400, `Member ${options.member}: result status false`));
                }

                let callback = options.callback,
                    $push;

                const $set = {
                    _waitingForResultStatus: null,
                    _waitingForResultStatusCb: null
                };

                const communications = memberDb.communications || [];

                for (let i = 0, len = communications.length; i < len; i++) {
                    $set[`communications.${i}.checkResult`] = null;
                }

                if (callback.success === true) {
                    if (!memberDb._lock && memberDb._log instanceof Array) {
                        $set[`_log.${memberDb._log.length -1}.callSuccessful`] = true;
                        $set[`_log.${memberDb._log.length -1}.callState`] = 2;
                    } else {
                        const activeMember = application.AutoDialer.getMemberFromActiveDialer(options.dialer, options.member);
                        if (activeMember) {
                            activeMember.setCallSuccessful(true);
                        } else {
                            log.warn(`Not found active member ${options.member}`);
                        }
                    }
                    $set._endCause = "NORMAL_CLEARING";
                    $set.callSuccessful = true;
                    $set._nextTryTime = null;
                    for (let i = 0, len = communications.length; i < len; i++) {
                        $set[`communications.${i}.state`] = 2;
                    }
                } else {

                    // TODO bug if 0 - set default
                    if (+callback.next_after_sec >= 0) {
                        $set._nextTryTime = Date.now() + (+callback.next_after_sec * 1000);
                    } else {
                        $set._nextTryTime = (dialerDb.parameters && dialerDb.parameters.intervalTryCount) * 1000 + Date.now()
                    }

                    if (callback.stop_communications) {
                        let all = callback.stop_communications === 'all',
                            arrNumbers = callback.stop_communications instanceof Array ? callback.stop_communications : [callback.stop_communications];

                        for (let i = 0, len = communications.length; i < len; i++) {
                            if (all || ~arrNumbers.indexOf(communications[i].number)) {
                                $set[`communications.${i}.state`] = 2;
                                $set[`communications.${i}.stopCommunication`] = true;
                                communications[i].state = 2;
                            }
                        }
                    }

                    if (callback.reset_retries === true) {
                        $set._probeCount = 0;
                        $set._endCause = null;
                        for (let key in communications) {
                            $set[`communications.${key}._probe`] = 0;
                        }
                    } else if (memberDb._probeCount >= memberDb._maxTryCount) {
                        $set._endCause = END_CAUSE.MAX_TRY;
                        for (let key in communications) {
                            $set[`communications.${key}.state`] = 2;
                        }
                    }

                    if (!$set._endCause) {
                        let noCommunications = true;
                        for (let comm of communications) {
                            if (comm.state === 0) {
                                noCommunications = false;
                                break;
                            }
                        }

                        if (noCommunications)
                            $set._endCause = END_CAUSE.NO_COMMUNICATIONS;
                    }

                    if (callback.next_communication && memberDb.communications) {
                        $set._endCause = null;
                        $push = {
                            "communications": {
                                number: callback.next_communication,
                                // TODO
                                priority: 100,
                                status: 0,
                                state: 0
                            }
                        };
                    }
                }

                $set['_log.0.callback'] = {
                    from: caller.id,
                    time: Date.now(),
                    data: callback
                };

                const q = {
                    $set: $set,
                    // $push: {
                    //     '_callback': {
                    //         from: caller.id,
                    //         time: Date.now(),
                    //         data: callback
                    //     }
                    // }
                };

                dbDialer._updateMember(
                    {_id: memberDb._id},
                    q,
                    {},
                    (e, r) => {
                        if (e)
                            return cb(e);

                        if ($push) {
                            dbDialer._updateMember({_id: memberDb._id}, {$push}, {}, cb);
                        } else {
                            return cb(null, r);
                        }
                    }
                );

                if ($set._endCause) {
                    const lastNumber = isFinite(memberDb._lastNumberId) && memberDb.communications[memberDb._lastNumberId]
                        ? memberDb.communications[memberDb._lastNumberId]
                        : null
                    ;

                    const event = {
                        "Event-Name": "CUSTOM",
                        "Event-Subclass": "engine::dialer_member_end",
                        // TODO
                        "variable_domain_name": dialerDb.domain,
                        "dialerId": memberDb.dialer,
                        "dialerName": dialerDb.name,
                        "id": memberDb._id.toString(),
                        "name": memberDb.name,
                        "currentProbe": memberDb._probeCount,
                        "endCause": $set._endCause,
                        "reason": "callback",
                        "callback_user_id": caller.id
                    };

                    if (lastNumber) {
                        event.currentNumber = lastNumber.number;
                        event.dlr_member_number_description = lastNumber.description || ''
                    }

                    for (let key in memberDb.variables) {
                        if (memberDb.variables.hasOwnProperty(key))
                            event[`variable_${key}`] = memberDb.variables[key]
                    }
                    console.log(event);
                    application.Broker.publish(application.Broker.Exchange.FS_EVENT, `.CUSTOM.engine%3A%3Adialer_member_end..`, event);

                }
            }, true); //TODO...
        },

        _updateById: (id, doc, cb) => {
            let db = application.DB._query.dialer;
            return db._updateMember({_id: id}, doc, null, cb);
        },

        _updateByFilter: (filter, update, cb) => {
            let db = application.DB._query.dialer;
            return db._updateMember(filter, update, null, cb);
        },

        _updateByIdFix: (id, doc, cb) => {
            let db = application.DB._query.dialer;
            return db._updateMemberFix(id, doc, cb);
        },

        _updateMember (filter, doc, sort, cb) {
            let db = application.DB._query.dialer;
            return db._updateMember(filter, doc, sort, cb);
        },
        
        _updateMultiMembers (filter, update, cb) {
            let db = application.DB._query.dialer;
            return db._updateMultiMembers(filter, update, cb);
        },


        _lockCount (dialerId, cb) {
            let db = application.DB._query.dialer;
            return db._lockCount(dialerId, cb);
        },

        _aggregate (agg, cb) {
            let db = application.DB._query.dialer;
            return db._aggregateMembers(agg, cb);
        },

        _getCursor: (filter, projection) => {
            const db = application.DB._query.dialer;
            return db._getCursor(filter, projection);
        },

        _updateOneMember: (filter, update, cb) => {
            const db = application.DB._query.dialer;
            return db._updateOneMember(filter, update, cb);
        }
    },
    
    agents: {
        list: function (caller, option, cb) {
            checkPermissions(caller, 'dialer', 'r', function (err) {
                if (err)
                    return cb(err);

                if (!option || !option.dialer)
                    return cb(new CodeError(400, "Bad request options"));

                let domain = validateCallerParameters(caller, option['domain']);
                if (!domain) {
                    return cb(new CodeError(400, 'Bad request: domain is required.'));
                }

                let db = application.DB._query.dialer;

                db.findById(option.dialer, domain, (err, data) => {
                    if (err)
                        return cb(err);


                });
            })
        },

        stats: function (caller, options, cb) {
            checkPermissions(caller, 'dialer', 'r', function (err) {
                if (err)
                    return cb(err);

                if (!options || !options.dialer)
                    return cb(new CodeError(400, "Bad request options"));

                const domain = validateCallerParameters(caller, options['domain']);
                if (!domain) {
                    return cb(new CodeError(400, 'Bad request: domain is required.'));
                }

                if (!(options.agents instanceof Array)) {
                    return cb(new CodeError(400, 'Bad request: agent is required.'));
                }
                if (!(options.skills instanceof Array)) {
                    return cb(new CodeError(400, 'Bad request: skills is required.'));
                }

                application.PG.getQuery('agents').getAgentStats(
                    options.dialer,
                    options.agents,
                    options.skills,
                    domain,
                    cb
                );
            })
        }
    },

    templates: {
        list: (caller, options, cb) => {
            checkPermissions(caller, 'dialer/templates', 'r', function (err) {
                if (err)
                    return cb(err);

                if (!options)
                    return cb(new CodeError(400, "Bad request options"));

                application.PG.getQuery('dialer').templates.list(options, cb);
            });
        },

        item: (caller, options, cb) => {
            checkPermissions(caller, 'dialer/templates', 'r', function (err) {
                if (err)
                    return cb(err);

                if (!options)
                    return cb(new CodeError(400, "Bad request options"));

                application.PG.getQuery('dialer').templates.item(options.dialerId, options.id, cb);
            });
        },

        create: (caller, options, cb) => {
            checkPermissions(caller, 'dialer/templates', 'c', function (err) {
                if (err)
                    return cb(err);

                if (!options)
                    return cb(new CodeError(400, "Bad request options"));

                if (options.cron) {
                    const e = testCronFormat(options.cron);
                    if (e) {
                        return cb(e)
                    }
                }

                application.PG.getQuery('dialer').templates.create(options, (err, res) => {
                    if (err)
                        return cb(err);

                    if (options.cron) {
                        cronJobs.recreate(res.id, options.cron, {
                            dialerId: options.dialerId,
                            id: res.id,
                        });
                    }

                    return cb(null, res);
                });
            });
        },

        update: (caller, options, cb) => {
            checkPermissions(caller, 'dialer/templates', 'u', function (err) {
                if (err)
                    return cb(err);

                if (!options)
                    return cb(new CodeError(400, "Bad request options"));

                application.PG.getQuery('dialer').templates.update(options.dialerId, options.id, options, (err, data) => {
                    if (err)
                        return cb(err);

                    if (data) {
                        if (data.cron) {
                            cronJobs.recreate(data.id, data.cron, {
                                dialerId: options.dialerId,
                                id: options.id,
                            });
                        } else {
                            cronJobs.cancel(data.id)
                        }
                    }
                    return cb(err, data);
                });
            });
        },

        remove: (caller, options, cb) => {
            checkPermissions(caller, 'dialer/templates', 'd', function (err) {
                if (err)
                    return cb(err);

                if (!options)
                    return cb(new CodeError(400, "Bad request options"));

                application.PG.getQuery('dialer').templates.remove(options.dialerId, options.id, (err, res) => {
                    if (err) {
                        return cb(err);
                    }
                    if (res)
                        cronJobs.cancel(res.id);
                    return cb(null, res);
                });
            });
        },

        startExecute: (caller, options, cb) => {
            checkPermissions(caller, 'dialer/templates', 'r', function (err) {
                if (err)
                    return cb(err);

                if (!options)
                    return cb(new CodeError(400, "Bad request options"));

                return _startExecute(options, cb);
            });
        },

        endExecute: (caller, options, cb) => {
            checkPermissions(caller, 'dialer/templates', 'r', function (err) {
                if (err)
                    return cb(err);

                if (!options)
                    return cb(new CodeError(400, "Bad request options"));

                let data = null;
                if (options.body instanceof Object) {
                    try {
                        data = Buffer.from(JSON.stringify(options.body));
                        if (data.length === 2) {
                            data = null;
                        }
                    } catch (e) {
                        log.error(e)
                    }
                }

                application.PG.getQuery('dialer').templates.endExecute(options, data, (err, res) => {
                    if (err)
                        return cb(err);

                    if (!res) {
                        return cb(new CodeError(406, `Bad proccess id ${options.pid}`))
                    }

                    if (options.success && res.next_process_id && res.dialer_id) {
                        _startExecute({
                            dialerId: res.dialer_id,
                            id: res.next_process_id,
                        }, (err) => {
                            if (err)
                                log.error(err);
                        })
                    }

                    return cb(null, res);
                })
            })
        },

        _initJobs: (cb) => {
            cronJobs.jobs.forEach((job, key) => {
                cronJobs.cancel(key);
            });

            application.PG.getQuery('dialer').templates.getNoEmptyCron((err, jobs) => {
                if (err) {
                    return cb(err);
                }

                jobs.forEach(job => {
                    cronJobs.recreate(job.id, job.cron, {
                        dialerId: job.dialer_id,
                        id: job.id
                    });
                });

                return cb(null);
            })
        }
    }
};

module.exports = Service;

function replaceExpression(obj) {
    if (obj)
        for (var key in obj) {
            if (typeof obj[key] === "object")
                replaceExpression(obj[key]);
            else if (typeof obj[key] !== "function" && key === "expression") {
                obj["sysExpression"] = expVal(obj[key]);
            }
        }
}

function _startExecute(options = {}, cb) {
    application.PG.getQuery('dialer').templates.setExecute(options.dialerId, options.id, (err, res) => {
        if (err)
            return cb(err);

        if (!res) {
            return cb(new CodeError(400, `Process is working`))
        }

        if (res.before_delete && res.action === "import") {
            application.DB._query.dialer.removeMemberByDialerId(res.dialer_id, (err) => {
                if (err) {
                    log.error(err);
                    application.PG.getQuery('dialer').templates.rollback(res.dialer_id, res.id, {
                        process_start: null,
                        process_state: ""
                    }, (err) => {
                        if (err)
                            return log.error(err);
                    });
                    return cb(err)
                }

                executeTemplate(res, cb)
            });
        } else {
            executeTemplate(res, cb)
        }

    });
}

function executeTemplate(res, cb) {
    switch (res.type) {
        case "SQL":
        case "WEB":
            return sendRequestTemplate(res, cb);
        default:
            application.PG.getQuery('dialer').templates.rollback(res.dialer_id, res.id, {
                process_start: null,
                process_state: ""
            }, (err) => {
                if (err)
                    return log.error(err);
            });
            return cb(new CodeError(400, `Not implement type ${res.type}`));
    }
}

function sendRequestTemplate(record, cb) {
    const template = record.template || {};
    const options = {
        method: template.method,
        uri: template.url, //TODO
        body: null,
        headers: template.headers || {}
    };

    if (!template.body) {
        template.body = {}
    }

    // todo lower case
    if (!options.headers['Content-Type']) {
        options.headers['Content-Type'] = 'application/json';

        if (record.success_data) {
            if (record.success_data instanceof Buffer) {
                template.body.user_data = record.success_data;
                try {
                    template.body.user_data = JSON.parse(template.body.user_data)
                } catch (e) {

                }
            }
        }

        try {
            options.body = JSON.stringify(template.body);
        } catch (e) {
            //TODO ???
            return cb(new CodeError(400, e.message))
        }
    }

    options.headers['X-Action'] = record.action;
    options.headers['X-Response-Host'] = BASE_URI;
    options.headers['X-Response-Path'] = `/api/v2/dialer/${record.dialer_id}/templates/${record.id}/end/${record.process_id}`;
    options.headers['X-Dialer-Id'] = record.dialer_id;
    options.headers['X-Template-Id'] = record.id;

    request(options, (err, res) => {
        if (err) {
            log.error(err);
            application.PG.getQuery('dialer').templates.rollback(record.dialer_id, record.id, {
                process_start: null,
                process_state: `ERROR_STAGE_1`,
                last_response_text: err.message
            }, (err) => {
                if (err)
                    return log.error(err);

            });
            return cb(err);
        }

        log.debug(`Execute template ${record.id} response code ${res.statusCode} text: ${res.body}`);
        if (res.statusCode !== 200) {
            application.PG.getQuery('dialer').templates.rollback(record.dialer_id, record.id, {
                process_start: null,
                process_state: `ERROR_STAGE_1`,
                last_response_text: res.body ? res.body + "" : ""
            }, (err) => {
                if (err)
                    return log.error(err);
            });

            return cb(new CodeError(res.statusCode, res.body + ""));
        }

        return cb(null, res.body)
    });
}

function testCronFormat(format) {
    try {
        cronParser.parseExpression(format);
        return null;
    } catch (e) {
        return new CodeError(400, e.message)
    }
}