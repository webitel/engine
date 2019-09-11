var Call = /** @class */ (function () {
    function Call(client) {
        this.client = client;
    }
    /* Call control */
    Call.prototype.hangup = function () { };
    Call.prototype.answer = function () { };
    Call.prototype.hold = function () { };
    Call.prototype.unHold = function () { };
    Call.prototype.toggleCall = function () { };
    Call.prototype.sendDTMF = function () { };
    Call.prototype.transfer = function () { };
    return Call;
}());
export { Call };
//# sourceMappingURL=call.js.map