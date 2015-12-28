var MongoClient = require("mongodb").MongoClient,
    mongoClient = new MongoClient(),
    config = require('../../conf'),
    log = require('../../lib/log')(module);

module.exports = initConnect;

function initConnect (server) {
    mongoClient.connect(config.get('mongodb:uri') ,function(err, db) {
        if (err) {
            log.error('Connect db error: %s', err.message);
            return server.emit('sys::connectDbError', err);
        };

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
            acl: require('./query/acl').addQuery(db)
        };

        server.emit('sys::connectDb', db);
        log.info('Connected db %s ', config.get('mongodb:uri'));
        db.on('close', function () {
            log.warn('close MongoDB');
        });

        db.on('error', function (err) {
            log.error('close MongoDB: ', err);
        });
    });
};