/**
 * Created by igor on 17.08.16.
 */

"use strict";

const eventsService = require(__appRoot + '/services/events'),
    statsService = require(__appRoot + '/services/stats'),
    log = require(__appRoot + '/lib/log')(module);

let event = eventsService.registered('SE:HEARTBEAT'),
    _subscribe = false;
    
module.exports = (application) => {

    const RK = `*.HEARTBEAT.*.*.*`,
        EXCHANGE = application.Broker.Exchange.FS_EVENT;

    let queueName;

    if (application.Broker.isConnect()) {
        subscribe();
    }
    application.Broker.on('init:broker', subscribe);

    event.domains.on('added', (_, key) => {
        if (key === 'root') {
            bindHeartbeat();
        }
    });

    event.domains.on('removed', (_, key) => {
        if (key === 'root') {
            unBindHeartbeat();
        }
    });

    function subscribe() {
        application.Broker.subscribe('', undefined, handleHeartbeat, (err, qName) => {
            if (err)
                return log.error(err);

            queueName = qName;
        });
    }
    
    function bindHeartbeat() {
        if (_subscribe || !queueName) return;
        application.Broker.bind(queueName, EXCHANGE, RK, (err) => {
            if (err)
                return log.error(err);
            log.trace(`bindHeartbeat - ok`);
            _subscribe = true;
        });
    }
    
    function unBindHeartbeat() {
        application.Broker.unbind(queueName, EXCHANGE, RK, (err) => {
            if (err)
                log.error(err);
            log.trace(`un bind heartbeat - ok`);
            _subscribe = false;
        });
    }

    function handleHeartbeat(msg) {
        try {
            let json = JSON.parse(msg.content.toString()),
                workerMemory = statsService.memoryUsage()
                ;

            json['Event-Name'] = 'SE:HEARTBEAT';
            json['engine_socket_count'] = application._getWSocketSessions();
            json['engine_online_count'] = application.Users.length();
            json['engine_online_max'] = statsService.maxUserSession();
            json['engine_domain_online'] = application.Domains.length();

            json['engine_uptime_sec'] = process.uptime();
            
            json['engine_free_mem'] = statsService.freeMemory();
            json['engine_mem_rss'] = workerMemory.rss;
            json['engine_mem_heap_used'] = workerMemory.heapUsed;
            json['engine_mem_heap_total'] = workerMemory.heapTotal;
            
            eventsService.fire('SE:HEARTBEAT', 'root', json);
        } catch (e) {
            log.error(e);
        }
    }
};