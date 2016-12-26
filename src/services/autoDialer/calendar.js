/**
 * Created by igor on 25.05.16.
 */

let dynamicSort = require(__appRoot + '/utils/sort').dynamicSort,
    moment = require('moment-timezone')
    ;

const CALENDAR_TYPE_REAPET = {
    NEVER: 0,
    EACH_YEAR: 1
};

module.exports = class Calendar {

    constructor (conf, communications) {
        this._calendar = conf;

        this.expire = false;
        this.sleepTime = 0;
        this.deadLineTime = 0;
        this._timezone = conf.timeZone && conf.timeZone.id;

        this._timer = null;
        this._currentTimeOfDay = 0;
        this._initTimer();

        this._currentCommunicationsRanges = [];
        this._communications = [];
        this._communicationsTypeCodes = [];
        this._initCommunications(communications);

        this._calendarMap = {
            deadLineTime: 0,
            exceptDates: [],
            currentDate: null

        };
        this.reCalc();
    }

    _initCommunications (communications) {
        if (communications && communications.types instanceof Array) {
            this._communications = communications.types
                .sort( (i,p) => {
                    return i.priority > p.priority
                })
                .filter( i => {
                    if (!(i.ranges instanceof Array)) {
                        return false
                    }

                    this._communicationsTypeCodes.push(i.code);
                    return true;
                })
            ;
            this._reCalcCommunications();
        }
    }

    _reCalcCommunications () {
        this._currentCommunicationsRanges = [];
        this._communications.forEach( communication => {

            for (let range of communication.ranges) {
                if (this._currentTimeOfDay >= range.startTime && this._currentTimeOfDay < range.endTime) {
                    this._currentCommunicationsRanges.push({
                        code: communication.code,
                        priority: communication.priority,
                        range: range,
                        rangeId: `${new Date().getDate()}_${range.startTime}_${range.endTime}`
                    });
                    break;
                }

            }
        });
        console.log(this.getCurrentTimeOfDay(), this.getCommunicationCodes());
    }

    getCommunicationCodes () {
        return this._currentCommunicationsRanges.map( i => i.code)
    }

    checkCommunicationInAllCode (code) {
        return !!~this._communicationsTypeCodes.indexOf(code)
    }

    getCommunicationByPosition (position) {
        return this._currentCommunicationsRanges[position]
    }

    getCurrentTimeOfDay () {
        return this._currentTimeOfDay;
    }

    _initTimer () {
        this._currentTimeOfDay = this.calcCurrentTimeOfDay();
        this._timer = setInterval(() => {
            if (++this._currentTimeOfDay > 1440)
                this._currentTimeOfDay = 1;

            this._reCalcCommunications();
        }, 60000);
    }

    stop () {
        clearInterval(this._timer);
    }

    calcCurrentTimeOfDay () {
        let current;

        if (this._timezone)
            current = moment().tz(this._timezone);
        else current = moment();

        return current.get('hours') * 60 + current.get('minutes')
    }

    reCalc () {
        this._calendarMap = {
            deadLineTime: 0,
            exceptDates: []
        };
        let calendar = this._calendar;


        if (calendar && calendar.accept instanceof Array) {
            let sort = calendar.accept.sort(dynamicSort('weekDay'));
            let getValue = function (v, last) {
                return {
                    startTime: v.startTime,
                    endTime: v.endTime,
                    next: last
                };
            };

            for (let i = 0, len = sort.length, last = i !== len - 1; i < len; i++) {
                if (this._calendarMap[sort[i].weekDay]) {
                    this._calendarMap[sort[i].weekDay].push(getValue(sort[i], last));
                    this._calendarMap[sort[i].weekDay].sort(dynamicSort('startTime'));
                } else {
                    this._calendarMap[sort[i].weekDay] = [getValue(sort[i], last)];
                }
            }
        } else {
            this.expire = true;
            return;
        }

        if (calendar.except instanceof Array) {
            for (let except of calendar.except) {
                let exceptDate = new Date(except.date);
                this._calendarMap.exceptDates.push(
                    {
                        year: exceptDate.getFullYear(),
                        date: exceptDate.getDate(),
                        month: exceptDate.getMonth(),
                        repeat: except.repeat
                    }
                )
            }
        }

        let current;
        if (calendar.timeZone && calendar.timeZone.id)
            current = moment().tz(calendar.timeZone.id);
        else current = moment();

        let currentTime = current.valueOf();

        // Check range date;
        if (calendar.startDate && currentTime < calendar.startDate) {
            this.sleepTime = new Date(calendar.startDate).getTime() - Date.now() + 1;
            return;
        } else if (calendar.endDate && calendar && currentTime > calendar.endDate) {
            this.expire = true;
            return
        }

        let currentWeek = current.isoWeekday();
        let currentTimeOfDay = current.get('hours') * 60 + current.get('minutes');

        let deadLineRes = getDeadlineMinuteFromSortMap(currentTimeOfDay, currentWeek, this._calendarMap);

        if (deadLineRes.active) {
            this.deadLineTime = (deadLineRes.minute * 60 * 1000) + Date.now();
            return;
        } else {
            this.sleepTime = deadLineRes.minute * 60 * 1000;
            return;
        }
    }
};


function getDeadlineMinuteFromSortMap (currentMinuteOfDay, currentWeek, map) {
    // TODO

    let i = parseInt(currentWeek),
        count = 0,
        result = {active: false, minute: null},
        offsetDay = 0
        ;

    let date = new Date();

    let isExcept = (date) => {
        for (let except of map.exceptDates) {
            if (
                except.date == date.getDate() &&
                except.month == date.getMonth() &&
                (except.repeat == CALENDAR_TYPE_REAPET.EACH_YEAR || (except.year == date.getFullYear()) )
            ) {
                return true
            }
        }
        return false;
    };

    while (1) {
        i = (i > 7) ? 1 : i;

        if (map[i] instanceof Array && !isExcept(date)) {
            for (let item of map[i]) {
                if (count === 0 && item.endTime > currentMinuteOfDay) {
                    if (item.startTime > currentMinuteOfDay) {
                        result.minute = item.startTime - currentMinuteOfDay;
                        return result;
                    } else {
                        result.minute = item.endTime - currentMinuteOfDay;
                        result.active = true;
                        return result;
                    }
                }

                if (count === 0 && item.endTime <= currentMinuteOfDay && item.startTime >= currentMinuteOfDay) {
                    break;
                }

                if (count !== 0) {
                    result.minute = offsetDay += item.startTime;
                    return result;
                }
            }
        }
        offsetDay += (count == 0 ? 1440 - currentMinuteOfDay : 1440);
        i++;
        count++;
        date.setDate(date.getDate() + 1);
        console.log(date.toString())
    }
}