/**
 * Created by Admin on 04.08.2015.
 */
'use strict';

var log = require(__appRoot + '/lib/log')(module);

module.exports = function (ws) {
    try {
        var userId = ws['upgradeReq']['webitelId'],
            user = application.Users.get(userId);
        if (user) {
            for (var key in user.ws) {
                if (user.ws[key].readyState == user.ws[key].CLOSED) {
                    user.ws.splice(key, 1);
                    if (user.ws.length == 0) {
                        application.Users.remove(user.id);
                        log.trace('disconnect: ', user.id);
                        log.debug('Users session: ', application.Users.length());
                    }
                    ;
                };
            };
        };
    } catch (e) {
        log.error(e);
    };
};