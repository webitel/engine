/**
 * Created by Igor Navrotskyj on 28.08.2015.
 */

'use strict';

var channelService = require(__appRoot + '/services/channel'),
    CodeError = require(__appRoot + '/lib/error');

module.exports = {
    addRoutes: addRoutes
};

function addRoutes(api) {
    api.get('/api/v2/channels', getChannels);
    api.post('/api/v2/channels', originate);
    api.post('/api/v2/channels/fake', fakeCall);
    api.delete('/api/v2/channels/:id', killUuid);

    // TODO add eavesdrop_whisper_aleg=true and eavesdrop_whisper_bleg=true channel variables to allow you to start eavesdrop in whisper mode of specific call leg
    api.post('/api/v2/channels/:id/eavesdrop', eavesdrop);

    api.delete('/api/v2/channels/domain/:domain', killChannelsFromDomain);

    api.put('/api/v2/channels/:id', changeState);
    api.patch('/api/v2/channels/:id', changeState);

    // V1
    api.post('/api/v1/channels', originateV1);
    api.post('/api/v1/fake', fakeCall);

    api.delete('/api/v1/channels', killChannelsFromDomain);
    api.delete('/api/v1/channels/domain/:domain', killChannelsFromDomain);

    api.delete('/api/v1/channels/:id', killUuid);
    api.put('/api/v1/channels/:id', changeState);
    api.patch('/api/v1/channels/:id', changeState);
};

function originateV1 (req, res, next) {
    var extension = req.body.calledId, // CALLE
        user = req.body.callerId || '', //CALLER
        auto_answer_param = req.body.auto_answer_param;

    var option = {
        "auto_answer_param": auto_answer_param,
        "extension": extension,
        "user": user || req.webitelUser.id
    };

    channelService.makeCall(req.webitelUser, option,
        function (err, result) {
            if (err) {
                return next(err);
            };

            return res
                .status(200)
                .json({
                    "status": "OK",
                    "info": result['body']
                });
        }
    );
};


function getChannels (req, res, next) {
    var option = {
        domain: req.query['domain']
    };
    channelService.channelsList(req.webitelUser, option,
        function (err, result) {
            if (err) {
                return next(err);
            };

            return res
                .status(200)
                .json(result);
        }
    );
};

function originate (req, res, next) {
    var extension = req.body.calledId, // CALLE
        user = req.body.callerId || '', //CALLER
        auto_answer_param = req.body.auto_answer_param;

    var option = {
        "auto_answer_param": auto_answer_param,
        "extension": extension,
        "user": user || req.webitelUser.id
    };

    channelService.makeCall(req.webitelUser, option,
        function (err, result) {
            if (err) {
                return next(err);
            };

            return res
                .status(200)
                .json({
                    "status": "OK",
                    "info": result['body']
                });
        }
    );
};

function fakeCall (req, res, next) {
    var number = req.body.number || '',
        displayNumber = req.body.displayNumber || '00000',
        domainName = number.split('@')[1] || '',
        dialString =  ''.concat('originate ', '{presence_data=@', domainName, '}[sip_h_X-Test=', number.split('@')[0], ',origination_callee_id_number=',displayNumber,
            ',origination_caller_id_number=', displayNumber, ']', 'user/', number,
            ' &bridge(sofia/external/test_terrasoft@switch-d1.webitel.com)');
    ;

    channelService.bgApi(dialString,
        function (err, result) {
            if (err) {
                return next(err);
            };

            return res
                .status(200)
                .json({
                    "status": "OK",
                    "info": result['body']
                });
        }
    );
};

function killUuid (req, res, next) {
    var option = {
        "channel-uuid": req.params['id'],
        "cause": req.query['cause']
    };

    channelService.hangup(req.webitelUser, option,
        function (err, result) {
            if (err) {
                return next(err);
            };

            return res
                .status(200)
                .json({
                    "status": "OK",
                    "info": result['body']
                });
        }
    );
};

function eavesdrop (req, res, next) {
    var option = {
        "user": req.body['user'],
        "channel-uuid": req.params['id'],
        "side": req.query['side']
    };

    channelService.eavesdrop(req.webitelUser, option,
        function (err, result) {
            if (err) {
                return next(err);
            };

            return res
                .status(200)
                .json({
                    "status": "OK",
                    "info": result['body']
                });
        }

    );
};

function killChannelsFromDomain (req, res, next) {
    var option = {
        "domain": req.params['domain'],
        "cause": req.query['cause']
    };

    channelService.hupAll(req.webitelUser, option,
        function (err, result) {
            if (err) {
                return next(err);
            };

            return res
                .status(200)
                .json({
                    "status": "OK",
                    "info": result['body']
                });
        }
    );
};

function changeState (req, res, next) {
    var channelUuid = req.params['id'],
        state = req.body['state'],
        caller = req.webitelUser;

    var cb = function (err, result) {
        if (err) {
            return next(new CodeError(400, err.message));
        };

        return res
            .status(200)
            .json({
                "status": "OK",
                "info": result['body']
            });
    };

    if (channelUuid && state) {
        var option = {
            "channel-uuid": channelUuid
        };
        switch (state) {
            case "hold":
                channelService.hold(caller, option, cb);
                break;
            case  "unhold":
                channelService.unHold(caller, option, cb);
                break;
            default :
                return next(new CodeError(400, 'Bad request.'));
                break;
        };
    } else {
        return next(new CodeError(400, 'Bad request'));
    };
};