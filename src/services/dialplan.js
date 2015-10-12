/**
 * Created by Igor Navrotskyj on 27.08.2015.
 */

'use strict';

var CodeError = require(__appRoot + '/lib/error'),
    validateCallerParameters = require(__appRoot + '/utils/validateCallerParameters'),
    checkPermissions = require(__appRoot + '/middleware/checkPermissions'),
    expVal = require(__appRoot + '/utils/validateExpression')
    ;

var Service = {
    
    /**
     *
     * @param domain
     * @param cb
     * @returns {*}
     */
    getExtensions: function (caller, domain, cb) {
        try {
            checkPermissions(caller, 'rotes/extension', 'r', function (err) {
                if (err)
                    return cb(err);

                domain = validateCallerParameters(caller, domain);
                if (!domain) {
                    return cb(new CodeError(400, 'Domain is required.'));
                };

                var dbDialplan = application.DB._query.dialplan;
                dbDialplan.getExtensions(domain, cb);
            });
        } catch (e) {
            cb(e);
        };
    },

    /**
     *
     * @param option
     * @param cb
     * @returns {*}
     */
    updateExtension: function (caller, _id, option, cb) {
        try {
            checkPermissions(caller, 'rotes/extension', 'u', function (err) {
                if (err)
                    return cb(err);

                option = option || {};
                var callflow = option['callflow'],
                    timezone = option['timezone'],
                    timezonename = option['timezonename'],
                    extension = {
                        "$set": {}
                    };

                if (!_id || option['destination_number'] || option['domain'] || option['userRef'] || option['version']) {
                    return cb(new CodeError(400, 'Bad request.'));
                }
                ;

                for (var key in option) {
                    if (key == 'callflow')
                        replaceExpression(option[key]);
                    if (option[key])
                        extension.$set[key] = option[key];
                };

                var dbDialplan = application.DB._query.dialplan;
                dbDialplan.updateExtension(_id, extension, cb);
            });
            return 1;
        } catch (e) {
            cb(e);
        };
    },
    
    _removeExtension: function (userId, domain, cb) {
        if (!domain || !userId) {
            return cb(new CodeError(400, "Domain is required."));
        };
        var dbDialplan = application.DB._query.dialplan;
        dbDialplan.removeExtensionFromUser(userId, domain, cb);
    },
    
    _createExtension: function (dialplan, cb) {
        var dbDialplan = application.DB._query.dialplan;
        dbDialplan.createExtension(dialplan, cb);
    },

    /**
     *
     * @param number
     * @param domain
     * @param cb
     * @returns {*}
     * @private
     */
    _existsExtension: function (number, domain, cb) {
        if (!number || !domain)
            return cb(new CodeError(400, "Domain or number is require."));

        var dbDialplan = application.DB._query.dialplan;
        dbDialplan.existsExtension(number, domain, cb);
    },
    
    _updateOrInsertExtension: function (userId, number, domain, cb) {
        if (!userId ||!number || !domain)
            return cb(new CodeError(400, "Domain, userId or number is require."));

        var _userExtension = getTemplateExtension(userId, number, domain);
        var dbDialplan = application.DB._query.dialplan;
        dbDialplan.updateOrInsertExtension(number, domain, _userExtension, cb);
    },

    /**
     *
     * @param domain
     * @param dialplan
     * @param cb
     * @returns {*}
     */
    createDefault: function (caller, domain, dialplan, cb) {
        checkPermissions(caller, 'rotes/default', 'c', function (err) {
            if (err)
                return cb(err);

            domain = validateCallerParameters(caller, domain);
            if (!domain) {
                return cb(new CodeError(400, 'Bad request.'));
            }
            ;
            try {
                replaceExpression(dialplan);
                dialplan['createdOn'] = new Date().toString();
                dialplan['domain'] = domain;
                if (!dialplan['order']) {
                    dialplan['order'] = 0;
                }
                ;
                dialplan['version'] = 2;
                var dbDialplan = application.DB._query.dialplan;
                dbDialplan.createDefault(dialplan, cb);
            } catch (e) {
                return cb(e);
            };
        });
    },

    /**
     *
     * @param domain
     * @param cb
     * @returns {*}
     */
    getDefault: function (caller, domain, cb) {
        try {
            checkPermissions(caller, 'rotes/default', 'r', function (err) {
                if (err)
                    return cb(err);

                domain = validateCallerParameters(caller, domain);
                if (!domain) {
                    return cb(new CodeError(400, 'Domain is required.'));
                }
                ;

                var dbDialplan = application.DB._query.dialplan;
                dbDialplan.getDefault(domain, cb);
            });
        } catch (e) {
            cb(e);
        };
    },

    /**
     *
     * @param _id
     * @param cb
     * @returns {*}
     */
    removeDefault: function (caller, _id, cb) {
        checkPermissions(caller, 'rotes/default', 'd', function (err) {
            if (err)
                return cb(err);

            if (!_id) {
                return cb(new CodeError(400, 'Id is required.'));
            }
            ;
            var dbDialplan = application.DB._query.dialplan;
            dbDialplan.removeDefault(_id, cb);
        });
    },

    /**
     *
     * @param _id
     * @param option
     * @param cb
     * @returns {*}
     */
    updateDefault: function (caller, _id, option, cb) {
        checkPermissions(caller, 'rotes/default', 'u', function (err) {
            if (err)
                return cb(err);

            if (!_id) {
                return cb(new CodeError(400, 'Id is required.'));
            }
            ;
            option = option || {};
            option['domain'] = validateCallerParameters(caller, option['domain']);

            if (!option['domain']) {
                return cb(new CodeError(400, 'Domain is required.'));
            }
            ;

            replaceExpression(option);
            option['version'] = 2;
            var dbDialplan = application.DB._query.dialplan;
            return dbDialplan.updateDefault(_id, option, cb);
        });
    },

    /**
     *
     * @param domain
     * @param option
     * @param cb
     * @returns {*}
     */
    incOrderDefault: function (caller, domain, option, cb) {
        checkPermissions(caller, 'rotes/default', 'u', function (err) {
            if (err)
                return cb(err);

            domain = validateCallerParameters(caller, domain);
            if (!domain || !option || isNaN(option['inc']) || isNaN(option['start'])) {
                return cb(new CodeError(400, 'Bad request.'));
            }
            ;
            var dbDialplan = application.DB._query.dialplan;
            return dbDialplan.incOrderDefault(domain, option, cb);
        });
    },

    /**
     *
     * @param _id
     * @param option
     * @param cb
     * @returns {*}
     */
    setOrderDefault: function (caller, _id, option, cb) {
        checkPermissions(caller, 'rotes/default', 'u', function (err) {
            if (err)
                return cb(err);

            if (!_id || !option || isNaN(option['order'])) {
                return cb(new CodeError(400, 'Bad request.'));
            }
            ;
            var dbDialplan = application.DB._query.dialplan;
            return dbDialplan.setOrderDefault(_id, option, cb);
        });
    },

    /**
     *
     * @param domain
     * @param dialplan
     * @param cb
     * @returns {*}
     */
    createPublic: function (caller, domain, dialplan, cb) {
        checkPermissions(caller, 'rotes/public', 'c', function (err) {
            if (err)
                return cb(err);

            domain = validateCallerParameters(caller, domain);
            if (!domain || !dialplan || !dialplan['destination_number']) {
                return cb(new CodeError(400, 'Bad request.'));
            }
            ;
            try {
                replaceExpression(dialplan);
                dialplan['createdOn'] = new Date().toString();
                dialplan['domain'] = domain;
                if (!dialplan['order']) {
                    dialplan['order'] = 0;
                }
                ;
                dialplan['version'] = 2;
                var dbDialplan = application.DB._query.dialplan;
                dbDialplan.createPublic(dialplan, cb);
            } catch (e) {
                return cb(e);
            }
            ;
        });
    },

    /**
     *
     * @param domain
     * @param cb
     * @returns {*}
     */
    getPublic: function (caller, domain, cb) {
        try {
            checkPermissions(caller, 'rotes/public', 'r', function (err) {
                if (err)
                    return cb(err);

                domain = validateCallerParameters(caller, domain);
                if (!domain) {
                    return cb(new CodeError(400, 'Domain is required.'));
                }
                ;

                var dbDialplan = application.DB._query.dialplan;
                dbDialplan.getPublic(domain, cb);
            });
        } catch (e) {
            cb(e);
        };
    },

    /**
     *
     * @param _id
     * @param cb
     * @returns {*}
     */
    removePublic: function (caller, _id, cb) {
        checkPermissions(caller, 'rotes/public', 'd', function (err) {
            if (err)
                return cb(err);

            if (!_id) {
                return cb(new CodeError(400, 'Id is required.'));
            }
            ;
            var dbDialplan = application.DB._query.dialplan;
            dbDialplan.removePublic(_id, cb);
        });
    },

    /**
     *
     * @param _id
     * @param option
     * @param cb
     * @returns {*}
     */
    updatePublic: function (caller, _id, option, cb) {
        checkPermissions(caller, 'rotes/public', 'u', function (err) {
            if (err)
                return cb(err);

            if (!_id || !option) {
                return cb(new CodeError(400, 'Id is required.'));
            }
            ;
            option = option || {};
            option['domain'] = validateCallerParameters(caller, option['domain']);

            if (!option['domain']) {
                return cb(new CodeError(400, 'Domain is required.'));
            }
            ;

            replaceExpression(option);
            option['version'] = 2;
            var dbDialplan = application.DB._query.dialplan;
            return dbDialplan.updatePublic(_id, option, cb);
        });
    },

    /**
     *
     * @param caller
     * @param domain
     * @param cb
     */
    getDomainVariable: function (caller, domain, cb) {
        checkPermissions(caller, 'rotes/extension', 'r', function (err) {
            if (err)
                return cb(err);

            domain = validateCallerParameters(caller, domain);
            if (!domain) {
                return cb(new CodeError(400, 'Domain is required.'));
            }
            ;
            var dbDialplan = application.DB._query.dialplan;
            dbDialplan.getDomainVariable(domain, cb);
        });
    },

    /**
     *
     * @param caller
     * @param domain
     * @param variables
     * @param cb
     * @returns {*}
     */
    insertOrUpdateDomainVariable: function (caller, domain, variables, cb) {
        checkPermissions(caller, 'rotes/extension', 'c', function (err) {
            if (err)
                return cb(err);

            domain = validateCallerParameters(caller, domain);
            if (!domain || !variables) {
                return cb(new CodeError(400, 'Bad request.'));
            }
            ;
            var dbDialplan = application.DB._query.dialplan;
            dbDialplan.insertOrUpdateDomainVariable(domain, variables, cb);
        });
    },
    
    _removeDefaultByDomain: function (domain, cb) {
        if (!domain) {
            return cb(new CodeError(400, 'Domain is required.'));
        }
        var dbDialplan = application.DB._query.dialplan;
        dbDialplan.removeDefaultByDomain(domain, cb);
    },

    _removePublicByDomain: function (domain, cb) {
        if (!domain) {
            return cb(new CodeError(400, 'Domain is required.'));
        }
        var dbDialplan = application.DB._query.dialplan;
        dbDialplan.removePublicByDomain(domain, cb);
    },

    _removeVariablesByDomain: function (domain, cb) {
        if (!domain) {
            return cb(new CodeError(400, 'Domain is required.'));
        }
        var dbDialplan = application.DB._query.dialplan;
        dbDialplan.removeVariablesByDomain(domain, cb);
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

// @@private
function getTemplateExtension(id, number, domain) {
    return {
        "destination_number": number,
        "domain": domain,
        "userRef": id + '@' + domain,
        "name": "ext_" + number,
        "version": 2,
        "callflow": [
            {
                "setVar": [ "ringback=$${us-ring}", "transfer_ringback=$${uk-ring}","hangup_after_bridge=true",
                    "continue_on_fail=true"]
            },
            {
                "recordSession": "start"
            },
            {
                "bridge": {
                    "endpoints": [{
                        "name": number,
                        "type": "user"
                    }]
                }
            },
            {
                "recordSession": "stop"
            },
            {
                "answer": ""
            },
            {
                "sleep": "1000"
            },
            {
                "voicemail": {
                    "user": number
                }
            }
        ]
    }
};