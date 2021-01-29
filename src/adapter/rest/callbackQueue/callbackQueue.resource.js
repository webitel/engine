/**
 * Created by igor on 05.07.17.
 */

'use strict';

module.exports = {
    addRoutes: addRoutes
};

const parseQueryToObject = require(__appRoot + '/utils/parse').parseQueryToObject,
    callbackService = require(__appRoot + '/services/callback'),
    getRequest = require(__appRoot + '/utils/helper').getRequest,
    getIp = require(__appRoot + '/utils/ip')
;


function addRoutes(api) {
    api.get('/api/v2/callback', list);
    api.get('/api/v2/callback/:id', get);
    api.post('/api/v2/callback', create);
    api.put('/api/v2/callback/:id', update);
    api.delete('/api/v2/callback/:id', del);

    //members
    api.get('/api/v2/callback/:queueId/members', listMembers);
    api.get('/api/v2/callback/:queueId/members/:id', getMember);

    api.post('/callback/members', createMemberPublic);

    api.post('/api/v2/callback/:queueId/members', createMember);
    api.put('/api/v2/callback/:queueId/members/:id', updateMember);
    api.delete('/api/v2/callback/:queueId/members/:id', delMember);

    api.post('/api/v2/callback/:queueId/members/:id/comments', addComment);
    api.delete('/api/v2/callback/:queueId/members/:id/comments/:commentId', removeComment);
    api.put('/api/v2/callback/:queueId/members/:id/comments/:commentId', updateComment);

    api.post('/api/v2/callback/:queueId/members/:id/call', makeCall);
}

function list(req, res, next) {
    if (req.query.view) {
        const options = {
            type: req.query.view,
            domain: req.query.domain,
            filter: req.query.filter,
            limit: parseInt(req.query.limit, 10) || 40,
            offset: 0
        };

        options.offset = (parseInt(req.query.page, 10) || 0) * options.limit;

        return callbackService.members.view(req.webitelUser, options, (err, result) => {
            if (err)
                return next(err);

            return res.status(200).json({
                "status": "OK",
                "data": result
            });
        })
    }

    callbackService.list(req.webitelUser, getRequest(req), (err, result) => {
        if (err)
            return next(err);

        return res.status(200).json({
            "status": "OK",
            "data": result
        });
    })
}

function listMembers(req, res, next) {
    const requestList = getRequest(req);
    requestList.queue = req.params.queueId;
    callbackService.members.list(req.webitelUser, requestList , (err, result) => {
        if (err)
            return next(err);

        return res.status(200).json({
            "status": "OK",
            "data": result
        });
    })
}

function getMember(req, res, next) {
    let options = {
        id: req.params.id,
        domain: req.query.domain,
        queue: req.params.queueId
    };

    callbackService.members.get(req.webitelUser, options, (err, result) => {
        if (err)
            return next(err);

        if (!result)
            return res.status(404).json({
                "status": "error",
                "info": `Not found ${options.id}`
            });

        return res.status(200).json({
            "status": "OK",
            "data": result
        });
    });
}
function updateMember(req, res, next) {
    let options = {
        id: req.params.id,
        domain: req.query.domain,
        queue: req.params.queueId,
        data: req.body
    };

    callbackService.members.update(req.webitelUser, options, (err, result) => {
        if (err)
            return next(err);

        return res.status(200).json({
            "status": "OK",
            "data": result
        });
    });
}
function delMember(req, res, next) {
    let options = {
        id: req.params.id,
        queue: req.params.queueId,
        domain: req.query.domain
    };

    callbackService.members.remove(req.webitelUser, options, (err, result) => {
        if (err)
            return next(err);

        return res.status(200).json({
            "status": "OK",
            "data": result
        });
    });
}

function createMemberPublic(req, res, next) {
    const options = req.body;

    if (req.query['domain'])
        options.domain = req.query['domain'];

    let relip = getIp(req) || '';
    const cflaer = relip.indexOf(',');
    if (cflaer > 0) {
        relip = relip.substr(0, cflaer)
    }

    callbackService.members.createPublic({ip: relip, widget: req.query.widget}, options, (err, result) => {
        if (err) {
            return next(err);
        }

        return res
            .status(200)
            .json({
                "status": "OK",
                "data": result,
            });
    });
}

function createMember(req, res, next) {
    const options = req.body;

    if (req.query['domain'])
        options.domain = req.query['domain'];

    options.queue = req.params.queueId;

    callbackService.members.create(req.webitelUser, options, (err, result) => {
        if (err) {
            return next(err);
        }

        return res
            .status(200)
            .json({
                "status": "OK",
                "data": result,
            });
    });
}
function addComment(req, res, next) {
    const options = {
        data: req.body,
        queue: req.params.queueId,
        id: req.params.id,
        domain: req.query['domain']
    };

    callbackService.members.createComment(req.webitelUser, options, (err, result) => {
        if (err) {
            return next(err);
        }

        return res
            .status(200)
            .json({
                "status": "OK",
                "data": result,
            });
    });
}
function removeComment(req, res, next) {
    const options = {
        commentId: req.params.commentId,
        queue: req.params.queueId,
        id: req.params.id,
        domain: req.query['domain']
    };

    callbackService.members.removeComment(req.webitelUser, options, (err, result) => {
        if (err) {
            return next(err);
        }

        return res
            .status(200)
            .json({
                "status": "OK",
                "data": result,
            });
    });
}
function updateComment(req, res, next) {
    const options = {
        commentId: req.params.commentId,
        queue: req.params.queueId,
        text: req.body.text,
        id: req.params.id,
        domain: req.query['domain']
    };

    callbackService.members.updateComment(req.webitelUser, options, (err, result) => {
        if (err) {
            return next(err);
        }

        return res
            .status(200)
            .json({
                "status": "OK",
                "data": result,
            });
    });
}

function get(req, res, next) {
    let options = {
        id: req.params.id,
        domain: req.query.domain
    };

    callbackService.get(req.webitelUser, options, (err, result) => {
        if (err)
            return next(err);

        if (!result)
            return res.status(404).json({
                "status": "error",
                "info": `Not found ${options.id}`
            });

        return res.status(200).json({
            "status": "OK",
            "data": result
        });
    });
}
function create(req, res, next) {
    const options = req.body;

    if (req.query['domain'])
        options.domain = req.query['domain'];

    callbackService.create(req.webitelUser, options, (err, result) => {
        if (err) {
            return next(err);
        }

        return res
            .status(200)
            .json({
                "status": "OK",
                "data": result,
            });
    });
}
function update(req, res, next) {
    let options = {
        id: req.params.id,
        domain: req.query.domain,
        data: req.body
    };

    callbackService.update(req.webitelUser, options, (err, result) => {
        if (err)
            return next(err);

        return res.status(200).json({
            "status": "OK",
            "data": result
        });
    });
}
function del(req, res, next) {
    let options = {
        id: req.params.id,
        domain: req.query.domain
    };

    callbackService.remove(req.webitelUser, options, (err, result) => {
        if (err)
            return next(err);

        return res.status(200).json({
            "status": "OK",
            "data": result
        });
    });
}

function makeCall(req, res, next) {
    let options = {
        id: req.params.id,
        queue: req.params.queueId,
        domain: req.query.domain
    };

    callbackService.members.call(req.webitelUser, options, (err, result) => {
        if (err)
            return next(err);

        return res.status(200).json({
            "status": "OK",
            "data": result
        });
    });
}
