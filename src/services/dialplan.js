/**
 * Created by Igor Navrotskyj on 27.08.2015.
 */

'use strict';

const CodeError = require(__appRoot + '/lib/error'),
    validateCallerParameters = require(__appRoot + '/utils/validateCallerParameters'),
    channelServices = require('./channel'),
    deleteDomainFromStr = require(__appRoot + '/utils/parse').deleteDomainFromStr,
    checkPermissions = require(__appRoot + '/middleware/checkPermissions'),
    expVal = require(__appRoot + '/utils/validateExpression')
    ;

const Service = {
    
    /**
     *
     * @param caller
     * @param request
     * @param cb
     * @returns {*}
     */
    listExtension: function (caller, request, cb) {
        checkPermissions(caller, 'rotes/extension', 'r', function (err) {
            if (err)
                return cb(err);

            const domain = validateCallerParameters(caller, request.domain);
            if (!domain) {
                return cb(new CodeError(400, 'Domain is required.'));
            }
            request.domain = domain;
            application.PG.getQuery('dialplan').listExtension(request, cb);
        });
    },

    /**
     *
     * @param caller
     * @param options
     * @param cb
     */
    itemExtension: function (caller, options = {}, cb) {
        checkPermissions(caller, 'rotes/extension', 'r', function (err) {
            if (err)
                return cb(err);

            const domain = validateCallerParameters(caller, options.domain);
            if (!domain) {
                return cb(new CodeError(400, 'Domain is required.'));
            }

            if (!options.id) {
                return cb(new CodeError(400, 'Id is required.'));
            }

            application.PG.getQuery('dialplan').itemExtension(options.id, domain, cb);
        });
    },

    /**
     *
     * @param option
     * @param cb
     * @returns {*}
     */
    updateExtension: function (caller, options, cb) {
        checkPermissions(caller, 'rotes/extension', 'u', function (err) {
            if (err)
                return cb(err);

            if (!options.id) {
                return cb(new CodeError(400, 'Id is required.'));
            }

            const domain = validateCallerParameters(caller, options.domain);

            if (!domain) {
                return cb(new CodeError(400, 'Domain is required.'));
            }
            replaceExpression(options);
            application.PG.getQuery('dialplan').updateExtension(options.id, domain, options, cb);
        });
    },
    
    _removeExtension: function (userId, domain, cb) {
        if (!domain || !userId) {
            return cb(new CodeError(400, "Domain is required."));
        }
        application.PG.getQuery('dialplan').removeExtensionFromUser(userId, domain, cb);
    },
    
    _createExtension: function (dialplan, cb) {
        application.PG.getQuery('dialplan').createExtension(dialplan, cb);
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

        application.PG.getQuery('dialplan').existsExtension(number, domain, cb);
    },
    
    _updateOrInsertExtension: function (userId, number, domain, cb) {
        if (!userId ||!number || !domain)
            return cb(new CodeError(400, "Domain, userId or number is require."));

        var _userExtension = getTemplateExtension(userId, number, domain);
        var dbDialplan = application.DB._query.dialplan;
        dbDialplan.updateOrInsertExtension(userId, domain, _userExtension, cb);
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

            replaceExpression(dialplan);
            dialplan.domain = domain;
            dialplan.version = 2;
            application.PG.getQuery('dialplan').createDefault(dialplan, cb);
        });
    },

    /**
     *
     * @param domain
     * @param cb
     * @returns {*}
     */
    listDefault: function (caller, request, cb) {
        checkPermissions(caller, 'rotes/default', 'r', function (err) {
            if (err)
                return cb(err);

            const domain = validateCallerParameters(caller, request.domain);
            if (!domain) {
                return cb(new CodeError(400, 'Domain is required.'));
            }
            request.domain = domain;
            application.PG.getQuery('dialplan').listDefault(request, cb);
        });
    },

    itemDefault: function (caller, options, cb) {
        checkPermissions(caller, 'rotes/default', 'r', function (err) {
            if (err)
                return cb(err);

            const domain = validateCallerParameters(caller, options.domain);
            if (!domain) {
                return cb(new CodeError(400, 'Domain is required.'));
            }

            application.PG.getQuery('dialplan').itemDefault(options.id, domain, cb);
        });
    },

    /**
     *
     * @param _id
     * @param cb
     * @returns {*}
     */
    removeDefault: function (caller, options = {}, cb) {
        checkPermissions(caller, 'rotes/default', 'd', function (err) {
            if (err)
                return cb(err);

            if (!options.id) {
                return cb(new CodeError(400, 'Id is required.'));
            }
            const domain = validateCallerParameters(caller, options.domain);

            if (!domain) {
                return cb(new CodeError(400, 'Domain is required.'));
            }

            application.PG.getQuery('dialplan').deleteDefault(options.id, domain, cb);
        });
    },

    /**
     *
     * @param _id
     * @param option
     * @param cb
     * @returns {*}
     */
    updateDefault: function (caller, options = {}, cb) {
        checkPermissions(caller, 'rotes/default', 'u', function (err) {
            if (err)
                return cb(err);

            if (!options.id) {
                return cb(new CodeError(400, 'Id is required.'));
            }

            const domain = validateCallerParameters(caller, options.domain);

            if (!domain) {
                return cb(new CodeError(400, 'Domain is required.'));
            }
            replaceExpression(options);
            application.PG.getQuery('dialplan').updateDefault(options.id, domain, options, cb);
        });
    },

    move: function (caller, options, cb) {
        checkPermissions(caller, 'rotes/default', 'u', function (err) {
            if (err)
                return cb(err);

            if (!options.id) {
                return cb(new CodeError(400, 'Id is required.'));
            }

            const domain = validateCallerParameters(caller, options.domain);

            if (!domain) {
                return cb(new CodeError(400, 'Domain is required.'));
            }
            application.PG.getQuery('dialplan').moveDefault(options.id, domain, options.up, cb);
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

            replaceExpression(dialplan);
            dialplan['domain'] = domain;
            dialplan['version'] = 2;
            application.PG.getQuery('dialplan').createPublic(dialplan, cb);
        });
    },

    /**
     *
     * @param domain
     * @param cb
     * @returns {*}
     */
    listPublic: function (caller, request = {}, cb) {
        checkPermissions(caller, 'rotes/public', 'r', function (err) {
            if (err)
                return cb(err);

            const domain = validateCallerParameters(caller, request.domain);
            if (!domain) {
                return cb(new CodeError(400, 'Domain is required.'));
            }
            request.domain = domain;
            application.PG.getQuery('dialplan').listPublic(request, cb);
        });
    },

    itemPublic: function (caller, options = {}, cb) {
        checkPermissions(caller, 'rotes/public', 'r', function (err) {
            if (err)
                return cb(err);

            const domain = validateCallerParameters(caller, options.domain);
            if (!domain) {
                return cb(new CodeError(400, 'Domain is required.'));
            }

            if (!options.id) {
                return cb(new CodeError(400, 'Id is required.'));
            }

            application.PG.getQuery('dialplan').itemPublic(options.id, domain, cb);
        });
    },

    /**
     *
     * @param _id
     * @param cb
     * @returns {*}
     */
    removePublic: function (caller, options = {}, cb) {
        checkPermissions(caller, 'rotes/public', 'd', function (err) {
            if (err)
                return cb(err);

            if (!options.id) {
                return cb(new CodeError(400, 'Id is required.'));
            }
            const domain = validateCallerParameters(caller, options.domain);

            if (!domain) {
                return cb(new CodeError(400, 'Domain is required.'));
            }

            application.PG.getQuery('dialplan').deletePublic(options.id, domain, cb);
        });
    },

    /**
     *
     * @param _id
     * @param option
     * @param cb
     * @returns {*}
     */
    updatePublic: function (caller, options, cb) {
        checkPermissions(caller, 'rotes/public', 'u', function (err) {
            if (err)
                return cb(err);

            if (!options.id) {
                return cb(new CodeError(400, 'Id is required.'));
            }

            const domain = validateCallerParameters(caller, options.domain);

            if (!domain) {
                return cb(new CodeError(400, 'Domain is required.'));
            }
            replaceExpression(options);
            application.PG.getQuery('dialplan').updatePublic(options.id, domain, options, cb);
        });
    },

    /**
     *
     * @param caller
     * @param request
     * @param cb
     */
    listDomainVariables: function (caller, request = {}, cb) {
        checkPermissions(caller, 'rotes/domain', 'r', function (err) {
            if (err)
                return cb(err);

            const domain = validateCallerParameters(caller, request.domain);
            if (!domain) {
                return cb(new CodeError(400, 'Domain is required.'));
            }
            request.domain = domain;
            application.PG.getQuery('dialplan').listDomainVariables(request, cb);
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
        checkPermissions(caller, 'rotes/domain', 'c', function (err) {
            if (err)
                return cb(err);

            domain = validateCallerParameters(caller, domain);
            if (!domain || !variables) {
                return cb(new CodeError(400, 'Bad request.'));
            }
            application.PG.getQuery('dialplan').insertOrUpdateDomainVariable(domain, variables, cb);
        });
    },
    
    _removeDefaultByDomain: function (domain, cb) {
        if (!domain) {
            return cb(new CodeError(400, 'Domain is required.'));
        }
        application.PG.getQuery('dialplan').removeDefaultByDomain(domain, cb);
    },

    _removePublicByDomain: function (domain, cb) {
        if (!domain) {
            return cb(new CodeError(400, 'Domain is required.'));
        }
        application.PG.getQuery('dialplan').removePublicByDomain(domain, cb);
    },

    _removeVariablesByDomain: function (domain, cb) {
        if (!domain) {
            return cb(new CodeError(400, 'Domain is required.'));
        }
        application.PG.getQuery('dialplan').deleteVariables(domain, cb);
    },

    debugPublic: function (caller, options, cb) {
        checkPermissions(caller, 'rotes/public', 'r', function (err) {
            if (err)
                return cb(err);

            return makeDebugAcrCall(caller, options, 'public', cb)
        })
    },
    
    debugDefault: function (caller, options, cb) {
        checkPermissions(caller, 'rotes/default', 'r', function (err) {
            if (err)
                return cb(err);

            return makeDebugAcrCall(caller, options, 'default', cb)
        })
    }

};


function makeDebugAcrCall(caller, options = {}, context, cb) {
    if (!options.number) {
        return cb(new CodeError(400, 'Number is required.'))
    }

    if (!options.uuid) {
        return cb(new CodeError(400, 'Uuid is required.'))
    }

    if (!options.from) {
        return cb(new CodeError(400, 'From is required.'))
    }

    if (!caller.domain) {
        if (!options.domain) {
            return cb(new CodeError(400, "From or domain is required."))
        }

        return channelServices.bgApi(`originate [domain_name=${options.domain},origination_uuid=${options.uuid},origination_caller_id_number=${options.number},` +
            `webitel_direction=debug,webitel_debug_acr=true]user/${options.from} ${options.number} XML ${context}`, cb);

    } else {

        return channelServices.bgApi(`originate [domain_name=${caller.domain},origination_uuid=${options.uuid},origination_caller_id_number=${options.number},` +
            `webitel_direction=debug,webitel_debug_acr=true]user/${deleteDomainFromStr(options.from)}@${caller.domain} ${options.number} XML ${context}`, cb);

    }
}

module.exports = Service;

function replaceExpression(obj) {
    if (obj)
        for (let key in obj) {
            if (typeof obj[key] === "object")
                replaceExpression(obj[key]);
            else if (typeof obj[key] !== "function" && key === "expression") {
                obj["sysExpression"] = expVal(obj[key]);
            }
        }
    return;
}

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
}