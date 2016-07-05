/**
 * Created by igor on 26.04.16.
 */

'use strict';


var dialerService = require(__appRoot + '/services/dialer')
    ;

module.exports = {
    addRoutes: addRoutes
};

function addRoutes (api) {
    api.get('/api/v2/dialer', list);
    api.post('/api/v2/dialer', create);
    api.get('/api/v2/dialer/:id', item);
    api.put('/api/v2/dialer/:id', update);
    api.put('/api/v2/dialer/:id/state/:state', setState);
    api.delete('/api/v2/dialer/:id', remove);

    api.get('/api/v2/dialer/:dialer/members', listMembers);
    api.get('/api/v2/dialer/:dialer/members/count', countMembers);
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

function listMembers (req, res, next) {
    let options = {
        dialer: req.params.dialer,
        limit: req.query.limit,
        pageNumber: req.query.page,
        domain: req.query.domain,
        columns: {},
        sort: {},
        filter: {}
    };

    if (req.query.columns)
        req.query.columns.split(',')
            .forEach( (i) => options.columns[i] = 1 );

    if (req.query.sort) {
        let _s = req.query.sort.split('=');
        if (_s.length == 2)
            options.sort[_s[0]] = parseInt(_s[1]);
    }

    // TODO
    if (req.query.filter) {
        let _s = req.query.filter.split(',');
        _s.forEach( (item) => {
            let _f = item.split('=');
            if (_f.length == 2) {
                if (/^\^/.test(_f[1]))
                    options.filter[_f[0]] = {$regex: _f[1]};
                else if (/^true$|^false$/.test(_f[1])) {
                    options.filter[_f[0]] = {$exists: _f[1] === "true"};
                } else options.filter[_f[0]] = isNaN(parseInt(_f[1])) ?_f[1] : parseInt(_f[1]);
            }
        });
    }

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
        filter: {}
    };

    // TODO
    if (req.query.filter) {
        let _s = req.query.filter.split(',');
        _s.forEach( (item) => {
            let _f = item.split('=');
            if (_f.length == 2) {
                if (/^\^/.test(_f[1]))
                    options.filter[_f[0]] = {$regex: _f[1]};
                else options.filter[_f[0]] = isNaN(parseInt(_f[1])) ?_f[1] : parseInt(_f[1]);
            }
        });
    }

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
    console.log(req.body);
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
        data: req.body
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