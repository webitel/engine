
export interface ILog {
    debug(primaryMessage: string, ...supportingData: any[]): void;
    warn(primaryMessage: string, ...supportingData: any[]): void;
    error(primaryMessage: string, ...supportingData: any[]): void;
    info(primaryMessage: string, ...supportingData: any[]): void;
}

export class Log implements ILog {
    public debug(msg: string, ...supportingDetails: any[]): void {
        this.emitLogMessage("debug", msg, supportingDetails);
    }
    public info(msg: string, ...supportingDetails: any[]): void {
        this.emitLogMessage("info", msg, supportingDetails);
    }
    public warn(msg: string, ...supportingDetails: any[]): void {
        this.emitLogMessage("warn", msg, supportingDetails);
    }
    public error(msg: string, ...supportingDetails: any[]): void {
        this.emitLogMessage("error", msg, supportingDetails);
    }

    private emitLogMessage(msgType: "debug" | "info" | "warn" | "error", msg: string, supportingDetails: any[] ) {
        if (supportingDetails.length > 0) {
            console[msgType](msg, ...supportingDetails)
        } else {
            console[msgType](msg)
        }
    }
}