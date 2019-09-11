import {Client} from './client'

export interface ICallCallback {
    Ringing? : (call: Call) => void;
    Answer? : (call: Call) => void;
    State? : (call: Call) => void;
    Hangup? : (call: Call) => void;
}


export class Call {
    private destination : string;
    constructor(protected client: Client) {

    }

    /* Call control */
    public hangup() {}
    public answer() {}
    public hold() {}
    public unHold() {}
    public toggleCall() {}
    public sendDTMF() {}
    public transfer() {}
}