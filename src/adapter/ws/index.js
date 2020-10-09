'use strict';

var WebSocketServer = require('ws').Server,
    conf = require('../../conf'),
    handleMessage = require('./handleMessage'),
    handleEsl = require('./handleEslEvent'),
    handleConsole = require('./handleWConsoleEvent'),
    handleBroadcast = require('./eslEvents/broadcast'),
    wsOriginAllow = conf.get('server:socket:originHost').toString() !== 'false';

module.exports = createWSS;

function createWSS(express, application) {
    let option = {
        server: express
    };
    if (wsOriginAllow) {
        option['origin'] = conf.get('server:socket:originHost').toString()
    }

    let wss = new WebSocketServer(option);

    handleMessage(wss, application);
    handleEsl(application);
    handleConsole(application);
    handleBroadcast(application);
}

// @private
