/**
 * Created by i.navrotskyj on 21.12.2015.
 */
'use strict';

var aclServices = require(__appRoot + '/services/acl');

module.exports = {
    addRoutes: addRoutes
};

function addRoutes (api) {
    api.get('/api/v2/acl/roles', listRoles);
    api.get('/api/v2/acl/roles/:name', whatResources);
    api.put('/api/v2/acl/roles/:name', updateResources);
    api.delete('/api/v2/acl/roles/:name', removeRole);
    api.post('/api/v2/acl/roles', addRole);

    //api.get('/api/v2/acl/roles/:name/:resources', allowedPermissions);
}

function listRoles (req, res, next) {
    aclServices.getRoles(req.webitelUser, {}, (err, result) => {
        if (err) return next(err);

        return res
            .status(200)
            .json({
                "status": "OK",
                "info": result
            })
            ;
    })
};

function addRole (req, res, next) {
    aclServices.addRole(req.webitelUser, req.body, (err, result) => {
        if (err) return next(err);

        return res
            .status(200)
            .json({
                "status": "OK",
                "info": result
            })
            ;
    })
};

function whatResources (req, res, next) {
    aclServices.whatResources(req.webitelUser, req.params.name, (err, result) => {
        if (err) return next(err);

        return res
            .status(200)
            .json({
                "status": "OK",
                "info": result
            })
            ;
    })
};

function updateResources (req, res, next) {
    aclServices.updateResources(req.webitelUser, {"roles": req.params.name, "allows": req.body}, (err, result) => {
        if (err) return next(err);

        return res
            .status(200)
            .json({
                "status": "OK",
                "info": result
            })
            ;
    })
};

function allowedPermissions (req, res, next) {
    aclServices.allowedPermissions(req.webitelUser, {roleName: req.params.name, resources: req.params.resources}, (err, result) => {
        if (err) return next(err);

        return res
            .status(200)
            .json({
                "status": "OK",
                "info": result
            })
            ;
    })
};

function removeRole (req, res, next) {
    aclServices.removeRole(req.webitelUser, req.params.name, (err, result) => {
        if (err) return next(err);

        return res
            .status(200)
            .json({
                "status": "OK",
                "info": result
            })
            ;
    })
};