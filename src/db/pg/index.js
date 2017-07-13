/**
 * Created by igor on 06.07.17.
 */


const pg = require('pg'),
    log = require(__appRoot + '/lib/log')(module),
    initQueue = require('./init'),
    async = require('async'),
    conf = require(__appRoot + '/conf');

// create a config to configure both pooling behavior
// and client options
// note: all config is optional and the environment variables
// will be read if the config is not present

const config = conf.get('pg');

//this initializes a connection pool
//it will keep idle connections open for 30 seconds
//and set a limit of maximum 10 idle clients
const pool = new pg.Pool(config);

const query = new Map();
query.set('widget', require('./query/widget')(pool));
query.set('callback', require('./query/callback')(pool));

function initData(err) {
    if (err) {
        log.error(err);
        setTimeout(() => {
            init(pool, initData)
        }, 1000)
    }
}

init(pool, initData);

pool.on('error', function (err, client) {
    // if an error is encountered by a client while it sits idle in the pool
    // the pool itself will emit an error event with both the error and
    // the client which emitted the original error
    // this is a rare occurrence but can happen if there is a network partition
    // between your application and the database, the database restarts, etc.
    // and so you might want to handle it and at least log it out
    log.error('idle client error', err.message, err.stack);
});

module.exports.getQuery = name => query.get(name);

//export the query method for passing queries to the pool
module.exports.query = function (text, values, callback) {
    return pool.query(text, values, callback);
};

// the pool also supports checking out a client for
// multiple operations, such as a transaction
module.exports.connect = function (callback) {
    return pool.connect(callback);
};

function init(pool, cb) {
    async.eachSeries(
        initQueue,
        (sql, cb) => {
            pool.query(sql, cb);
        },
        cb
    )
}