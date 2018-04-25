/**
 * Created by Igor Navrotskyj on 06.08.2015.
 */

'use strict';

var log = require(__appRoot + '/lib/log')(module);

var auth = require('./auth');
var calls = require('./calls');
var accounts = require('./accounts');
var domains = require('./domains');
var events = require('./events');
var users = require('./users');
var cdr = require('./cdr');
var blackList = require('./blackList');
var callCentre = require('./callCentre');
var gateway = require('./gateway');
var license = require('./license');
var hotdesk = require('./hotdesk');

module.exports = function (application) {
    var controller  = {};
    registerController(auth, controller, application);
    registerController(domains, controller, application);
    registerController(accounts, controller, application);
    registerController(calls, controller, application);
    registerController(events, controller, application);
    registerController(users, controller, application);
    registerController(cdr, controller, application);
    registerController(blackList, controller, application);
    registerController(callCentre, controller, application);
    registerController(gateway, controller, application);
    registerController(license, controller, application);
    registerController(hotdesk, controller, application);

    return controller;
};

function registerController(commands, controller, application) {
    for (var key in commands) {
        if (commands.hasOwnProperty(key)) {
            if (controller[key]) {
                var _e = new Error('Command ' + key + ' already exists');
                log.error(_e);
                application.stop(_e);
            };
            log.info('Registered command: %s', key);
            controller[key] = commands[key];
        };
    };
};