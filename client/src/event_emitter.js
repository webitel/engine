function assert(condition, message) {
    if (!condition)
        throw new TypeError(message);
}
function assertEvent(event) {
    assert(typeof event === "string" || typeof event === "symbol", "'event' must be an string or symbol");
}
function assertListener(listener) {
    assert(typeof listener === "function", "'listener' must be a function");
}
var Item = /** @class */ (function () {
    function Item(listener, context, once) {
        this.listener = listener;
        this.context = context;
        this.once = once;
    }
    return Item;
}());
export { Item };
function addListener(emitter, event, listener, context, prepend, once) {
    if (context === void 0) { context = emitter; }
    assertEvent(event);
    assertListener(listener);
    var listeners = emitter._listeners;
    listener = new Item(listener, context, once);
    if (listeners[event]) {
        if (prepend)
            listeners[event].unshift(listener);
        else
            listeners[event].push(listener);
    }
    else
        listeners[event] = [listener];
    return emitter;
}
// tslint:disable-next-line:max-classes-per-file
var EventEmitter = /** @class */ (function () {
    function EventEmitter() {
        this._listeners = {};
        this.on = this.addListener;
        this.off = this.removeListener;
    }
    EventEmitter.prototype.eventNames = function () {
        var listeners = this._listeners;
        return Object.keys(listeners).filter(function (event) { return listeners[event] !== undefined; });
    };
    EventEmitter.prototype.rawListeners = function (event) {
        assertEvent(event);
        return this._listeners[event] || [];
    };
    EventEmitter.prototype.listeners = function (event) {
        var listeners = this.rawListeners(event);
        var length = listeners.length;
        if (!length)
            return [];
        var ret = new Array(length);
        for (var i = 0; i < length; i++) {
            ret[i] = listeners[i].listener;
        }
        return ret;
    };
    EventEmitter.prototype.listenerCount = function (event) {
        return this.rawListeners(event).length;
    };
    EventEmitter.prototype.emit = function (event) {
        var args = [];
        for (var _i = 1; _i < arguments.length; _i++) {
            args[_i - 1] = arguments[_i];
        }
        var listeners = this.rawListeners(event);
        var length = listeners.length;
        if (!length) {
            if (event === "error")
                throw args[0];
            return false;
        }
        for (var i = 0; i < length; i++) {
            var _a = listeners[i], listener = _a.listener, context = _a.context, once = _a.once;
            if (once)
                this.removeListener(event, listener);
            listener.apply(context, args);
        }
        return true;
    };
    EventEmitter.prototype.addListener = function (event, listener, context) {
        return addListener(this, event, listener, context, false, false);
    };
    EventEmitter.prototype.once = function (event, listener, context) {
        return addListener(this, event, listener, context, false, true);
    };
    EventEmitter.prototype.prependListener = function (event, listener, context) {
        return addListener(this, event, listener, context, true, false);
    };
    EventEmitter.prototype.prependOnceListener = function (event, listener, context) {
        return addListener(this, event, listener, context, true, true);
    };
    EventEmitter.prototype.removeAllListeners = function (event) {
        assert(event === undefined ||
            typeof event === "string" ||
            typeof event === "symbol", "'event' must be an string, symbol or undefined");
        if (event === undefined) {
            this._listeners = {};
            return this;
        }
        if (!this._listeners[event])
            return this;
        this._listeners[event] = undefined;
        return this;
    };
    EventEmitter.prototype.removeListener = function (event, listener) {
        assertListener(listener);
        var listeners = this.rawListeners(event);
        var length = listeners.length;
        if (!length)
            return this;
        if (length === 1) {
            if (listener !== listeners[0].listener)
                return this;
            this._listeners[event] = undefined;
            return this;
        }
        listeners = listeners.slice(0);
        var index = length - 1;
        for (; index >= 0; index--) {
            if (listeners[index].listener === listener)
                break;
        }
        if (index < 0)
            return this;
        if (index === 0)
            listeners.shift();
        else
            listeners.splice(index, 1);
        this._listeners[event] = listeners;
        return this;
    };
    return EventEmitter;
}());
export default EventEmitter;
//# sourceMappingURL=event_emitter.js.map