/**
 * Created by i.navrotskyj on 25.01.2016.
 */
'use strict';

var log = require(__appRoot + '/lib/log')(module),
    checkPermissions = require(__appRoot + '/middleware/checkPermissions'),
    plainTableToJSONArray = require(__appRoot + '/utils/parse').plainTableToJSONArray,
    validateEslResponse = require(__appRoot + '/utils/parse').validateEslResponse,
    domainService = require('./domain'),
    CodeError = require(__appRoot + '/lib/error')
    ;

var Service = {
    /**
     *
     * @param caller
     * @param cb
     */
    list: function (caller, cb) {
        checkPermissions(caller, 'license', 'r', (e) => {
            if (e)
                return cb(e);

            if (caller.domain) {
                domainService.item(caller, {name: caller.domain}, (err, res) => {
                    if (err)
                        return cb(err);

                    _exec('status', `${res.variable_customer_id}`, (e, txt) => {
                        if (e)
                            return cb(e);

                        return plainTableToJSONArray(txt, cb);
                    });
                })
            } else {
                _exec('status', '', (e, txt) => {
                    if (e)
                        return cb(e);

                    return plainTableToJSONArray(txt, cb);
                });
            }
        });
    },

    /**
     *
     * @param caller
     * @param option
     * @param cb
     * @returns {*}
     */
    item: function (caller, option, cb) {
        let operation = caller && caller.domain ? 'ro' : 'r',
            cid = option && option.cid
        ;
        if (!cid)
            return cb(new CodeError(400, 'Bad request, cid is required.'));

        checkPermissions(caller, 'license', operation, (e) => {
            if (e)
                return cb(e);

            _exec('status', cid, (e, txt) => {
                if (e)
                    return cb(e);

                return plainTableToJSONArray(txt, cb);
            });
        });
    },

    /**
     *
     * @param caller
     * @param option
     * @param cb
     * @returns {*}
     */
    upload: function (caller, option, cb) {
        let operation = caller && caller.domain ? 'uo' : 'u',
            token = option && option.token
            ;

        if (!token)
            return cb(new CodeError(400, 'Bad request, token is required.'));

        checkPermissions(caller, 'license', operation, (e) => {
            if (e)
                return cb(e);

            return _exec('upload', token, cb);
        });
    },

    /**
     *
     * @param caller
     * @param option
     * @param cb
     * @returns {*}
     */
    remove: function (caller, option, cb) {
        let operation = caller && caller.domain ? 'do' : 'd',
            cid = option && option.cid
            ;
        if (!cid)
            return cb(new CodeError(400, 'Bad request, cid is required.'));

        checkPermissions(caller, 'license', operation, (e) => {
            if (e)
                return cb(e);

            return _exec('delete', cid, cb);
        });
    }
};

module.exports = Service;

function _exec (operation, arg, cb) {
    let _operation = operation || '',
        _arg = arg || ''
        ;
    if (!application && !application.WConsole)
        return cb(new CodeError(500, 'Connect to WConsole established.'))
            ;
    application.WConsole.api(`license ${_operation} ${_arg}`, (res) => {
        validateEslResponse(res && res.body, (e, txt) => {
            if (e)
                return cb(e);

            cb(null, txt)
        });
    });
};