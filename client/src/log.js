var Log = /** @class */ (function () {
    function Log() {
    }
    Log.prototype.debug = function (msg) {
        var supportingDetails = [];
        for (var _i = 1; _i < arguments.length; _i++) {
            supportingDetails[_i - 1] = arguments[_i];
        }
        this.emitLogMessage("debug", msg, supportingDetails);
    };
    Log.prototype.info = function (msg) {
        var supportingDetails = [];
        for (var _i = 1; _i < arguments.length; _i++) {
            supportingDetails[_i - 1] = arguments[_i];
        }
        this.emitLogMessage("info", msg, supportingDetails);
    };
    Log.prototype.warn = function (msg) {
        var supportingDetails = [];
        for (var _i = 1; _i < arguments.length; _i++) {
            supportingDetails[_i - 1] = arguments[_i];
        }
        this.emitLogMessage("warn", msg, supportingDetails);
    };
    Log.prototype.error = function (msg) {
        var supportingDetails = [];
        for (var _i = 1; _i < arguments.length; _i++) {
            supportingDetails[_i - 1] = arguments[_i];
        }
        this.emitLogMessage("error", msg, supportingDetails);
    };
    Log.prototype.emitLogMessage = function (msgType, msg, supportingDetails) {
        if (supportingDetails.length > 0) {
            console[msgType].apply(console, [msg].concat(supportingDetails));
        }
        else {
            console[msgType](msg);
        }
    };
    return Log;
}());
export { Log };
//# sourceMappingURL=log.js.map