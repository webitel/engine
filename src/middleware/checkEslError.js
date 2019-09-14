/**
 * Created by Igor Navrotskyj on 03.09.2015.
 */

'use strict';

var CodeError = require(__appRoot + '/lib/error');

module.exports = function (eslResponse) {
    if (!eslResponse)
        return new CodeError(500, "ESL response undefined");

    if (eslResponse && typeof eslResponse['body'] === 'string')
        if (eslResponse['body'].indexOf('-ERR') === 0 || eslResponse['body'].indexOf('-USAGE') === 0)
            return new CodeError(500, eslResponse['body']);

    return false;
};