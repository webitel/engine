/**
 * Created by i.navrotskyj on 21.12.2015.
 */
'use strict';

var log = require(__appRoot + '/lib/log')(module),
    //application = require(__appRoot + '/application'),
    checkPermissions = require(__appRoot + '/middleware/checkPermissions'),
    CodeError = require(__appRoot + '/lib/error')
    ;

var Service = {
    _init: function (application, cb) {
        var Acl = require('acl'),
            ACL_CONF = require(__appRoot + '/conf/acl'),
            acl
        ;

        var aclQuery = application.DB._query.acl;

        aclQuery.getRoles((e, r) => {
            if (e)
                return application.stop(e);

            if (!r || r.length == 0) {
                log.info('Insert defaults roles [root, admin, user]');
                aclQuery.insert(ACL_CONF, (e) => {
                    if (e)
                        return application.stop(e);
                    setup(null, ACL_CONF);
                });
            } else {
                // TODO skip root role;
                let _r = r.map( (i) => i.roles === 'root' ? ACL_CONF[0] : i );
                setup(null, _r);
            }
        });

        function setup(e, aclConf) {
            if (e)
                return application.stop(e);

            application.acl = acl = new Acl(new Acl.memoryBackend());
            application.acl.allow(parseAclRoles(aclConf), function (err) {
                if (err)
                    return application.stop(err);
            });

            aclConf.forEach((item) => {
                if (item.hasOwnProperty('parents')) {
                    acl.addRoleParents(item.roles, item.parents);
                }
                log.debug('Register role %s', item.roles);
            });
            log.info('Load roles.');
            if (cb) {
                cb();
            };
        };
    },

    getRoles: function (caller, option, cb) {
        checkPermissions(caller, 'acl/roles', 'r', function (err) {
            if (err)
                return cb(err);

            var acl = application.acl,
                data = acl.backend._buckets,
                dataRoles = data && data.meta.roles,
                resultRoles = [];


            if (dataRoles instanceof Array)
                dataRoles.forEach((i) => {
                    if (i == 'root') return;
                    resultRoles.push(i);
                });

            cb(null, {
                "roles": resultRoles,
                "parents": data.parents
            })
        });
    },
    
    _deleteById: function (id) {
        var aclQuery = application.DB._query.acl;
        aclQuery.removeById(id, (e) => {
            if (e)
                log.error(e);
        });
    },

    addRole: function (caller, option, cb) {
        checkPermissions(caller, 'acl/roles', 'c', function (err) {
            if (err)
                return cb(err);
            if (!option || !(option.allows instanceof Object))
                return cb(new CodeError(400, 'Bad request'));

            if (!option.roles)
                return cb(new CodeError(400, 'Roles is required'));

            let data = {
                "roles": option.roles,
                "allows": option.allows,
                "parents": option.parents
            };

            let parents = option.parents;

            var aclQuery = application.DB._query.acl;
            aclQuery.insert(data, (e, r) => {
                if (e)
                    return cb(e);

                try {
                    var acl = application.acl;
                    acl.allow(parseAclRoles([data]), (e) => {
                        if (e) {
                            log.error(e);
                            Service._deleteById(data['_id']);
                            return cb(e);
                        }
                        ;

                        if (typeof parents === 'string' && parents.length > 0) {
                            acl.addRoleParents(data.roles, parents, (e) => {
                                if (e)
                                    return log.error(e);
                                return null;
                            });
                        }
                        cb(null, 'Created');
                    });
                } catch (e) {
                    log.error(e);
                    Service._deleteById(data['_id']);
                    return cb(e);
                }
            });
        });
    },

    whatResources: function (caller, roleName, cb) {
        checkPermissions(caller, 'acl/roles', 'r', function (err) {
            if (err)
                return cb(err);

            Service._whatResources(roleName, cb);
        });
    },
    
    updateResources: function (caller, option, cb) {
        checkPermissions(caller, 'acl/roles', 'u', function (err) {
            if (err)
                return cb(err);

            if (!option || option.roles == 'root')
                return cb(new CodeError(400, 'Bad request.'));

            var allow = option.allows,
                roles = option.roles
            ;

            if (!roles || !(allow instanceof Object))
                return cb(new CodeError(400, 'Roles or allow is required.'));

            var acl = application.acl,
                aclQuery = application.DB._query.acl,
                query = {};
            try {
                for (let key in allow) {
                    acl.removeAllow(roles, key);
                    acl.allow(roles, key, allow[key]);

                    query['allows.' + key] = allow[key];
                };

                aclQuery.update(
                    {"roles": roles},
                    {
                        "$set": query
                    },
                    (e) => {

                        if (e) {
                            log.error(e);
                            return cb(e);
                        };

                        return Service._whatResources(roles, cb);
                    }
                );
            } catch (e) {
                return cb(e)
            };
        });
    },

    allowedPermissions: function (caller, option, cb) {
        if (!option || !option.roleName || !option.resources)
            return cb(new CodeError(400, "Role name or resource is required"));

        var acl = application.acl;
        acl.allowedPermissions(option.roleName, option.resources, cb);
    },
    
    removeRole: function (caller, roleName, cb) {
        checkPermissions(caller, 'acl/roles', 'd', function (err) {
            if (err)
                return cb(err);

            if (!roleName || roleName == 'root')
                return cb(new CodeError(400, 'Bad request.'));

            var acl = application.acl,
                aclQuery = application.DB._query.acl
            ;
            // TODO parent destroy;
            acl.removeRole(roleName, (e) => {
                if (e) {
                    log.error(e);
                    return cb(e);
                };

                aclQuery.removeByName(roleName, (e) => {
                    if (e) {
                        log.error(e);
                        return cb(e);
                    };

                    return cb(null, 'Destroyed')

                })
            });
        });
    },

    _whatResources: function (roleName, cb) {
        if (!roleName)
            return cb(new CodeError(401, "Role name is required"));
        var acl = application.acl;
        acl.whatResources(roleName, cb)
    }
};

module.exports = Service;


function parseAclRoles (dataRoles) {
    var result = [],
        t;
    for (let item of dataRoles) {
        t = {
            "roles": item.roles,
            "allows": []
        };
        for (let key in item.allows) {
            t.allows.push({
                "resources": key,
                "permissions": item.allows[key] instanceof Array ? item.allows[key] : [item.allows[key]]
            })
        }
        result.push(t);
    };

    return result;
};