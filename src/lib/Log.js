var winston = require('winston');
var conf = require('../conf');
require('winston-logstash');

function getLogger(module) {
    var pathDirectory = module.filename.split('//').slice(-2).join('//').split('\\');
    var path = pathDirectory.slice(pathDirectory.length - 3).join('\\') + '(' + process.pid + ')';

    var logLevels = {
        levels: {
            trace: 0,
            debug: 1,
            warn: 2,
            error: 3,
            info: 4
        },
        colors: {
            trace: 'yellow',
            debug: 'yellow',
            info: 'green',
            warn: 'yellow',
            error: 'red'
        }
    };
    winston.addColors(logLevels.colors);
    var logger = new (winston.Logger)({
        levels: logLevels.levels,
        transports: [
            new winston.transports.Console({
                colorize: true,
                level: conf.get('application:loglevel'),
                label: path,
                'timestamp': true
            })
        ]
    });
    if (conf.get('application:logstash:enabled').toString() == 'true') {
        logger.add(winston.transports.Logstash, {
            port: conf.get('application:logstash:port'),
            node_name: conf.get('application:logstash:node_name'),
            host: conf.get('application:logstash:host'),
            level: conf.get('application:loglevel')
        });
    };
    //(\"secret\"\:\"[^\"]*\")|(password=[^,|"]*)|(\bauth\b[^.]*)
    logger.addFilter(function (msg, meta) {
        return maskSecrets(msg, meta);
    });
    return logger;
};

function maskSecrets(msg, meta) {
    if (/secret|password|\bauth\b/) {
        msg = msg.replace(/(\"secret\"\:\"[^\"]*\")|(password=[^,|"]*)|(\sauth\s[^.]*)|("password","value":"[^"]*)/g, '*****');
    };

    return {
        msg: msg,
        meta: meta
    };
}

module.exports = getLogger;