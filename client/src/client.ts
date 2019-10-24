import {Socket, IMessage} from './socket'
import {Log} from './log'
import EventEmitter from './event_emitter'
import {Call, CallData, CallEventHandler, CallState} from './call'

import {SipPhone, SipConfiguration} from './sip'

export interface Config {
    endpoint: string;
    token? : string;
    logLvl? : "debug" | "info" | "warn" | "error";
    phone?: number;
}

interface PromiseCallback {
    resolve: (res : Object) => void,
    reject: (err : Object) => void
}

export interface OutboundCallRequest {
    toNumber: string
    toName?: string
    variables?: Map<string,string>
}

const WEBSOCKET_AUTHENTICATION_CHALLENGE  = "authentication_challenge";

const WEBSOCKET_MAKE_OUTBOUND_CALL  = "call_invite";
const WEBSOCKET_EVENT_HELLO = "hello";
const WEBSOCKET_EVENT_CALL = "call";

export enum Response {
    STATUS_FAIL = "FAIL",
    STATUS_OK = "OK"
}

export interface Session {
    id: string
    expire: Number
    user_id: Number
    role_ids: Array<Number>
}

export interface ConnectionInfo {
    sock_id: string
    server_build_commit: string
    server_node_id: string
    server_version: string
    server_time: Number
    session: Session
}

export class Client {
    public phone : SipPhone;
    private socket : Socket;
    private connectionInfo: ConnectionInfo;
    private req_seq : number = 0;
    private queueRequest : Map<number, PromiseCallback> = new Map<number, PromiseCallback>();
    private log : Log;
    private eventHandler: EventEmitter;
    private callStore: Map<string, Call>;

    constructor(protected readonly _config : Config) {
        this.log = new Log();
        this.eventHandler = new EventEmitter();
        this.callStore = new Map<string, Call>();
    }

    public async test() {
        return await this.request("test")
    }

    public async connect() {
        await this.connectToSocket();
    }

    public async disconnect() {
        await this.socket.close()
    }

    //TODO check count
    public async subscribe(action : string, handler: CallEventHandler, data? : Object) : Promise<null | Error> {
        const res = await this.request(`subscribe_${action}`, data);
        this.eventHandler.on(action, handler);
        return res
    }
    //TODO check count
    public async unSubscribe(action : string, handler: CallEventHandler, data? : Object) : Promise<null | Error> {
        const res = await this.request(`un_subscribe_${action}`, data);
        this.eventHandler.off(action, handler);
        return res;
    }

    public allCall() : Call[] {
        return Array.from(this.callStore.values())
    }

    public callById(id : string) : Call | null {
        if (this.callStore.has(id)) {
            return this.callStore.get(id);
        }
        return null;
    }

    public async auth() {
        return this.request(WEBSOCKET_AUTHENTICATION_CHALLENGE, {token: this._config.token})
    }

    public sessionInfo() : Session {
        return this.connectionInfo.session
    }

    public get version() : string {
        return this.connectionInfo.server_version
    }

    public get instanceId() : string {
        return this.connectionInfo.sock_id
    }

    public invite(req: OutboundCallRequest) {
        return this.request(WEBSOCKET_MAKE_OUTBOUND_CALL, req)
    }

    public answer(id: string) : boolean {
        return this.phone.answer(id);
    }

    public request(action : string, data? : Object) : Promise<null | Error>  {
        return new Promise<null | Error>((resolve: () => void, reject: () => void) => {
            this.queueRequest.set(++this.req_seq, {resolve, reject});
            this.socket.send({
                seq: this.req_seq,
                action,
                data
            });
        });
    }

    private onMessage(message : IMessage) {
        this.log.debug("receive message: ", message);
        if (message.seq_reply > 0 ) {
            if (this.queueRequest.has(message.seq_reply)) {
                const promise = this.queueRequest.get(message.seq_reply);
                this.queueRequest.delete(message.seq_reply);
                if (message.status == Response.STATUS_OK) {
                    promise.resolve(message.data)
                } else {
                    promise.reject(message.error)
                }
            }
        } else {
            switch (message.event) {
                case WEBSOCKET_EVENT_HELLO:
                    this.connected(message.data as ConnectionInfo);
                    this.log.debug(`opened session ${this.connectionInfo.sock_id} for userId=${this.connectionInfo.session.user_id}`)
                    break;
                case WEBSOCKET_EVENT_CALL:
                    this.handleCallEvents(message.data as CallData);
                    break;

                case "sip":
                    this.eventHandler.emit("sip", message.data)
                    break;
                default:
                    this.log.error(`event ${message.event} not handler`)
            }
        }
    }

    private connected(info : ConnectionInfo) {
        this.connectionInfo = info;

        this.phone = new SipPhone(this.instanceId);
        this.phone.register(this.deviceSettings);

        // @ts-ignore
        window.cli = this;
    }

    private get deviceSettings() : SipConfiguration | null {
        if (this.connectionInfo) {
            // return {
            //     uri: "sip:400@webitel.lo",
            //     authorization_user: "300",
            //     realm: "webitel.lo",
            //     ha1 : '221e05baa91888c3ca176e6be5d30c5c'
            // }
            return {
                uri: "sip:100@webitel.lo",
                authorization_user: "100",
                realm: "webitel.lo",
                ha1 : 'ce6c7bf5194816a4e2aa65289d82c55a'
            }
        }
        return null
    }

    private connectToSocket() : Promise<Error | null> {
        return new Promise<Error | null>((resolve, reject) => {
            try {
                this.socket = new Socket(this._config.endpoint);
                this.socket.connect(this._config.token)
            } catch (e) {
                reject(e);
                return
            }

            this.socket.on("message", this.onMessage.bind(this));
            this.socket.on("close", (code: number) => {
                this.log.error("socket close code: ", code)
            });
            this.socket.on("open", () => {
                resolve(null);
            });
        });
    }

    private handleCallEvents(event: CallData) {
        let call : Call;
        if (this.callStore.has(event.id)) {
            call = this.callStore.get(event.id);
            call.setState(event);
        } else {
            call = new Call(this, event as CallData);
            this.callStore.set(event.id, call);
        }

        if (call.state == CallState.Hangup) {
            this.callStore.delete(call.id)
        }

        this.eventHandler.emit(WEBSOCKET_EVENT_CALL, call);
    }
}