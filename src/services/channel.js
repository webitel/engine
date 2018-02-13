/**
 * Created by Igor Navrotskyj on 06.08.2015.
 */

'use strict';

const log = require(__appRoot + '/lib/log')(module),
    validateCallerParameters = require(__appRoot + '/utils/validateCallerParameters'),
    CodeError = require(__appRoot + '/lib/error'),
    checkEslError = require(__appRoot + '/middleware/checkEslError');

const VAR_SEPARATOR = String.fromCharCode(27);

var Service = {
    bgApi: function (execString, cb) {
        log.debug('Exec: %s', execString);

        application.Esl.bgapi(
            execString,
            function (res) {
                var err = checkEslError(res);
                if (err)
                    return cb && cb(err, res);

                return cb && cb(null, res);
            }
        );
    },

    makeCall: function (caller, options, cb) {
        const _extension = options && options['extension'];
        const user = options['user'] || "";
        if (!_extension)
            return cb(new Error('Bad parameters: extension is required'));

        if (typeof user !== "string") {
            return cb(new Error('Bad parameters: user is required'));
        }

        let _originatorParam = ['w_jsclient_originate_number=' + _extension];
        if (options['params'] instanceof Array) {
            _originatorParam = _originatorParam.concat(options['params']);
        }

        if (typeof options['auto_answer_param'] === "string" && options['auto_answer_param']) {
            _originatorParam.push(options['auto_answer_param'])
        }

        if (caller.domain) {
            _originatorParam.push(`domain_name=${caller.domain}`);
        } else if (~user.indexOf('@')) { //TODO from root
            _originatorParam.push(`domain_name=${user.split('@')[1]}`);
        }

        _originatorParam.push(
            'webitel_direction=outbound',
            'ignore_early_media=true'
        );

        const dialString = `originate [${_originatorParam.join(',')}]user/${options['user']} ${_extension} XML default ${_extension} ${_extension}`;

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
        let _extension = options && options['extension'];
        const user = options['user'] || "";
        if (!_extension)
            return cb(new Error('Bad parameters: extension is required'));

        _extension = _extension.split('@')[0];

        if (typeof user !== "string") {
            return cb(new Error('Bad parameters: user is required'));
        }

        let _originatorParam = ['w_jsclient_originate_number=' + _extension, 'w_jsclient_xtransfer=' + options['parent_call_uuid']];

        if (typeof options['auto_answer_param'] === "string" && options['auto_answer_param']) {
            _originatorParam.push(options['auto_answer_param'])
        }

        if (caller.domain) {
            _originatorParam.push(`domain_name=${caller.domain}`);
        } else if (~user.indexOf('@')) { //TODO from root
            _originatorParam.push(`domain_name=${user.split('@')[1]}`);
        }

        _originatorParam.push(
            'webitel_direction=outbound',
            'ignore_early_media=true'
        );

        const dialString = `originate {${_originatorParam.join(',')}}user/${options['user']} ${_extension} XML default ${_extension} ${_extension}`;

        Service.bgApi(dialString, cb);
    },

    attXferBridge: function (caller, options, cb) {
        Service.bgApi('uuid_setvar ' + options['channel-uuid-leg-d'] + ' w_transfer_result confirmed');
        Service.bgApi('uuid_setvar ' + options['channel-uuid-leg-a'] + ' w_transfer_result confirmed');
        Service.bgApi(
            'uuid_bridge ' + options['channel-uuid-leg-b'] + ' ' + options['channel-uuid-leg-c'],
            (err, res) => {
                if (err) {
                    Service.bgApi('uuid_setvar ' + options['channel-uuid-leg-d'] + ' w_transfer_result error');
                    Service.bgApi('uuid_setvar ' + options['channel-uuid-leg-a'] + ' w_transfer_result error');
                }
                return cb && cb(err, res);
            }
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
            side = options['side'],
            displayValue = options['display'],
            vars = options['variables'],
            variables = []
        ;
        if (displayValue) {
            variables.push(`origination_callee_id_number=${displayValue}`)
        }

        if (vars instanceof Array && vars.length > 0) {
            variables = variables.concat(vars)
        }

        if (options['channel-uuid'] === 'all' && caller.id !== 'root') {
            return cb(new Error('Permission denied.'));
        }

        if (caller.domain) {
            user = (user + '').split('@')[0] + '@' + caller.domain;
        }

        if (!side) {
            side = user;
        }

        Service.bgApi(
            'originate ' + (variables.length > 0 ? `[^^${VAR_SEPARATOR}${variables.join(VAR_SEPARATOR)}]` : '')  + 'user/' + user + ' &eavesdrop(' + (options['channel-uuid'] || '')
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

    spy: function (caller, options, cb) {
        var user = options['user'] || caller.id,
            spy = options.spy,
            side = options['side'],
            displayValue = options['display'],
            domain = validateCallerParameters(caller, options['domain']),
            variables = ''
            ;
        if (!spy) {
            return cb(new CodeError(400, "Bad spy user"));
        }
        if (!domain)
            return cb(new CodeError(400, "Bad domain name"));


        if (displayValue) {
            variables = `[origination_callee_id_number=${displayValue},origination_caller_id_name=${displayValue}]`
        }

        user = (user + '').split('@')[0] + '@' + domain;
        spy = (spy + '').split('@')[0] + '@' + domain;

        if (!side) {
            side = user;
        }

        Service.bgApi(
            'originate ' + variables + 'user/' + user + ' &userspy(' + spy
            + ') XML default ' + side + ' ' + side,
            cb
        );
    },

    channelsList: function (caller, options, cb) {
        var _domain = validateCallerParameters(caller, options['domain']),
            _item = '';
        if (_domain) {
            _item = ' like ' + _domain;
        }

        application.Esl.show('channels' + _item, 'json', function (err, data) {
            if (err) {
                return cb(err);
            };
            return cb(null, data);
        });
    },

    callList: function (caller, options, cb) {
        let _domain = validateCallerParameters(caller, options['domain']),
            _item = '';
        if (_domain) {
            _item = ' like ' + _domain;
        }

        application.Esl.show('calls' + _item, 'json', function (err, data) {
            if (err) {
                return cb(err);
            };
            return cb(null, data);
        });
    },

    channelsByUser: function (caller, options, cb) {
        const _domain = validateCallerParameters(caller, options['domain']);
        if (!_domain) {
            return cb(new CodeError(400, `Domain is required`))
        }

        if (!options.userId) {
            return cb(new CodeError(400, 'User id is required'))
        }

        application.PG.getQuery('channels').listByPresence(`${options.userId}@${_domain}`, cb);
    },

    /**
     *
     * @param caller
     * @param options
     * @param cb
     * @returns {*}
     */
    hupAll: function (caller, options, cb) {
        var _domain = validateCallerParameters(caller, options['domain']),
            _api = 'hupall ' + (options['cause'] || "MANAGER_REQUEST")
        ;

        // hupall normal_clearing  domain_name  10.10.10.144
        if (_domain) {
            _api += ' domain_name ' + _domain;
        };

        return Service.bgApi(_api, cb);
    },

    _countChannels: function (cb) {
        Service.bgApi('show channels count', cb)
    }
};

module.exports = Service;