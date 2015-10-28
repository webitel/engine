/**
 * Created by Igor Navrotskyj on 04.09.2015.
 */

'use strict';

var log = require(__appRoot + '/lib/log')(module),
    authService = require(__appRoot + '/services/auth'),
    dialplanService = require(__appRoot + '/services/dialplan'),
    blackListService = require(__appRoot + '/services/blacklist'),
    contactBookService = require(__appRoot + '/services/contactBook'),
    emailService = require(__appRoot + '/services/email'),
    statusService = require(__appRoot + '/services/userStatus')
    ;

module.exports = function (event) {
    log.debug(event.type);
    switch (event.type) {
        case "USER_CREATE":
            onUserCreate(event.getHeader('User-ID'), event.getHeader('User-Domain'));
            break;

        case "USER_DESTROY":
            var userDomain = event.getHeader('User-Domain');
            var userId = event.getHeader('User-ID') + '@' + userDomain;
            onUserDelete(userId, userDomain);
            break;

        case "DOMAIN_CREATE":
            break;

        case "DOMAIN_DESTROY":
            onDomainDelete(event.getHeader('Domain-Name'));
            break;

        case "USER_STATE":
            onUserState(event);
            break;

        case "ACCOUNT_STATUS":
            onUserStatus(event);
            break;
    };
    var _e = event.serialize('json', 1);
    _e['webitel-event-name'] = _e['Event-Name'];
    application.broadcastInDomain(_e, event.getHeader('Event-Domain'));
    // TODO
    application.broadcast(_e);
};

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
};

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
        log.debug('Default routes destroy %s from domain %s', result && result.n, domainName);
    });

    dialplanService._removePublicByDomain(domainName, function (err, result) {
        if (err) {
            return log.error(err);
        };
        log.debug('Public routes destroy %s from domain %s', result && result.n, domainName);
    });

    dialplanService._removeVariablesByDomain(domainName, function (err, result) {
        if (err) {
            return log.error(err);
        };
        log.debug('Route variables destroy %s from domain %s', result && result.n, domainName);
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
};

function onUserState (event) {
    // TODO delete event
    /*
    let domainName = event.getHeader('User-Domain'),
        userId = event.getHeader('User-ID') + '@' + domainName,
        state = event.getHeader('User-State'),
        user = application.Users.get(userId);

    if (user) {
        //user.setState(state);
    };
    */
};

function onUserStatus (event) {
    var jsonEvent = event.serialize('json', true);
    let domainName = jsonEvent['Account-Domain'],
        userId = jsonEvent['Account-User'] + '@' + domainName,
        status = jsonEvent['Account-Status'],
        state = jsonEvent['Account-User-State'],
        user = application.Users.get(userId);
    if (user) {
        user.setState(state, status);
    };

    var data = {
        "domain": domainName,
        "account": jsonEvent['Account-User'],
        "status": status,
        "state": state,
        // TODO
        "description": jsonEvent['Account-Status-Descript'] ? decodeURI(decodeURI(jsonEvent['Account-Status-Descript'])) : '',
        "online": !!user
    };

    statusService.insert(data);
};

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