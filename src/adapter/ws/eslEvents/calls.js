/**
 * Created by igor on 23.06.16.
 */

let eventsService = require(__appRoot + '/services/events'),
    log = require(__appRoot + '/lib/log')(module),
    application = require(__appRoot + '/application'),
    validateCallerParameters = require(__appRoot + '/utils/validateCallerParameters'),
    getDomainFromSwitchEvent = require(__appRoot + '/utils/helper').getDomainFromSwitchEvent
    ;

const _srvEvents = [
    "CHANNEL_CREATE",
    "CHANNEL_DESTROY",
    "CHANNEL_CALLSTATE",
    "CHANNEL_STATE",
    "CHANNEL_ANSWER",
    "CHANNEL_HANGUP_COMPLETE",
    "CHANNEL_HANGUP",
    "CHANNEL_HOLD",
    "CHANNEL_UNHOLD",
    "CHANNEL_BRIDGE",
    "CHANNEL_UNBRIDGE",
    "DTMF"
];


module.exports = (application) => {
    for (var i = 0, len = _srvEvents.length; i < len; i++) {
        let e = eventsService.registered('SE:' + _srvEvents[i].toUpperCase());

        e.domains.on('removed', (_, domainName) => {
            if (domainName === 'root') {
                application.Broker.unBindDomainEvent(e.name.replace('SE:', ''), '*');
            } else {
                application.Broker.unBindDomainEvent(e.name.replace('SE:', ''), domainName);
            }
        });

        application.on(`unsubscribe::SE:${_srvEvents[i]}`, (args, caller, eventName) => {
            if (args && caller) {
                application.Broker.unBindDomainEvent(eventName.replace('SE:', ''), caller.domain || '*', (err) => {
                    if (err) {
                        log.error(err);
                    }
                });
            }
        });

        application.on(`subscribe::SE:${_srvEvents[i]}`, (args, caller, eventName) => {
            if (args && caller) {
                application.Broker.bindDomainEvent(eventName.replace('SE:', ''), caller.domain || '*', (err) => {
                    if (err) {
                        log.error(err);
                    }
                })
            }
        });
    }

    application.Broker.on('callDomainEvent', (e) => {
        if (e['Channel-Presence-Data'] || e['Channel-Presence-ID']) {
            e['Event-Name'] = 'SE:' + e['Event-Name'];
            eventsService.fire(e['Event-Name'], getDomainFromSwitchEvent(e), e);
            eventsService.fire(e['Event-Name'] , 'root', e);
        }
    });
};