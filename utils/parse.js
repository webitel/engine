/**
 * Created by Igor Navrotskyj on 07.08.2015.
 */

'use strict';

var CodeError = require(__appRoot + '/lib/error');

var const_DataSeparator = '=================================================================================================';

var Parser = {
    validateEslResponse: function (res, cb) {
        if (typeof res != 'string') {
            return cb(new CodeError(400, "Bad response type"));
        };

        if (res.indexOf('-ERR') > -1 || res.indexOf('-USAGE') > -1) {
            return cb(new CodeError(400, res));
        };

        return cb(null, res);
    },

    plainTableToJSON: function (data, domain, cb) {
        if (!data) {
            cb('Data is undefined!');
            return
        };
        try {
            var _line,
                _head,
                _json = {},
                _id;

            _line = data.split('\n');
            _head = _line[0].split('\t');
            for (var i = 2; i < _line.length && _line[i] != const_DataSeparator; i++) {
                _id = '';
                _line[i].split('\t').reduce(function (_json, line, index) {
                    if (index == 0) {
                        _id = line.trim(); // + '@' + domain;
                        _json[_id] = {
                            id: _id
                        };
                    } else {
                        _json[_id][_head[index].trim()] = line.trim();
                    };
                    return _json;
                }, _json);
            };
            cb(null, _json);
        } catch (e) {
            cb(e);
        };
    },

    plainTableToJSONArray: function (data, cb, _separator) {
        if (!data) {
            cb('Data is undefined!');
            return
        };
        try {
            var _line,
                _head,
                _json = [],
                _item,
                _lineItems,
                _headCounts,
                separator = _separator || '\t'
                ;

            _line = data.split('\n');
            _head = _line[0].split(separator);
            _headCounts = _head.length;
            for (var i = 1; i < _line.length; i++) {
                _lineItems = _line[i].split(separator);
                if (_line[i] == "" || _line[i] == const_DataSeparator || _lineItems.length != _headCounts) continue;
                _item = {};
                _lineItems.reduce(function (_item, line, index) {
                    _item[_head[index].trim()] = line.trim();
                    return _item;
                }, _item);

                _json.push(_item);
            };
            cb(null, _json);
        } catch (e) {
            cb(e);
        };
    },

    plainCollectionToJSON: function (data, cb) {
        if (!data) {
            cb('Data is undefined!');
            return
        };

        try {

            var _json = {},
                lines = data.split('\n'),
                line,
                attribute,
                separatorId;

            for (var i = 0, len = lines.length; i < len; i++) {
                line = lines[i];
                separatorId = line.indexOf('=');
                attribute = line.substring(0, separatorId);
                if (attribute === '')
                    continue;
                _json[attribute] = line.substring(separatorId + 1);
            };

            cb(null, _json);
        } catch (e) {
            cb(e['message']);
        }
    },

    getDomainFromStr: function (str) {
        if (typeof str != 'string')
            return false;
        return str.substring(str.indexOf('@') + 1);
    }
};

module.exports = Parser;