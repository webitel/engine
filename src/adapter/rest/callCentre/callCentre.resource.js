/**
 * Created by Igor Navrotskyj on 03.09.2015.
 */

'use strict';

var ccServices = require(__appRoot + '/services/callCentre');

module.exports = {
    addRoutes: addRoutes
};

function addRoutes(api) {
    // TODO rename to /api/v2/callcenter/queues/:queue/agents (Kibana usage api)
    api.get('/api/v2/callcenter/queues/:queue/tiers', getTiers);
    // TODO error...
    api.get('/api/v2/callcenter/queues/:queue/tiers_', getTiersByQueue);
    api.get('/api/v2/callcenter/queues/:queue/members', getMembers);
    api.get('/api/v2/callcenter/queues/:queue/members/count', getMembersCount);

    // TODO

    api.get('/api/v2/callcenter/queues', queuesList);
    api.post('/api/v2/callcenter/queues', queueCreate);
    api.get('/api/v2/callcenter/queues/:name', queueItem);
    api.put('/api/v2/callcenter/queues/:name', queueUpdate);
    api.patch('/api/v2/callcenter/queues/:name/:state', queueSetState);
    // For Supertest.
    api.put('/api/v2/callcenter/queues/:name/:state', queueSetState);
    api.delete('/api/v2/callcenter/queues/:name', queueDelete);
    api.post('/api/v2/callcenter/queues/:queue/tiers', createTier);

    api.patch('/api/v2/callcenter/queues/:queue/tiers/:agent/level', setTierLevel);
    // For Supertest.
    api.put('/api/v2/callcenter/queues/:queue/tiers/:agent/level', setTierLevel);

    api.patch('/api/v2/callcenter/queues/:queue/tiers/:agent/position', setTierPosition);
    // For Supertest.
    api.put('/api/v2/callcenter/queues/:queue/tiers/:agent/position', setTierPosition);

    api.delete('/api/v2/callcenter/queues/:queue/tiers/:agent', deleteTier);

    // WTEL-323
    api.post('/api/v2/callcenter/agent/:id/status', agentSetStatus);
    api.post('/api/v2/callcenter/agent/:id/state', agentSetState);
    api.get('/api/v2/callcenter/agent/:id', agentGetParams);
}

function getTiersByQueue (req, res, next) {
    var options = {
        "domain": req.query['domain'],
        "queue": req.params['queue']
    };
    ccServices.getTiersByFilter(req.webitelUser, options, function (err, arr) {
        if (err)
            return next(err);

        res
            .status(200)
            .json({
                "status": "OK",
                "info": arr
            })
    });
}

function getTiers (req, res, next) {
    var options = {
        "domain": req.query['domain'],
        "queue": req.params['queue']
    };
    ccServices.getTiers(req.webitelUser, options, function (err, arr) {
        if (err)
            return next(err);

        res
            .status(200)
            .json({
                "status": "OK",
                "info": arr
            })
    });
}

function getMembers (req, res, next) {
    var options = {
        "domain": req.query['domain'],
        "queue": req.params['queue']
    };
    ccServices.getMembers(req.webitelUser, options, function (err, arr) {
        if (err)
            return next(err);

        res
            .status(200)
            .json({
                "status": "OK",
                "info": arr
            })
    });
}

function getMembersCount (req, res, next) {
    var options = {
        "domain": req.query['domain'],
        "queue": req.params['queue'],
        "count": true
    };
    ccServices.getMembers(req.webitelUser, options, function (err, count) {
        if (err)
            return next(err);

        res
            .status(200)
            .json({
                "status": "OK",
                "info": parseInt(count)
            })
    });
}

function queuesList (req, res, next) {
    var option = {
        "domain": req.query['domain']
    };
    ccServices.queuesList(req.webitelUser, option, function (err, result) {
        if (err)
            return next(err);

        res
            .status(200)
            .json({
                "status": "OK",
                "info": result
            })
    });
}

function queueCreate (req, res, next) {
    var option = req.body;
    if (req.query['domain']) {
        option['domain'] = req.query['domain'];
    };

    ccServices.queueCreate(req.webitelUser, option, function (err, result) {
        if (err)
            return next(err);

        res
            .status(200)
            .json({
                "status": "OK",
                "info": result
            })
    });
}

function queueItem (req, res, next) {
    var option = {
        "name": req.params['name'],
        "domain": req.query['domain']
    };

    ccServices.queueItem(req.webitelUser, option, function (err, result) {
        if (err)
            return next(err);

        res
            .status(200)
            .json({
                "status": "OK",
                "info": result
            })
    });
}

function queueUpdate (req, res, next) {
    var option = {
        "name": req.params['name'],
        "domain": req.query['domain'],
        "params": req.body
    };

    ccServices.queueUpdate(req.webitelUser, option, function (err, result) {
        if (err)
            return next(err);

        res
            .status(200)
            .json({
                "status": "OK",
                "info": result
            })
    });
}

function queueSetState (req, res, next) {
    var option = {
        "name": req.params['name'],
        "domain": req.query['domain'],
        "state": req.params['state']
    };

    ccServices.queueSetState(req.webitelUser, option, function (err, result) {
        if (err)
            return next(err);

        res
            .status(200)
            .json({
                "status": "OK",
                "info": result
            })
    });
}

function queueDelete (req, res, next) {
    var option = {
        "name": req.params['name'],
        "domain": req.query['domain']
    };

    ccServices.queueDelete(req.webitelUser, option, function (err, result) {
        if (err)
            return next(err);

        res
            .status(200)
            .json({
                "status": "OK",
                "info": result
            })
    });
}

function createTier (req, res, next) {
    var option = {
        "queue": req.params['queue'],
        "agent": req.body['agent'],
        "domain": req.query['domain'],
        "level": req.body['level'],
        "position": req.body['position']
    };

    ccServices.tierCreate(req.webitelUser, option, function (err, result) {
        if (err)
            return next(err);

        res
            .status(200)
            .json({
                "status": "OK",
                "info": result
            })
    });
}

function setTierLevel (req, res, next) {
    var option = {
        "queue": req.params['queue'],
        "agent": req.params['agent'],
        "domain": req.query['domain'],
        "level": req.body['level']
    };

    ccServices.tierSetLevel(req.webitelUser, option, function (err, result) {
        if (err)
            return next(err);

        res
            .status(200)
            .json({
                "status": "OK",
                "info": result
            })
    });
}

function setTierPosition (req, res, next) {
    var option = {
        "queue": req.params['queue'],
        "agent": req.params['agent'],
        "domain": req.query['domain'],
        "position": req.body['position']
    };

    ccServices.tierSetPosition(req.webitelUser, option, function (err, result) {
        if (err)
            return next(err);

        res
            .status(200)
            .json({
                "status": "OK",
                "info": result
            })
    });
}

function deleteTier (req, res, next) {
    var option = {
        "queue": req.params['queue'],
        "agent": req.params['agent'],
        "domain": req.query['domain']
    };

    ccServices.tierDelete(req.webitelUser, option, function (err, result) {
        if (err)
            return next(err);

        res
            .status(200)
            .json({
                "status": "OK",
                "info": result
            })
    });
}

function agentSetStatus(req, res, next) {
    let options = {
        domain: req.query['domain'],
        agent: req.params['id'],
        status: req.body.status
    };

    ccServices.setAgentStatus(req.webitelUser, options, (err, result) => {
        if (err)
            return next(err);

        res
            .status(200)
            .json({
                "status": "OK",
                "info": result
            })
    });
}

function agentSetState(req, res, next) {
    let options = {
        domain: req.query['domain'],
        agent: req.params['id'],
        state: req.body.state
    };

    ccServices.setAgentState(req.webitelUser, options, (err, result) => {
        if (err)
            return next(err);

        res
            .status(200)
            .json({
                "status": "OK",
                "info": result
            })
    });
}

function agentGetParams(req, res, next) {
    let options = {
        domain: req.query['domain'],
        id: req.params['id']
    };

    ccServices.getAgentParams(req.webitelUser, options, (err, result) => {
        if (err)
            return next(err);

        res
            .status(200)
            .json({
                "status": "OK",
                "info": result
            })
    });
}