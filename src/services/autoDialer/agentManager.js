/**
 * Created by igor on 24.05.16.
 */

'use strict';

let EventEmitter2 = require('eventemitter2').EventEmitter2,
    ccService = require(__appRoot + '/services/callCentre'),
    accountService = require(__appRoot + '/services/account'),
    async = require('async'),
    log = require(__appRoot + '/lib/log')(module),
    AGENT_STATE = require('./const').AGENT_STATE,
    AGENT_STATUS = require('./const').AGENT_STATUS,
    AGENT_STRATEGY = require('./const').AGENT_STRATEGY
    ;

class AgentManager extends EventEmitter2 {

    constructor () {
        super();
    }

    checkSetAvailableTime (cb) {
        application.DB._query.dialer._getAgentCount(
            {
                setAvailableTime: {$lte: Date.now()}
            },
            (err, res) => {
                if (err)
                    return cb(err);

                if (res > 0) {
                    return async.times(res, (n, next) => {
                        this.setAvailableTime(next)
                    }, cb);
                }

                return cb();
            }
        )
    }

    setAvailableTime (cb) {
        application.DB._query.dialer._findAndModifyAgent(
            {"setAvailableTime": {$lte: Date.now()}},
            {},
            {$set: {setAvailableTime: null}},
            (err, res) => {
                if (err)
                    return cb(err);

                const {value: agent} = res;

                if (agent && agent.state === AGENT_STATE.Reserved) {
                    this.setAgentStatus(agent, AGENT_STATE.Waiting, e => {
                        if (e)
                            log.error(e);
                    });
                    return cb(null, true)
                } else {
                    log.warn('no agents');
                }
                return cb(null, false)
            }
        )
    }

    setAgentStatus (agent, status, cb) {
        // TODO if err remove agent ??
        log.trace(`try set new state ${agent.agentId} -> ${status}`);
        ccService._setAgentState(agent.agentId, status, cb);
    }

    setNoAnswerAgent (agent, cb) {
        ccService._setAgentStatus(agent.agentId, AGENT_STATUS.OnBreak, (err) => {
            if (err) {
                return log.error(err);
            }

            return this.setAgentStatus(agent, AGENT_STATE.Waiting, cb);
        })
    }

    initAgents (dialer, callback) {

        async.waterfall(
            [
                (cb) => {
                    accountService._listByDomain(dialer._domain, cb);
                },

                (agents, cb) => {
                    let _agents = [];

                    for (let agent in agents) {
                        if (agents.hasOwnProperty(agent) && agents[agent].agent === 'true')
                            _agents.push(agents[agent]);
                    }

                    cb(null, _agents)
                }
            ],
            (err, agents) => {
                if (err)
                    return log.error(err);

                async.eachSeries(agents,
                    (agent, cb) => {
                        let agentId = `${agent.id}@${agent.domain}`;

                        ccService._getAgentParams(agentId, (err, res) => {
                            if (err)
                                return cb(err);
                            let agentParams = res && res[0];
                            if (!agentParams) {
                                log.warn(`Bad agent ${agentId}`);
                                return cb();
                            }
                            if (agentParams) {
                                log.trace(`Upsert agent parameters - ${agentId}`, agentParams);
                                application.DB._query.dialer._initAgent(agentId, dialer._domain, agentParams, agent.skills ? agent.skills.split(',') : [], (e, res) => {
                                    if (e)
                                        return cb(e);

                                    application.DB._query.dialer._initAgentInDialer(agentId, dialer._objectId, cb);
                                });
                            } else {
                                return cb();
                            }
                        });
                    },
                    (err, res) => {
                        callback(err, res);
                    }
                );
            }
        );
    }

    setActiveAgentsInDialer (dialerId, active, cb) {
        application.DB._query.dialer._setActiveAgents(dialerId, active, cb);
    }

    getAvailableCount (dialerId, agents, skills, cb) {
        application.DB._query.dialer._getAgentCount(getAvailableAgentFilter(dialerId, agents, skills), cb);
    }

    getAllLoggedAgent (dialerId, agents, skills, cb) {
        application.DB._query.dialer._getAgentCount({
            status: {
                $nin: [AGENT_STATUS.LoggedOut, AGENT_STATUS.OnBreak]
            },
            dialer: {$elemMatch: {_id: dialerId}},
            $or: [
                {
                    agentId: {$in: agents}
                },
                {
                    skills: {$in: skills}
                }
            ]
        }, cb);
    }

    resetAgentsStats (dialerId, cb) {
        application.DB._query.dialer._updateAgentMulti(
            {
                dialer: {$elemMatch: {_id: dialerId}}
            },
            {
                $set: {
                    "dialer.$.lastStatus": "reset status",
                    "dialer.$.callCount": 0,
                    "dialer.$.missedCall": 0,
                    "dialer.$.gotCallCount": 0,
                    "dialer.$.callTimeSec": 0,
                    "dialer.$.lastBridgeCallTimeStart": 0,
                    "dialer.$.lastBridgeCallTimeEnd": 0,
                    "dialer.$.connectedTimeSec": 0,
                    "dialer.$.idleSec": 0,
                    "dialer.$.Available": 0,
                    "dialer.$.On Break": 0,
                    "dialer.$.Logged Out": 0,
                    "dialer.$.Available (On Demand)": 0,
                }
            },
            cb
        );
    }

    rollbeckAgent (agentId, dialerId, cb) {
        application.DB._query.dialer._findAndModifyAgent(
            {
                agentId: agentId,
                dialer: {$elemMatch: {_id: dialerId}}
            },
            {},
            {
                $set: {"dialer.$.process": null},
                $currentDate: { lastModified: true }
            },
            cb
        )
    }

    huntingAgent (dialerId, agents, skills, strategy, dialerReadyOn, member, cb) {
        const filter = getAvailableAgentFilter(dialerId, agents, skills);

        // console.dir(filter, {depth: 10, colors: true});

        const sort = {
            // "dialer._id": 1
        };

        switch (strategy) {
            case AGENT_STRATEGY.RANDOM:
                // filter.randomPoint = { $near : [Math.random(), 0] };
                sort.randomValue =  Math.random() > 0.5 ? 1 : -1;
                break;
            case AGENT_STRATEGY.WITH_FEWEST_CALLS:
                sort["dialer.callCount"] = 1;
                break;
            case AGENT_STRATEGY.WITH_LEAST_TALK_TIME:
                sort["dialer.callTimeSec"] = 1;
                break;

            case AGENT_STRATEGY.LONGEST_IDLE_AGENT:
                sort["dialer.idleSec"] = -1;
                break;

            case AGENT_STRATEGY.TOP_DOWN:
                //TODO
                // break;
            default:
                sort["dialer.callCount"] = 1;
        }

        application.DB._query.dialer._findAndModifyAgentByHunting(
                dialerId,
                filter,
                sort,
                {
                    $set: {"dialer.$.process": "active" , "dialer.$.lastStatus": `hunting for ${member._id}`},
                    $currentDate: { lastModified: true }
                },
                (err, res) => {
                    if (err)
                        return cb(err);

                    if (!res.value)
                        return cb();

                    const agent = res.value;
                    agent._idleTime = this.getIdleTimeSecAgent(agent, dialerReadyOn);

                    if (member.processEnd) {
                        return this.rollbeckAgent(agent.agentId, dialerId, e => {
                            if (e) {
                                log.error(`bad rollback agent ${agent.agentId}`);
                                return cb(e);
                            } else {
                                return cb();
                            }
                        });
                    }

                    this.setAgentStatus(agent, AGENT_STATE.Reserved, err => {
                        if (err) {
                            log.error(err);
                            this.rollbeckAgent(agent.agentId, dialerId, e => {
                                if (e) {
                                    log.error(`bad rollback agent ${agent.agentId}`);
                                    return cb(e);
                                } else {
                                    return cb();
                                }
                            })
                        } else {
                            return cb(null, agent)
                        }

                    })
                }
            )
    }

    // TODO
    getIdleTimeSecAgent (agent = {}, dialerReadyTime = 0) {
        if (!agent) {
            log.error("getIdleTimeSecAgent -> No agent!!!");
            return -1;
        }

        if (agent._idleTime >= 0) {
            return agent._idleTime
        }

        if (agent.state !== AGENT_STATE.Waiting ) {
            log.error("Agent no waiting -> ", agent);
            return -1
        }

        const lastChange = Math.max(agent.lastStateChange, agent.lastStatusChange);
        if (!lastChange) {
            log.error("Agent no lastStateChange -> ", agent);
            return -1
        }

        let idle = -1;
        if (dialerReadyTime < lastChange) {
            idle = Math.round((Date.now() - lastChange) / 1000);
        } else {
            idle = Math.round((Date.now() - dialerReadyTime) / 1000);
        }

        if (idle < 0 ) {
            log.error(`Agent idle = ${idle} -> `, agent);
            return 0
        }

        return idle

    }
    
    setAgentStats (agentId, dialerId, params = {}, cb) {
        const $set = {};
        const $inc = {};

        if (params.hasOwnProperty('process'))
            $set["dialer.$.process"] = params.process;

        if (params.hasOwnProperty('lastStatus'))
            $set["dialer.$.lastStatus"] = params.lastStatus;

        if (params.hasOwnProperty('setAvailableTime'))
            $set["setAvailableTime"] = params.setAvailableTime;

        if (params.hasOwnProperty('lastBridgeCallTimeStart'))
            $set["dialer.$.lastBridgeCallTimeStart"] = params.lastBridgeCallTimeStart;

        if (params.hasOwnProperty('lastBridgeCallTimeEnd'))
            $set["dialer.$.lastBridgeCallTimeEnd"] = params.lastBridgeCallTimeEnd;

        if (params.gotCall === true)
            $inc["dialer.$.gotCallCount"] = 1;

        if (params.call === true)
            $inc["dialer.$.callCount"] = 1;

        if (params.hasOwnProperty('callTimeSec'))
            $inc["dialer.$.callTimeSec"] = params.callTimeSec;

        if (params.hasOwnProperty('connectedTimeSec'))
            $inc["dialer.$.connectedTimeSec"] = params.connectedTimeSec;

        if (params.noAnswer === true)
            $inc["noAnswerCount"] = 1;

        if (params.idleSec)
            $inc["dialer.$.idleSec"] = params.idleSec;

        if (params.clearNoAnswer === true)
            $set["noAnswerCount"] = 0;

        if (params.hasOwnProperty('minNextCallTime'))
            $set.minNextCallTime = params.minNextCallTime;

        if (params.missedCall === true)
            $inc["dialer.$.missedCall"] = 1;

        
        const update = {
            $set,
            $currentDate: { lastModified: true }
        };

        if (Object.keys($inc).length > 0) {
            update.$inc = $inc;
        }

        application.DB._query.dialer._findAndModifyAgent(
            {
                agentId: agentId,
                dialer: {$elemMatch: {_id: dialerId}}
            },
            {},
            update,
            (err, res) => {
                if (err)
                    return cb(err);

                if (!res.value) {
                    log.error('Bad response setAgentStats update:', res);
                    return cb(new Error('Bad response setAgentStats update'));
                }

                return cb(err, res);
            }
        )
    }

    flushAgentProcess (agentId, dialerId, wrap, cb) {
        application.DB._query.dialer._findAndModifyAgent(
            {
                agentId: agentId,
                dialer: {$elemMatch: {_id: dialerId}}
            },
            {},
            {
                $set: {"dialer.$.process": null},
                $currentDate: { lastModified: true }
            },
            (err, res) => {
                if (err)
                    return cb(err);

                if (!res.value) {
                    log.error(`Bad response`, res);
                    return cb(new Error('Bad response flushAgentProcess query'));
                }

                return cb(err, res);
            }
        )
    }
}

module.exports = AgentManager;

function getAvailableAgentFilter(dialerId, agents, skills) {
    return {
        status: {
            $in: [AGENT_STATUS.Available, AGENT_STATUS.AvailableOnDemand]
        },
        state: AGENT_STATE.Waiting,
        setAvailableTime: null,
        dialer: {$elemMatch: {_id: dialerId}},
        "dialer.process": {$ne: "active"},
        $or: [
            {
                agentId: {$in: agents}
            },
            {
                skills: {$in: skills}
            }
        ]
    };
}