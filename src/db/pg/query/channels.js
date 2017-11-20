/**
 * Created by I. Navrotskyj on 20.11.17.
 */

"use strict";

const log = require(__appRoot + '/lib/log')(module),
    CodeError = require(__appRoot + '/lib/error'),
    buildQuery = require('./utils').buildQuery;

module.exports = add;


function add(pool) {
    return {
        listByPresence: (userId, cb) => {
            pool.query(
                `SELECT *
                FROM channels WHERE presence_id = $1`,
                [userId],
                (err, res) => {
                    if (err)
                        return cb(err);

                    if (res && res.rowCount) {
                        return cb(null, res.rows)
                    } else {
                        return cb(null, []);
                    }
                }
            )
        }
    }
}

