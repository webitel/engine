/**
 * Created by Igor Navrotskyj on 06.08.2015.
 */

'use strict';

var WebitelCommandTypes = require(__appRoot + '/const').WebitelCommandTypes,
    getCommandResponseJSON = require('./responseTemplate').getCommandResponseJSON,
    callService = require(__appRoot + '/services/channel'),
    chatService = require(__appRoot + '/services/chat'),
    log = require(__appRoot +  '/lib/log')(module);


module.exports = callsCtrl();

function callsCtrl () {
    var controller = {};
    controller[WebitelCommandTypes.Call.name] = call;
    controller[WebitelCommandTypes.Hangup.name] = hangup;
    controller[WebitelCommandTypes.AttendedTransfer.name] = attendedTransfer;
    controller[WebitelCommandTypes.Transfer.name] = transfer;
    controller[WebitelCommandTypes.Bridge.name] = bridge;
    controller[WebitelCommandTypes.VideoRefresh.name] = videoRefresh;
    controller[WebitelCommandTypes.ToggleHold.name] = toggleHold;
    controller[WebitelCommandTypes.Hold.name] = hold;
    controller[WebitelCommandTypes.UnHold.name] = unHold;
    controller[WebitelCommandTypes.Dtmf.name] = dtmf;
    controller[WebitelCommandTypes.Broadcast.name] = broadcast;
    controller[WebitelCommandTypes.AttXfer.name] = attXfer;
    controller[WebitelCommandTypes.AttXfer2.name] = attXfer2;
    controller[WebitelCommandTypes.AttXferBridge.name] = attXferBridge;
    controller[WebitelCommandTypes.AttXferCancel.name] = attXferCancel;
    controller[WebitelCommandTypes.Dump.name] = dump;
    controller[WebitelCommandTypes.GetVar.name] = getVar;
    controller[WebitelCommandTypes.SetVar.name] = setVar;
    controller[WebitelCommandTypes.Eavesdrop.name] = eavesdrop;
    controller[WebitelCommandTypes.Displace.name] = displace;
    controller[WebitelCommandTypes.Show.Channel.name] = showChannels;
    controller[WebitelCommandTypes.Chat.Send.name] = sendChat;
    return controller;
};

function call (caller, execId, args, ws) {
    callService.makeCall(caller, args, function (err, res) {
        return getCommandResponseJSON(ws, execId, res);
    });
};

function hangup (caller, execId, args, ws) {
    callService.hangup(caller, args, function (err, res) {
        return getCommandResponseJSON(ws, execId, res);
    });
};

function attendedTransfer (caller, execId, args, ws) {
    callService.attendedTransfer(caller, args, function (err, res) {
        return getCommandResponseJSON(ws, execId, res);
    });
};

function transfer (caller, execId, args, ws) {
    callService.transfer(caller, args, function (err, res) {
        return getCommandResponseJSON(ws, execId, res);
    });
};

function bridge (caller, execId, args, ws) {
    callService.bridge(caller, args, function (err, res) {
        return getCommandResponseJSON(ws, execId, res);
    });
};

function videoRefresh (caller, execId, args, ws) {
    callService.videoRefresh(caller, args, function (err, res) {
        return getCommandResponseJSON(ws, execId, res);
    });
};

function toggleHold (caller, execId, args, ws) {
    callService.toggleHold(caller, args, function (err, res) {
        return getCommandResponseJSON(ws, execId, res);
    });
};

function hold (caller, execId, args, ws) {
    callService.hold(caller, args, function (err, res) {
        return getCommandResponseJSON(ws, execId, res);
    });
};

function unHold (caller, execId, args, ws) {
    callService.unHold(caller, args, function (err, res) {
        return getCommandResponseJSON(ws, execId, res);
    });
};

function dtmf (caller, execId, args, ws) {
    callService.dtmf(caller, args, function (err, res) {
        return getCommandResponseJSON(ws, execId, res);
    });
};

function broadcast (caller, execId, args, ws) {
    callService.broadcast(caller, args, function (err, res) {
        return getCommandResponseJSON(ws, execId, res);
    });
};

function attXfer (caller, execId, args, ws) {
    callService.attXfer(caller, args, function (err, res) {
        return getCommandResponseJSON(ws, execId, res);
    });
};

function attXfer2 (caller, execId, args, ws) {
    callService.attXfer2(caller, args, function (err, res) {
        return getCommandResponseJSON(ws, execId, res);
    });
};

function attXferBridge (caller, execId, args, ws) {
    callService.attXferBridge(caller, args, function (err, res) {
        return getCommandResponseJSON(ws, execId, res);
    });
};

function attXferCancel (caller, execId, args, ws) {
    callService.attXferCancel(caller, args, function (err, res) {
        return getCommandResponseJSON(ws, execId, res);
    });
};

function dump (caller, execId, args, ws) {
    callService.dump(caller, args, function (err, res) {
        return getCommandResponseJSON(ws, execId, res);
    });
};

function getVar (caller, execId, args, ws) {
    callService.getVar(caller, args, function (err, res) {
        return getCommandResponseJSON(ws, execId, res);
    });
};

function setVar (caller, execId, args, ws) {
    callService.setVar(caller, args, function (err, res) {
        return getCommandResponseJSON(ws, execId, res);
    });
};

function eavesdrop (caller, execId, args, ws) {
    callService.eavesdrop(caller, args, function (err, res) {
        return getCommandResponseJSON(ws, execId, res);
    });
};

function displace (caller, execId, args, ws) {
    callService.displace(caller, args, function (err, res) {
        return getCommandResponseJSON(ws, execId, res);
    });
};

function showChannels (caller, execId, args, ws) {
    var option = {
        "domain": args['domain']
    };

    callService.channelsList(caller, option, function (err, res) {
        return getCommandResponseJSON(ws, execId, res);
    });
};

function sendChat (caller, execId, args, ws) {
    chatService.send(caller, args, function (err, res) {
        return getCommandResponseJSON(ws, execId, res);
    });
};