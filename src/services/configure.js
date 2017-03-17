/**
 * Created by i.navrotskyj on 09.12.2015.
 */
'use strict';

let log = require(__appRoot + '/lib/log')(module),
    checkPermissions = require(__appRoot + '/middleware/checkPermissions'),
    CodeError = require(__appRoot + '/lib/error'),
    checkEslError = require(__appRoot + '/middleware/checkEslError'),
    application = require(__appRoot + '/application')
    ;

const ALLOW_RELOAD_MODULE = ["amqp", "callcenter"];
const ALLOW_RELOAD_CACHE = ["clear", "remove"];

let Service = {
    reloadXml: function (caller, cb) {
        checkPermissions(caller, 'system/reload', 'u', function (err) {
            if (err)
                return cb(err);

            application.WConsole.reloadXml(caller, (err, res) => {
                if (err)
                    return cb(err);

                return cb(null, res);
            });
        });
    },
    
    reloadMod: function (caller, moduleName, cb) {
        checkPermissions(caller, 'system/reload', 'u', function (err) {
            if (err)
                return cb(err);

            if (!~ALLOW_RELOAD_MODULE.indexOf(moduleName))
                return cb(new CodeError(400, `Not allow reload module: ${moduleName}`));


            let execString = `reload mod_${moduleName}`;
            log.warn('Exec: %s', execString);

            application.Esl.bgapi(
                execString,
                function (res) {
                    var err = checkEslError(res);
                    if (err)
                        return cb && cb(err, res.body);

                    return cb && cb(null, res.body);
                }
            );
        });
    },

    cache: function (caller, options, cb) {
        checkPermissions(caller, 'system/reload', 'u', function (err) {
            if (err)
                return cb(err);



            const execString = `http_clear_cache`;
            log.warn('Exec: %s', execString);

            application.Esl.bgapi(
                execString,
                function (res) {
                    var err = checkEslError(res);
                    if (err)
                        return cb && cb(err, res.body);

                    return cb && cb(null, res.body);
                }
            );
        });
    }
};

module.exports = Service;