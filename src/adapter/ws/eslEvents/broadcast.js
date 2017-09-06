/**
 * Created by I. Navrotskyj on 05.09.17.
 */

"use strict";

const eventsService = require(__appRoot + '/services/events'),
    EVENT_NAME = 'SE::BROADCAST',
    log = require(__appRoot + '/lib/log')(module),
    mapEvents = new Map(),
    application = require(__appRoot + '/application');

eventsService.registered(EVENT_NAME);

module.exports = app => {

    app.Broker.on('init:broker', () => mapEvents.clear());

    function addLogRoute(id) {
        if (mapEvents.has(id)) {
            mapEvents.set(id, mapEvents.get(id) + 1)
        } else {
            application.Broker.bind(
                application.Broker.systemBroadcastQueueName,
                application.Broker.Exchange.ENGINE,
                `*.broadcast.message.${id}`,
                e => {
                    if (e) {
                        log.error(e)
                    } else {
                        log.trace(`Add handle *.broadcast.message.${id}`)
                    }
                }
            );

            mapEvents.set(id, 1)
        }
    }

    function minusLogRoute(id) {
        if (mapEvents.has(id)) {
            let val = mapEvents.get(id);
            val--;
            if (val === 0) {
                application.Broker.unbind(
                    application.Broker.systemBroadcastQueueName,
                    application.Broker.Exchange.ENGINE,
                    `*.broadcast.message.${id}`,
                    e => {
                        if (e) {
                            log.error(e)
                        } else {
                            log.trace(`Remove handle *.broadcast.message.${id}`)
                        }
                    }
                );

                mapEvents.delete(id);
            } else {
                mapEvents.set(id, val);
            }
        }
    }

    app.on(`unsubscribe::${EVENT_NAME}`, (eventConfig, caller, eventName) => {
        if (!eventConfig) {
            return
        }
        const args = eventConfig.args || {};

        if (args.name === "log") {
            if (args.id) {
                minusLogRoute(args.id);
            }
        }
    });

    app.on(`subscribe::${EVENT_NAME}`, (args, caller, eventName) => {
        if (args) {
            if (args.name === "log") {
                if (args.id) {
                    addLogRoute(args.id)
                }
            }
        }
    });


    app.Broker.on("broadcast.message", json => {
        json['Event-Name'] = 'SE::BROADCAST';
        eventsService.fire('SE::BROADCAST', json['domain'], json);
        eventsService.fire('SE::BROADCAST', 'root', json);
    })

};