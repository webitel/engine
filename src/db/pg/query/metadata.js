/**
 * Created by I. Navrotskyj on 30.10.17.
 */

"use strict";

const log = require(__appRoot + '/lib/log')(module),
    CodeError = require(__appRoot + '/lib/error');

const sqlItem = `
    SELECT *
    FROM metadata
    WHERE domain = $1 AND object_name = $2;
`;

const sqlCreateOrReplaceItem = `
INSERT INTO metadata (domain, object_name, data)
VALUES($1, $2, $3)
ON CONFLICT (domain, object_name) DO UPDATE SET data = EXCLUDED.data
RETURNING *;
`;


const sqlDelete = `
DELETE FROM metadata
WHERE  domain = $1 AND object_name = $2
RETURNING *;
`;

function add(pool) {
    return {
        item: (domain, object_name, cb) => {
            pool.query(
                sqlItem,
                [domain, object_name],
                (err, res) => {
                    if (err)
                        return cb(err);

                    if (res && res.rowCount) {
                        return cb(null, res.rows[0])
                    } else {
                        return cb(new CodeError(404, `Not found ${object_name}`));
                    }
                }
            )
        },

        createOrReplace: (domain, object_name, data, cb) => {
            try {
                pool.query(
                    sqlCreateOrReplaceItem,
                    [domain, object_name, JSON.stringify(data)],
                    (err, res) => {
                        if (err)
                            return cb(err);

                        if (res && res.rowCount) {
                            return cb(null, res.rows[0])
                        } else {
                            return cb(new CodeError(404, `Not found ${object_name}`));
                        }
                    }
                )
            } catch (e) {
                return cb(new CodeError(400, e.message))
            }
        },

        remove: (domain, object_name, cb) => {
            pool.query(
                sqlDelete,
                [domain, object_name],
                (err, res) => {
                    if (err)
                        return cb(err);

                    if (res && res.rowCount) {
                        return cb(null, res.rows[0])
                    } else {
                        return cb(new CodeError(404, `Not found ${object_name}`));
                    }
                }
            )
        }
    }
}

module.exports = add;