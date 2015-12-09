/**
 * Created by i.navrotskyj on 09.12.2015.
 */
'use strict';

var log = require(__appRoot + '/lib/log')(module),
    checkPermissions = require(__appRoot + '/middleware/checkPermissions'),
    application = require(__appRoot + '/application')
    ;

var Service = {
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
    }
};

module.exports = Service;