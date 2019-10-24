import {formatWebSocketUri} from './utils'
import EventEmitter from './event_emitter'

const SOCKET_URL_SUFFIX = 'websocket';

interface ISocketEvents {
    message: IMessage;
    close: number;
    open: Socket;
}

export interface IMessage {
    event?: string,
    status?: string,
    seq?: number,
    seq_reply?: number,
    data: Object
    error?: Map<string, any>
}

export interface IRequest {
    seq: number;
    action: string;
    data?: Object;
}

export class Socket extends EventEmitter<ISocketEvents> {
    private socket : WebSocket;

    constructor(private host : string) {
        super();
    }

    public connect(token: string) {
        this.socket = new WebSocket(`${formatWebSocketUri(this.host)}/${SOCKET_URL_SUFFIX}?access_token=${token}`);

        this.socket.onclose = e => this.onClose(e.code);

        this.socket.onmessage = e => this.onMessage(e.data);
        this.socket.onopen = () => this.onOpen();
    }

    public send(request : IRequest) : never | null {
        this.socket.send(JSON.stringify(request));
        return null
    }

    public close(code? : number) {
        this.socket.close(code);
        this.socket = null;
    }

    private onOpen() {
        this.emit("open", this)
    }

    private onClose(code : number) {
        this.emit("close", code);
    }

    private onMessage(data : string) {
        const message = <IMessage>JSON.parse(data);
        //console.log(JSON.stringify(message, null, 4));
        this.emit("message", message)
    }
}