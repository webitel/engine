/**
 * Created by igor on 05.07.17.
 */

'use strict';

const log = require(__appRoot + '/lib/log')(module),
    validateCallerParameters = require(__appRoot + '/utils/validateCallerParameters'),
    checkPermissions = require(__appRoot + '/middleware/checkPermissions'),
    CodeError = require(__appRoot + '/lib/error');


const Service = {
    list: (caller, option = {}, cb) => {
        checkPermissions(caller, 'callback', 'r', function (err) {
            if (err)
                return cb(err);

            if (!option)
                return cb(new CodeError(400, "Bad request options"));

            let domain = validateCallerParameters(caller, option['domain']);
            if (!domain) {
                return cb(new CodeError(400, 'Bad request: domain is required.'));
            }
            option.domain = domain;
            application.PG.getQuery('callback').list(option, cb);
        });
    },

    get: (caller, option = {}, cb) => {
        checkPermissions(caller, 'callback', 'r', function (err) {
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

            application.PG.getQuery('callback').findById(option.id, domain, cb);
        });
    },

    create: (caller, option = {}, cb) => {
        option.domain = validateCallerParameters(caller, option.domain);

        if (!option.domain)
            return cb(new CodeError(400, 'Domain is required.'));

        checkPermissions(caller, 'callback', 'c', (e) => {
            if (e)
                return cb(e);

            application.PG.getQuery('callback').create(option, cb);
        })
    },

    update: (caller, option = {}, cb) => {
        checkPermissions(caller, 'callback', 'u', function (err) {
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

            application.PG.getQuery('callback').update(option.id, domain, option.data, cb);

        });
    },

    remove: function (caller, option, cb) {
        checkPermissions(caller, 'callback', 'd', function (err) {
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

            application.PG.getQuery('callback').delete(option.id, domain, cb);
        });
    },

    members: {
        list: (caller, option = {}, cb) => {
            checkPermissions(caller, 'callback/members', 'r', function (err) {
                if (err)
                    return cb(err);

                if (!option)
                    return cb(new CodeError(400, "Bad request options"));

                if (!option.queue)
                    return cb(new CodeError(400, "Bad request queue is required."));

                let domain = validateCallerParameters(caller, option['domain']);
                if (!domain) {
                    return cb(new CodeError(400, 'Bad request: domain is required.'));
                }
                option.domain = domain;

                if (!option.filter)
                    option.filter = {};

                option.filter.queue_id = +option.queue;
                option.filter.domain = option.domain;
                application.PG.getQuery('callback').members.list(option, cb);
            });
        },

        get: (caller, option = {}, cb) => {
            checkPermissions(caller, 'callback/members', 'r', function (err) {
                if (err)
                    return cb(err);

                if (!option)
                    return cb(new CodeError(400, "Bad request options"));

                if (!option.queue)
                    return cb(new CodeError(400, "Bad request queue is required."));

                let domain = validateCallerParameters(caller, option['domain']);
                if (!domain) {
                    return cb(new CodeError(400, 'Bad request: domain is required.'));
                }

                if (!option.id)
                    return cb(new CodeError(400, 'Bad request: id is required.'));

                application.PG.getQuery('callback').members.findById(option.id, option.queue, domain, cb);
            });
        },

        create: (caller, option = {}, cb) => {
            checkPermissions(caller, 'callback/members', 'c', function (err) {
                if (err)
                    return cb(err);

                if (!option)
                    return cb(new CodeError(400, "Bad request options"));

                if (!option.queue)
                    return cb(new CodeError(400, "Bad request queue is required."));

                let domain = option['domain'] = validateCallerParameters(caller, option['domain']);
                if (!domain) {
                    return cb(new CodeError(400, 'Bad request: domain is required.'));
                }

                application.PG.getQuery('callback').members.create(option, (err, data) => {
                    if (err)
                        return cb(err);
                    application.Broker.emit('hookEvent', 'CUSTOM', domain, getJson('callback_member_add', domain, data));
                    return cb(null, data);
                });
            });
        },

        createPublic: (callerInfo, option = {}, cb) => {

            if (!application.Esl.connected)
                return cb(new CodeError(500, 'No live connect to FreeSWITCH'));

            if (!option.domain)
                return cb(new CodeError(400, 'Domain is required.'));

            if (!option.number)
                return cb(new CodeError(400, 'Number is required.'));

            if (!callerInfo.widget)
                return cb(new CodeError(400, 'Widget is required.'));

            const tryCall = !option.callback_time;
            if (tryCall)
                option.callback_time = Date.now();

            option.request_ip = callerInfo.ip;
            application.PG.getQuery('callback').members.createPublic(callerInfo.widget, option, (err, info) => {
                if (err)
                    return cb(err);

                if (tryCall) {
                    //TODO webitel_widget_id webitel_widget_name
                    const dialString = `originate [^^:cc_queue='${info.queueName}':call_timeout=${info.callTimeout || 60}:domain_name='${option.domain}':ignore_early_media=true:loopback_bowout=false:hangup_after_bridge=true]loopback/${option.number}/default '${info.destinationNumber}' XML public ${option.number} ${option.number}`;
                    log.trace(`Exec: ${dialString}`);
                    application.Esl.bgapi(dialString, (res) => {
                        if (/^-ERR|^-USAGE/.test(res.body)) {
                            log.error(res.body);
                        } else {
                            log.trace(`Call: ${res.body}`);
                        }
                        if (info.member) {
                            application.Broker.emit('hookEvent', 'CUSTOM', option.domain, getJson('callback_member_add', option.domain, info.member));
                        }
                        return cb(null, "Success");
                    });
                }
            });
        },

        update: (caller, option = {}, cb) => {
            checkPermissions(caller, 'callback/members', 'u', function (err) {
                if (err)
                    return cb(err);

                if (!option)
                    return cb(new CodeError(400, "Bad request options"));


                if (!option.id)
                    return cb(new CodeError(400, 'Bad request: id is required.'));

                if (!option.queue)
                    return cb(new CodeError(400, "Bad request queue is required."));

                if (!option.data)
                    return cb(new CodeError(400, 'Bad request: data is required.'));

                let domain = validateCallerParameters(caller, option['domain']);
                if (!domain) {
                    return cb(new CodeError(400, 'Bad request: domain is required.'));
                }

                let sendHook = false;
                if (option.data.done === true) {
                    option.data.done_at = Date.now();
                    option.data.done_by = caller.id;
                    sendHook = true;
                }
                application.PG.getQuery('callback').members.update(option.id, option.queue, domain, option.data, (err, res) => {
                    if (err)
                        return cb(err);

                    if (sendHook && res) {
                        application.Broker.emit('hookEvent', 'CUSTOM', domain, getJson('callback_member_done', domain, res));
                    }

                    return cb(null, res);
                });

            });
        },

        remove: function (caller, option, cb) {
            checkPermissions(caller, 'callback/members', 'd', function (err) {
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

                if (!option.queue)
                    return cb(new CodeError(400, "Bad request queue is required."));

                application.PG.getQuery('callback').members.delete(option.id, option.queue, domain, cb);
            });
        },

        createComment: (caller, option = {}, cb) => {
            checkPermissions(caller, 'callback/members', 'u', function (err) {
                if (err)
                    return cb(err);

                if (!option)
                    return cb(new CodeError(400, "Bad request options"));


                if (!option.id)
                    return cb(new CodeError(400, 'Bad request: id is required.'));

                if (!option.queue)
                    return cb(new CodeError(400, "Bad request queue is required."));

                if (!option.data)
                    return cb(new CodeError(400, 'Bad request: data is required.'));

                let domain = validateCallerParameters(caller, option['domain']);
                if (!domain) {
                    return cb(new CodeError(400, 'Bad request: domain is required.'));
                }


                option.data.created_on = Date.now();
                option.data.created_by = caller.id;

                application.PG.getQuery('callback').members.addComment(option.id, option.data, (err, res) => {
                    if (err)
                        return cb(err);

                    if (res) {
                        application.Broker.emit('hookEvent', 'CUSTOM', domain, {
                            "Event-Name": "CUSTOM",
                            "Event-Subclass": "engine::callback_member_comment",
                            "variable_domain_name": domain,
                            "created_by": res.created_by,
                            "created_on": res.created_on,
                            "comment_id": res.id,
                            "member_id": res.member_id,
                            "comment": res.text
                        });
                    }

                    return cb(null, res);
                });
            });
        },

        removeComment: (caller, option = {}, cb) => {
            checkPermissions(caller, 'callback/members', 'u', function (err) {
                if (err)
                    return cb(err);

                if (!option)
                    return cb(new CodeError(400, "Bad request options"));


                if (!option.id)
                    return cb(new CodeError(400, 'Bad request: id is required.'));

                if (!option.queue)
                    return cb(new CodeError(400, "Bad request queue is required."));

                if (!option.commentId)
                    return cb(new CodeError(400, 'Bad request: commentId is required.'));

                let domain = validateCallerParameters(caller, option['domain']);
                if (!domain) {
                    return cb(new CodeError(400, 'Bad request: domain is required.'));
                }

                application.PG.getQuery('callback').members.removeComment(option.id, option.queue, domain, option.commentId, cb);
            });
        },

        updateComment: (caller, option = {}, cb) => {
            checkPermissions(caller, 'callback/members', 'u', function (err) {
                if (err)
                    return cb(err);

                if (!option)
                    return cb(new CodeError(400, "Bad request options"));


                if (!option.id)
                    return cb(new CodeError(400, 'Bad request: id is required.'));

                if (!option.queue)
                    return cb(new CodeError(400, "Bad request queue is required."));

                if (!option.commentId)
                    return cb(new CodeError(400, 'Bad request: commentId is required.'));

                if (!option.text)
                    return cb(new CodeError(400, 'Bad request: text is required.'));

                let domain = validateCallerParameters(caller, option['domain']);
                if (!domain) {
                    return cb(new CodeError(400, 'Bad request: domain is required.'));
                }

                application.PG.getQuery('callback').members.updateComment(option.id, option.queue, domain, option.commentId, option.text, cb);

            });
        }
    }
};

module.exports = Service;

function getJson(eventName, domain, member) {
    const e = {
        "Event-Name": "CUSTOM",
        "Event-Subclass": `engine::${eventName}`,
        "variable_domain_name": domain
    };

    for (let key in member) {
        e[key] = member[key]
    }

    return e
}