/**
 * Created by igor on 26.04.16.
 */

'use strict';


var dialerService = require(__appRoot + '/services/dialer'),
    parseQueryToObject = require(__appRoot + '/utils/parse').parseQueryToObject
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
    const options = {
        id: req.params.id,
        domain: req.query.domain
    };

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