/**
 * Created by igor on 05.07.17.
 */

"use strict";

const log = require(__appRoot + '/lib/log')(module),
    fsExtra = require('fs-extra'),
    fs = require('fs'),
    conf = require(__appRoot + '/conf'),
    publicWebRtc = conf.get('widget:publicWebRtc'),
    publicPostApi = conf.get('widget:publicPostApi'),
    WIDGET_PATH = conf.get('widget:basePath'),
    WIDGET_URI = conf.get('widget:baseUri'),
    validateCallerParameters = require(__appRoot + '/utils/validateCallerParameters'),
    checkPermissions = require(__appRoot + '/middleware/checkPermissions'),
    CodeError = require(__appRoot + '/lib/error');


const Service = {
    list: (caller, option = {}, cb) => {
        checkPermissions(caller, 'widget', 'r', function (err) {
            if (err)
                return cb(err);

            if (!option)
                return cb(new CodeError(400, "Bad request options"));

            let domain = validateCallerParameters(caller, option['domain']);
            if (!domain) {
                return cb(new CodeError(400, 'Bad request: domain is required.'));
            }
            option.domain = domain;
            application.PG.getQuery('widget').list(option, cb);
        });
    },

    get: (caller, option = {}, cb) => {
        checkPermissions(caller, 'widget', 'r', function (err) {
            if (err)
                return cb(err);

            if (!option)
                return cb(new CodeError(400, "Bad request options"));

            let domain = validateCallerParameters(caller, option['domain']);
            if (!domain) {
                return cb(new CodeError(400, 'Bad request: domain is required.'));
            }

            if (!option.id)
                return cb(new CodeError(400, 'Bad request: id is required.'));

            application.PG.getQuery('widget').findById(option.id, domain, {}, (err, res) => {
                if (err)
                    return cb(err);

                if (res) {
                    res._widgetBaseUri = WIDGET_URI;
                }
                return cb(err, res)
            });
        });
    },

    create: (caller, option = {}, cb) => {
        const domain = option.domain = validateCallerParameters(caller, option.domain);

        if (!domain)
            return cb(new CodeError(400, 'Domain is required.'));

        if (!(option.config instanceof Object))
            return cb(new CodeError(400, 'Bad config.'));

        checkPermissions(caller, 'widget', 'c', (e) => {
            if (e)
                return cb(e);

            application.PG.getQuery('widget').create(option, (err, id) => {
                if (err)
                    return cb(err);

                generateWidgetFile(id, domain, option.config, (err, path) => {
                    if (err)
                        return cb(err);

                    application.PG.getQuery('widget')._setFilePath(id, path, e => {
                        if (e)
                            return log.error(e)
                    });
                    return cb(null, [id])
                });
            });
        })
    },

    update: (caller, option = {}, cb) => {
        checkPermissions(caller, 'widget', 'u', function (err) {
            if (err)
                return cb(err);

            if (!option)
                return cb(new CodeError(400, "Bad request options"));


            if (!option.id)
                return cb(new CodeError(400, 'Bad request: id is required.'));

            if (!option.data)
                return cb(new CodeError(400, 'Bad request: data is required.'));

            let domain = validateCallerParameters(caller, option['domain']);
            if (!domain) {
                return cb(new CodeError(400, 'Bad request: domain is required.'));
            }

            generateWidgetFile(option.id, domain, option.data.config, (err, path) => {
                if (err)
                    return cb(err);

                option.data._filePath = path;
                application.PG.getQuery('widget').update(option.id, domain, option.data, cb);
            });

        });
    },

    remove: function (caller, option, cb) {
        checkPermissions(caller, 'widget', 'd', function (err) {
            if (err)
                return cb(err);

            if (!option)
                return cb(new CodeError(400, "Bad request options"));

            let domain = validateCallerParameters(caller, option['domain']);
            if (!domain) {
                return cb(new CodeError(400, 'Bad request: domain is required.'));
            }

            if (!option.id)
                return cb(new CodeError(400, 'Bad request: id is required.'));

            return application.PG.getQuery('widget').delete(option.id, domain, (err, id, filePath) => {
                if (err)
                    return cb(err);

                if (filePath) {
                    removeWidgetFile(filePath)
                }

                return cb(null, id)
            });
        });
    },
};

module.exports = Service;

function generateWidgetFile(id, domain, config = {}, cb) {
    fsExtra.ensureDir(`${WIDGET_PATH}/domains/${domain}`, err => {
        if (err)
            return cb(err);

        config.publicPostApi = publicPostApi;
        config.publicWebRtc = publicWebRtc;
        const text = getWidgetConfig(config);
        if (!text) {
            return cb(new CodeError(400, `Bad config.`))
        }
        const path = `${WIDGET_PATH}/domains/${domain}/${id}.js`;
        log.trace(`try save widget file: ${path}`);
        fs.writeFile(path, text, err => {
            if (err)
                return cb(err);

            return cb(null, path)
        })
    })
}

function removeWidgetFile(path) {
    fs.lstat(path, (err, stat) => {
        if (err)
            return log.error(err);

        if (!stat.isFile()) {
            return log.error(`Bad file ${path}`)
        }

        fs.unlink(path, err => {
            if (err)
                return log.error(err)
        })
    })
}

function getWidgetConfig(obj) {
    try {
        return `
(function(w) {
    if (typeof w.WebitelCallbackInit === 'function') {
    w.WebitelCallbackInit(${JSON.stringify(obj)})
}
})(window);
        `
    } catch (e) {
        return null
    }
}