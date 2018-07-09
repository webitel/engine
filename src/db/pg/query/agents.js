"use strict";

const log = require(__appRoot + '/lib/log')(module),
    CodeError = require(__appRoot + '/lib/error');


function add(pool) {

    return {

        getAgentStats: (dialerId, agents, skills, domainName, cb) => {
            pool.query(`
                SELECT a.name, a.ready_time, wrap_up_time, last_bridge_end, a.state, a.status, a.logged_in, a.logged_out,
                    extract(EPOCH FROM a.last_set_stats)::INT as last_set_stats, 
                    ad.call_count, ad.missed_call, ad.call_time_sec, ad.connected_time_sec, ad.idle_sec, ad.on_break_sec, ad.wrap_time_sec,
                    extract(EPOCH FROM ad.active)::BIGINT as active, ad.bridged_count, ad.last_offered_call as last_offered_call
                FROM agents a
                LEFT JOIN agent_in_dialer ad on a.name = ad.agent_name AND ad.dialer_id = $1
                WHERE a.name like $2
                    AND a.name = ANY($3)
                ORDER BY a.name;
                `,
                [dialerId.toString(), '%' + domainName, agents],
                (err, res) => {
                    if (err)
                        return cb(err);

                    if (res) {
                        return cb(null, res.rows)
                    }
                    return cb(null, [])
                }
            )
        },

        resetAgentStats: (dialerId, cb) => {
            pool.query(
                `DELETE FROM agent_in_dialer WHERE dialer_id = $1`,
                [dialerId.toString()],
                cb
            )
        },

        setActiveAgents: (dialerId, active, agents, cb) => {
            if (active > 0) {
                pool.query(`
                    WITH upd AS (
                        UPDATE agent_in_dialer
                        SET active = NOW() at time zone 'utc'
                        WHERE dialer_id = $1
                    )
                    INSERT INTO agent_in_dialer (agent_name, active, dialer_id)
                    SELECT a.name, NOW() at time zone 'utc', $2
                    FROM agents a
                    WHERE a.name = ANY($3) AND NOT EXISTS (select * FROM agent_in_dialer as ad WHERE ad.agent_name = a.name and ad.dialer_id = $4)
                `,
                    [
                        dialerId.toString(), dialerId.toString(), agents, dialerId.toString()
                    ],
                    cb
                );
            } else {
                //TODO calc last status
                pool.query(`UPDATE agent_in_dialer
                SET active = NULL
                WHERE dialer_id = $1`,
                    [
                        dialerId.toString()
                    ],
                    cb
                );
            }
        },

        getAvailableCount: (dialerId, agents, skills, cb) => {
            const epoch = getEpoch();
            pool.query(`SELECT count(a.name)::int as Count
                  FROM agents a
                    WHERE a.status in ('Available', 'Available (On Demand)')
                    AND a.state = 'Waiting'
                    AND a.last_bridge_end < ${epoch} - a.wrap_up_time
                    AND a.ready_time <= ${epoch}
                    AND a.name = ANY($1)`,
                [
                    agents
                ],
                (err, res) => {
                    return callbackCount(err, res, cb)
                }
            );
        },

        getAllLoggedAgent: (dialerId, agents, skills, cb) => {
            pool.query(`SELECT count(*)::int as Count
                        FROM agents a
                        WHERE NOT a.status in ('Logged Out', 'On Break')
                          AND a.name = ANY($1)`,
                [
                    agents
                ],
                (err, res) => {
                    return callbackCount(err, res, cb)
                }
            );
        },


        rollback: (agentId, dialerId, cb) => {
            pool.query(`UPDATE agents
                SET state = 'Waiting'
                WHERE name = $1
            `,
                [
                    agentId
                ],
                (err, res) => {
                    if (err)
                        return cb(err);

                    if (!res.rowCount) {
                        console.log('ERROR');
                    }
                    cb()
                }
            );
        },

        reset: (dialerId, cb) => {
            pool.query(`UPDATE agent_in_dialer
                SET process = null
                WHERE dialer_id = $1
            `,
                [
                    dialerId + ""
                ],
                cb
            );
        },

        setStatus: (agent, dialerId, params = {}, cb) => {
            const fieldsAgentInDialer = [];
            const fieldsAgent = [];
            //region agent

            if (params.noAnswer === true) {
                fieldsAgent.push(`no_answer_count = no_answer_count + 1`)
            }

            if (params.wrapTime > 0) {
                fieldsAgent.push(`ready_time = ${Math.round(Date.now() / 1000) + params.wrapTime}`)
            }

            if (params.bridged && agent.no_answer_count > 0) {
                fieldsAgent.push(`no_answer_count = 0`)
            }
            //endregion

            //region ad
            // if (params.hasOwnProperty('process')) {
            //     if (!params.process) {
            //         fieldsAgentInDialer.push(`process=NULL`);
            //     } else {
            //         fieldsAgentInDialer.push(`process="${params.process}"`);
            //     }
            //
            // }

            if (params.wrapTime > 0 && params.bridged === true) {
                fieldsAgentInDialer.push(`wrap_time_sec = wrap_time_sec + ${params.wrapTime}`);
            }

            if (params.hasOwnProperty('lastStatus')) {
                fieldsAgentInDialer.push(`last_status='${params.lastStatus}'`);
            }

            if (params.bridged === true) {
                fieldsAgentInDialer.push(`bridged_count = bridged_count + 1`);
            }

            if (params.call === true) {
                fieldsAgentInDialer.push(`call_count = call_count + 1`);
                fieldsAgentInDialer.push(`last_offered_call = ${Math.round(Date.now()/1000)}`)
            }

            if (params.hasOwnProperty('callTimeSec')) {
                fieldsAgentInDialer.push(`call_time_sec = call_time_sec + ${params.callTimeSec}`);
            }

            if (params.hasOwnProperty('connectedTimeSec')) {
                fieldsAgentInDialer.push(`connected_time_sec = connected_time_sec + ${params.connectedTimeSec}`);
            }

            if (params.missedCall === true) {
                fieldsAgentInDialer.push(`missed_call = missed_call + 1`);
            }

            //endregion

            if (fieldsAgentInDialer.length === 0) {
                return cb && cb(new Error("Bad update agent parameters"))
            }

            if (fieldsAgent.length > 0) {
                pool.query(
                    `UPDATE agents
                      SET ${fieldsAgent.join(',\n')}
                    WHERE name = $1`,
                    [agent.name],
                    (err, res) => {
                        if (err)
                            return cb(err);

                        if (!res.rowCount) {
                            console.log('ERROR');
                        }

                        pool.query(
                            `UPDATE agent_in_dialer
                                    SET ${fieldsAgentInDialer.join(',\n')}
                              WHERE agent_name = $1 and dialer_id = $2`,
                            [agent.name, dialerId.toString()],
                            (err) => {
                                if (err)
                                    return cb(err);

                                cb(null, agent)
                            }
                        )

                    }
                )
            } else {
                pool.query(
                    `UPDATE agent_in_dialer
                        SET ${fieldsAgentInDialer.join(',\n')}
                    WHERE agent_name = $1 and dialer_id = $2`,
                    [agent.name, dialerId.toString()],
                    (err) => {
                        if (err)
                            return cb(err);

                        cb(null, agent)
                    }
                )
            }
        },

        //todo  add def active dialer ATOMIC!!!
        huntingAgent: (dialerId, agents, skills, orderBy, member, cb) => {
            const epoch = getEpoch();

            pool.query(
                `WITH cte AS (
                     SELECT a.name, ad.id as ad_id
                     FROM agents a
                        left join agent_in_dialer ad on ad.agent_name = a.name and ad.dialer_id = $1
                     WHERE  a.status in ('Available', 'Available (On Demand)')
                      AND a.state = 'Waiting'
                      AND a.last_bridge_end < $2 - a.wrap_up_time
                      AND a.ready_time <= $3
                      AND a.name = ANY ($4)
                     ORDER BY ${orderBy}
                     LIMIT 1
                     FOR UPDATE OF a SKIP LOCKED
                )
                UPDATE agents a
                SET    state = 'Reserved'
                FROM   cte
                WHERE  a.name = cte.name AND NOT EXISTS (
                    SELECT 1 FROM agents WHERE name = cte.name AND state = 'Reserved'                
                )
                
                RETURNING a.name, a.contact, a.status, a.state, a.max_no_answer, a.wrap_up_time, a.reject_delay_time, a.no_answer_delay_time, a.no_answer_count, cte.ad_id`,
                [dialerId.toString(), epoch, epoch, agents],
                (err, res) => {
                    if (err)
                        return cb(err);


                    if (res && res.rows.length > 0 && res.rows[0].name) {
                        if (!res.rows[0].ad_id) {
                            pool.query(`INSERT INTO agent_in_dialer (agent_name, dialer_id, active)
                                VALUES ($1, $2, NOW() at time zone 'utc')`,
                                [res.rows[0].name, dialerId.toString()],
                                err => {
                                    if (err)
                                        return log.error(err);
                                }
                            )
                        }
                        return cb(null, res.rows[0])
                    }
                    return cb(null, null) //TODO
                }
            );
            return;
            /*
                        pool.query(
                            `
                            SELECT * from dialer_hunting_agent($1, $2, $3, $4, $5)
                            as (name VARCHAR(255), contact VARCHAR(1024), status VARCHAR(255), state VARCHAR(255), max_no_answer int, wrap_up_time int, reject_delay_time int, no_answer_delay_time int, no_answer_count int);
                            `,
                            [epoch, dialerId.toString(), agents, skills, orderBy],
                            (err, res) => {
                                if (err)
                                    return cb(err);
                                //console.log(res.rows);
                                if (res && res.rows.length > 0 && res.rows[0].name) {
                                    return cb(null, res.rows[0])
                                }
                                return cb(null, null) //TODO
                            }
                        )



                        pool.query(`
                            WITH ag AS (
                              SELECT *
                              FROM agents a
                              LEFT JOIN agent_in_dialer ad on a.name = ad.agent_name AND ad.dialer_id = $1
                                WHERE a.status in ('Available', 'Available (On Demand)')
                                AND a.state = 'Waiting'
                                AND a.last_bridge_end < ${epoch} - a.wrap_up_time
                                AND a.ready_time <= ${epoch}
                                AND (
                                    a.name = ANY($2)
                                    OR exists(
                                         SELECT *
                                         FROM unnest(a.skills) s
                                         WHERE s = ANY($3)
                                     )
                                    )
                                  --TODO add dialer ???

                                AND NOT EXISTS(SELECT * from agent_in_dialer a1 where a1.agent_name = a.name AND a1.process = 'active')
                              ORDER BY ${orderBy}
                              LIMIT 1
                            ), upd AS (
                              UPDATE agent_in_dialer
                              SET process = 'active',
                                last_offered_call = extract(EPOCH FROM now() )::BIGINT
                              WHERE agent_name = (SELECT ag.name FROM ag)
                              RETURNING *
                            ), ins AS (
                              INSERT INTO agent_in_dialer (agent_name, process, dialer_id, active, last_offered_call)
                              SELECT ag.name, 'active', $4, NOW() at time zone 'utc', extract(EPOCH FROM now() )::BIGINT FROM ag
                              WHERE NOT EXISTS(SELECT *
                                               FROM upd)
                            )
                            SELECT *
                            FROM ag;
                        `,
                            [
                                dialerId.toString(),
                                agents,
                                skills,
                                dialerId.toString()
                            ],
                            (err, res) => {
                                if (err)
                                    return cb(err);

                                if (res && res.rows.length > 0) {
                                    return cb(null, res.rows[0])
                                }
                                return cb(null, null) //TODO
                            });

                            */
        },

        setUserStats: (event = {}, cb) => {
            pool.query(
                `with old as (
                    SELECT *
                    FROM user_stats
                    WHERE id = $1
                    LIMIT 1
                    FOR UPDATE
                ), upd as (
                  INSERT INTO user_stats (id, state, status, description, cc, ws, updated_at)
                  VALUES ($1, $2, $3, $4, $5, $6, (extract(EPOCH FROM now() AT TIME ZONE 'UTC') * 1000)::BIGINT)
                  ON CONFLICT (id)
                  DO UPDATE SET state = $2, status = $3, description = $4, cc = $5, ws = $6, updated_at = (extract(EPOCH FROM now() AT TIME ZONE 'UTC') * 1000)::BIGINT
                )
                SELECT * FROM old`,
                [event['presence_id'], event['Account-User-State'], event['Account-Status'], event['Account-Status-Descript'] || "", event['cc'], event['ws']],
                (err, res) => {
                    if (err)
                        return cb(err);

                    if (res.rowCount) {
                        return cb(null, res.rows[0])
                    }
                    return cb(null, null)
                }
            )
        }
    }
}

module.exports = add;

function callbackCount(err, res, cb) {
    if (err)
        return cb(err);

    if (res && res.rows.length > 0) {
        return cb(null, res.rows[0].count)
    }
    return cb(null, 0) //TODO
}

function getEpoch() {
    return Math.round(Date.now() / 1000);
}