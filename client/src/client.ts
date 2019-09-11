import {Socket, IMessage} from './socket'
import {Log} from './log'

import {SipPhone} from './sip'

export interface IConfig {
    endpoint: string;
    token? : string;
    logLvl? : "debug" | "info" | "warn" | "error";
}

interface IPromiseCallback {
    resolve: (res : Map<string, any>) => void,
    reject: (err : Map<string, any>) => void
}

const WEBSOCKET_AUTHENTICATION_CHALLENGE  = "authentication_challenge";

export enum Response {
    // STATUS_FAIL = "FAIL",
    STATUS_OK = "OK"
}

export class Client {
    private socket : Socket;
    private req_seq : number = 0;
    private queueRequest : Map<number, IPromiseCallback> = new Map<number, IPromiseCallback>();
    private log : Log;

    constructor(protected readonly _config : IConfig) {
        this.log = new Log();
        new SipPhone();
    }

    public async connect() {
        await this.connectToSocket();
    }

    public async disconnect() {
        await this.socket.close()
    }

    public subscribe(action : string, data? : Object) : Promise<null | Error> {
        return new Promise<null | Error>((resolve: () => void, reject: () => void) => {
            this.queueRequest.set(++this.req_seq, {resolve, reject});

            this.socket.send({
                seq: this.req_seq,
                action,
                data
            });
        });
    }

    public auth() {
        return this.request(WEBSOCKET_AUTHENTICATION_CHALLENGE, {token: this._config.token})
    }

    protected request(action : string, data? : Object) : Promise<null | Error>  {
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
            // message.data.delete("debug");
            output(syntaxHighlight(JSON.stringify(message, undefined, 4)))
        }
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
}



function syntaxHighlight(json : any) {
    json = json.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
    return json.replace(/("(\\u[a-zA-Z0-9]{4}|\\[^u]|[^\\"])*"(\s*:)?|\b(true|false|null)\b|-?\d+(?:\.\d*)?(?:[eE][+\-]?\d+)?)/g, function (match : any) {
        var cls = 'number';
        if (/^"/.test(match)) {
            if (/:$/.test(match)) {
                cls = 'key';
            } else {
                cls = 'string';
            }
        } else if (/true|false/.test(match)) {
            cls = 'boolean';
        } else if (/null/.test(match)) {
            cls = 'null';
        }
        return '<span class="' + cls + '">' + match + '</span>';
    });
}

function output(inp :any) {
    document.body.appendChild(document.createElement('pre')).innerHTML = inp;
}