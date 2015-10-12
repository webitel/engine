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
            "domain" : 1,
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
            "_name": "cdrFields",
            "_": {
                "callflow.times.created_time" : -1,
                "variables.uuid" : 1,
                "callflow.caller_profile.caller_id_name" : 1,
                "callflow.caller_profile.caller_id_number" : 1,
                "callflow.caller_profile.callee_id_number" : 1,
                "callflow.caller_profile.callee_id_name" : 1,
                "callflow.caller_profile.destination_number" : 1,
                "callflow.times.answered_time" : 1,
                "callflow.times.bridged_time" : 1,
                "callflow.times.hangup_time" : 1,
                "variables.duration" : 1,
                "variables.hangup_cause" : 1,
                "variables.billsec" : 1,
                "variables.direction" : 1
            }
        }
    ],
    "collectionFile": [{
        "uuid": 1,
        "domain": 1
    }],
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
    }]
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
                    .ensureIndex(_index, {unique: item['_'] ? true : false, name: item['_name']}, function (err, res) {
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

