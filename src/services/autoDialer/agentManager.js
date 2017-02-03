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
    DIFF_CHANGE_MSEC = require('./const').DIFF_CHANGE_MSEC,
    AGENT_STRATEGY = require('./const').AGENT_STRATEGY
    ;

class AgentManager extends EventEmitter2 {

    constructor () {
        super();
    }


    taskUnReserveAgent (agent, timeSec, gotAgentCall, dilerAgentParams = {}) {
        if (agent.lock === true) {
            agent.lock = false;

            let wrapTime = Date.now() + (timeSec * 1000),
                agentId = agent.id
                ;

            agent.lockTime = wrapTime + DIFF_CHANGE_MSEC;
            // TODO
            if (agent.availableTime > agent.lockTime)
                agent.availableTime = Infinity;

            if (gotAgentCall) {
                agent.callCount++;
                agent.lastBridgeCallTimeEnd = Date.now();
                agent.callTimeMs += agent.lastBridgeCallTimeEnd - agent.lastBridgeCallTimeStart;
            }

            agent.unIdleTime = wrapTime;
            log.trace(`Set agent lock time ${timeSec} sec`);

            let maxNoAnswer = isFinite(dilerAgentParams.maxNoAnswer) ? dilerAgentParams.maxNoAnswer : agent.maxNoAnswer;
            if (maxNoAnswer != 0 && agent._noAnswerCallCount >= maxNoAnswer) {
                this.setNoAnswerAgent(agent, (err) => {
                    if (err)
                        return log.error(err);
                    agent.lockTime = 0;
                    agent._noAnswerCallCount = 0;
                    return log.trace(`change  ${agentId} status to no answer`);
                });
            }
        }
    }

    setAgentStatus (agent, status, cb) {
        // TODO if err remove agent ??
        log.trace(`try set new state ${agent.agentId} -> ${status}`);
        ccService._setAgentState(agent.agentId, status, cb);
    }

    setNoAnswerAgent (agent, cb) {
        ccService._setAgentStatus(agent.id, AGENT_STATUS.OnBreak, (err) => {
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
                                log.debug(`Upsert agent parameters - ${agentId}`);
                                application.DB._query.dialer._initAgent(agentId,  agentParams, agent.skills ? agent.skills.split(',') : [], (e, res) => {
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

    getAvailableCount (dialerId, agents, skills, cb) {
        application.DB._query.dialer._getAgentCount(getAvailableAgentFilter(dialerId, agents, skills), cb);
    }

    rollbeckAgent (agentId, dialerId, cb) {
        application.DB._query.dialer._findAndModifyAgent(
            {
                agentId: agentId,
                dialer: {$elemMatch: {_id: dialerId}}
            },
            {},
            {
                $set: {"dialer.$.process": null, "dialer.$.setAvailableTime": null},
                $currentDate: { lastModified: true }
            },
            cb
        )
    }

    huntingAgent (dialerId, agents, skills, strategy, cb) {
        const filter = getAvailableAgentFilter(dialerId, agents, skills);

        // console.dir(filter, {depth: 10, colors: true});

        const sort = {};

        switch (strategy) {
            case AGENT_STRATEGY.RANDOM:
                filter.randomPoint = { $near : [Math.random(), 0] };
                break;
            case AGENT_STRATEGY.WITH_FEWEST_CALLS:
                sort["dialer.callCount"] = 1;
                break;
            case AGENT_STRATEGY.WITH_LEAST_TALK_TIME:
                sort["dialer.callTimeSec"] = 1;
                break;

            case AGENT_STRATEGY.TOP_DOWN:
                //TODO
                // break;
            default:
                sort["dialer.callCount"] = 1;
        }

        application.DB._query.dialer._findAndModifyAgent(
                filter,
                sort,
                {
                    $set: {"dialer.$.process": "active"},
                    $currentDate: { lastModified: true }
                },
                (err, res) => {
                    if (err)
                        return cb(err);

                    if (!res.value)
                        return cb();

                    const agent = res.value;
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

    setAgentStats (agentId, dialerId, params = {}, cb) {
        const $set = {};
        const $inc = {};

        if (params.hasOwnProperty('process'))
            $set["dialer.$.process"] = params.process;

        if (params.hasOwnProperty('lastStatus'))
            $set["dialer.$.lastStatus"] = params.lastStatus;

        if (params.hasOwnProperty('setAvailableTime'))
            $set["dialer.$.setAvailableTime"] = params.setAvailableTime;

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

                if (!res.value)
                    throw res;

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
                $set: {"dialer.$.process": null, "dialer.$.setAvailableTime": wrap},
                $currentDate: { lastModified: true }
            },
            (err, res) => {
                if (err)
                    return cb(err);

                if (!res.value)
                    throw res;

                return cb(err, res);
            }
        )
    }

    getAgentById (id) {
    }
}

module.exports = AgentManager;

function getAvailableAgentFilter(dialerId, agents, skills) {
    return {
        status: {
            $in: [AGENT_STATUS.Available, AGENT_STATUS.AvailableOnDemand]
        },
        state: AGENT_STATE.Waiting,
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