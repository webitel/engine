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
    INSERT INTO dialer_templates(dialer_id, name, type, action, template, description, before_delete, cron, next_process_id)
    VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
    RETURNING id;
`;

const sqlTemplateUpdate = `
    UPDATE dialer_templates
        SET name = $1
            ,type = $2
            ,action = $3
            ,template = $4
            ,description = $5
            ,before_delete = $6
            ,cron = $7
            ,next_process_id = $8
    WHERE id = $9 AND dialer_id = $10
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
                        //dialer_id, name, type, action, template, description, before_delete, cron, next_process_id
                        [
                            data.dialerId,
                            data.name,
                            data.type,
                            data.action,
                            data.template ? JSON.stringify(data.template) : null,
                            data.description,
                            data.before_delete ? 1 : 0,
                            data.cron,
                            data.next_process_id ? +data.next_process_id : null,
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
                            data.before_delete ? 1 : 0,
                            data.cron,
                            data.next_process_id ? +data.next_process_id : null,
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

            rollback: (dialerId, id, data = {}, cb) => {
                pool.query(
                    `
                    UPDATE dialer_templates
                    SET process_start = NULL
                        ,process_id = NULL
                        ,process_state = COALESCE($1, process_state)
                        ,last_response_text = COALESCE($2, last_response_text)
                    WHERE id = $3 AND NOT process_start is NULL AND dialer_id = $4
                    RETURNING *;
                    `,
                    [
                        data.process_state,
                        data.last_response_text,
                        +id,
                        dialerId
                    ],
                    (err, res) => {
                        if (err)
                            return cb(err);

                        if (res) {
                            return cb(null, res.rows[0])
                        } else {
                            return cb(new CodeError(404, `Not found ${id}@${dialerId}`));
                        }
                    }
                )
            },

            endExecute: (options = {}, buf, cb) => {
                pool.query(
                    `
                    UPDATE dialer_templates
                    SET process_start = NULL
                        ,process_state = 'END'
                        ,process_id = NULL
                        ,success_data = case when $1 then $2 else success_data end
                        ,last_response_text = $3
                    WHERE id = $4 AND NOT process_start is NULL AND dialer_id = $5 
                        AND process_id = $6
                    RETURNING *;
                    `,
                    [
                        options.success,
                        buf,
                        options.message,
                        +options.id,
                        options.dialerId,
                        options.pid
                    ],
                    (err, res) => {
                        if (err)
                            return cb(err);

                        if (res) {
                            return cb(null, res.rows[0])
                        } else {
                            return cb(new CodeError(404, `Not found ${id}@${dialerId}`));
                        }
                    }
                )
            },

            setExecute: (dialerId, id, cb) => {
                pool.query(
                    `
                    UPDATE dialer_templates
                    SET process_start = extract(EPOCH from now())::INT
                        ,process_state = 'CHECK_RESPONSE'
                        ,process_id = substring(md5(clock_timestamp()::text), 0, 10)
                    WHERE id = $1 AND process_start is NULL AND dialer_id = $2 
                        AND NOT EXISTS (SELECT 1 FROM dialer_templates WHERE dialer_id = $2 AND NOT process_start is NULL)
                    RETURNING *;
                    `,
                    [
                        +id,
                        dialerId
                    ],
                    (err, res) => {
                        if (err)
                            return cb(err);

                        if (res) {
                            return cb(null, res.rows[0])
                        } else {
                            return cb(new CodeError(404, `Not found ${id}@${dialerId}`));
                        }
                    }
                )
            },

            getActiveTemplates: (dialerId, cb) => {
                pool.query(
                    `SELECT count(*) as count FROM dialer_templates where dialer_id = $1 AND NOT process_start is NULL`,
                    [dialerId],
                    (err, res) => {
                        if (err)
                            return cb(err);

                        if (res) {
                            return cb(null, res.rows[0].count)
                        } else {
                            return cb(new CodeError(404, `Not found ${dialerId}`));
                        }
                    }
                )
            },

            getNoEmptyCron: (cb) => {
                pool.query(`
                    SELECT id, cron, dialer_id FROM dialer_templates
                    where not cron is null AND cron != '';
                    `,
                    [],
                    (err, res) => {
                        if (err)
                            return cb(err);

                        if (res) {
                            return cb(null, res.rows)
                        }
                        return cb(null, []);
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