/**
 * Created by I. Navrotskyj on 08.09.17.
 */

"use strict";

const log = require(__appRoot + '/lib/log')(module),
    CodeError = require(__appRoot + '/lib/error'),
    buildQuery = require('./utils').buildQuery;

// TODO move config ???
const DEFAULT_TABLE_NAME = 'callflow_default';
const PUBLIC_TABLE_NAME = 'callflow_public';
const EXTENSION_TABLE_NAME = 'callflow_extension';
const VARIABLES_TABLE_NAME = 'callflow_variables';

const sqlInsertDefault = `
INSERT INTO ${DEFAULT_TABLE_NAME} (destination_number, name, domain, fs_timezone, callflow,
 callflow_on_disconnect, cf_diagram, version, description, disabled, debug)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING id;
`;


const sqlUpdateDefault = `
UPDATE ${DEFAULT_TABLE_NAME} 
    SET destination_number = $1, 
        name = $2, 
        domain = $3, 
        fs_timezone = $4, 
        callflow = $5, 
        callflow_on_disconnect = $6, 
        cf_diagram = $7, 
        description = $8, 
        disabled = $9, 
        debug = $10
WHERE id = $11 AND domain = $12
RETURNING *;
`;

const sqlItemDefault = `
    SELECT * FROM ${DEFAULT_TABLE_NAME} WHERE id = $1 AND domain = $2;
`;

const sqlDeleteDefault = `
    DELETE FROM ${DEFAULT_TABLE_NAME} WHERE id = $1 AND domain = $2
    RETURNING *;
`;

const sqlMoveDownDefault = `
    WITH t1 as (
        SELECT id, "order", domain
        FROM ${DEFAULT_TABLE_NAME}
        WHERE id = $1 and domain = $2
    ), mov as (
        SELECT id, "order"
        FROM ${DEFAULT_TABLE_NAME}
        WHERE domain = $2 AND "order" > (SELECT t1."order"
                         FROM t1
                         LIMIT 1)
        ORDER BY "order" asc
        LIMIT 1
    )
    UPDATE ${DEFAULT_TABLE_NAME}
      set "order" = case WHEN (id = $1) THEN (SELECT "order"
                                         FROM mov LIMIT 1)
               ELSE (SELECT "order"
                     FROM t1 LIMIT 1) end
    WHERE (id = $1 AND exists(SELECT * FROM mov))
          or id = (SELECT id FROM mov LIMIT 1);
`;

const sqlMoveUpDefault = `
    WITH t1 as (
        SELECT id, "order", domain
        FROM ${DEFAULT_TABLE_NAME}
        WHERE id = $1 and domain = $2
    ), mov as (
        SELECT id, "order"
        FROM ${DEFAULT_TABLE_NAME}
        WHERE domain = $2 AND "order" < (SELECT t1."order"
                         FROM t1
                         LIMIT 1)
        ORDER BY "order" desc
        LIMIT 1
    )
    UPDATE ${DEFAULT_TABLE_NAME}
      set "order" = case WHEN (id = $1) THEN (SELECT "order"
                                         FROM mov LIMIT 1)
               ELSE (SELECT "order"
                     FROM t1 LIMIT 1) end
    WHERE (id = $1 AND exists(SELECT * FROM mov))
          or id = (SELECT id FROM mov LIMIT 1);
`;

const sqlDeleteDefaultByDomain = `
    DELETE FROM ${DEFAULT_TABLE_NAME} WHERE domain = $1;
`;


const sqlInsertPublic = `
    INSERT INTO ${PUBLIC_TABLE_NAME} (destination_number, name, domain, fs_timezone, callflow, callflow_on_disconnect, cf_diagram, disabled, debug, description)
    VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
    RETURNING id;
`;


const sqlItemPublic = `
    SELECT * FROM ${PUBLIC_TABLE_NAME} WHERE id = $1 AND domain = $2;
`;


const sqlDeletePublic = `
    DELETE FROM ${PUBLIC_TABLE_NAME} WHERE id = $1 AND domain = $2
    RETURNING *;
`;

const sqlUpdatePublic = `
    UPDATE ${PUBLIC_TABLE_NAME} 
    SET destination_number = $1, 
        name = $2, 
        domain = $11, 
        fs_timezone = $3, 
        callflow = $4, 
        callflow_on_disconnect = $5, 
        cf_diagram = $6, 
        disabled = $7, 
        debug = $8, 
        description = $9
    WHERE id = $10 AND domain = $11    
    RETURNING *;
`;

const sqlDeletePublicByDomain = `
    DELETE FROM ${PUBLIC_TABLE_NAME} WHERE domain = $1;
`;

const sqlItemExtension = `
    SELECT * FROM ${EXTENSION_TABLE_NAME}
    WHERE id = $1 AND domain = $2;
`;

const sqlExistsExtension = `
    SELECT count(id) FROM ${EXTENSION_TABLE_NAME}
    WHERE domain = $1 AND destination_number in ($2) ;
`;

const sqlInsertExtension = `
    INSERT INTO ${EXTENSION_TABLE_NAME} (destination_number, domain, user_id, name, callflow, callflow_on_disconnect, cf_diagram)
    VALUES ($1, $2, $3, $4, $5, $6, $7)
    RETURNING *;
`;

const sqlUpdateExtension = `
    UPDATE ${EXTENSION_TABLE_NAME}
    SET destination_number = $1,
      name = $2,
      fs_timezone = $3,
      callflow = $4,
      callflow_on_disconnect = $5,
      cf_diagram = $6
    WHERE id = $7 and domain = $8
    RETURNING *;
`;

const sqlDeleteExtensionByUserId = `
    DELETE FROM ${EXTENSION_TABLE_NAME} WHERE domain = $1 AND user_id = $2;
`;

const sqlDeleteExtensionById = `
    DELETE FROM ${EXTENSION_TABLE_NAME} WHERE domain = $1 AND id = $2
    RETURNING id;
`;

const sqlUpsertDomainVariables = `
    with upsert as (
	  update ${VARIABLES_TABLE_NAME}
	  set variables = $1
	  where domain = $2
	  returning *
	)
	INSERT INTO ${VARIABLES_TABLE_NAME} (domain, variables)
	select $2, $1 
	WHERE NOT EXISTS (SELECT * FROM upsert);
`;

const sqlDeleteDomainVariables = `
    DELETE FROM ${VARIABLES_TABLE_NAME} WHERE domain = $1;
`;

function add(pool) {
    return {

        //region default query
        createDefault: (dialPlan = {}, cb) => {
            try {
                pool.query(
                    sqlInsertDefault,
                    [
                        dialPlan.destination_number,
                        dialPlan.name,
                        dialPlan.domain,
                        dialPlan.fs_timezone,
                        dialPlan.callflow ? JSON.stringify(dialPlan.callflow) : null,
                        dialPlan.callflow_on_disconnect ? JSON.stringify(dialPlan.callflow_on_disconnect) : null,
                        dialPlan.cf_diagram ? JSON.stringify(dialPlan.cf_diagram) : null,
                        dialPlan.version,
                        dialPlan.description,
                        dialPlan.disabled,
                        dialPlan.debug
                    ],
                    (err, res) => {
                        if (err) {
                            return cb(err);
                        }
                        if (res && res.rowCount) {
                            return cb(null, res.rows[0])
                        } else {
                            return cb(new CodeError(500, `Bad response`));
                        }
                    }
                )
            } catch (e) {
                return cb(new CodeError(400, e.message))
            }
        },

        listDefault: (request, cb) => {
            buildQuery(pool, request, DEFAULT_TABLE_NAME, cb);
        },

        itemDefault: (id, domain, cb) => {
            pool.query(
                sqlItemDefault,
                [
                    +id,
                    domain
                ],
                (err, res) => {
                    if (err) {
                        return cb(err);
                    }
                    if (res && res.rowCount) {
                        return cb(null, res.rows[0])
                    } else {
                        return cb(new CodeError(404, `Not found ${id}`));
                    }
                }
            );
        },

        updateDefault: (id, domain, dialPlan = {}, cb) => {
            try {
                pool.query(
                    sqlUpdateDefault,
                    [
                        dialPlan.destination_number,
                        dialPlan.name,
                        dialPlan.domain,
                        dialPlan.fs_timezone,
                        dialPlan.callflow ? JSON.stringify(dialPlan.callflow) : null,
                        dialPlan.callflow_on_disconnect ? JSON.stringify(dialPlan.callflow_on_disconnect) : null,
                        dialPlan.cf_diagram ? JSON.stringify(dialPlan.cf_diagram) : null,
                        dialPlan.description,
                        dialPlan.disabled,
                        dialPlan.debug,
                        +id,
                        domain
                    ],
                    (err, res) => {
                        if (err) {
                            return cb(err);
                        }
                        if (res && res.rowCount) {
                            return cb(null, res.rows[0])
                        } else {
                            return cb(new CodeError(500, `Bad response`));
                        }
                    }
                );
            } catch (e) {
                return cb(new CodeError(400, e.message))
            }
        },

        deleteDefault: (id, domain, cb) => {
            pool.query(
                sqlDeleteDefault,
                [
                    +id,
                    domain
                ],
                (err, res) => {
                    if (err) {
                        return cb(err);
                    }
                    if (res && res.rowCount) {
                        return cb(null, res.rows[0])
                    } else {
                        return cb(new CodeError(404, `Not found ${id}`));
                    }
                }
            );
        },

        moveDefault: (id, domain, up, cb) => {
            pool.query(
                up ? sqlMoveUpDefault : sqlMoveDownDefault,
                [
                    +id,
                    domain
                ],
                cb
            );
        },

        removeDefaultByDomain: (domain, cb) => {
            pool.query(
                sqlDeleteDefaultByDomain,
                [
                    domain
                ],
                (err, res) => {
                    if (err)
                        return cb(err);

                    return cb(null, (res && res.rowCount) || 0);
                }
            );
        },

        //endregion

        //region public query

        listPublic: (request, cb) => {
            buildQuery(pool, request, PUBLIC_TABLE_NAME, cb);
        },

        itemPublic: (id, domain, cb) => {
            pool.query(
                sqlItemPublic,
                [
                    +id,
                    domain
                ],
                (err, res) => {
                    if (err) {
                        return cb(err);
                    }
                    if (res && res.rowCount) {
                        return cb(null, res.rows[0])
                    } else {
                        return cb(new CodeError(404, `Not found ${id}`));
                    }
                }
            );
        },

        createPublic: (dialPlan = {}, cb) => {
            try {
                pool.query(
                    sqlInsertPublic,
                    //destination_number, name, domain, fs_timezone, callflow, callflow_on_disconnect, cf_diagram, disabled, debug, description
                    [
                        dialPlan.destination_number,
                        dialPlan.name,
                        dialPlan.domain,
                        dialPlan.fs_timezone,
                        dialPlan.callflow ? JSON.stringify(dialPlan.callflow) : null,
                        dialPlan.callflow_on_disconnect ? JSON.stringify(dialPlan.callflow_on_disconnect) : null,
                        dialPlan.cf_diagram ? JSON.stringify(dialPlan.cf_diagram) : null,
                        dialPlan.disabled,
                        dialPlan.debug,
                        dialPlan.description || ""
                    ],
                    (err, res) => {
                        if (err) {
                            return cb(err);
                        }
                        if (res && res.rowCount) {
                            return cb(null, res.rows[0])
                        } else {
                            return cb(new CodeError(500, `Bad response`));
                        }
                    }
                )
            } catch (e) {
                return cb(new CodeError(400, e.message))
            }
        },

        deletePublic: (id, domain, cb) => {
            pool.query(
                sqlDeletePublic,
                [
                    +id,
                    domain
                ],
                (err, res) => {
                    if (err) {
                        return cb(err);
                    }
                    if (res && res.rowCount) {
                        return cb(null, res.rows[0])
                    } else {
                        return cb(new CodeError(404, `Not found ${id}`));
                    }
                }
            );
        },

        updatePublic: (id, domain, dialPlan = {}, cb) => {
            try {
                pool.query(
                    sqlUpdatePublic,
                    [
                        dialPlan.destination_number,
                        dialPlan.name,
                        dialPlan.fs_timezone,
                        dialPlan.callflow ? JSON.stringify(dialPlan.callflow) : null,
                        dialPlan.callflow_on_disconnect ? JSON.stringify(dialPlan.callflow_on_disconnect) : null,
                        dialPlan.cf_diagram ? JSON.stringify(dialPlan.cf_diagram) : null,
                        dialPlan.disabled,
                        dialPlan.debug,
                        dialPlan.description || "",
                        +id,
                        domain
                    ],
                    (err, res) => {
                        if (err) {
                            return cb(err);
                        }
                        if (res && res.rowCount) {
                            return cb(null, res.rows[0])
                        } else {
                            return cb(new CodeError(404, `Not found ${id}`));
                        }
                    }
                );
            } catch (e) {
                return cb(new CodeError(400, e.message))
            }
        },

        removePublicByDomain: (domain, cb) => {
            pool.query(
                sqlDeletePublicByDomain,
                [
                    domain
                ],
                (err, res) => {
                    if (err)
                        return cb(err);

                    return cb(null, (res && res.rowCount) || 0);
                }
            );
        },

        //endregion


        //region extensions

        createExtension: (dialPlan = {}, cb) => {
            try {
                pool.query(
                    sqlInsertExtension,
                    //destination_number, domain, user_id, name, callflow, callflow_on_disconnect, cf_diagram
                    [
                        dialPlan.destination_number,
                        dialPlan.domain,
                        dialPlan.userRef,
                        dialPlan.name,
                        dialPlan.callflow ? JSON.stringify(dialPlan.callflow) : null,
                        dialPlan.callflow_on_disconnect ? JSON.stringify(dialPlan.callflow_on_disconnect) : null,
                        dialPlan.cf_diagram ? JSON.stringify(dialPlan.cf_diagram) : null
                    ],
                    (err, res) => {
                        if (err) {
                            return cb(err);
                        }
                        if (res && res.rowCount) {
                            return cb(null, res.rows[0])
                        } else {
                            return cb(new CodeError(500, `Bad response`));
                        }
                    }
                )
            } catch (e) {
                return cb(new CodeError(400, e.message))
            }
        },

        listExtension:  (request, cb) => {
            buildQuery(pool, request, EXTENSION_TABLE_NAME, cb);
        },

        itemExtension: (id, domain, cb) => {
            pool.query(
                sqlItemExtension,
                [
                    +id,
                    domain
                ],
                (err, res) => {
                    if (err) {
                        return cb(err);
                    }
                    if (res && res.rowCount) {
                        return cb(null, res.rows[0])
                    } else {
                        return cb(new CodeError(404, `Not found ${id}`));
                    }
                }
            );
        },

        existsExtension: (number, domain, cb) => {
            let n;
            if (number instanceof Array) {
                n = number
            } else {
                n = [number]
            }

            pool.query(
                sqlExistsExtension,
                [
                    domain,
                    n
                ],
                (err, res) => {
                    if (err) {
                        return cb(err);
                    }
                    if (res && res.rowCount > 0) {
                        return cb(null, res.rows[0].count > 0)
                    } else {
                        return cb(new CodeError(404, `Not found ${number}`));
                    }
                }
            );
        },

        removeExtensionFromUser: (userId, domain, cb) => {
            pool.query(
                sqlDeleteExtensionByUserId,
                [
                    domain,
                    userId
                ],
                (err, res) => {
                    if (err) {
                        return cb(err);
                    }
                    if (res && res.rowCount) {
                        return cb(null, res.rows[0])
                    } else {
                        return cb(new CodeError(404, `Not found ${userId}`));
                    }
                }
            );
        },

        updateExtension: (id, domain, dialPlan = {}, cb) => {
            try {
                pool.query(
                    sqlUpdateExtension,
                    [
                        dialPlan.destination_number,
                        dialPlan.name,
                        dialPlan.fs_timezone,
                        dialPlan.callflow ? JSON.stringify(dialPlan.callflow) : null,
                        dialPlan.callflow_on_disconnect ? JSON.stringify(dialPlan.callflow_on_disconnect) : null,
                        dialPlan.cf_diagram ? JSON.stringify(dialPlan.cf_diagram) : null,
                        +id,
                        domain
                    ],
                    (err, res) => {
                        if (err) {
                            return cb(err);
                        }
                        if (res && res.rowCount) {
                            return cb(null, res.rows[0])
                        } else {
                            return cb(new CodeError(404, `Not found ${id}`));
                        }
                    }
                );
            } catch (e) {
                return cb(new CodeError(400, e.message))
            }
        },

        removeExtensionById: (id, domain, cb) => {
            pool.query(
                sqlDeleteExtensionById,
                [domain, +id],
                (err, res) => {
                    if (err) {
                        return cb(err);
                    }
                    if (res && res.rowCount) {
                        return cb(null, res.rows[0])
                    } else {
                        return cb(new CodeError(404, `Not found ${id}`));
                    }
                }

            )
        },

        //endregion


        //region domain variables
        listDomainVariables: (request, cb) => {
            buildQuery(pool, request, VARIABLES_TABLE_NAME, cb);
        },

        insertOrUpdateDomainVariable: (domain, variables = {}, cb) => {
            try {
                pool.query(
                    sqlUpsertDomainVariables,
                    [
                        JSON.stringify(variables),
                        domain
                    ],
                    (err, res) => {
                        if (err) {
                            return cb(err);
                        }

                        return cb(null, variables)
                    }
                );
            } catch (e) {
                return cb(new CodeError(400, e.message))
            }
        },

        deleteVariables: (domain, cb) => {
            pool.query(
                sqlDeleteDomainVariables,
                [
                    domain
                ],
                (err, res) => {
                    if (err)
                        return cb(err);

                    return cb(null, (res && res.rowCount) || 0);
                }
            );
        }
        //endregion
    }
}

module.exports = add;