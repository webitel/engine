import * as SipClient from 'jssip'
import EventEmitter from "./event_emitter";
import {Log} from "./log";

export interface SipConfiguration {
    realm: string;
    uri: string;
    authorization_user: string;
    ha1?: string;
    //registrar_server: string;
    //registrar_server: "sip:sip1.webitel.com",
}


export class SipPhone extends EventEmitter {
    static readonly userAgent = "Webitel-Phone/0.0.1";
    private ua : SipClient.UA;
    private sessionCache = new Map<string,any>();
    private log : Log;

    constructor(private instanceId: string) {
        super();
        SipClient.debug.enable('*');
        this.log = new Log();

        this.on("unregistered", () => {
            this.ua.removeAllListeners();
            this.ua = null;
            this.sessionCache.clear();
        })
    }

    private get callOption() : Object {
        return {
            sessionTimersExpires: 120,
            pcConfig1: {
                'iceServers': [
                    { 'urls':['stun:stun.l.google.com:19302']}
                ]
            },
            mediaConstraints: {
                audio: true,
                video: false
            }
        }
    }

    protected removeSession(id : string) : boolean {
        if (this.sessionCache.has(id)) {
            this.sessionCache.delete(id);
            return true;
        }
        return false
    }

    protected storeSession(id: string, session : any) {
        if (this.sessionCache.has(id)) {
            throw "Session already store"
        }
        this.sessionCache.set(id, session);
    }

    public get allSession () : any[] {
        return Array.from(this.sessionCache.values())
    }

    public getSession(id : string) : any | null {
        if (this.sessionCache.has(id)) {
            return this.sessionCache.get(id);
        }
        return null
    }

    public hasSession(id : string) : boolean {
        return this.sessionCache.has(id);
    }

    public answer(id : string) : boolean {
        if (this.sessionCache.has(id)) {
            this.sessionCache.get(id).answer(this.callOption);
            return true
        }
        return false
    }

    public async register(sipConf: SipConfiguration) {
        var socket = new SipClient.WebSocketInterface('ws://192.168.177.9:5080');

        var configuration = {
            ...sipConf,
            user_agent: SipPhone.userAgent,
            sockets  : [ socket ],
            session_timers: true,
            register_expires: 300,
            connection_recovery_min_interval: 5,
            connection_recovery_max_interval: 60,
            instance_id: "8f1fa16a-1165-4a96-8341-785b1ef24f13"
        };

        var ua = this.ua = new SipClient.UA(configuration);

        ua.on('connected', (e : any) => {
            this.log.error('connected', e);
        });

        ua.on('newRTCSession', (e : any) => {
            const session = e.session;
            let id = e.request.getHeader('X-Webitel-Uuid') || session.id;

            this.storeSession(id, session);

            session.on("ended",() => {
                // this handler will be called for incoming calls too
                this.removeSession(id)
            });

            session.on("failed",() => {
                // this handler will be called for incoming calls too
                this.removeSession(id)
            });

            session.on("accepted",function(){
                // the call has answered
            });

            session.on("confirmed",function(){
                // this handler will be called for incoming calls too
            });

            session.on('addstream', function(e : any){
                // set remote audio stream (to listen to remote audio)
                // remoteAudio is <audio> element on page
                const remoteAudio = document.createElement("audio")
                remoteAudio.src = window.URL.createObjectURL(e.stream);
                remoteAudio.play();
            });

            if (session.direction == "incoming" && e.request.getHeader('X-Webitel-Sock-Id') === this.instanceId) {
                session.answer(this.callOption);
            }
        });

        ua.on('disconnected', (e : any) => {
            this.log.error('disconnected', e);
            this.emit("unregistered")
        });

        ua.on('registered', (e : any) => {
            this.log.error('registered', e);
        });

        ua.on('unregistered', (e : any) => {
            this.log.error('unregistered', e);
            this.emit("unregistered")
        });

        ua.on('registrationFailed', (e : any) => {
            this.log.error('registrationFailed', e);
        });

        ua.on('error', (e : any) => {
            this.emit("error", e);
            this.log.error('error', e);
        });

        ua.start()
    }

    public async unregister() {
        if (this.ua) {
            this.ua.unregister()
        }
    }
}