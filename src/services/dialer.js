/**
 * Created by igor on 26.04.16.
 */

'use strict';

var CodeError = require(__appRoot + '/lib/error'),
    validateCallerParameters = require(__appRoot + '/utils/validateCallerParameters'),
    log = require(__appRoot + '/lib/log')(module),
    checkPermissions = require(__appRoot + '/middleware/checkPermissions'),
    END_CAUSE = require('./autoDialer/const').END_CAUSE,
    expVal = require(__appRoot + '/utils/validateExpression')
    ;

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
            if (option.state == 1) {
                return application.AutoDialer.runDialerById(option.id, domain, cb);
            } else if (option.state == 3) {
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

            let db = application.DB._query.dialer;
            return db.resetProcessStatistic(option.id, domain, cb);

        });
    },


    _removeByDomain: function (domainName, cb) {
        const db = application.DB._query.dialer;
        db._removeByDomain(domainName, cb);
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

                let member = option.data;
                member.dialer = option.dialer;
                member.createdOn = Date.now();
                member.randomValue = Math.random();
                member._score = member.createdOn + (member.priority || 0);

                if (!(member.communications instanceof Array) || member.communications.length == 0)
                    return cb(new CodeError(400, 'Bad communications (must array)'));

                for (let comm of member.communications) {
                    if (!comm.number)
                        return cb(new CodeError(400, `Bad communication number`));
                    comm.state = 0;
                }

                let db = application.DB._query.dialer;
                return db.createMember(member, (err, res) => {
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

                application.DB._query.dialer._resetMembers(options.dialer, options.resetLog, options.fromDate, caller.id, cb);
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

                if (memberDb._waitingForResultStatusCb !== 1)
                    return cb(new CodeError(400, `Member ${options.member}: result status false`));

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

                const q = {
                    $set: $set,
                    $push: {
                        '_callback': {
                            from: caller.id,
                            time: Date.now(),
                            data: callback
                        }
                    }
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
                        : {}
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
        }
    }
};

module.exports = Service;

function replaceExpression(obj) {
    if (obj)
        for (var key in obj) {
            if (typeof obj[key] == "object")
                replaceExpression(obj[key]);
            else if (typeof obj[key] != "function" && key == "expression") {
                obj["sysExpression"] = expVal(obj[key]);
            };
        };
    return;
};