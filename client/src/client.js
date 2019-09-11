var __awaiter = (this && this.__awaiter) || function (thisArg, _arguments, P, generator) {
    return new (P || (P = Promise))(function (resolve, reject) {
        function fulfilled(value) { try { step(generator.next(value)); } catch (e) { reject(e); } }
        function rejected(value) { try { step(generator["throw"](value)); } catch (e) { reject(e); } }
        function step(result) { result.done ? resolve(result.value) : new P(function (resolve) { resolve(result.value); }).then(fulfilled, rejected); }
        step((generator = generator.apply(thisArg, _arguments || [])).next());
    });
};
var __generator = (this && this.__generator) || function (thisArg, body) {
    var _ = { label: 0, sent: function() { if (t[0] & 1) throw t[1]; return t[1]; }, trys: [], ops: [] }, f, y, t, g;
    return g = { next: verb(0), "throw": verb(1), "return": verb(2) }, typeof Symbol === "function" && (g[Symbol.iterator] = function() { return this; }), g;
    function verb(n) { return function (v) { return step([n, v]); }; }
    function step(op) {
        if (f) throw new TypeError("Generator is already executing.");
        while (_) try {
            if (f = 1, y && (t = op[0] & 2 ? y["return"] : op[0] ? y["throw"] || ((t = y["return"]) && t.call(y), 0) : y.next) && !(t = t.call(y, op[1])).done) return t;
            if (y = 0, t) op = [op[0] & 2, t.value];
            switch (op[0]) {
                case 0: case 1: t = op; break;
                case 4: _.label++; return { value: op[1], done: false };
                case 5: _.label++; y = op[1]; op = [0]; continue;
                case 7: op = _.ops.pop(); _.trys.pop(); continue;
                default:
                    if (!(t = _.trys, t = t.length > 0 && t[t.length - 1]) && (op[0] === 6 || op[0] === 2)) { _ = 0; continue; }
                    if (op[0] === 3 && (!t || (op[1] > t[0] && op[1] < t[3]))) { _.label = op[1]; break; }
                    if (op[0] === 6 && _.label < t[1]) { _.label = t[1]; t = op; break; }
                    if (t && _.label < t[2]) { _.label = t[2]; _.ops.push(op); break; }
                    if (t[2]) _.ops.pop();
                    _.trys.pop(); continue;
            }
            op = body.call(thisArg, _);
        } catch (e) { op = [6, e]; y = 0; } finally { f = t = 0; }
        if (op[0] & 5) throw op[1]; return { value: op[0] ? op[1] : void 0, done: true };
    }
};
import { Socket } from './socket';
import { Log } from './log';
import { SipPhone } from './sip';
var WEBSOCKET_AUTHENTICATION_CHALLENGE = "authentication_challenge";
export var Response;
(function (Response) {
    // STATUS_FAIL = "FAIL",
    Response["STATUS_OK"] = "OK";
})(Response || (Response = {}));
var Client = /** @class */ (function () {
    function Client(_config) {
        this._config = _config;
        this.req_seq = 0;
        this.queueRequest = new Map();
        this.log = new Log();
        new SipPhone();
    }
    Client.prototype.connect = function () {
        return __awaiter(this, void 0, void 0, function () {
            return __generator(this, function (_a) {
                switch (_a.label) {
                    case 0: return [4 /*yield*/, this.connectToSocket()];
                    case 1:
                        _a.sent();
                        return [2 /*return*/];
                }
            });
        });
    };
    Client.prototype.disconnect = function () {
        return __awaiter(this, void 0, void 0, function () {
            return __generator(this, function (_a) {
                switch (_a.label) {
                    case 0: return [4 /*yield*/, this.socket.close()];
                    case 1:
                        _a.sent();
                        return [2 /*return*/];
                }
            });
        });
    };
    Client.prototype.subscribe = function (action, data) {
        var _this = this;
        return new Promise(function (resolve, reject) {
            _this.queueRequest.set(++_this.req_seq, { resolve: resolve, reject: reject });
            _this.socket.send({
                seq: _this.req_seq,
                action: action,
                data: data
            });
        });
    };
    Client.prototype.auth = function () {
        return this.request(WEBSOCKET_AUTHENTICATION_CHALLENGE, { token: this._config.token });
    };
    Client.prototype.request = function (action, data) {
        var _this = this;
        return new Promise(function (resolve, reject) {
            _this.queueRequest.set(++_this.req_seq, { resolve: resolve, reject: reject });
            _this.socket.send({
                seq: _this.req_seq,
                action: action,
                data: data
            });
        });
    };
    Client.prototype.onMessage = function (message) {
        this.log.debug("receive message: ", message);
        if (message.seq_reply > 0) {
            if (this.queueRequest.has(message.seq_reply)) {
                var promise = this.queueRequest.get(message.seq_reply);
                this.queueRequest.delete(message.seq_reply);
                if (message.status == Response.STATUS_OK) {
                    promise.resolve(message.data);
                }
                else {
                    promise.reject(message.error);
                }
            }
        }
        else {
            // message.data.delete("debug");
            output(syntaxHighlight(JSON.stringify(message, undefined, 4)));
        }
    };
    Client.prototype.connectToSocket = function () {
        var _this = this;
        return new Promise(function (resolve, reject) {
            try {
                _this.socket = new Socket(_this._config.endpoint);
                _this.socket.connect(_this._config.token);
            }
            catch (e) {
                reject(e);
                return;
            }
            _this.socket.on("message", _this.onMessage.bind(_this));
            _this.socket.on("close", function (code) {
                _this.log.error("socket close code: ", code);
            });
            _this.socket.on("open", function () {
                resolve(null);
            });
        });
    };
    return Client;
}());
export { Client };
function syntaxHighlight(json) {
    json = json.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
    return json.replace(/("(\\u[a-zA-Z0-9]{4}|\\[^u]|[^\\"])*"(\s*:)?|\b(true|false|null)\b|-?\d+(?:\.\d*)?(?:[eE][+\-]?\d+)?)/g, function (match) {
        var cls = 'number';
        if (/^"/.test(match)) {
            if (/:$/.test(match)) {
                cls = 'key';
            }
            else {
                cls = 'string';
            }
        }
        else if (/true|false/.test(match)) {
            cls = 'boolean';
        }
        else if (/null/.test(match)) {
            cls = 'null';
        }
        return '<span class="' + cls + '">' + match + '</span>';
    });
}
function output(inp) {
    document.body.appendChild(document.createElement('pre')).innerHTML = inp;
}
//# sourceMappingURL=client.js.map