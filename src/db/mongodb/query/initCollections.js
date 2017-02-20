/**
 * Created by i.navrotskyj on 12.10.2015.
 */
'use strict';
var Collections = require(__appRoot + '/conf/config.json').mongodb,
    log = require(__appRoot + '/lib/log')(module)
    ;

var Indexes = {
    "collectionPublic": [{
        "_unique": true,
        "_": {
            "destination_number" : 1
        }
    }],
    "collectionDefault": [{
        "_unique": true,
        "_": {
            "domain" : 1,
            "destination_number" : 1
        }
    }],
    "collectionExtension": [{
        "_unique": true,
        "_": {
            "domain" : 1,
            "destination_number" : 1
        }
    }],
    "collectionDomainVar": [{
        "domain" : 1
    }],
    "collectionAuth": [{
        "_unique": true,
        "_": {
            "key" : 1,
            "domain" : 1,
            "username" : 1,
            "expires" : 1
        }
    }],
    "collectionCDR": [
        {
            "variables.domain_name" : 1
        },
        {
            "variables.uuid" : 1
        },
        {
            "variables.loopback_leg" : 1
        },
        {
            "callflow.times.created_time" : 1
        },
        {
            "callflow.caller_profile.caller_id_name" : 1
        },
        {
            "callflow.caller_profile.caller_id_number" : 1
        },
        {
            "callflow.caller_profile.callee_id_number" : 1
        },
        {
            "callflow.caller_profile.callee_id_name" : 1
        },
        {
            "callflow.caller_profile.destination_number" : 1
        },
        {
            "callflow.times.answered_time" : 1
        },
        {
            "callflow.times.bridged_time" : 1
        },
        {
            "callflow.times.hangup_time" : 1
        },
        {
            "variables.duration" : 1
        },
        {
            "variables.hangup_cause" : 1
        },
        {
            "variables.billsec" : 1
        },
        {
            "variables.webitel_direction" : 1
        },
        {
            "_elasticExportError" : 1
        }
    ],
    "collectionFile": [
        {
            "uuid": 1,
            "domain": 1
        },
        {
            "createdOn": 1
        },
        {
            "size": 1
        },
        {
            "domain": 1
        },
        {
            "_unique": true,
            "_": {
                "domain": 1,
                "path": 1
            }
        }
    ],
    "collectionDomain": [
        {
            "_unique": true,
            "_": {
                "name" : 1
            }
        }
    ],
    "collectionEmail": [{
        "domain": 1,
        "provider": 1
    }],
    "collectionBlackList": [{
        "_unique": true,
        "_": {
            "domain" : 1,
            "name" : 1,
            "number" : 1
        }
    }],
    "collectionConference": [{
            "email": 1,
            "confirmed": 1
        },
        {
            "_createdOn": 1
        }
    ],
    "collectionLocation": [{
        "sysOrder": 1,
        "sysLength": 1,
        "code": 1
    }],
    "collectionHook": [
        {
            "_id": 1,
            "domain": 1
        },
        {
            "enable": 1,
            "domain": 1,
            "event": 1
        }
    ],
    "collectionAclPermissions": [{
        "_unique": true,
        "_": {
            "roles": 1
        }
    }],
    "collectionMedia": [{
        "_unique": true,
        "_": {
            "name": 1,
            "domain": 1,
            "type": 1
        }
    }],
    "collectionDialerMembers": [
        {
            "_probeCount" : 1
        },
        {
            "_endCause" : 1
        },
        {
            "_nextTryTime" : 1
        },
        {
            "dialer" : 1,
            "_lock" : 1
        },
        {
            "name" : 1
        },
        {
            "_waitingForResultStatus" : 1
        },
        {
            "priority" : 1
        },
        {
            "_unique": true,
            "_name": "number_unique",
            "_": {
                "communications.number" : 1,
                "dialer": 1
            }
        },
        {
            "_name": "hunting",
            "_": {
                _nextTryTime: -1,
                priority: -1,
                _id: -1,
                dialer: 1,
                _endCause: 1,
                _lock: 1,
                "communications.state": 1,
                "communications.gatewayPositionMap": 1
            }

        },
        {
            "_name": "communicationsRangeId",
            "_unique": false,
            "_": {
                "communications._range.rangeId" : 1
            }

        },
        {
            "_name": "communicationsType",
            "_unique": false,
            "_": {
                "communications.type" : 1
            }

        },
        {
            "_name": "communicationsAttempts",
            "_unique": false,
            "_": {
                "communications._range.attempts" : 1
            }

        }
    ],
    "collectionDialerAgents": [
        {
            "randomPoint" : "2d"
        }
    ],
    "collectionAgentStatus": [
        {
            "domain" : 1
        },
        {
            "account" : 1
        },
        {
            "status" : 1
        },
        {
            "state" : 1
        },
        {
            "online" : 1
        },
        {
            "date" : 1
        },
        {
            "endDate" : 1
        }
    ],
    "collectionGateway": [
        {
            "name" : 1
        },
        {
            "params.realm" : 1
        },
        {
            "domain" : 1
        }
    ]
};

function Init (db) {
    for (let key in Collections) {
        if (Collections.hasOwnProperty(key) && Indexes.hasOwnProperty(key)) {
            var index = Indexes[key];
            index.forEach(function (item) {
                var _index = (item['_'] instanceof Object)
                    ? item['_']
                    : item
                ;
                db
                    .collection(Collections[key])
                    .ensureIndex(_index, {unique: item['_'] && item['_unique'] !== false ? true : false, name: item['_name']}, function (err, res) {
                        if (err)
                            return log.error(err);
                        return log.debug("Ensure index %s [%s]", res, Collections[key]);
                    })
                ;
            });
        };
    };
};

module.exports = Init;

