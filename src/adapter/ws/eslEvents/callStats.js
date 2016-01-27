/**
 * Created by Igor Navrotskyj on 04.09.2015.
 */

// TODO: Delete me
'use strict';

var log = require(__appRoot + '/lib/log')(module),
    HashCollection = require(__appRoot + '/lib/collection'),
    eventsService = require(__appRoot + '/services/events');

const _EVENTS = {
    //CALL_EVENT_COUNT_NAME: 'SERVER::CALL-INFO',
    //CHANNEL_CREATE: 'CALL::CHANNEL_CREATE',
    //CHANNEL_DESTROY: 'CALL::CHANNEL_DESTROY'
};

for (var key in _EVENTS) {
    eventsService.registered(_EVENTS[key]);
};

var calls = new HashCollection('uuid');
var domains = new HashCollection('id');

// TODO (delete) for 1ec.
module.exports = function (jsonEvent) {
    if (jsonEvent['Event-Name'] == 'CHANNEL_DESTROY') {
        onHandleCallDestroy(jsonEvent);
        //eventsService.fire(_EVENTS.CHANNEL_DESTROY, jsonEvent["variable_domain_name"], jsonEvent);
    } else if (jsonEvent['Event-Name'] == 'CHANNEL_CREATE') {
        onHandleCallCreate(jsonEvent);
        //eventsService.fire(_EVENTS.CHANNEL_CREATE, jsonEvent["variable_domain_name"], jsonEvent);
    };
};

function onHandleCallCreate (e) {
    var callId = getCallIdFromEvent(e),
        call = calls.get(callId);

    if (!call && e["variable_domain_name"]) {
        call = {
            "domain": e["variable_domain_name"],
            "callerId": e["Caller-Caller-ID-Number"],
            "destination_number": e["Caller-Destination-Number"],
            "Event-Name": _EVENTS.CALL_EVENT_COUNT_NAME
        };
        calls.add(callId, call);
        var domain = domains.get(e["variable_domain_name"]);
        call['countCall'] = domain['countCall'];
        eventsService.fire(_EVENTS.CALL_EVENT_COUNT_NAME, e["variable_domain_name"], call);
        log.debug('ON NEW CALL %s, all call %s', e["variable_domain_name"], call['countCall']);
    };
};

function onHandleCallDestroy (e) {
    var callId = getCallIdFromEvent(e),
        call = calls.get(callId);

    if (call)
        calls.remove(callId);
};

calls.on('added', function (e) {
    var domain_id = e['domain'],
        domain = domains.get(domain_id);
    if (domain) {
        domain['countCall'] ++;
    } else {
        domain = {
            "countCall": 1
        };
        domains.add(domain_id, domain);
    };
});

calls.on('removed', function (e) {
    var domain_id = e['domain'],
        domain = domains.get(domain_id);
    if (domain) {
        domain['countCall'] --;
        if (domain['countCall'] == 0) {
            domains.remove(domain_id);
        };
    };
});

function getCallIdFromEvent(e) {
    return e["variable_w_account_origination_uuid"] || e["Channel-Call-UUID"];
};