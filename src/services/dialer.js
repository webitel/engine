/**
 * Created by igor on 26.04.16.
 */

'use strict';

var CodeError = require(__appRoot + '/lib/error'),
    validateCallerParameters = require(__appRoot + '/utils/validateCallerParameters'),
    log = require(__appRoot + '/lib/log')(module),
    checkPermissions = require(__appRoot + '/middleware/checkPermissions')
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
        checkPermissions(caller, 'dialer', 'd', function (err) {
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

            let db = application.DB._query.dialer;
            return db.update(option.id, domain, option.data, cb);

        });
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

        setCallback: (caller, options = {}, cb) => {

            if (!options.dialer)
                return cb(new CodeError(400, "Dialer id is required."));

            if (!options.member)
                return cb(new CodeError(400, "Member id is required."));

            if (!(options.callback instanceof Object))
                return cb(new CodeError(400, "Bad callback options"));

            let dbDialer = application.DB._query.dialer;

            dbDialer.memberById(options.member, options.dialer, (err, memberDb) => {
                if (err)
                    return cb(err);

                if (!memberDb)
                    return cb(new CodeError(404, `Not found ${options.member} in dialer ${options.dialer}`));

                let callback = options.callback;

                if (callback.success === true) {
                    memberDb._endCause = "NORMAL_CLEARING";
                    memberDb._nextTryTime = null;
                } else {
                    memberDb._endCause = null;
                    if (+callback.next_after_sec > 0) {
                        memberDb._nextTryTime = Date.now() + (+callback.next_after_sec * 1000);
                    }

                    if (callback.reset_retries === true) {
                        memberDb._probeCount = 0;
                        for (let key in memberDb.communications) {
                            memberDb.communications[key]._probe = 0;
                        }
                    }

                    if (callback.stop_communications && memberDb.communications) {
                        let all = callback.stop_communications === 'all',
                            arrNumbers = callback.stop_communications instanceof Array ? callback.stop_communications : [];

                        for (let i = 0, len = memberDb.communications.length; i < len; i++) {
                            if (memberDb.communications[i] && (all || ~arrNumbers.indexOf(memberDb.communications[i].number))) {
                                memberDb.communications[i].state = 2;
                            }
                        }
                    }

                    if (callback.next_communication && memberDb.communications) {
                        memberDb.communications.push({
                            number: callback.next_communication,
                            // TODO
                            priority: 100,
                            status: 0,
                            state: 0
                        });
                    }
                }

                dbDialer._updateMember(
                    {_id: memberDb._id},
                    memberDb,
                    {},
                    cb
                );
            })
        },

        _updateById: (id, doc, cb) => {
            let db = application.DB._query.dialer;
            return db._updateMember({_id: id}, doc, null, cb);
        },

        _updateMember (filter, doc, sort, cb) {
            let db = application.DB._query.dialer;
            return db._updateMember(filter, doc, sort, cb);
        },

        _aggregate (agg, cb) {
            let db = application.DB._query.dialer;
            return db._aggregateMembers(agg, cb);
        }
    }
};

module.exports = Service;