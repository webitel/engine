/**
 * Created by igor on 10.03.17.
 */

"use strict";

const parser = require('cron-parser'),
    log = require(`${__appRoot}/lib/log`)(module)
    ;

class Scheduler {
    constructor(cronFormat, fn, options = {}) {
        this.timerId = null;
        this.name = fn.name ? fn.name : `unknown-${cronFormat}`;
        this.isLog = !!options.log;
        this.interval = parser.parseExpression(cronFormat,  {tz: options.timezone});
        if (typeof fn === 'function') {
            this._fn = fn;
        } else {
            throw `Bad constructor parameters fn`
        }

        log.info(`Create job: ${this.name} -> ${cronFormat}`);
        this.next();
    }

    next (intervalMs) {
        if (!intervalMs)
            intervalMs = this.getNextIntervalMs();

        clearTimeout(this.timerId);

        if (intervalMs > 0x7FFFFFFF) {//setTimeout limit is MAX_INT32=(2^31-1)
            log.debug(`job ${this.name}: sleep (divide segment ${msToHMS(intervalMs)})`);
            this.timerId = setTimeout(() => {
                this.next(intervalMs - 0x7FFFFFFF);
            }, 0x7FFFFFFF);
        } else {
            if (this.isLog) {
                log.debug(`jod ${this.name}: next schedule after ${msToHMS(intervalMs)} ms`);
            }
            this.timerId = setTimeout(() => {
                if (this.isLog) {
                    log.trace(`job ${this.name}: execute`);
                }
                this._fn.apply(null, [() => {
                    this.next()
                }]);
            }, intervalMs);
        }
    }

    getNextIntervalMs () {
        let n = -1;
        let nextJob;
        do {
            nextJob = this.interval.next();
            n = nextJob.getTime() - Date.now()
        } while (n < 0);

        return n;
    }

    cancel() {
        clearTimeout(this.timerId);
        log.debug(`job ${this.name} canceled`);
    }
}

function pad(i) {
    if (i < 10) {
        return `0${i}`
    } else {
        return `${i}`
    }
}

function msToHMS( ms ) {
    let seconds = ms / 1000;
    const hours = parseInt( seconds / 3600 );
    seconds = seconds % 3600;
    const minutes = parseInt( seconds / 60 );
    seconds = seconds % 60;
    return `${pad(hours)}:${pad(minutes)}:${pad(seconds)}`
}

module.exports = Scheduler ;