/**
 * Created by I. Navrotskyj on 04.10.17.
 */

"use strict";

const log = require(__appRoot + '/lib/log')(module),
    CodeError = require(__appRoot + '/lib/error'),
    buildQuery = require('./utils').buildQuery;


const create = `
    INSERT INTO tcp_dump (duration, filter, description) 
    VALUES ($1, $2, $3)
    RETURNING id;             
`;
const get = `
    SELECT * FROM tcp_dump WHERE id = $1;             
`;

const remove = `
    DELETE FROM tcp_dump WHERE id = $1
    RETURNING *;             
`;

const update = `
    UPDATE tcp_dump 
        SET description = $1
    WHERE id = $2
    RETURNING *;
`;

function add(pool) {
    return {
        list: (request, cb) => {
            buildQuery(pool, request, "tcp_dump", cb);
        },

        create: (data = {}, cb) => {
            pool.query(
                create,
                [data.duration, data.filter || "", data.description || ""],
                (err, res) => {
                    if (err)
                        return cb(err);

                    if (res && res.rowCount) {
                        return cb(null, res.rows[0].id)
                    } else {
                        log.error('bad response', res);
                        return cb(new Error('Bad db response'));
                    }
                }
            )
        },

        get: (id, cb) => {
            pool.query(
                get,
                [+id],
                (err, res) => {
                    if (err)
                        return cb(err);

                    if (res && res.rowCount) {
                        return cb(null, res.rows[0])
                    } else {
                        return cb(new CodeError(404, `Not found ${id}`));
                    }
                }
            )
        },

        remove: (id, cb) => {
            pool.query(
                remove,
                [+id],
                (err, res) => {
                    if (err)
                        return cb(err);

                    if (res && res.rowCount) {
                        return cb(null, res.rows[0])
                    } else {
                        return cb(new CodeError(404, `Not found ${id}`));
                    }
                }
            )
        },

        update: (id, data = {}, cb) => {
            pool.query(
                update,
                [data.description || "", +id],
                (err, res) => {
                    if (err)
                        return cb(err);

                    if (res && res.rowCount) {
                        return cb(null, res.rows[0])
                    } else {
                        return cb(new CodeError(404, `Not found ${id}`));
                    }
                }
            )
        }
    }
}


module.exports = add;