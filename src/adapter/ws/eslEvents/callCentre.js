/**
 * Created by Igor Navrotskyj on 03.09.2015.
 */

'use strict';

var eventsService = require(__appRoot + '/services/events'),
    log = require(__appRoot + '/lib/log')(module),
    getDomainFromStr = require(__appRoot + '/utils/parse').getDomainFromStr
    ;

var _srvEvents = [
    'agent-offering',
    'bridge-agent-start',
    'bridge-agent-end',
    'bridge-agent-fail',
    'members-count',
    'member-queue-start',
    'member-queue-end',
    'agent-status-change',
    'agent-state-change'
];

for (var i = 0, len = _srvEvents.length; i < len; i++) {
    eventsService.registered('CC::' + _srvEvents[i].toUpperCase());
};

module.exports = function (event) {
    try {
        var domain = getDomainFromStr(event['CC-Queue']),
            eventName = "CC::" + event['CC-Action'].toUpperCase()
            ;
        event['Event-Name'] = eventName;
        eventsService.fire(eventName, domain, event);
        // TODO
        eventsService.fire(eventName, 'root', event);
    } catch(e) {
        log.error(e);
    };
};