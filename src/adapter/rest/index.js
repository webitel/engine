'use strict';

var express = require('express');
var morgan = require('morgan');
var api = express();
var log = require(__appRoot + '/lib/log')(module);
var bodyParser = require('body-parser');
var favicon = require('serve-favicon');
var path = require('path');
var conf = require(__appRoot + '/conf');
const getIp = require(__appRoot + '/utils/ip');

api.use(favicon(__appRoot + '/public/static/favicon.ico'));

morgan.token('webitelUser', req => (req.webitelUser && req.webitelUser.id) || 'NOT REGISTER');
morgan.token('colorStatus', (req, res) => {
    const status = res._header
        ? res.statusCode
        : undefined;

    const color = status >= 500 ? 31 // red
        : status >= 400 ? 33 // yellow
        : status >= 300 ? 36 // cyan
        : status >= 200 ? 32 // green
        : 0 // no color

    return `\x1b[${color}m${status}\x1b[0m`
});

morgan.token('realIp', getIp);

api.use(morgan('api: [:webitelUser] > [:colorStatus] :realIp :method :url :response-time ms :res[content-length] ":user-agent"', {
    skip: req => req.method === 'OPTIONS'
}));

require('./cdr/cdr.resource').addRoutes(api);

api.use(bodyParser.json({limit: '2mb'}));

if (conf.get('conference:enable').toString() === 'true') {
    api.use('/', express.static(path.join(__appRoot, '/public/conference')));
    require('./verto/verto.resource').addRoutes(api);
}

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
require('./license/license.resource').addRoutes(api);
require('./hook/hook.resource').addRoutes(api);
require('./calendar/calendar.resource').addRoutes(api);
require('./dialer/dialer.resource').addRoutes(api);
require('./vmail/vmail.resource').addRoutes(api);
require('./widget/widget.resource').addRoutes(api);
require('./callbackQueue/callbackQueue.resource').addRoutes(api);

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