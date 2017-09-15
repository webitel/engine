/**
 * Created by igor on 06.07.17.
 */

"use strict";

const log = require(__appRoot + '/lib/log')(module),
    CodeError = require(__appRoot + '/lib/error'),
    buildQuery = require('./utils').buildQuery;

const create = `
    INSERT INTO widget (name, description, config, domain, queue_id, limit_by_number, limit_by_ip, _file_path, blacklist, language, callflow_id)
    VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
    RETURNING id;             
`;

const sqlItem = `
    SELECT * FROM widget WHERE id = $1 AND domain = $2
`;

const sqlDelete = `
    DELETE FROM widget WHERE id = $1 AND domain = $2
    RETURNING id, _file_path;
`;

function add(pool) {
    return {
        list: (request, cb) => {
            buildQuery(pool, request, "widget", cb);
        },

        findById: (_id, domainName, options, cb) => {
            pool.query(
                sqlItem,
                [
                    +_id,
                    domainName
                ], (err, res) => {
                    if (err) {
                        return cb(err);
                    }
                    if (res && res.rowCount) {
                        return cb(null, res.rows[0])
                    } else {
                        return cb(new CodeError(404, `Not found ${_id}@${domainName}`));
                    }
                }
            )
        },

        _setFilePath: (id, path, cb) => {
            pool.query(
                `UPDATE widget SET _file_path = $1 WHERE id = $2`,
                [
                    path,
                    +id
                ], (err) => {
                    if (err)
                        return cb(err);

                    return cb(null)
                }
            )
        },

        create: (doc, cb) => {
            pool.query(
                create,
                [
                    doc.name, //$1
                    doc.description, //$2
                    JSON.stringify(doc.config), //$3
                    doc.domain, //$4
                    doc.queue_id, //$5
                    doc.limit_by_number, //$6
                    doc.limit_by_ip, //$7
                    doc._filepath || '', //$8
                    doc.blacklist && doc.blacklist.length > 0 ? `{${doc.blacklist.join(',')}` : null, //$9
                    doc.language, //10
                    doc.callflow_id, //11
                ], (err, res) => {
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

        delete: (id, domain, cb) => {
            pool.query(
                sqlDelete,
                [
                   id,
                   domain
                ], (err, res) => {
                    if (err)
                        return cb(err);

                    if (res && res.rowCount) {
                        return cb(null, res.rows[0].id, res.rows[0]._file_path)
                    } else {
                        return cb(new CodeError(404, `Not found ${id}@${domain}`));
                    }
                }
            )
        },

        update: (_id, domainName, doc = {}, cb) => {
            const values = [];
            const params = [];

            if (doc.config instanceof Object) {
                doc.config = JSON.stringify(doc.config)
            }

            for (let field of allowUpdateFields) {
                if (doc.hasOwnProperty(field)) {
                    values.push(`${field} = $` + params.push(doc[field]));
                }
            }

            let update = `UPDATE widget SET ${values.join(',')} WHERE id = $${params.length + 1} AND domain = $${params.length + 2} RETURNING *`;
            params.push(+_id);
            params.push(domainName);
            pool.query(
                update,
                params, (err, res) => {
                    if (err)
                        return cb(err);

                    if (res && res.rowCount) {
                        return cb(null, res.rows[0])
                    } else {
                        return cb(new CodeError(404, `Not found ${_id}@${domainName}`));
                    }
                }
            );
        }
    }
}

module.exports = add;

const allowUpdateFields = ['name', 'description', 'config', 'limit_by_number', 'limit_by_ip', 'blacklist', 'language', 'callflow_id', 'queue_id'];