/**
 * Created by Igor Navrotskyj on 06.08.2015.
 */

'use strict';

var log = require(__appRoot + '/lib/log')(module),
    validateCallerParameters = require(__appRoot + '/utils/validateCallerParameters'),
    checkEslError = require(__appRoot + '/middleware/checkEslError');

var Service = {
    bgApi: function (execString, cb) {
        log.debug('Exec: %s', execString);

        application.Esl.bgapi(
            execString,
            function (res) {
                var err = checkEslError(res);
                if (err)
                    return cb(err, res);

                return cb(null, res);
            }
        );
    },

    makeCall: function (caller, options, cb) {
        var _originatorParam = new Array('w_jsclient_originate_number=' + options['extension']);
        if (options['params'] instanceof Array) {
            _originatorParam = _originatorParam.concat(options['params']);
        };

        var _autoAnswerParam = [].concat( options['auto_answer_param'] || []),
            _param = '[' + _originatorParam.concat(_autoAnswerParam).join(',') + ']',
            dialString = ('originate ' + _param + 'user/' + options['user'] + ' ' + options['extension']
                + ' xml default ' + options['user'] + ' ' + options['user'])
            ;

        Service.bgApi(dialString, cb);
    },

    hangup: function (caller, options, cb) {
        if (options['variable'] && options['value']) {
            Service.setVar(caller, options, function () {

            });
        };
        Service.bgApi(
            'uuid_kill ' + options['channel-uuid'] + ' ' + (options['cause'] || ''),
            cb
        );
    },

    attendedTransfer: function (caller, options, cb) {
        var _originatorParam = new Array('w_jsclient_originate_number=' + options['destination'],
            'w_jsclient_xtransfer=' + options['call_uuid']);
        if (options['is_webrtc']) {
            _originatorParam.push('sip_h_Call-Info=answer-after=1');
        };
        var _autoAnswerParam = [].concat( options['auto_answer_param'] || []),
            _param = '[' + _originatorParam.concat(_autoAnswerParam).join(',') + ']';

        var dialString = ('originate ' + _param + 'user/' + options['user'] + ' ' + options['destination'] +
            ' xml default ' + options['user'] + ' ' + options['user']);

        Service.bgApi(dialString, cb);
    },

    transfer: function (caller, option, cb) {
        Service.bgApi(
            'uuid_transfer ' + option['channel-uuid'] + ' ' + option['destination'],
            cb
        );
    },
    
    bridge: function (caller, options, cb) {
        Service.bgApi(
            'uuid_bridge ' + options['channel_uuid_A'] + ' ' + options['channel_uuid_B'],
            cb
        );
    },
    
    videoRefresh: function (caller, options, cb) {
        Service.bgApi(
            'uuid_video_refresh ' + options['uuid'],
            cb
        );
    },

    toggleHold: function (caller, options, cb) {
        Service.bgApi(
            'uuid_hold toggle ' + options['channel-uuid'],
            cb
        );
    },
    
    hold: function (caller, options, cb) {
        Service.bgApi(
            'uuid_hold ' + options['channel-uuid'],
            cb
        );
    },

    unHold: function (caller, options, cb) {
        Service.bgApi(
            'uuid_hold off ' + options['channel-uuid'],
            cb
        );
    },
    
    dtmf: function (caller, options, cb) {
        var _digits = options['digits'];
        Service.bgApi(
            'uuid_recv_dtmf ' + options['channel-uuid'] + ' ' + _digits,
            function (err, res) {
                try {
                    if (res && (res['body'].indexOf('-ERR no reply') == 0 || res['body'] == '')) {
                        res['body'] = '+OK ' + _digits;
                    };
                    cb(null, res);
                } catch (e) {
                    return cb(e);
                };
            }
        );
    },
    // TODO delete
    attXfer: function (caller, options, cb) {
        var _account = options['user'].split('@')[0];

        Service.bgApi(
            ('uuid_broadcast ' + options['channel-uuid'] + ' att_xfer::{origination_cancel_key=#,origination_caller_id_name='
                + _account + ',origination_caller_id_number=' + _account
                + ',webitel_att_xfer=true}user/' + options['destination'] + ''),
            cb
        );
    },
    // TODO delete
    attXfer2: function (caller, options, cb) {
        var _originatorParam = new Array('w_jsclient_originate_number=' + options['extension'], 'w_jsclient_xtransfer=' + options['parent_call_uuid']),
            _autoAnswerParam = [].concat( options['auto_answer_param'] || []),
            _param = '{' + _originatorParam.concat(_autoAnswerParam).join(',') + '}',
            dialString = ('originate ' + _param + 'user/' + options['user'] + ' ' + options['extension'].split('@')[0]
                + ' xml default ' + options['user'].split('@')[0] + ' ' + options['user'].split('@')[0]);

        Service.bgApi(dialString, cb);
    },

    attXferBridge: function (caller, options, cb) {
        Service.bgApi(
            'uuid_bridge ' + options['channel-uuid-leg-c'] + ' ' + options['channel-uuid-leg-b'],
            cb
        );
    },

    attXferCancel: function (caller, options, cb) {
        Service.bgApi(
            'uuid_kill ' + options['channel-uuid-leg-c'],
            cb
        );
    },
    
    dump: function (caller, options, cb) {
        Service.bgApi(
            'uuid_dump ' + options['uuid'] + ' json',
            cb
        );
    },

    getVar: function (caller, options, cb) {
        Service.bgApi(
            'uuid_getvar ' + options['channel-uuid'] + ' ' + options['variable'] + ' ' + (options['inleg'] || ''),
            cb
        );
    },

    setVar: function (caller, options, cb) {
        Service.bgApi(
            'uuid_setvar ' + options['channel-uuid'] + ' ' + options['variable'] + ' ' + options['value']
            + ' ' + (options['inleg'] || ''),
            cb
        );
    },

    eavesdrop: function (caller, options, cb) {
        var user = options['user'] || caller.id,
            side = options['side'];
        if (options['channel-uuid'] == 'all' && caller.id != 'root') {
            return cb(new Error('Permission denied.'));
        };
        if (caller.domain) {
            user = (user + '').split('@')[0] + '@' + caller.domain;
        };

        if (!side) {
            side = user;
        };

        Service.bgApi(
            'originate user/' + user + ' &eavesdrop(' + (options['channel-uuid'] || '')
                + ') XML default ' + side + ' ' + side,
            cb
        );
    },

    displace: function (caller, options, cb) {
        var _play = options['record'] == 'start'
            ? 'start'
            : 'stop';
        _play += ' silence_stream://0 3';

        Service.bgApi(
            'uuid_displace ' + options['channel-uuid'] + ' ' + _play,
            cb
        );
    },

    channelsList: function (caller, options, cb) {
        var _domain = validateCallerParameters(caller['domain'], options['domain']),
            _item = '';
        if (_domain) {
            _item = ' like ' + _domain;
        };

        application.Esl.show('channels' + _item, 'json', function (err, data) {
            if (err) {
                return cb(err);
            };
            return cb(null, data);
        });
    }
};

module.exports = Service;