/**
 * Created by Igor Navrotskyj on 07.08.2015.
 */

'use strict';

var parsePlainTableToJSON = require(__appRoot + '/utils/parse').plainTableToJSON,
    plainTableToJSONArray = require(__appRoot + '/utils/parse').plainTableToJSONArray,
    plainCollectionToJSON = require(__appRoot + '/utils/parse').plainCollectionToJSON,
    checkPermissions = require(__appRoot + '/middleware/checkPermissions'),
    validateCallerParameters = require(__appRoot + '/utils/validateCallerParameters'),
    log = require(__appRoot + '/lib/log')(module),
    CodeError = require(__appRoot + '/lib/error'),
    channelService = require('./channel')
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

            const domain = validateCallerParameters(caller, options['domain']);
            
            options['domain'] = domain;
            application.WConsole.createSipGateway(caller, options, (err, res) => {
                if (err)
                    return cb(err);

                Service._updateOrInsertItemByName(options.name, options.domain);
                return cb(null, res);
            })
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

            Service._listGateway(_domain, type, cb);
        });
    },
    
    _listGateway: function (domain, type, cb) {
        application.WConsole.showSipGateway(null, domain, function (err, res) {
            if (err)
                return cb(err);
            // todo DEL
            if (type == 'plain') {
                return cb(null, res);
            }
            return plainTableToJSONArray(res, cb);
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

                plainCollectionToJSON(res, (err, data) => {
                    if (err)
                        return cb(err);

                    Service._updateOrInsert(gatewayName, null, data);
                    return cb(null, data);
                });
            });
        });
    },

    /**
     *
     * @param caller
     * @param option
     * @param cb
     */
    varGateway: function (caller, option, cb) {
        checkPermissions(caller, 'gateway', 'r', function (err) {
            if (err)
                return cb(err);

            if (!option || !option.name) {
                return cb(new CodeError(400, 'Gateway name is required.'));
            };

            return application.WConsole.gatewayVars(option.name, option.direction, function (err, res) {
                if (err)
                    return cb(err);

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
            }

            application.WConsole.changeSipGateway(caller, gatewayName, 'params', {}, function (err, res) {
                if (err)
                    return cb(err);

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
    deleteGateway: function (caller, gatewayName, cb) {
        checkPermissions(caller, 'gateway', 'd', function (err) {
            if (err)
                return cb(err);

            if (!gatewayName) {
                return cb(new CodeError(400, 'Gateway name is required.'));
            };

            application.WConsole.removeSipGateway(caller, gatewayName, (err, res) => {
                if (err)
                    return cb(err);


                const db = application.DB._query.gateway;
                db.removeByName(gatewayName, e => {
                    if (e)
                        return log.error(e);
                    log.trace(`remove gateway ${gatewayName} - success`);
                });
                return cb(null, res);
            });
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
    },

    _initGatewayLineCounter: function (app) {

        const checkCleanActiveChannels = () => {
            channelService._countChannels((err, res) => {
                if (err)
                    return log.error(err);

                if (parseInt(res.body) === 0) {
                    log.debug(`clean active lines`);
                    application.DB._query.gateway.cleanActiveChannels();
                }
            });
        };

        app.once('sys::wConsoleConnect', Service._reloadGatewayToDb);

        app.on('sys::errorConnectFsApi', () => {
            application.DB._query.gateway.cleanActiveChannels();
        });

        app.on('sys::connectFsApi', checkCleanActiveChannels);

        app.Schedule(5 * 60 * 1000, function () {
            log.debug('Schedule active channels.');
            checkCleanActiveChannels()
        });

    },
    
    _reloadGatewayToDb: function () {
        Service._listGateway('', null, (err, res) => {
            if (err)
                return log.error(err);

            log.trace(`Start upgrade gateway list`);
            const gws = [];
            if (res instanceof Array) {
                res.forEach( i => {
                    const name = i.Gateway.split('::').pop();
                    const domain = i.Domain;
                    gws.push(name);
                    Service._updateOrInsertItemByName(name, domain);
                })
            }

            if (gws.length > 0) {
                const db = application.DB._query.gateway;
                db.removeNotNames(gws, (err) => {
                    if (err)
                        log.error(err);
                    log.trace(`removed old gateways success`);
                })
            }
        });
    },

    _updateOrInsertItemByName: (name, domain) => {
        application.WConsole.changeSipGateway(null, name, "params", {}, function (err, res) {
            if (err)
                return log.error(err);

            plainCollectionToJSON(res, (err, gw) => {
                if (err)
                    return log.error(err);

                Service._updateOrInsert(name, domain, gw)
            });
        });
    },

    _updateOrInsert: (name, domain, params = {}) => {
        const q = {
            name: name,
            params
        };

        if (domain)
            q.domain = domain;

        const db = application.DB._query.gateway;

        db.insertOrUpdate(name, {$set: q, $max: {stats: {callsIn: 0, callsOut: 0, active: 0}}}, (err, res) => {
            if (err)
                return log.error(err);

            if (res.result.upserted) {
                log.trace(`insert gateway ${name} - success`);
            } else if (res.result.nModified > 0) {
                log.trace(`updated gateway ${name} - success`);
            } else {
                log.trace(`no changes gateway ${name}`);
            }
        })
    },

    _onChannel: (e) => {
        if (!e['variable_sofia_profile_name'] || e['variable_sofia_profile_name'] === 'internal')
            return;

        // console.log(e);

        const sipGatewayName = e['variable_sip_gateway'] || e['variable_sip_gateway_name'];

        if (sipGatewayName) {
            Service._changeGatewayLineByName(sipGatewayName, e['Call-Direction'], e['Event-Name'] === 'CHANNEL_CREATE', e['Channel-Call-UUID']);
        } else {
            Service._changeGatewayLineByRealm(e['variable_sip_from_host'] + ':' + e['variable_sip_from_port'],
                e['Call-Direction'], e['Event-Name'] === 'CHANNEL_CREATE', e['Channel-Call-UUID']);
        }
    },

    _changeGatewayLineByName: (gatewayName, direction, isNew, uuid) => {
        if (isNew) {
            setupChannelGatewayName(uuid, {name: gatewayName});
            application.DB._query.gateway.incrementLineByName(gatewayName, 1, direction, (e, res) => {
                if (e)
                    return log.error(e);
                log.trace(`Add line to gateway ${gatewayName}`);
            })
        } else {
            application.DB._query.gateway.incrementLineByName(gatewayName, -1, direction, (e, res) => {
                if (e)
                    return log.error(e);
                log.trace(`Minus line to gateway ${gatewayName}`);
            })
        }
    },

    _changeGatewayLineByRealm: (realm, direction, isNew, uuid) => {
        if (isNew) {
            application.DB._query.gateway.incrementLineByRealm(realm, 1, direction, (e, res) => {
                if (e)
                    return log.error(e);

                setupChannelGatewayName(uuid, res && res.value);
                log.trace(`Add line to gateway ${realm}`);
            })
        } else {
            application.DB._query.gateway.incrementLineByRealm(realm, -1, direction, (e, res) => {
                if (e)
                    return log.error(e);

                log.trace(`Minus line to gateway ${realm}`);
            })
        }
    }
};

module.exports = Service;

function setupChannelGatewayName(uuid, gw) {
    if (!gw)
        return;

    channelService.bgApi(`uuid_setvar ${uuid} webitel_gateway ${gw.name}`);
}