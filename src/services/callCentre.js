/**
 * Created by Igor Navrotskyj on 03.09.2015.
 */

'use strict';

var CodeError = require(__appRoot + '/lib/error'),
    checkPermissions = require(__appRoot + '/middleware/checkPermissions'),
    checkEslError = require(__appRoot + '/middleware/checkEslError'),
    parsePlainTableToJSONArray = require(__appRoot + '/utils/parse').plainTableToJSONArray,
    validateCallerParameters = require(__appRoot + '/utils/validateCallerParameters');

var Service = {
    /**
     *
     * @param caller
     * @param option
     * @param cb
     */
    // TODO rename to getAgents
    getTiers: function (caller, option, cb) {
        checkPermissions(caller, 'cc/tiers', 'r', function (err) {
            if (err)
                return cb(err);
            option = option || {};
            var domain = validateCallerParameters(caller, option['domain']);

            if (!option['queue'] || !domain)
                return cb(new CodeError(400, "Bad request."));

            application.Esl.bgapi('callcenter_config queue list agents ' + option['queue'] + '@' + domain,
                function (res) {
                    var err = checkEslError(res);
                    if (err)
                        return cb(err);

                    parsePlainTableToJSONArray(res['body'], function (err, arr) {
                        if (err)
                            return cb(err);

                        return cb(null, arr);
                    }, '|');
                }
            );

        });
    },
    // TODO rename to getTiers
    getTiersByQueue: function (caller, option, cb) {
        checkPermissions(caller, 'cc/tiers', 'r', function (err) {
            if (err)
                return cb(err);
            option = option || {};
            var domain = validateCallerParameters(caller, option['domain']);

            if (!option['queue'] || !domain)
                return cb(new CodeError(400, "Bad request."));

            application.Esl.bgapi('callcenter_config queue list tiers ' + option['queue'] + '@' + domain,
                function (res) {
                    var err = checkEslError(res);
                    if (err)
                        return cb(err);

                    parsePlainTableToJSONArray(res['body'], function (err, arr) {
                        if (err)
                            return cb(err);

                        return cb(null, arr);
                    }, '|');
                }
            );

        });
    },

    /**
     *
     * @param caller
     * @param option
     * @param cb
     */
    getTiersFromCaller: function (caller, option, cb) {
        checkPermissions(caller, 'cc/tiers', 'r', function (err) {
            if (err)
                return cb(err);

            let tiers = application.Agents.get(caller.id) || [];
            return cb(null, tiers);
        });
    },

    /**
     *
     * @param caller
     * @param option
     * @param cb
     */
    getMembers: function (caller, option, cb) {
        checkPermissions(caller, 'cc/members', 'r', function (err) {
            if (err)
                return cb(err);
            option = option || {};
            var param = option['count']
                ? 'count'
                : 'list';

            var domain = validateCallerParameters(caller, option['domain']);

            if (!option['queue'] || !domain)
                return cb(new CodeError(400, "Bad request."));

            application.Esl.bgapi('callcenter_config queue ' + param + ' members ' + option['queue'] + '@' + domain,
                function (res) {
                    var err = checkEslError(res);
                    if (err)
                        return cb(err);

                    if (!option['count']) {
                        parsePlainTableToJSONArray(res['body'], function (err, arr) {
                            if (err)
                                return cb(err);

                            return cb(null, arr);
                        }, '|');
                    } else {
                        return cb(null, res['body']);
                    }
                }
            );
        });
    },

    /**
     *
     * @param caller
     * @param option
     * @param cb
     */
    login: function (caller, option, cb) {
        let status = option['status']
            ? " '" + option['status'] + "'"
            : " 'Available'"
            ;
        if (caller['cc-logged']) {
            return cb(null, '+OK');
        };

        application.Esl.bgapi('callcenter_config agent set status ' + caller['id'] + ' ' + status, function (res) {
            if (getResponseOK(res)) {
                caller['cc-logged'] = true;
                application.loggedOutAgent.remove(caller.id);
                return cb(null, res.body);
            } else {
                return cb(new Error(res && res.body));
            }
        });
    },

    /**
     *
     * @param caller
     * @param option
     * @param cb
     */
    logout: function (caller, option, cb) {
        application.Esl.bgapi('callcenter_config agent set status ' + caller['id'] + " 'Logged Out'", function (res) {
            if (getResponseOK(res)) {
                caller['cc-logged'] = false;
                return cb(null, res.body);
            } else {
                return cb(new Error(res && res.body));
            }
        });
    },



    // TODO Deprecated (update to mod_crm)
    /**
     * 
     * @param caller
     * @param option
     * @param cb
     */
    queuesList: function (caller, option, cb) {
        checkPermissions(caller, 'cc/queue', 'r', function (err) {
            if (err)
                return cb(err);
            option = option || {};
            var domain = validateCallerParameters(caller, option['domain']);

            if (!domain) {
                return cb(new CodeError(400, 'Domain is required.'));
            };

            return application.WConsole.queueList(
                null,
                {
                    "domain": domain
                },
                cb
            );
        });
    },

    /**
     * 
     * @param caller
     * @param queue
     * @param cb
     */
    queueCreate: function (caller, queue, cb) {
        checkPermissions(caller, 'cc/queue', 'c', function (err) {
            if (err)
                return cb(err);
            queue = queue || {};
            var domain = validateCallerParameters(caller, queue['domain']);
            queue['domain'] = domain;

            if (!domain) {
                return cb(new CodeError(400, 'Domain is required.'));
            };

            if (!queue['name']) {
                return cb(new CodeError(400, 'Name is required.'));
            };

            return application.WConsole.queueCreate(
                null,
                queue,
                cb
            );
        });
    },

    /**
     *
     * @param caller
     * @param option
     * @param cb
     */
    queueItem: function (caller, option, cb) {
        checkPermissions(caller, 'cc/queue', 'r', function (err) {
            if (err)
                return cb(err);

            if (!option || !option['name']) {
                return cb(new CodeError(400, 'Name is required.'));
            };

            var domain = validateCallerParameters(caller, option['domain']);

            if (!domain) {
                return cb(new CodeError(400, 'Domain is required.'));
            };
            option['domain'] = domain;

            return application.WConsole.queueItem(
                null,
                option,
                cb
            );
        });
    },

    /**
     *
     * @param caller
     * @param option
     * @param cb
     */
    queueUpdate: function (caller, option, cb) {
        checkPermissions(caller, 'cc/queue', 'u', function (err) {
            if (err)
                return cb(err);

            if (!option || !option['name'] || !option['params']) {
                return cb(new CodeError(400, 'Name or params is required.'));
            };

            var domain = validateCallerParameters(caller, option['domain']);

            if (!domain) {
                return cb(new CodeError(400, 'Domain is required.'));
            };
            option['domain'] = domain;

            return application.WConsole.queueUpdateItem(
                null,
                option,
                cb
            );
        });
    },

    /**
     *
     * @param caller
     * @param option
     * @param cb
     */
    queueSetState: function (caller, option, cb) {
        checkPermissions(caller, 'cc/queue', 'u', function (err) {
            if (err)
                return cb(err);

            if (!option || !option['name'] || !option['state']) {
                return cb(new CodeError(400, 'Name or state is required.'));
            };

            var domain = validateCallerParameters(caller, option['domain']);

            if (!domain) {
                return cb(new CodeError(400, 'Domain is required.'));
            };
            option['domain'] = domain;

            return application.WConsole.queueUpdateItemState(
                null,
                option,
                cb
            );
        });
    },

    /**
     * 
     * @param caller
     * @param option
     * @param cb
     */
    queueDelete: function (caller, option, cb) {
        checkPermissions(caller, 'cc/queue', 'd', function (err) {
            if (err)
                return cb(err);

            if (!option || !option['name']) {
                return cb(new CodeError(400, 'Name is required.'));
            };

            var domain = validateCallerParameters(caller, option['domain']);

            if (!domain) {
                return cb(new CodeError(400, 'Domain is required.'));
            };
            option['domain'] = domain;

            return application.WConsole.queueDelete(
                null,
                option,
                cb
            );
        });
    },

    /**
     *
     * @param caller
     * @param option
     * @param cb
     */
    tierCreate: function (caller, option, cb) {
        checkPermissions(caller, 'cc/queue', 'c', function (err) {
            if (err)
                return cb(err);

            if (!option || !option['queue'] || !option['agent']) {
                return cb(new CodeError(400, 'Queue, agent is required.'));
            };

            var domain = validateCallerParameters(caller, option['domain']);

            if (!domain) {
                return cb(new CodeError(400, 'Domain is required.'));
            };
            option['domain'] = domain;
            let _agentId = option['agent'] + '@' + domain,
                _tier = {
                    "agent": _agentId,
                    "level": "1",
                    "position": "1",
                    "queue": option['queue'] + '@' + domain,
                    "state": "Ready"
                };

            return application.WConsole.tierCreate(
                null,
                option,
                function (err, res) {
                    if (!err) {
                        let agent = application.Agents.get(_agentId);
                        if (!agent) {
                            let _tmp = {};
                            _tmp[_tier.queue] = _tier;
                            application.Agents.add(_agentId, _tmp)
                        } else {
                            agent[_tier.queue] = _tier
                        };
                    };
                    return cb(err, res);
                }
            );
        });
    },

    /**
     * 
     * @param caller
     * @param option
     * @param cb
     */
    tierSetLevel: function (caller, option, cb) {
        checkPermissions(caller, 'cc/queue', 'u', function (err) {
            if (err)
                return cb(err);

            if (!option || !option['queue'] || !option['agent'] || !option['level']) {
                return cb(new CodeError(400, 'Queue, agent, level is required.'));
            };

            var domain = validateCallerParameters(caller, option['domain']);

            if (!domain) {
                return cb(new CodeError(400, 'Domain is required.'));
            };
            option['domain'] = domain;

            return application.WConsole.tierSetLvl(
                null,
                option,
                cb
            );
        });
    },

    /**
     *
     * @param caller
     * @param option
     * @param cb
     */
    tierSetPosition: function (caller, option, cb) {
        checkPermissions(caller, 'cc/queue', 'u', function (err) {
            if (err)
                return cb(err);

            if (!option || !option['queue'] || !option['agent'] || !option['position']) {
                return cb(new CodeError(400, 'Queue, agent, position is required.'));
            };

            var domain = validateCallerParameters(caller, option['domain']);

            if (!domain) {
                return cb(new CodeError(400, 'Domain is required.'));
            };
            option['domain'] = domain;

            return application.WConsole.tierSetPos(
                null,
                option,
                cb
            );
        });
    },

    /**
     *
     * @param caller
     * @param option
     * @param cb
     */
    tierDelete: function (caller, option, cb) {
        checkPermissions(caller, 'cc/queue', 'd', function (err) {
            if (err)
                return cb(err);

            if (!option || !option['queue'] || !option['agent']) {
                return cb(new CodeError(400, 'Queue, agent is required.'));
            };

            var domain = validateCallerParameters(caller, option['domain']);

            if (!domain) {
                return cb(new CodeError(400, 'Domain is required.'));
            };
            option['domain'] = domain;
            let agentId = option['agent'] + '@' + domain;
            let queueId = option['queue'] + '@' + domain;

            return application.WConsole.tierRemove(
                null,
                option,
                function (err, res) {
                    if (!err) {
                        let agent = application.Agents.get(agentId);
                        if (agent && agent.hasOwnProperty(queueId)) {
                            delete agent[queueId];
                        };
                    };
                    return cb(err, res);
                }
            );
        });
    }

};

function getResponseOK (res) {
    return res['body'] && res['body'].indexOf('+OK') == 0
};

module.exports = Service;