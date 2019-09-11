import { Client } from './client';
export interface ICallCallback {
    Ringing?: (call: Call) => void;
    Answer?: (call: Call) => void;
    State?: (call: Call) => void;
    Hangup?: (call: Call) => void;
}
export declare class Call {
    protected client: Client;
    private desctination;
    constructor(client: Client);
    hangup(): void;
    answer(): void;
    hold(): void;
    unHold(): void;
    transfer(): void;
}
