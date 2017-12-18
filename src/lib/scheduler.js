/**
 * Created by igor on 10.03.17.
 */

"use strict";

const parser = require('cron-parser'),
    log = require(`${__appRoot}/lib/log`)(module)
    ;

class Scheduler {
    constructor (cronFormat, fn, params = {}) {
        let _timer = null;
        const interval = parser.parseExpression(cronFormat,  {tz: params.timezone});
        const _c = cronFormat;

        log.info(`Create job: ${fn.name || ''} > ${cronFormat}`);

        this.cancel = () => clearTimeout(_timer);

        (function shed() {
            if (_timer)
                clearTimeout(_timer);

            let n = -1;
            do {
                n = interval.next().getTime() - Date.now()
            } while (n < 0);
            if (params.log)
                log.trace(`Next exec schedule: ${fn.name || ''} ${n}`);

            _timer = setTimeout( function tick() {
                if (params.log)
                    log.trace(`Exec schedule: ${fn.name || ''} ${_c}`);
                fn.apply(null, [shed]);
            }, n);
        })()
    }
}

module.exports = Scheduler;