/**
 * Created by Igor Navrotskyj on 04.09.2015.
 */

'use strict';

var log = require(__appRoot + '/lib/log')(module),
    authService = require(__appRoot + '/services/auth'),
    dialplanService = require(__appRoot + '/services/dialplan'),
    blackListService = require(__appRoot + '/services/blacklist'),
    contactBookService = require(__appRoot + '/services/contactBook'),
    callCenterService = require(__appRoot + '/services/callCentre'),
    dialerService = require(__appRoot + '/services/dialer'),
    calendarService = require(__appRoot + '/services/calendar'),
    emailService = require(__appRoot + '/services/email'),
    statusService = require(__appRoot + '/services/userStatus'),
    domainService = require(__appRoot + '/services/domain')
    ;

module.exports = function (event) {

    switch (event['Event-Name']) {
        case "USER_CREATE":
            onUserCreate(event['User-ID'], event['User-Domain']);
            break;
        case "USER_DESTROY":
            let userDomain = event['User-Domain'],
                userId = event['User-ID'] + '@' + userDomain
                ;
            onUserDelete(userId, userDomain);
            break;
        case "DOMAIN_CREATE":
            break;
        case "USER_MANAGED":
            // TODO
            return;
        case "DOMAIN_DESTROY":
            onDomainDelete(event['Domain-Name']);
            break;
        case "ACCOUNT_STATUS":
            onUserStatus(event);
            break;
    }
    event['webitel-event-name'] = event['Event-Name'];
    application.broadcastInDomain(event, event['Event-Domain']);
    application.broadcast(event);
};
// TODO move to response command
function onUserDelete (userId, domain) {
    var user = application.Users.get(userId);
    if (user) {
        user.disconnect();
    };

    authService.removeFromUserName(userId, domain, function (err) {
        if (err) {
            return log.error(err);
        };
        log.debug('Token destroyed %s', userId);
    });

    dialplanService._removeExtension(userId, domain, function (err) {
        if (err) {
            return log.error(err);
        };
        log.debug('Extension destroy %s', userId);
    });

    statusService._removeByUserId(userId, domain, function (err) {
        if (err) {
            return log.error(err);
        };
        log.debug('Statuses destroyed %s', userId);
    });

    callCenterService._removeById(userId, function (err) {
        if (err) {
            return log.error(err);
        }

        log.debug('Agent destroyed %s', userId);
    });
};

// TODO move to response command
function onUserCreate (login, domain) {
    var userId = login + '@' + domain;
    var extension = getTemplateExtension(login, domain);

    dialplanService._createExtension(extension, function (err) {
        if (err) {
            return log.error(err);
        };
        log.debug('Extension created %s', userId);
    });

};

// TODO move to response command
function onDomainDelete (domainName) {
    blackListService._removeByDomain(domainName, function (err, result) {
        if (err) {
            return log.error(err);
        };
        log.debug('Black list destroy %s from domain %s', result && result.n, domainName);
    });

    dialplanService._removeDefaultByDomain(domainName, function (err, result) {
        if (err) {
            return log.error(err);
        };
        log.debug('Default routes destroy %s from domain %s', result, domainName);
    });

    dialplanService._removePublicByDomain(domainName, function (err, result) {
        if (err) {
            return log.error(err);
        };
        log.debug('Public routes destroy %s from domain %s', result, domainName);
    });

    dialplanService._removeVariablesByDomain(domainName, function (err, result) {
        if (err) {
            return log.error(err);
        };
        log.debug('Route variables destroy %s from domain %s', result, domainName);
    });

    contactBookService._removeByDomain(domainName, function (err, result) {
        if (err) {
            return log.error(err);
        };
        log.debug('Contact book destroy %s from domain %s', result && result.n, domainName);
    });

    emailService._removeByDomain(domainName, function (err, result) {
        if (err) {
            return log.error(err);
        };
        log.debug('EMail settings destroy %s from domain %s', result && result.n, domainName);
    });

    domainService.settings._remove(domainName, function (err, result) {
        if (err) {
            return log.error(err);
        }
        log.debug('Domain settings destroy %s from domain %s', result && result.n, domainName);
    });
    
    callCenterService._removeByDomainName(domainName, function (err, result) {
        if (err) {
            return log.error(err);
        }
        log.debug('Agent settings destroy %s from domain %s', result && result.n, domainName);
    });

    dialerService._removeByDomain(domainName, function (err, result) {
        if (err) {
            return log.error(err);
        }
        log.debug('Dialer settings destroy %s from domain %s', result && result.n, domainName);
    });

    calendarService._removeByDomain(domainName, function (err, result) {
        if (err) {
            return log.error(err);
        }
        log.debug('Calendar settings destroy %s from domain %s', result && result.n, domainName);
    });
}

function onUserState (event) {

};

function onUserStatus (jsonEvent) {

    let domainName = jsonEvent['Account-Domain'],
        userId = jsonEvent['Account-User'] + '@' + domainName,
        status = jsonEvent['Account-Status'],
        state = jsonEvent['Account-User-State'],
        description = jsonEvent['Account-Status-Descript'] || "",
        user = application.Users.get(userId)
        ;

    if (description) {
        description = decodeURI(description);
        jsonEvent['Account-Status-Descript'] = description;
    }

    if (user) {
        user.setState(state, status, description);
        jsonEvent['Account-Online'] = true;
        jsonEvent['cc_logged'] = !!user['cc-logged'];
    } else {
        jsonEvent['Account-Online'] = false;
        jsonEvent['cc_logged'] = false;
    }

    jsonEvent['ws'] = !!user;
    jsonEvent['presence_id'] = userId;

    statusService.insert(jsonEvent);
}

// @private
function getTemplateExtension(number, domain) {
    return {
        "destination_number": number,
        "domain": domain,
        "userRef": number + '@' + domain,
        "name": "ext_" + number,
        "version": 2,
        "callflow": [

            {
                "setVar": "all:bridge_generate_comfort_noise=true"
            },
            {
                "ringback": {
                    "call": {
                        "name": "%(800,3200,425)",
                        "type": "tone"
                    },
                    "transfer": {
                        "name": "%(800,3200,425)",
                        "type": "tone"
                    }
                }
            },
            {
                "setVar": [
                    "hangup_after_bridge=true",
                    "continue_on_fail=true"
                ]
            },
            {
                "userData": {
                    "name": `${number}`,
                    "var": "effective_caller_id_name",
                    "setVar": "effective_callee_id_name"
                }
            },
            {
                "recordSession": {
                    "action": "start",
                    "stereo": false,
                    "type": "mp3"
                }
            },
            {
                "setVar": "all:origination_callee_id_name=${effective_callee_id_name}"
            },
            {
                "bridge": {
                    "endpoints": [
                        {
                            "name": `${number}`,
                            "type": "user",
                            "parameters": [
                                "leg_timeout=40",
                                `origination_callee_id_number=${number}`
                            ]
                        }
                    ]
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
                    "user": `${number}`
                }
            },
            {
                "hangup": ""
            }
        ]
    }
}