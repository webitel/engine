/**
 * Created by I. Navrotskyj on 07.11.17.
 */

"use strict";

const log = require(__appRoot + '/lib/log')(module),
    CodeError = require(__appRoot + '/lib/error'),
    buildQuery = require('./utils').buildQuery;

module.exports = add;

const TABLE_HISTORY = 'hook_queue';

function add(pool) {
    return {
        list: (request, cb) => {
            buildQuery(pool, request, TABLE_HISTORY, cb);
        }
    }
}