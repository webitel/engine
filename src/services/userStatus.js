/**
 * Created by Igor Navrotskyj on 16.09.2015.
 */

'use strict';

const CodeError = require(__appRoot + '/lib/error'),
    log = require(__appRoot + '/lib/log')(module),
    conf = require(__appRoot + '/conf')
    ;

const noWriteStatus = `${conf.get('application:writeUserStatus')}` !== 'true';


if (noWriteStatus) {
    log.warn(`Disabled application:writeUserStatus`)
} else {
    log.debug(`Enabled application:writeUserStatus`)
}

const Service = {
    insert: function (data) {
        if (noWriteStatus) return;


        if (!data['Account-User-State'] || !data['Account-Status'] || !data['Account-User']) {
            return log.warn('Caller %s status or state undefined.', data['Account-User']);
        }

        if (data['Account-Agent-Status']) {
            data['Account-Agent-Status'] = decodeURIComponent(data['Account-Agent-Status']);
            data['cc'] = data['Account-Agent-Status'] === "Available" ||
                data['Account-Agent-Status'] === "Available (On Demand)" ||
                data['Account-Agent-Status'] === "On Break";
        } else {
            data['cc'] = false;
        }

        const now = Date.now();

        application.PG.getQuery('agents').setUserStats(data, (err, res) => {
            if (err) {
                return log.error(err);
            }

            if (!res) {
                return;
            }

            application.Broker.publish(application.Broker.Exchange.STORAGE_COMMANDS,
                `log.user.${application.Broker.encodeRK(data['Account-User'])}.${application.Broker.encodeRK(data['Account-Domain'])}.status`,
                {
                    "presence_id": data['presence_id'],
                    "domain": data['Account-Domain'],
                    "extension": data['Account-User'],
                    "account": data['Account-User-Name'] ? decodeURIComponent(data['Account-User-Name']) : "", //
                    "display_status": getDisplayStatus(res['description'], res['state'], res['status']),
                    "status": res['status'],
                    "state": res['state'],
                    "description": res['description'],
                    "ws": res['ws'],
                    "cc": res['cc'],
                    "created_time": +res['updated_at'],
                    "end_time": now,
                    "duration": Math.round((now - +res['updated_at']) / 1000)
                }, e => {
                    if (e)
                        return log.error(e)
                }
            );

            //TODO
            if (data['cc'] !== res['cc']) {
                application.Broker.publish(application.Broker.Exchange.STORAGE_COMMANDS,
                    `log.user.${application.Broker.encodeRK(data['Account-User'])}.${application.Broker.encodeRK(data['Account-Domain'])}.status`,
                    {
                        "presence_id": data['presence_id'],
                        "domain": data['Account-Domain'],
                        "extension": data['Account-User'],
                        "account": data['Account-User-Name'] ? decodeURIComponent(data['Account-User-Name']) : "", //
                        "display_status": data['cc'] ? "Logged In" : "Logged Out",
                        "ws": res['ws'],
                        "cc": res['cc'],
                        "created_time": +res['updated_at']
                    }, e => {
                        if (e)
                            return log.error(e)
                    }
                );
            }

        });
    },
    
    _removeByUserId: function (userId, domain, cb) {
        if (!domain || !userId) {
            return cb(new CodeError(400, "Domain is required."));
        }

        const dbUserStatus = application.DB._query.userStatus;
        return dbUserStatus._removeByUserId(domain, userId, cb);
    }
};

module.exports = Service;


function getDisplayStatus(description, state, status) {
    if (description)
        return description;
    return status === "NONE" ? state : status;
}