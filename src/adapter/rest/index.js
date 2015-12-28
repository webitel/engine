'use strict';

var express = require('express');
var api = express();
var log = require(__appRoot + '/lib/log')(module);
var bodyParser = require('body-parser');
var favicon = require('serve-favicon');
var path = require('path');
var conf = require(__appRoot + '/conf');

api.use(favicon(__appRoot + '/public/static/favicon.ico'));
api.use(bodyParser.json());

require('./logger').addRoutes(api);
if (conf.get('conference:enable').toString() == 'true') {
    api.use('/', express.static(path.join(__appRoot, '/public/conference')));
    require('./verto/verto.resource').addRoutes(api);
};

api.use('/', express.static(path.join(__appRoot, '/public/static')));

require('./cors').addRoutes(api);

require('./auth/auth.resource').addRoutes(api);
require('./acl/acl.resource').addRoutes(api);

require('./domain/domain.resource').addRoutes(api);
require('./account/account.resource').addRoutes(api);
require('./stats/stats.resource').addRoutes(api);
require('./cdr/cdr.resource').addRoutes(api);
require('./dialplan/dialplan.resource').addRoutes(api);
require('./dialplan/blackList.resource').addRoutes(api);
require('./email/email.resource').addRoutes(api);
require('./channels/channels.resource').addRoutes(api);
require('./callCentre/callCentre.resource').addRoutes(api);
require('./outbound/outbound.resource').addRoutes(api);
require('./contactBook/contactBook.resource').addRoutes(api);
require('./location/number.resource').addRoutes(api);
require('./gateway/gateway.resource').addRoutes(api);
require('./configure/configure.resource').addRoutes(api);

// Error handle
require('./error').addRoutes(api);

var route, methods;

api._router.stack.forEach(function(middleware){
    if(middleware.route){ // routes registered directly on the app
        route = middleware.route;
        methods = Object.keys(route.methods);
        log.info('Add: [%s]: %s', (methods.length > 1) ? 'ALL' : methods[0].toUpperCase(), route.path);

    } else if(middleware.name === 'router'){ // router middleware
        middleware.handle.stack.forEach(function(handler){
            route = handler.route;
            methods = Object.keys(route.methods);
            log.info('Add: [%s]: %s', (methods.length > 1) ? 'ALL' : methods[0].toUpperCase(), route.path);
        });
    }
});

module.exports = api;