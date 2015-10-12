/**
 * Created by Igor Navrotskyj on 27.08.2015.
 */

'use strict';

var keywords = 'function|case|if|return|new|switch|var|this|typeof|for|while|break|do|continue';

function push(arr, e) {
    arr.push(e);
    return arr.length - 1;
};

module.exports = function (expression) {
    var all = [];

    expression = expression
        // GLOBAL
        .replace(/\$\$\{([\s\S]*?)\}/gi, function (a, b) {
            return 'sys.getGlbVar(\"' + b + '\")';
        })
        // ChannelVar
        .replace(/\$\{([\s\S]*?)\}/gi, function (a, b) {
            return 'sys.getChnVar(\"' + b + '\")';
        })
        .replace(/(\/(\\\/|[^\/\n])*\/[gim]{0,3})|(([^\\])((?:'(?:\\'|[^'])*')|(?:"(?:\\"|[^"])*")))/g, function(m, r, d1, d2, f, s, b, bb)
        {
            if (r != null && r != '') {
                s = r.replace(/\\/g, '\u0001');
                m = '\0B';
            } else {
                s = s;
                m = f + '\0B';
            }
            return m + push(all, s) + '\0';
        })
        .replace(new RegExp('\\b(' + keywords + ')\\b', 'gi'), '')
        // WEBITEL COMMANDS
        .replace(/\&match\(([\s\S]*?)\)/gi, function (f, param) {
            var _params = param.split(',', 2);
            return 'sys.match(\"' + _params[0] + '\", ' + _params[1] + ')';
        })

        // TODO оптимизировать
        .replace(/\&year\(([\s\S]*?)\)/gi, function (f, param) {
            return 'sys.year(\"' + param + '\")';
        })
        .replace(/\&yday\(([\s\S]*?)\)/gi, function (f, param) {
            return 'sys.yday(\"' + param + '\")';
        })
        .replace(/\&mon\(([\s\S]*?)\)/gi, function (f, param) {
            return 'sys.mon(\"' + param + '\")';
        })
        .replace(/\&mday\(([\s\S]*?)\)/gi, function (f, param) {
            return 'sys.mday(\"' + param + '\")';
        })
        .replace(/\&week\(([\s\S]*?)\)/gi, function (f, param) {
            return 'sys.week(\"' + param + '\")';
        })
        .replace(/\&mweek\(([\s\S]*?)\)/gi, function (f, param) {
            return 'sys.mweek(\"' + param + '\")';
        })
        .replace(/\&wday\(([\s\S]*?)\)/gi, function (f, param) {
            return 'sys.wday(\"' + param + '\")';
        })
        .replace(/\&hour\(([\s\S]*?)\)/gi, function (f, param) {
            return 'sys.hour(\"' + param + '\")';
        })
        .replace(/\&minute\(([\s\S]*?)\)/gi, function (f, param) {
            return 'sys.minute(\"' + param + '\")';
        })
        .replace(/\&minute_of_day\(([\s\S]*?)\)/gi, function (f, param) {
            return 'sys.minute_of_day(\"' + param + '\")';
        })
        .replace(/\&time_of_day\(([\s\S]*?)\)/gi, function (f, param) {
            return 'sys.time_of_day(\"' + param + '\")';
        })
        // END COMMANDS
        .replace(/\0B(\d+)\0/g, function(m, i) {
            return all[i];
        });
    return expression;
};