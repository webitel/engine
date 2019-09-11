import EventEmitter from './event_emitter';
interface ISocketEvents {
    message: IMessage;
    close: Socket;
    open: Socket;
}
export interface IMessage {
    event?: string;
    status?: string;
    seq?: number;
    seq_reply?: number;
    data: Map<string, any>;
    error: Map<string, any>;
}
export interface IRequest {
    seq: number;
    action: string;
    data?: Object;
}
export declare class Socket extends EventEmitter<ISocketEvents> {
    private socket;
    constructor(host: string);
    send(request: IRequest): never | null;
    private onOpen;
    private onClose;
    private onMessage;
}
export {};
