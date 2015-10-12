/**
 * Created by Igor Navrotskyj on 26.08.2015.
 */

'use strict';

var util = require('util');

function CodeError(status, message) {
    this.status = status;
    this.message = message;
    Error.captureStackTrace(this, CodeError);
};
util.inherits(CodeError, Error);
CodeError.prototype.name = "CodeError";

module.exports = CodeError;