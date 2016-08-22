/**
 * Created by Admin on 03.08.2015.
 */

'use strict';

var os = require('os'),
    conf = require(__appRoot + '/conf');

module.exports = {
    addRoutes: addRoutes
};

/**
 * Adds routes to the api.
 */
function addRoutes(api) {
    api.get('/api/v2/status', applicationStatus);
    api.get('/api/v2/status/config', applicationStatusConfig);
    api.get('/api/v1/status', applicationStatus);
}

function applicationStatusConfig (req, res) {
    // TODO add config
    res.status(200).end();
}

function applicationStatus(req, res) {
    if (application.Esl && !application.Esl['connecting']) {
        application.Esl.api('status', function(response) {
            res.json(getResult(response));
        });
    } else {
        res.json(getResult(false));
    }
}

function getResult (freeSwitchStatus) {
    return {
        "version": process.env['VERSION'] || '',
        "nodeMemory": getMemoryUsage(),
        "processId": process.pid,
        "socketSessions": application._getWSocketSessions(),
        "userSessions": application.Users.length(),
        "maxUserSessions": application.Users._maxSession,
        "domainSessions": application.Domains.length(),
        "processUpTimeSec": process.uptime(),
        "wConsole": getWConsoleInfo(),
        "system": getOsInfo(),
        "crashCount": process.env['CRASH_WORKER_COUNT'] || 0,
        "freeSWITCH": (freeSwitchStatus) ? freeSwitchStatus['body'] : 'Connect server error.',
        "nodeVersion": process.version
    }
}

function getMemoryUsage () {
    var memory = process.memoryUsage();
    return {
        "rss": memory['rss'],
        "heapTotal": memory['heapTotal'],
        "heapUsed": memory['heapUsed']
    }
}

function getOsInfo () {
    return {
        "totalMemory": os.totalmem(),
        "freeMemory": os.freemem(),
        "platform": os.platform(),
        "name": os.type(),
        "architecture": os.arch()
    };
}

function getCpuInfo () {
    var res = {};
    var cpus = os.cpus();
    for(var i = 0, len = cpus.length; i < len; i++) {
        res['CPU' + i] = {};
        var cpu = cpus[i], total = 0;
        for(var type in cpu.times)
            total += cpu.times[type];

        for(type in cpu.times)
            res['CPU' + i][type] = Math.round(100 * cpu.times[type] / total)
    }
    return res;
}

function getWConsoleInfo () {
    var wConsole = application.WConsole;
    if (!wConsole)
        return {status: "Internal Error"};

    return {
        "status": wConsole._status == 1 ? "Connected": "Offline",
        "apiQueue": wConsole.apiCallbackQueue.length,
        "cmdQueue": wConsole.cmdCallbackQueue.length,
        // move package
        "version": wConsole.version || '',
        "sid": wConsole._serverId
    }
}