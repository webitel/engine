import {Client} from './client'

export type CallEventHandler = (call: Call) => void

export enum CallState {
    Ringing = "ringing",
    Active  = "active",
    Hold    = "hold",
    Hangup  = "hangup"
}

export enum CallDirection {
    Inbound = "inbound",
    Outbound = "outbound",
    Internal = "internal"
}

export interface CallData {
    id: string,
    destination: string,
    direction: "outbound" | "inbound" | "internal", //TODO ?
    domain_id: Number,
    event: string,
    from_name: string,
    from_number: string,
    node_name: string,
    parent_id?: string,
    state: number,
    state_name: string,
    to_name: string,
    to_number: string,
    user_id: Number,
    hangup_cause?: string
}

export class Call {
    private answered_at : Number = 0;
    private _muted: boolean = false;
    constructor(protected client: Client, protected data: CallData) {

    }

    public get id() : string {
        return this.data.id
    }

    public get direction(): string {
        return this.data.direction
    }

    public get destination(): string {
        return this.data.destination
    }

    public get userId() : Number {
        return this.data.user_id;
    }

    public get toNumber() : string {
        return this.data.to_number;
    }

    public get toName() : string {
        return this.data.to_name;
    }

    public get fromNumber() : string {
        return this.data.from_number;
    }

    public get fromName() : string {
        return this.data.from_name;
    }

    public get state() : string {
        return this.data.event;
    }

    public get parentCallId() : string | null {
        if (this.data.parent_id) {
            return this.data.parent_id
        }
        return null
    }

    public get hangupCause() : string  {
        return this.data.hangup_cause
    }

    public get nodeId() : string {
        return this.data.node_name
    }

    public setState(event: CallData) : void {
        this.data = event;
        switch (this.state) {
            case "active":
                if (this.answered_at === 0) {
                    this.answered_at = Date.now()
                }
                break;
            default:
                // throw "FIXME";
        }
    }

    public toString() : string {
        return `[${this.data.node_name}:${this.id}${this.parentCallId ? '<'+ this.parentCallId + '>' : ''}] ${this.state} ${this.fromNumber} (${this.fromName}) ${this.direction} to: ${this.toNumber} ${this.toName}`
    }

    public get muted() : boolean {
        return this._muted
    }

    /* Call control */
    public answer() : boolean {
        let sessionId = null;
        if (this.client.phone.hasSession(this.id)) {
            sessionId = this.id;
        } else if (this.client.phone.hasSession(this.parentCallId)) {
            sessionId = this.parentCallId
        }

        if (sessionId) {
            return this.client.phone.answer(sessionId)
        }
        return false;
    }

    public async hangup(cause?: string) {
        if (this.answered_at === 0 && !cause) {
            cause = this.direction === CallDirection.Inbound ? "USER_BUSY" : "ORIGINATOR_CANCEL"
        }

        return await this.client.request("call_hangup", {
            id: this.id,
            node_id: this.data.node_name,
            cause
        })
    }

    public async toggleHold() {
        if (this.state === CallState.Hold) {
            return await this.unHold()
        } else {
            return await this.hold()
        }
    }

    public async hold() {
        if (this.state === CallState.Hold) {
            throw "Call is hold"
        }
        return await this.client.request("call_hold", {
            id: this.id,
            node_id: this.data.node_name
        })
    }

    public async unHold() {
        if (this.state !== CallState.Hold) {
            throw "Call is active"
        }
        return await this.client.request("call_unhold", {
            id: this.id,
            node_id: this.data.node_name
        })
    }

    public async sendDTMF(dtmf : string) {
        return await this.client.request("call_dtmf", {
            id: this.id,
            node_id: this.data.node_name,
            dtmf
        })
    }

    public async blindTransfer(destination: string) {
        if (!this.parentCallId) {
            throw "Not allow one leg"
        }
        return await this.client.request("call_blind_transfer", {
            id: this.parentCallId,
            node_id: this.data.node_name,
            destination
        })
    }

    public async mute(mute : boolean = false) {
        const res = await this.client.request("call_mute", {
            id: this.id,
            node_id: this.data.node_name,
            mute
        });
        this._muted = mute;
        return res;
    }

    public bridgeTo(call : Call) {
        return this.client.request("call_bridge", {
            id: this.id,
            node_id: this.data.node_name,
            to: {
                id: call.id,
                node_id: call.nodeId
            }
        })
    }
}