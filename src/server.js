'use strict';
// Include the cluster module
var cluster = require('cluster'),
    log = require('./lib/log')(module),
    conf = require('./conf'),
    startPort = conf.get('server:port'),
    count = parseInt(process.env['COUNT']) || 1,
    crashCount = 1;

// Code to run if we're in the master process
if (cluster.isMaster) {

    for (let i = 0; i < count; i++) {
        cluster.fork({
            "server:port": startPort++
        });
    };

    // Listen for dying workers
    cluster.on('exit', function (worker) {

        // Replace the dead worker, we're not sentiment
        log.error('Worker ' + worker.id + ' died.');
        cluster.fork({
            "CRASH_WORKER_COUNT": (crashCount++)
        });
    });

// Code to run if we're in a worker process
} else {
    require('./worker');
    log.info('Worker ' + cluster.worker.id + ' running!');
};