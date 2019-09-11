export declare class Item {
    readonly listener: (...args: any[]) => void;
    readonly context: any;
    readonly once: boolean;
    constructor(listener: (...args: any[]) => void, context: any, once: boolean);
}
declare namespace EventEmitter {
    type DefaultListener = (...args: any[]) => void;
    type DefaultEvents = {
        [E in string | symbol | any]: EventEmitter.DefaultListener;
    };
    type Event<Events extends {}> = Extract<keyof Events, string | symbol>;
    type EmitArgs<T> = [T] extends [(...args: infer U) => any] ? U : [T] extends [void] ? [] : [T];
    type Listener<E extends {}, K extends keyof E> = (...args: EmitArgs<E[K]>) => void;
    interface Listeners {
        [event: string]: Item[] | undefined;
    }
}
interface EventEmitter<Events extends {} = EventEmitter.DefaultEvents> {
    on(event: "error", listener: (error: Error) => void, context?: any): this;
    on<K extends EventEmitter.Event<Events>>(event: K, listener: EventEmitter.Listener<Events, K>, context?: any): this;
    off(event: "error", listener: (error: Error) => void): this;
    off<K extends EventEmitter.Event<Events>>(event: K, listener: EventEmitter.Listener<Events, K>): this;
}
declare class EventEmitter<Events extends {} = EventEmitter.DefaultEvents> {
    protected _listeners: EventEmitter.Listeners;
    constructor();
    eventNames(): Array<EventEmitter.Event<Events>>;
    rawListeners(event: "error"): Item[];
    rawListeners<K extends EventEmitter.Event<Events>>(event: K): Item[];
    listeners(event: "error"): EventEmitter.DefaultListener[];
    listeners<K extends EventEmitter.Event<Events>>(event: K): Array<EventEmitter.Listener<Events, K>>;
    listenerCount(event: "error"): number;
    listenerCount<K extends EventEmitter.Event<Events>>(event: K): number;
    emit(event: "error", error: Error): boolean;
    emit<K extends EventEmitter.Event<Events>>(event: K, ...args: EventEmitter.EmitArgs<Events[K]>): boolean;
    addListener(event: "error", listener: (error: Error) => void, context?: any): this;
    addListener<K extends EventEmitter.Event<Events>>(event: K, listener: EventEmitter.Listener<Events, K>, context?: any): this;
    once(event: "error", listener: (error: Error) => void, context?: any): this;
    once<K extends EventEmitter.Event<Events>>(event: K, listener: EventEmitter.Listener<Events, K>, context?: any): this;
    prependListener(event: "error", listener: (error: Error) => void, context?: any): this;
    prependListener<K extends EventEmitter.Event<Events>>(event: K, listener: EventEmitter.Listener<Events, K>, context?: any): this;
    prependOnceListener(event: "error", listener: (error: Error) => void, context?: any): this;
    prependOnceListener<K extends EventEmitter.Event<Events>>(event: K, listener: EventEmitter.Listener<Events, K>, context?: any): this;
    removeAllListeners(event?: "error"): this;
    removeAllListeners<K extends EventEmitter.Event<Events>>(event: K): this;
    removeListener(event: "error", listener: (error: Error) => void): this;
    removeListener<K extends EventEmitter.Event<Events>>(event: K, listener: EventEmitter.Listener<Events, K>): this;
}
export default EventEmitter;
