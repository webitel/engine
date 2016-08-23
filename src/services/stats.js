/**
 * Created by igor on 23.08.16.
 */

"use strict";

const os = require('os')
    ;
    
const Service = module.exports = {

    allStats: (cb) => {
        return Service.switchStatus((err, freeSwitchStatus) => {
            return cb(
                null,
                {
                    "version": process.env['VERSION'] || '',
                    "nodeMemory": Service.memoryUsage(),
                    "processId": process.pid,
                    "socketSessions": application._getWSocketSessions(),
                    "userSessions": application.Users.length(),
                    "maxUserSessions": Service.maxUserSession(),
                    "domainSessions": application.Domains.length(),
                    "processUpTimeSec": process.uptime(),
                    "wConsole": Service.consoleInfo(),
                    "system": Service.osInfo(),
                    "crashCount": process.env['CRASH_WORKER_COUNT'] || 0,
                    "freeSWITCH": (err) ? err.message : freeSwitchStatus,
                    "nodeVersion": process.version
                }
            )
        })
    },

    switchStatus: (cb) => {
        if (application.Esl && !application.Esl['connecting']) {
            application.Esl.api('status', function(response) {
                cb(null, response.body);
            });
        } else {
            cb(new Error('No ESL connect.'))
        }
    },

    maxUserSession: () => {
        return application.Users._maxSession
    },

    osInfo: () => {
        return {
            "totalMemory": os.totalmem(),
            "freeMemory": Service.freeMemory(),
            "platform": os.platform(),
            "name": os.type(),
            "architecture": os.arch()
        };
    },

    freeMemory: () => {
        return os.freemem();
    },

    memoryUsage: () => {
        let memory = process.memoryUsage();
        return {
            "rss": memory.rss,
            "heapTotal": memory.heapTotal,
            "heapUsed": memory.heapUsed
        }
    },

    consoleInfo: () => {
        let wConsole = application.WConsole;
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
    },

    cpuInfo: () => {
        let res = {},
            cpus = os.cpus()
            ;
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
};