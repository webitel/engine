import JsSIP from 'jssip';
var SipPhone = /** @class */ (function () {
    function SipPhone() {
        var _this = this;
        var socket = new JsSIP.WebSocketInterface('wss://dev.webitel.com'); //'wss://dev.webitel.com'
        var configuration = {
            sockets: [socket],
            registrar_server: "sip:sip1.webitel.com",
            uri: 'sip:400@webitel.lo',
            authorization_user: "300",
            ha1: '221e05baa91888c3ca176e6be5d30c5c',
            realm: 'webitel.lo',
            session_timers: true,
            //session_timers_refresh_method: 'invite',
            connection_recovery_min_interval: 5,
            connection_recovery_max_interval: 60,
            instance_id: "8f1fa16a-1165-4a96-8341-785b1ef24f13",
            debug: true
        };
        var ua = this.ua = new JsSIP.UA(configuration);
        ua.on('connected', function (e) {
            console.error('connected', e);
        });
        ua.on('newRTCSession', function (e) {
            console.error('newRTCSession', e);
        });
        ua.on('disconnected', function (e) {
            console.error('disconnected', e);
        });
        ua.on('registered', function (e) {
            console.error('registered', e);
            setTimeout(function () {
                _this.makeCall();
            }, 2000);
        });
        ua.on('unregistered', function (e) {
            console.error('unregistered', e);
        });
        ua.on('registrationFailed', function (e) {
            console.error('registrationFailed', e);
        });
        ua.start();
    }
    SipPhone.prototype.makeCall = function () {
        var session = null;
        var selfView = document.getElementById('a1');
        var remoteView = document.getElementById('a2');
        var eventHandlers = {
            'progress': function (e) { },
            'failed': function (e) { },
            'confirmed': function (e) {
                // Attach local stream to selfView
                selfView.srcObject = session.connection.getLocalStreams()[0]; //window.URL.createObjectURL(session.connection.getLocalStreams()[0]);
                selfView.play();
            },
            'addstream': function (e) {
                var stream = e.stream;
                debugger;
                // Attach remote stream to remoteView
                remoteView.srcObject = window.URL.createObjectURL(stream);
                remoteView.play();
            },
            'ended': function (e) { }
        };
        var options = {
            sessionTimersExpires: 120,
            'eventHandlers': eventHandlers,
            'extraHeaders': ['X-Foo: foo', 'X-Bar: bar'],
            'mediaConstraints': { 'audio': true, 'video': false }
        };
        // session = this.ua.call('sip:1@webitel.lo', options);
    };
    return SipPhone;
}());
export { SipPhone };
//# sourceMappingURL=sip.js.map