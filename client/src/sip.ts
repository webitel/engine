
// import JsSIP from 'jssip'

declare var jssip:any;

export class SipPhone {
    private ua :any;
    constructor() {
        var socket = new jssip.WebSocketInterface('wss://dev.webitel.com'); //'wss://dev.webitel.com'
        var configuration = {
            sockets  : [ socket ],
            registrar_server: "sip:sip1.webitel.com",
            uri      : 'sip:400@webitel.lo',
            authorization_user: "300",
            ha1 : '221e05baa91888c3ca176e6be5d30c5c',
            realm: 'webitel.lo',
            session_timers: true,
            //session_timers_refresh_method: 'invite',
            connection_recovery_min_interval: 5,
            connection_recovery_max_interval: 60,
            instance_id: "8f1fa16a-1165-4a96-8341-785b1ef24f13",
            debug: true
        };

        var ua = this.ua = new jssip.UA(configuration);

        ua.on('connected', (e : any) => {
            console.error('connected', e);
        });

        ua.on('newRTCSession', (e : any) => {
            console.error('newRTCSession', e);
        });
        ua.on('disconnected', (e : any) => {
            console.error('disconnected', e);
        });

        ua.on('registered', (e : any) => {

            console.error('registered', e);

            setTimeout(() => {
                this.makeCall()
            }, 2000)
        });

        ua.on('unregistered', (e : any) => {
            console.error('unregistered', e);
        });

        ua.on('registrationFailed', (e : any) => {
            console.error('registrationFailed', e);
        });

        ua.start();

    }
    makeCall() {
        var session = null as any;
        var selfView =   document.getElementById('a1') as HTMLVideoElement;
        var remoteView =  document.getElementById('a2') as HTMLVideoElement;


        var eventHandlers = {
            'progress':   function(){ /* Your code here */ },
            'failed':     function(){ /* Your code here */ },
            'confirmed':  function(){
                // Attach local stream to selfView
                selfView.srcObject = session.connection.getLocalStreams()[0] ; //window.URL.createObjectURL(session.connection.getLocalStreams()[0]);
                selfView.play()
            },
            'addstream':  function(e : any) {
                var stream = e.stream;
                debugger
                // Attach remote stream to remoteView
                remoteView.src = window.URL.createObjectURL(stream) ;
                remoteView.play()
            },
            'ended':      function(){ /* Your code here */ }
        };

        var options = {
            sessionTimersExpires: 120,
            'eventHandlers': eventHandlers,
            'extraHeaders': [ 'X-Foo: foo', 'X-Bar: bar' ],
            'mediaConstraints': {'audio': true, 'video': false}
        };

        // session = this.ua.call('sip:1@webitel.lo', options);

    }
}