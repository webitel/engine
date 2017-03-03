/**
 * Created by igor on 03.03.17.
 */

"use strict";

const log = require(__appRoot + '/lib/log')(module),
    async = require('async'),
    moment = require('moment-timezone')
    ;
    
function checkDialerDeadline(dialerManager, dialerDb, calendarDb, cb) {
    dialerDb._getActiveDialer({calendar: 1, domain: 1, state: 1}, (err, res) => {
        if (err)
            return log.error(err);

        if (res instanceof Array) {
            async.forEachOf(res, (dialer, key, callback) => {
                const calendarId = dialer.calendar && dialer.calendar.id;
                if (!calendarId)
                    return callback();

                calendarDb.findById(dialer.domain, calendarId, (err, calendar) => {
                    if (err) {
                        return callback(err);
                    }

                    const currentTimeOfDay = getCurrentTimeOfDay(calendar);
                    dialerManager.emit('changeDialerState', dialer, calendar, currentTimeOfDay);
                    callback();
                });
            }, cb);
        }
    })
}

function getCurrentTimeOfDay(calendar) {
    let current;

    if (calendar.timeZone && calendar.timeZone.id)
        current = moment().tz(calendar.timeZone.id);
    else current = moment();

    const currentTime = current.valueOf();

    // Check range date;
    if (calendar.startDate && currentTime < calendar.startDate)
        return null;
    else if (calendar.endDate && currentTime > calendar.endDate)
        return null;

    //Check work
    let isAccept = false;
    const currentTimeOfDay = current.get('hours') * 60 + current.get('minutes');

    if (calendar.accept instanceof Array) {
        let currentWeek = current.isoWeekday();

        for (let i = 0, len = calendar.accept.length; i < len; i++) {
            isAccept = currentWeek === calendar.accept[i].weekDay && between(currentTimeOfDay, calendar.accept[i].startTime, calendar.accept[i].endTime);
            if (isAccept)
                break;
        }

    } else {
        return null;
    }

    if (!isAccept)
        return null;

    // Check holiday
    if (calendar.except instanceof Array) {
        const currentDay = current.get('date'),
            currentMonth = current.get('month'),
            currentYear = current.get('year')
            ;

        for (let i = 0, len = calendar.except.length; i < len; i++) {
            const exceptDate = moment(calendar.except[i].date);
            if (exceptDate.get('date') == currentDay && exceptDate.get('month') == currentMonth &&
                (calendar.except[i].repeat === 1 || (calendar.except[i].repeat === 0 && exceptDate.get('year') == currentYear)) )
                return null;
        }
    }

    return currentTimeOfDay;
}

module.exports = {
    checkDialerDeadline: checkDialerDeadline,
    getCurrentTimeOfDay: getCurrentTimeOfDay
};


function between(x, min, max) {
    return x >= min && x <= max;
}