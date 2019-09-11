export interface IConfig {
    endpoint: string;
    token?: string;
}
export declare enum Response {
    STATUS_OK = "OK"
}
export declare class Client {
    protected readonly _config: IConfig;
    private socket;
    private req_seq;
    private queueRequest;
    constructor(_config: IConfig);
    connect(): Promise<void>;
    subscribe(action: string): Promise<null | Error>;
    protected request(): void;
    private onMessage;
    private connectToSocket;
}
