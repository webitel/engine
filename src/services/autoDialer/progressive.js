/**
 * Created by igor on 25.05.16.
 */

let Dialer = require('./dialer'),
    log = require(__appRoot + '/lib/log')(module),
    Gw = require('./gw'),
    DIALER_TYPES = require('./const').DIALER_TYPES;

module.exports = class Progressive extends Dialer {
    constructor (config, calendarConf) {
        super(DIALER_TYPES.ProgressiveDialer, config, calendarConf);

        this._am = config.agentManager;
        this._gw = new Gw({}, null, this._variables);
        this._agentReserveCallback = [];
        this._agents = [];

        if (config.agents instanceof Array)
            this._agents = [].concat(config.agents); //.map( (i)=> `${i}@${this._domain}`);


        if (this._limit > this._agents.length && this._skills.length === 0  )
            this._limit = this._agents.length;

        let getMembersFromEvent = (e) => {
            return this.members.get(e.getHeader('variable_dlr_member_id'))
        };

        let onChannelCreate = (e) => {
            let m = getMembersFromEvent(e);
            if (m) {
                m.channelsCount++;
                m.setCallUUID(e.getHeader('variable_uuid'));
            }
        };
        let onChannelDestroy = (e) => {
            let m = getMembersFromEvent(e);
            if (m && --m.channelsCount === 0) {
                if (m._agentNoAnswer !== true) {
                    m._agent._noAnswerCallCount = 0;
                    this._am.taskUnReserveAgent(m._agent, m._agent.getTime('wrapUpTime', this.defaultAgentParams), true, this.defaultAgentParams);
                } else {
                    m.nextTrySec = 1;
                    this._am.taskUnReserveAgent(m._agent, m._agent.getTime('noAnswerDelayTime', this.defaultAgentParams), false, this.defaultAgentParams);
                }
                this.addMemberCallbackQueue(m, e, m._agent.getTime('wrapUpTime', this.defaultAgentParams));
            }
        };

        this.once('end', () => {
            log.debug('Off channel events');
            application.Esl.off('esl::event::CHANNEL_DESTROY::*', onChannelDestroy);
            application.Esl.off('esl::event::CHANNEL_CREATE::*', onChannelCreate);
        });
        application.Esl.subscribe(['CHANNEL_CREATE', 'CHANNEL_DESTROY']);

        application.Esl.on('esl::event::CHANNEL_DESTROY::*', onChannelDestroy);
        application.Esl.on('esl::event::CHANNEL_CREATE::*', onChannelCreate);

    }

    dialMember (member) {
        log.trace(`try call ${member.sessionId}`);

        let gw = this._gw.fnDialString(member);

        this.findAvailAgents( (agent) => {
            member.log(`set agent: ${agent.id}`);
            let ds = gw(agent, null, null, this.defaultAgentParams);
            member._gw = gw;
            member._agent = agent;

            member.log(`dialString: ${ds}`);
            log.trace(`Call ${ds}`);
            member.inCall = true;
            application.Esl.bgapi(ds, (res) => {
                log.trace(`fs response: ${res && res.body}`);
                if (/^-ERR|^-USAGE/.test(res.body)) {
                    let error =  res.body.replace(/-ERR\s(.*)\n/, '$1');
                    member.log(`agent: ${error}`);
                    member.minusProbe();
                    member.nextTrySec = 1;
                    // TODO ??
                    this.nextTrySec = 0;
                    let delayTime = agent.getTime('rejectDelayTime', this.defaultAgentParams);
                    if (error == 'NO_ANSWER') {
                        agent._noAnswerCallCount++;
                        member._agentNoAnswer = true;
                        delayTime = agent.getTime('noAnswerDelayTime', this.defaultAgentParams);
                    }
                    member.end();
                    this._am.taskUnReserveAgent(agent, delayTime, false, this.defaultAgentParams);
                } else {
                    agent.lastBridgeCallTimeStart = Date.now();
                }
            });
        });

    }

    setAgent (agent) {
        this.checkSleep();
        if (this._agentReserveCallback.length === 0 || !this.isReady())
            return false;

        const fn = this._agentReserveCallback.shift();
        this._am.reserveAgent(agent, (err) => {
            if (err) {
                this._agentReserveCallback.push(fn);
                return log.error(err);
            }

            if(typeof fn === 'function')
                fn(agent);
        });
        return true;
    }

    findAvailAgents (cb) {
        var a = this._am.getFreeAgent(this._agents, this.agentStrategy);
        if (a) {
            this._am.reserveAgent(a, (err) => {
                if (err) {
                    log.error(err);
                    return this._agentReserveCallback.push(cb);
                }
                cb(a)
            })
        } else {
            this._agentReserveCallback.push(cb);
            console.log(`find agent... queue length ${this._agentReserveCallback.length}`);
        }
    }
};