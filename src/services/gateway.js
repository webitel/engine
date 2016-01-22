/**
 * Created by Igor Navrotskyj on 07.08.2015.
 */

'use strict';

var parsePlainTableToJSON = require(__appRoot + '/utils/parse').plainTableToJSON,
    plainTableToJSONArray = require(__appRoot + '/utils/parse').plainTableToJSONArray,
    plainCollectionToJSON = require(__appRoot + '/utils/parse').plainCollectionToJSON,
    checkPermissions = require(__appRoot + '/middleware/checkPermissions'),
    validateCallerParameters = require(__appRoot + '/utils/validateCallerParameters'),
    CodeError = require(__appRoot + '/lib/error')
    ;

var Service = {
    bgApi: function (execString, cb) {
        application.Esl.bgapi(
            execString,
            function (res) {
                return cb(null, res);
            }
        );
    },

    killGateway: function (caller, options, cb) {
        checkPermissions(caller, 'gateway', 'u', function (err) {
            if (err)
                return cb(err);

            Service.bgApi(
                'sofia profile ' + (options['profile'] || '') + ' killgw ' + (options['gateway'] || ''),
                cb
            );
        });
    },
    
    listSipProfile: function (caller, options, cb) {
        checkPermissions(caller, 'gateway/profile', 'r', function (err) {
            if (err)
                return cb(err);

            Service.bgApi(
                'sofia status',
                function (err, res) {
                    parsePlainTableToJSON(res['body'], null, cb);
                }
            );
        });
    },

    rescanSipProfile: function (caller, options, cb) {
        checkPermissions(caller, 'gateway/profile', 'u', function (err) {
            if (err)
                return cb(err);

            Service.bgApi(
                'sofia profile ' + (options['profile'] || '') + ' rescan',
                function (err, res) {
                    if (res['body'].indexOf('Invalid ') == 0)
                        return cb(new Error(res['body']));
                    return cb(null, res);
                }
            );
        });
    },

    // todo Deprecated
    /**
     *
     * @param caller
     * @param options
     * @param cb
     */
    createGateway: function (caller, options, cb) {
        checkPermissions(caller, 'gateway', 'c', function (err) {
            if (err)
                return cb(err);

            var domain = validateCallerParameters(caller, options['domain']);
            
            options['domain'] = domain;
            application.WConsole.createSipGateway(caller, options, cb)
        });
    },

    /**
     * 
     * @param caller
     * @param domain
     * @param cb
     */
    listGateway: function (caller, domain, cb, type) {
        checkPermissions(caller, 'gateway', 'r', function (err) {
            if (err)
                return cb(err);

            var _domain = validateCallerParameters(caller, domain);

            application.WConsole.showSipGateway(caller, _domain, function (err, res) {
                if (err)
                    return cb(err);
                // todo DEL
                if (type == 'plain') {
                    return cb(null, res);
                }
                return plainTableToJSONArray(res, cb);
            });
        });
    },

    /**
     * 
     * @param caller
     * @param gatewayName
     * @param type
     * @param option
     * @param cb
     */
    // todo del typeResponce
    changeGateway: function (caller, gatewayName, type, option, cb, typeResponce) {
        checkPermissions(caller, 'gateway', 'u', function (err) {
            if (err)
                return cb(err);

            if (!gatewayName || !type || !option) {
                return cb(new CodeError(400, 'Bad request.'));
            };

            application.WConsole.changeSipGateway(caller, gatewayName, type, option, function (err, res) {
                if (err)
                    return cb(err);
                if (typeResponce == 'plain')
                    return cb(null, res);

                return plainCollectionToJSON(res, cb);
            });
        });
    },

    /**
     *
     * @param caller
     * @param gatewayName
     * @param cb
     */
    itemGateway: function (caller, gatewayName, cb) {
        checkPermissions(caller, 'gateway', 'r', function (err) {
            if (err)
                return cb(err);

            if (!gatewayName) {
                return cb(new CodeError(400, 'Gateway name is required.'));
            };

            return Service.changeGateway(caller, gatewayName, 'params', {}, cb);
        });
    },

    /**
     *
     * @param caller
     * @param gatewayName
     * @param cb
     */
    deleteGateway: function (caller, gatewayName, cb) {
        checkPermissions(caller, 'gateway', 'd', function (err) {
            if (err)
                return cb(err);

            if (!gatewayName) {
                return cb(new CodeError(400, 'Gateway name is required.'));
            };

            application.WConsole.removeSipGateway(caller, gatewayName, cb);
        });
    },

    /**
     *
     * @param caller
     * @param gatewayName
     * @param profile
     * @param cb
     */
    upGateway: function (caller, gatewayName, profile, cb) {
        checkPermissions(caller, 'gateway', 'u', function (err) {
            if (err)
                return cb(err);

            if (!gatewayName) {
                return cb(new CodeError(400, 'Gateway name is required.'));
            };

            application.WConsole.upSipGateway(caller, gatewayName, profile, cb);
        });
    },

    /**
     *
     * @param caller
     * @param gatewayName
     * @param cb
     */
    downGateway: function (caller, gatewayName, cb) {
        checkPermissions(caller, 'gateway', 'u', function (err) {
            if (err)
                return cb(err);

            if (!gatewayName) {
                return cb(new CodeError(400, 'Gateway name is required.'));
            };

            application.WConsole.downSipGateway(caller, gatewayName, cb);
        });
    }
};

module.exports = Service;