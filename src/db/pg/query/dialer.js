/**
 * Created by I. Navrotskyj on 10.11.17.
 */

"use strict";

const log = require(__appRoot + '/lib/log')(module),
    CodeError = require(__appRoot + '/lib/error'),
    buildQuery = require('./utils').buildQuery;

module.exports = add;

const sqlTemplateItem = `
    SELECT * 
    FROM dialer_templates
    WHERE id = $1 AND dialer_id = $2
`;

const sqlTemplateCreate = `
    INSERT INTO dialer_templates(dialer_id, name, type, action, template, description)
    VALUES ($1, $2, $3, $4, $5, $6)
    RETURNING id;
`;

const sqlTemplateUpdate = `
    UPDATE dialer_templates
        SET name = $1
            ,type = $2
            ,action = $3
            ,template = $4
            ,description = $5
    WHERE id = $6 AND dialer_id = $7
    RETURNING *;
`;

const sqlRemoveTemplate = `
    DELETE FROM dialer_templates
    WHERE id = $1 AND dialer_id = $2
    RETURNING *;
`;

const sqlRemoveTemplateAllByDialer = `
    DELETE FROM dialer_templates
    WHERE dialer_id = $1;
`;

function add(pool) {
    return {
        templates: {

            list: (request, cb) => {
                buildQuery(pool, request, "dialer_templates", cb);
            },

            item: (dialerId, id, cb) => {
                pool.query(
                    sqlTemplateItem,
                    [+id, dialerId],
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

            create: (data = {}, cb) => {
                try {
                    pool.query(
                        sqlTemplateCreate,
                        //dialer_id, name, type, action, template, description
                        [
                            data.dialerId,
                            data.name,
                            data.type,
                            data.action,
                            data.template ? JSON.stringify(data.template) : null,
                            data.description
                        ],
                        (err, res) => {
                            if (err)
                                return cb(err);

                            if (res && res.rowCount) {
                                return cb(null, res.rows[0])
                            } else {
                                return cb(new CodeError(404, `Not found result`));
                            }
                        }
                    )
                } catch (e) {
                    return cb(new CodeError(400, e.message))
                }
            },

            update: (dialerId, id, data = {}, cb) => {
                try {
                    pool.query(
                        sqlTemplateUpdate,
                        [
                            data.name,
                            data.type,
                            data.action,
                            data.template ? JSON.stringify(data.template) : null,
                            data.description,
                            +id,
                            dialerId
                        ],
                        (err, res) => {
                            if (err)
                                return cb(err);

                            if (res && res.rowCount) {
                                return cb(null, res.rows[0])
                            } else {
                                return cb(new CodeError(404, `Not found ${id}@${dialerId}`));
                            }
                        }
                    )
                } catch (e) {
                    return cb(new CodeError(400, e.message));
                }
            },

            remove: (dialerId, id, cb) => {
                pool.query(
                    sqlRemoveTemplate,
                    [
                        +id,
                        dialerId
                    ],
                    (err, res) => {
                        if (err)
                            return cb(err);

                        if (res && res.rowCount) {
                            return cb(null, res.rows[0])
                        } else {
                            return cb(new CodeError(404, `Not found ${id}@${dialerId}`));
                        }
                    }
                )
            },

            //TODO
            removeAllByDialer: (dialerId, cb) => {
                pool.query(
                    sqlRemoveTemplateAllByDialer,
                    [dialerId],
                    (err) => cb(err)
                )
            }
        }
    }
}