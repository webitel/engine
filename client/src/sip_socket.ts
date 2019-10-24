import {Client} from "./client";


class SipSocket {
    public ondata : Function;
    public onconnect : Function;
    constructor(private sock : Client) {
        /*
        sock.eventHandler.on("sip", ({data}) => {
            console.error(data)
            this.ondata(data)
        })
         */
    }
    get url() {
        console.error("url")
        return "ws://192.168.177.13:5080";
    }

    get sip_uri() {
        console.error("sip_uri")
        return "sip:400@webitel.lo";
    }

    get via_transport() {
        console.error("via_transport")
        return "ws";
    }

    set via_transport(value) {
        console.error("set>via_transport", value)
    }

    connect() {
        console.error("connect")
        this.onconnect();
    }

    disconnect() {
        console.error("disconnect")
    }

    async send(message : string) {
        console.error("send", message)
        await this.sock.request("sip_proxy", {
            data: message
        })
        return true
    }

}
