'use strict';
// Include the cluster module
var cluster = require('cluster'),
    log = require('./lib/log')(module),
    conf = require('./conf'),
    startPort = conf.get('server:port'),
    count = parseInt(process.env['COUNT']) || 1
    ;

// Code to run if we're in the master process
if (cluster.isMaster === true) {

    var clusterMap = {},
        crashCount = 0;

    var forkWorker = function (port, crashCount, id) {
        var worker = cluster.fork({
            "server:port": port,
            "CRASH_WORKER_COUNT": crashCount,
            "WORKER_ID": id
        });
        worker.processId = id;

        worker.on('message', handleMessage);
        clusterMap[worker.id] = {
            port: port,
            worker: worker
        }
    };

    var handleMessage = function (msg) {
        for (let key in clusterMap) {
            try {
                if (this.id != clusterMap[key].worker.id)
                    clusterMap[key].worker.send(msg);
            } catch (e) {
                log.error(e);
            }
        };
    };

    for (let i = 0; i < count; i++)
        forkWorker(startPort, crashCount, i);

    // Listen for dying workers
    cluster.on('exit', function (worker) {

        // Replace the dead worker, we're not sentiment
        log.error('Worker ' + worker.id + ' died.');
        worker.removeListener('message', handleMessage);
        let port = clusterMap[worker.id].port;
        console.log('Close port: ' + port);
        delete clusterMap[worker.id];
        forkWorker(port, ++crashCount, worker.processId);
    });

// Code to run if we're in a worker process
} else {
    require('./worker');
    log.info('Worker ' + cluster.worker.id + ' running!');
}