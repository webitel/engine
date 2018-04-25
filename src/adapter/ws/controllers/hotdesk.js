
const getCommandResponseJSON = require('./responseTemplate').getCommandResponseJSON,
    getCommandResponseJSONError = require('./responseTemplate').getCommandResponseJSONError,
    hotdeskService = require(__appRoot + '/services/hotdesk'),
    log = require(__appRoot +  '/lib/log')(module)
;

module.exports = hotdeskCtrl();

function hotdeskCtrl () {
    return {
        "hotdesk_signin": signIn,
        "hotdesk_signout": signOut
    };
}

function signIn(caller, execId, args = {}, ws) {

    const sessionId = caller.getSession(ws);
    if (caller.checkHotdeskSession(args.address, sessionId)) {
        return getCommandResponseJSONError(ws, execId, new Error("You connected"));
    }

    hotdeskService.signIn(caller, args, (err, res) => {
        if (err)
            return getCommandResponseJSONError(ws, execId, err);

        caller.addHotdeskingSession(args.address, sessionId);
        return getCommandResponseJSON(ws, execId, JSON.stringify({status: 'OK', info: res}));
    })
}

function signOut(caller, execId, args, ws) {
    hotdeskService.signOut(caller, {}, (err, res) => {
        if (err)
            return getCommandResponseJSONError(ws, execId, err);

        return getCommandResponseJSON(ws, execId, JSON.stringify({status: 'OK', result: res}));
    })
}