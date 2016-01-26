/**
 * Created by Igor Navrotskyj on 15.09.2015.
 */

'use strict';
var generateUuid = require('node-uuid'),
    channelService = require('./channel'),
    log = require(__appRoot + '/lib/log')(module),
    STATUS = require(__appRoot + '/const/outboundQueue').STATUS,
    CodeError = require(__appRoot + '/lib/error'),
    checkPermissions = require(__appRoot + '/middleware/checkPermissions'),
    validateCallerParameters = require(__appRoot + '/utils/validateCallerParameters');

var _init = false;

var Service = {
    getCallee: function (caller, option, cb) {
        let dbOutQ = application.DB._query.oq;
        let dbOption = {
            "userId": caller.id,
            "limit": 5
        };

        return dbOutQ.getCallee(
            caller.domain,
            dbOption,
            function (err, res) {
                if (err)
                    return cb(err);

                if (!res || !res.value)
                    return cb(new Error('Not found callee.'));

                var channelUuid = generateUuid.v4(),
                    call = res.value;

                call['channelUuid'] = channelUuid;
                application.OutboundQuery.add(channelUuid, call);
            }
        );
    },

    getAvgBillSecUser: function (caller, cb) {
        let dbOutQ = application.DB._query.oq;
        dbOutQ.getAvgBillSecUser(caller.id, cb);
    },
    
    create: function (caller, data, cb) {
        checkPermissions(caller, 'outbound/list', 'c', function (err) {
            if (err)
                return cb(err);

            data = data || {};
            let domain = validateCallerParameters(caller, data['domain']);
            if (!domain) {
                return cb(new CodeError(400, 'Bad request: domain is required.'));
            };
            data['domain'] = domain;
            data['status'] = STATUS.IDLE;
            data['createdOn'] = new Date().getTime();

            let dbOutQ = application.DB._query.oq;
            dbOutQ.insert(data, cb);
        });
    },

    _init: function (application) {
        if (_init) return;

        application.OutboundQuery.on('added', function (call) {
            var caller = application.Users.get(call['userId']);
            if (caller) {
                channelService.makeCall(
                    caller,
                    {
                        "extension": call['number'],
                        "user": caller.id,
                        "params": [
                            'origination_uuid=' + call['channelUuid'],
                            'webitel_outbound_call=' + (call['name'] || '_unknown_'),
                            'webitel_outbound_recordId=' + call['_id'].toString()
                        ]
                    },
                    function (err, res) {
                        // TODO set status ACTIVE
                        //console.dir(arguments);
                    }
                );
            }
        });
        // TODO add reconnect
        application.Esl.on('esl::event::CHANNEL_HANGUP_COMPLETE::*', function (event) {
            let uuid = event.getHeader('variable_uuid');
            let call = application.OutboundQuery.get(uuid);

            if (call) {
                call['hangupCause'] = event.getHeader('variable_hangup_cause');
                call['status'] = call['hangupCause'] == 'NORMAL_CLEARING' ? STATUS.END : STATUS.IDLE;
                call['ModifiedOn'] = new Date().getTime();
                call['channelUuid'] = uuid;
                call['countOriginate'] = call['countOriginate'] > 0 ? call['countOriginate'] +  1 : 1;
                let _id = call['_id'];
                delete call['_id'];
                let dbOutQ = application.DB._query.oq;
                dbOutQ.updateItem(_id, call, function (err) {
                    if (err) {
                        log.error(err);
                    };
                });
            };
        });

        _init = true;
    }
};

module.exports = Service;