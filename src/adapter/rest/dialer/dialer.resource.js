/**
 * Created by igor on 26.04.16.
 */

'use strict';


const dialerService = require(__appRoot + '/services/dialer'),
    parseQueryToObject = require(__appRoot + '/utils/parse').parseQueryToObject,
    getIp = require(__appRoot + '/utils/ip'),
    getRequest = require(__appRoot + '/utils/helper').getRequest
;

module.exports = {
    addRoutes: addRoutes
};

function addRoutes (api) {
    api.get('/api/v2/dialer', list);
    api.post('/api/v2/dialer', create);
    api.get('/api/v2/dialer/:id', item);
    api.put('/api/v2/dialer/:id', update);
    api.put('/api/v2/dialer/:id/reset', resetProcess);

    api.put('/api/v2/dialer/:id/state/:state', setState);
    api.delete('/api/v2/dialer/:id', remove);

    api.get('/api/v2/dialer/:dialer/members', listMembers);
    api.get('/api/v2/dialer/:dialer/members/count', countMembers);
    api.put('/api/v2/dialer/:dialer/members/reset', resetMembers);
    api.post('/api/v2/dialer/:dialer/members/aggregate', aggregateMembers);
    api.post('/api/v2/dialer/:dialer/members', createMember);
    api.get('/api/v2/dialer/:dialer/members/:id', itemMember);
    api.delete('/api/v2/dialer/:dialer/members/:id', removeMember);
    api.delete('/api/v2/dialer/:dialer/members', removeMembers);
    api.put('/api/v2/dialer/:dialer/members/:id', updateMember);
    api.put('/api/v2/dialer/:dialer/members/:id/status', setStatusMember);

    api.get('/api/v2/dialer/:dialer/history', listHistory);

    api.post('/api/v2/dialer/:dialer/agents/stats', getAgentStats);

    api.get('/api/v2/dialer/:dialer/templates', listTemplates);
    api.get('/api/v2/dialer/:dialer/templates/:id', itemTemplate);
    api.post('/api/v2/dialer/:dialer/templates', createTemplate);
    api.put('/api/v2/dialer/:dialer/templates/:id', updateTemplate);
    api.delete('/api/v2/dialer/:dialer/templates/:id', removeTemplate);
}


function list (req, res, next) {
    let options = {
        limit: req.query.limit,
        pageNumber: req.query.page,
        domain: req.query.domain,
        columns: {}
    };

    if (req.query.columns)
        req.query.columns.split(',')
            .forEach( (i) => options.columns[i] = 1 );

    dialerService.list(req.webitelUser, options, (err, result) => {
        if (err)
            return next(err);

        return res.status(200).json({
            "status": "OK",
            "data": result
        });
    })
}

function item (req, res, next) {
    let options = {
        id: req.params.id,
        domain: req.query.domain
    };

    dialerService.item(req.webitelUser, options, (err, result) => {
        if (err)
            return next(err);

        return res.status(200).json({
            "status": "OK",
            "data": result
        });
    });
}

function remove (req, res, next) {
    let options = {
        id: req.params.id,
        domain: req.query.domain
    };

    dialerService.remove(req.webitelUser, options, (err, result) => {
        if (err)
            return next(err);

        return res.status(200).json({
            "status": "OK",
            "data": result
        });
    });
}

function update (req, res, next) {
    let options = {
        id: req.params.id,
        domain: req.query.domain,
        data: req.body
    };

    dialerService.update(req.webitelUser, options, (err, result) => {
        if (err)
            return next(err);

        return res.status(200).json({
            "status": "OK",
            "data": result
        });
    });
}

function setState (req, res, next) {
    let options = {
        id: req.params.id,
        state: +req.params.state,
        domain: req.query.domain
    };

    dialerService.setState(req.webitelUser, options, (err, result) => {
        if (err)
            return next(err);

        return res.status(200).json({
            "status": "OK",
            "data": result
        });
    });
}

function create (req, res, next) {
    let options = req.body;

    if (req.query.domain)
        options.domain = req.query.domain;

    dialerService.create(req.webitelUser, options, (err, result) => {
        if (err)
            return next(err);

        return res.status(200).json({
            "status": "OK",
            "data": result
        });
    });
}

function resetProcess (req, res, next) {
    const options = req.body || {};
    options.id = req.params.id;
    options.domain = req.query.domain;

    dialerService.resetProcessStatistic(req.webitelUser, options, (err, result) => {
        if (err)
            return next(err);

        return res.status(200).json({
            "status": "OK",
            "data": result
        });
    });
}

function listMembers (req, res, next) {
    let options = {
        dialer: req.params.dialer,
        limit: req.query.limit,
        pageNumber: req.query.page,
        domain: req.query.domain,
        columns: parseQueryToObject(req.query.columns),
        sort: parseQueryToObject(req.query.sort),
        filter: parseQueryToObject(req.query.filter)
    };

    dialerService.members.list(req.webitelUser, options, (err, result) => {
        if (err)
            return next(err);

        return res.status(200).json({
            "status": "OK",
            "data": result
        });
    })
}

function countMembers (req, res, next) {
    let options = {
        dialer: req.params.dialer,
        domain: req.query.domain,
        filter: parseQueryToObject(req.query.filter)
    };

    dialerService.members.count(req.webitelUser, options, (err, result) => {
        if (err)
            return next(err);

        return res.status(200).json({
            "status": "OK",
            "data": result
        });
    })
}

function aggregateMembers (req, res, next) {
    let options = {
        dialer: req.params.dialer,
        domain: req.query.domain,
        data: req.body
    };

    dialerService.members.aggregate(req.webitelUser, options, (err, result) => {
        if (err)
            return next(err);

        return res.status(200).json({
            "status": "OK",
            "data": result
        });
    })
}

function itemMember (req, res, next) {
    let options = {
        dialer: req.params.dialer,
        id: req.params.id,
        domain: req.query.domain
    };

    dialerService.members.item(req.webitelUser, options, (err, result) => {
        if (err)
            return next(err);

        return res.status(200).json({
            "status": "OK",
            "data": result
        });
    })
}

function updateMember (req, res, next) {
    let options = {
        id: req.params.id,
        dialer: req.params.dialer,
        domain: req.query.domain,
        data: req.body
    };

    dialerService.members.update(req.webitelUser, options, (err, result) => {
        if (err)
            return next(err);

        return res.status(200).json({
            "status": "OK",
            "data": result
        });
    });
}

function removeMember (req, res, next) {
    let options = {
        id: req.params.id,
        dialer: req.params.dialer,
        domain: req.query.domain
    };

    dialerService.members.remove(req.webitelUser, options, (err, result) => {
        if (err)
            return next(err);

        return res.status(200).json({
            "status": "OK",
            "data": result
        });
    });
}

function removeMembers (req, res, next) {
    let options = {
        filter: req.body,
        dialer: req.params.dialer,
        domain: req.query.domain
    };

    dialerService.members.removeByFilter(req.webitelUser, options, (err, result) => {
        if (err)
            return next(err);

        return res.status(200).json({
            "status": "OK",
            "data": result
        });
    });
}

function createMember (req, res, next) {
    let options = {
        dialer: req.params.dialer,
        domain: req.query.domain,
        data: req.body,
        autoRun: req.query.autoRun
    };

    dialerService.members.create(req.webitelUser, options, (err, result) => {
        if (err)
            return next(err);

        return res.status(200).json({
            "status": "OK",
            "data": result
        });
    });
}

function setStatusMember(req, res, next) {
    let options = {
        dialer: req.params.dialer,
        member: req.params.id,
        domain: req.query.domain,
        callerIp: getIp(req),
        callback: req.body
    };

    dialerService.members.setCallback(req.webitelUser, options, (err, result) => {
        if (err)
            return next(err);

        return res.status(200).json({
            "status": "OK",
            "info": "success"
        });
    });
}

function resetMembers(req, res, next) {
    let options = {
        dialer: req.params.dialer,
        resetLog: req.query._log === 'true',
        domain: req.query.domain,
        fromDate: req.query.from && +req.query.from
    };

    dialerService.members.resetMembers(req.webitelUser, options, (err, count) => {
        if (err)
            return next(err);

        return res.status(200).json({
            "status": "OK",
            "info": count
        });
    });
}

function listHistory(req, res, next) {
    let options = {
        dialer: req.params.dialer,
        limit: req.query.limit,
        pageNumber: req.query.page,
        domain: req.query.domain,
        columns: parseQueryToObject(req.query.columns),
        sort: parseQueryToObject(req.query.sort),
        filter: parseQueryToObject(req.query.filter)
    };

    dialerService.listHistory(req.webitelUser, options, (err, result) => {
        if (err)
            return next(err);

        return res.status(200).json({
            "status": "OK",
            "info": result
        });
    });
}

function getAgentStats(req, res, next) {
    const options = {
        dialer: req.params.dialer,
        agents: req.body.agents,
        skills: req.body.skills,
        domain: req.query.domain
    };

    dialerService.agents.stats(req.webitelUser, options, (err, result) => {
        if (err)
            return next(err);

        return res.status(200).json({
            "status": "OK",
            "data": result
        });
    });
}

function listTemplates(req, res, next) {
    const options = getRequest(req);
    if (!options.filter) {
        options.filter = {};
    }

    options.filter.dialer_id = req.params.dialer;
    delete options.domain;

    dialerService.templates.list(req.webitelUser, options, (err, result) => {
        if (err)
            return next(err);

        return res.status(200).json({
            "status": "OK",
            "data": result
        });
    })
}

function itemTemplate(req, res, next) {
    const options = {
        dialerId: req.params.dialer,
        id: req.params.id
    };

    dialerService.templates.item(req.webitelUser, options, (err, result) => {
        if (err)
            return next(err);

        return res.status(200).json({
            "status": "OK",
            "data": result
        });
    })
}

function createTemplate(req, res, next) {
    const options = req.body;
    options.dialerId = req.params.dialer;

    dialerService.templates.create(req.webitelUser, options, (err, result) => {
        if (err)
            return next(err);

        return res.status(200).json({
            "status": "OK",
            "data": result
        });
    })
}

function updateTemplate(req, res, next) {
    const options = req.body;
    options.dialerId = req.params.dialer;
    options.id = req.params.id;

    dialerService.templates.update(req.webitelUser, options, (err, result) => {
        if (err)
            return next(err);

        return res.status(200).json({
            "status": "OK",
            "data": result
        });
    })
}

function removeTemplate(req, res, next) {
    const options = {
        dialerId: req.params.dialer,
        id: req.params.id
    };

    dialerService.templates.remove(req.webitelUser, options, (err, result) => {
        if (err)
            return next(err);

        return res.status(200).json({
            "status": "OK",
            "data": result
        });
    })
}