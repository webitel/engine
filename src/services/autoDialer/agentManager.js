/**
 * Created by igor on 24.05.16.
 */

'use strict';

let EventEmitter2 = require('eventemitter2').EventEmitter2,
    ccService = require(__appRoot + '/services/callCentre'),
    async = require('async'),
    log = require(__appRoot + '/lib/log')(module),
    AGENT_STATE = require('./const').AGENT_STATE,
    AGENT_STATUS = require('./const').AGENT_STATUS,
    AGENT_STRATEGY = require('./const').AGENT_STRATEGY,
    Event = require(__appRoot + '/lib/modesl').Event
;

const agentEvent = new Event('CUSTOM', 'callcenter::info');
agentEvent.addHeader('CC-Action', 'agent-state-change');

class AgentManager extends EventEmitter2 {

    constructor () {
        super();
    }

    setAgentStatus (agent, status, cb) {
        log.trace(`try set new state ${agent.name} -> ${status}`);
        ccService._setAgentState(agent.name, status, cb);
    }

    setNoAnswerAgent (agent, cb) {
        ccService._setAgentStatus(agent.name, AGENT_STATUS.OnBreak, (err) => {
            if (err) {
                return log.error(err);
            }

            return this.setAgentStatus(agent, AGENT_STATE.Waiting, cb);
        })
    }

    setActiveAgentsInDialer (dialerId, active, agents, cb) {
        application.PG.getQuery('agents').setActiveAgents(dialerId, active, agents, cb);
    }

    getAvailableCount (dialerId, agents, skills, cb) {
        application.PG.getQuery('agents').getAvailableCount(dialerId, agents, skills, cb);
    }


    getAllLoggedAgent (dialerId, agents, skills, cb) {
        application.PG.getQuery('agents').getAllLoggedAgent(dialerId, agents, skills, cb);
    }


    resetAgentsStats (dialerId, cb) {
        application.PG.getQuery('agents').resetAgentStats(dialerId, cb);
    }

    rollbackAgent (agentId, dialerId, cb) {
        application.PG.getQuery('agents').rollback(agentId, dialerId, (err, res) => {
            if (err)
                return cb(err);

            this.sendEvent(agentId, dialerId, AGENT_STATE.Waiting, () => {
                return cb(err, res);
            });
        });
    }

    sendEvent (agentId, dialerId, state, cb) {
        agentEvent.addHeader('CC-Agent', agentId);
        agentEvent.addHeader('CC-Agent-State', state);
        agentEvent.addHeader('Dialer-Id', dialerId.toString());
        application.Esl.sendEvent(agentEvent, cb)
    }

    huntingAgent (dialer, member, cb) {
        const dialerId = dialer._id.toString(),
            agents = dialer._agents,
            skills = dialer._skills,
            strategy = dialer.agentStrategy
        ;


        let sort = '';
        switch (strategy) {
            case AGENT_STRATEGY.RANDOM:
                sort = 'random()';
                break;

            case AGENT_STRATEGY.WITH_LEAST_TALK_TIME:
                sort = 'ad.call_time_sec ASC NULLS FIRST';
                break;

            case AGENT_STRATEGY.LONGEST_IDLE_AGENT:
                sort = 'ad.idle_sec DESC NULLS FIRST';
                break;

            case AGENT_STRATEGY.WITH_LEAST_UTILIZATION:
                sort = '1 - (COALESCE(ad.idle_sec, 0) / GREATEST(COALESCE(ad.call_time_sec, 0) + COALESCE(ad.connected_time_sec, 0) + COALESCE(ad.wrap_time_sec,0) + COALESCE(ad.idle_sec, 0), 0.00001)::FLOAT) ASC';
                break;

            case WITH_HIGHEST_WAITING_TIME:
                sort = 'a.last_status_change ASC NULLS FIRST';
                break;
            //case AGENT_STRATEGY.TOP_DOWN:
            //TODO
            // break;
            //case AGENT_STRATEGY.WITH_FEWEST_CALLS:
            default:
                sort =  'ad.call_count ASC NULLS FIRST';
        }

        application.PG.getQuery('agents').huntingAgent(dialerId, agents, skills, sort, member, (err, agent) => {
            if (err)
                return cb(err);

            if (!agent)
                return cb();

            if (member.processEnd) {
                return this.rollbackAgent(agent.name, dialerId, err => {
                    if (err) {
                        log.error(`bad rollback agent ${agent.agentId}`);
                        return cb(err);
                    } else {
                        return cb();
                    }
                })
            }

            this.sendEvent(agent.name, dialerId, AGENT_STATE.Reserved, () => {
                return cb(null, agent);
            });
        });
    }


    setAgentStats (agent, dialerId, params = {}, cb) {
        application.PG.getQuery('agents').setStatus(agent, dialerId, params, (err, res) => {
            if (params.process === null && params.call === true && agent.status !== AGENT_STATUS.AvailableOnDemand) {
                this.rollbackAgent(agent.name, dialerId, err => {
                    if (err) {
                        log.error(`bad rollback agent ${agent.agentId}`);
                    }
                })

            }
            return cb(err, res)
        })
    }
}

module.exports = AgentManager;