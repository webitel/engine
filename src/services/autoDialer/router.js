/**
 * Created by igor on 25.05.16.
 */

let log = require(__appRoot + '/lib/log')(module),
    END_CAUSE = require('./const').END_CAUSE,
    Gw = require('./gw')
    ;

module.exports = class Router {

    constructor (resources, variables) {

        this._resourcePaterns = [];
        this._lockedGateways = [];
        this._limit = 0;
        this._variables = variables;

        if (resources instanceof Array) {

            var maxLimitGw = 0;
            resources.forEach((resource) => {
                try {
                    if (typeof resource.dialedNumber != 'string' || !(resource.destinations instanceof Array))
                        return;

                    let flags = resource.dialedNumber.match(new RegExp('^/(.*?)/([gimy]*)$'));
                    if (!flags)
                        flags = [null, resource.dialedNumber];

                    let regex = new RegExp(flags[1], flags[2]);
                    let gws = [];

                    resource.destinations.forEach( (i) => {
                        if (i.enabled !== true)
                            return;

                        // Check limit gw;
                        if (maxLimitGw !== -1)
                            if (i.limit === 0) {
                                maxLimitGw = -1;
                            } else {
                                maxLimitGw += i.limit
                            }

                        gws.push(new Gw(i, regex, this._variables));
                    });

                    this._resourcePaterns.push(
                        {
                            regexp: regex,
                            gws: gws
                        }
                    )
                } catch (e) {
                    log.error(e);
                }
            });

            this._limit = maxLimitGw;
        }
    }

    getDialStringFromMember (member) {
        let res = {
            found: false,
            dialString: false,
            cause: null,
            patternIndex: null,
            gw: null
        };

        for (let i = 0, len = this._resourcePaterns.length; i < len; i++) {
            if (this._resourcePaterns[i].regexp.test(member.number)) {
                res.found = true;
                for (let j = 0, lenGws = this._resourcePaterns[i].gws.length; j < lenGws; j++) {
                    let gatewayPositionMap = i + '>' + j;
                    // TODO...
                    if (member._currentNumber instanceof Object)
                        member._currentNumber.gatewayPositionMap = gatewayPositionMap;

                    member.setVariable('gatewayPositionMap', gatewayPositionMap);
                    if (~this._lockedGateways.indexOf(gatewayPositionMap))
                        continue; // Next gw check

                    res.dialString = this._resourcePaterns[i].gws[j].tryLock(member);
                    if (res.dialString) {
                        res.patternIndex = i; // Ok gw
                        res.gw = j;
                        break
                    } else {
                        this._lockedGateways.push(gatewayPositionMap) // Bad gw
                    }
                }
            }
        }
        if (!res.found)
            res.cause = END_CAUSE.NO_ROUTE;

        return res;
    }

    freeGateway (gw) {
        let gateway = this._resourcePaterns[gw.patternIndex].gws[gw.gw],
            gatewayPositionMap = gw.patternIndex + '>' + gw.gw;
        // Free
        if (gateway.unLock() && ~this._lockedGateways.indexOf(gatewayPositionMap))
            this._lockedGateways.splice(this._lockedGateways.indexOf(gatewayPositionMap), 1)

    }
}