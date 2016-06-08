'use strict';

var WebSocketServer = require('ws').Server,
    conf = require('../../conf'),
    handleMessage = require('./handleMessage'),
    handleEsl = require('./handleEslEvent'),
    handleConsole = require('./handleWConsoleEvent'),
    wsOriginAllow = conf.get('server:socket:originHost').toString() != 'false';

module.exports = createWSS;

function createWSS(express, application) {
    var option = {
        server: express
    };
    if (wsOriginAllow) {
        option['origin'] = conf.get('server:socket:originHost').toString()
    };

    var wss = new WebSocketServer(option);

    handleMessage(wss, application);
    handleEsl(application);
    handleConsole(application);
};

// @private