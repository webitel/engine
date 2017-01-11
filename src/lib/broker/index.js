/**
 * Created by i.navrotskyj on 20.03.2016.
 */


'use strict';

var log = require(__appRoot + '/lib/log')(module),
    util = require('util'),
    WebitelAmqp = require('./amqp'),
    WebitelEsl = require('./esl'),
    url = require('url')
    ;


class Broker {
    constructor(conf, app) {
        if (!conf)
            throw "Bad config broker";

        let configBroker;

        if (/^amqp:\/\//.test(conf.connectionString)) {
            configBroker = {
                "uri": conf.connectionString,
                "eventsExchange": conf.config.eventsExchange,
                "exchange": conf.config.exchange
            };
            return new WebitelAmqp(configBroker, app);

        } else if (/^esl:\/\//.test(conf.connectionString)) {
            let parseUri = url.parse(conf.connectionString);
            configBroker = {
                "host": parseUri.hostname,
                "port": +parseUri.port
            };

            if (parseUri.auth)
                configBroker.pwd = parseUri.auth.split(':')[1];
            
            return new WebitelEsl(configBroker, app);
        } else {
            app.stop(new Error("Broker config require."));
        };

    };

};

module.exports = Broker;