'use strict';

var path = require('path'),
    logo = require('./middleware/logo')
    ;

global.__appRoot = path.resolve(__dirname);
logo();

var APPLICATION = require('./application'),
    application = global.application = APPLICATION;

process.on('SIGINT', function() {
    console.log('SIGINT received ...');
    if (application) {
        application.stop();
    };
});

module.exports = application;