var MongoClient = require("mongodb").MongoClient,
    mongoClient = new MongoClient(),
    config = require('../../conf'),
    log = require('../../lib/log')(module);

module.exports = initConnect;

function initConnect (server) {
    var options = {
        server: {
            auto_reconnect: true,
            slaveOk: true,
            replset: {
                socketOptions: {
                    keepAlive: 1,
                    connectTimeoutMS : 30000 ,
                    socketTimeoutMS: 90000
                }
            }
            // socketOptions: { keepAlive: 1, connectTimeoutMS: 30000 },
            // replset: { socketOptions: { keepAlive: 1, connectTimeoutMS: 30000 } }
        }
    };
    mongoClient.connect(config.get('mongodb:uri'), options, function(err, db) {
        if (err) {
            log.error('Connect db error: %s', err.message);
            return server.emit('sys::connectDbError', err);
        }
        log.info('Connected db %s ', config.get('mongodb:uri'));
        require('./query/initCollections')(db);

        db._query = {
            email: require('./query/email').addQuery(db),
            auth: require('./query/auth').addQuery(db),
            dialplan: require('./query/dialplan').addQuery(db),
            blacklist: require('./query/blacklist').addQuery(db),
            book: require('./query/contactBook').addQuery(db),
            cdr: require('./query/cdr').addQuery(db),
            oq: require('./query/outboundQueue').addQuery(db),
            userStatus: require('./query/userStatus').addQuery(db),
            location: require('./query/location').addQuery(db),
            conference: require('./query/conference').addQuery(db),
            acl: require('./query/acl').addQuery(db),
            hook: require('./query/hook').addQuery(db),
            calendar: require('./query/calendar').addQuery(db),
            dialer: require('./query/dialer').addQuery(db),
            telegram: require('./query/telegram').addQuery(db),
            domain: require('./query/domain').addQuery(db)
        };

        server.emit('sys::connectDb', db);

        db.on('close', function () {
            log.warn('close MongoDB');
            server.emit('sys::closeDb', db);
        });
        db.on('reconnect', function () {
            log.info('Reconnect MongoDB');
            server.emit('sys::reconnectDb', db);
        });

        db.on('error', function (err) {
            log.error('err MongoDB: ', err);
        });
    });
}