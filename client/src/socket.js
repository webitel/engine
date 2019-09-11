var __extends = (this && this.__extends) || (function () {
    var extendStatics = function (d, b) {
        extendStatics = Object.setPrototypeOf ||
            ({ __proto__: [] } instanceof Array && function (d, b) { d.__proto__ = b; }) ||
            function (d, b) { for (var p in b) if (b.hasOwnProperty(p)) d[p] = b[p]; };
        return extendStatics(d, b);
    };
    return function (d, b) {
        extendStatics(d, b);
        function __() { this.constructor = d; }
        d.prototype = b === null ? Object.create(b) : (__.prototype = b.prototype, new __());
    };
})();
import { formatWebSocketUri } from './utils';
import EventEmitter from './event_emitter';
var SOCKET_URL_SUFFIX = 'websocket';
var Socket = /** @class */ (function (_super) {
    __extends(Socket, _super);
    function Socket(host) {
        var _this = _super.call(this) || this;
        _this.host = host;
        return _this;
    }
    Socket.prototype.connect = function (token) {
        var _this = this;
        this.socket = new WebSocket(formatWebSocketUri(this.host) + "/" + SOCKET_URL_SUFFIX + "?token=" + token);
        this.socket.onclose = function (e) { return _this.onClose(e.code); };
        this.socket.onmessage = function (e) { return _this.onMessage(e.data); };
        this.socket.onopen = function () { return _this.onOpen(); };
    };
    Socket.prototype.send = function (request) {
        this.socket.send(JSON.stringify(request));
        return null;
    };
    Socket.prototype.close = function (code) {
        this.socket.close(code);
        this.socket = null;
    };
    Socket.prototype.onOpen = function () {
        this.emit("open", this);
    };
    Socket.prototype.onClose = function (code) {
        this.emit("close", code);
    };
    Socket.prototype.onMessage = function (data) {
        var message = JSON.parse(data);
        console.log(JSON.stringify(message, null, '\t'));
        this.emit("message", message);
    };
    return Socket;
}(EventEmitter));
export { Socket };
//# sourceMappingURL=socket.js.map