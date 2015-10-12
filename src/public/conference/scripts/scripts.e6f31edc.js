/*
 * Verto HTML5/Javascript Telephony Signaling and Control Protocol Stack for FreeSWITCH
 * Copyright (C) 2005-2014, Anthony Minessale II <anthm@freeswitch.org>
 *
 * Version: MPL 1.1
 *
 * The contents of this file are subject to the Mozilla Public License Version
 * 1.1 (the "License"); you may not use this file except in compliance with
 * the License. You may obtain a copy of the License at
 * http://www.mozilla.org/MPL/
 *
 * Software distributed under the License is distributed on an "AS IS" basis,
 * WITHOUT WARRANTY OF ANY KIND, either express or implied. See the License
 * for the specific language governing rights and limitations under the
 * License.
 *
 * The Original Code is jquery.jsonrpclient.js modified for Verto HTML5/Javascript Telephony Signaling and Control Protocol Stack for FreeSWITCH
 *
 * The Initial Developer of the Original Code is
 * Textalk AB http://textalk.se/
 * Portions created by the Initial Developer are Copyright (C)
 * the Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * Anthony Minessale II <anthm@freeswitch.org>
 *
 * jquery.jsonrpclient.js - JSON RPC client code
 *
 */
/**
 * This plugin requires jquery.json.js to be available, or at least the methods $.toJSON and
 * $.parseJSON.
 *
 * The plan is to make use of websockets if they are available, but work just as well with only
 * http if not.
 *
 * Usage example:
 *
 *   var foo = new $.JsonRpcClient({ ajaxUrl: '/backend/jsonrpc' });
 *   foo.call(
 *     'bar', [ 'A parameter', 'B parameter' ],
 *     function(result) { alert('Foo bar answered: ' + result.my_answer); },
 *     function(error)  { console.log('There was an error', error); }
 *   );
 *
 * More examples are available in README.md
 */
(function($) {
  /**
   * @fn new
   * @memberof $.JsonRpcClient
   *
   * @param options An object stating the backends:
   *                ajaxUrl    A url (relative or absolute) to a http(s) backend.
   *                socketUrl  A url (relative of absolute) to a ws(s) backend.
   *                onmessage  A socket message handler for other messages (non-responses).
   *                getSocket  A function returning a WebSocket or null.
   *                           It must take an onmessage_cb and bind it to the onmessage event
   *                           (or chain it before/after some other onmessage handler).
   *                           Or, it could return null if no socket is available.
   *                           The returned instance must have readyState <= 1, and if less than 1,
   *                           react to onopen binding.
   */
    $.JsonRpcClient = function(options) {
        var self = this;
        this.options = $.extend({
            ajaxUrl       : null,
            socketUrl     : null, ///< The ws-url for default getSocket.
            onmessage     : null, ///< Other onmessage-handler.
            login         : null, /// auth login
            passwd        : null, /// auth passwd
            sessid        : null,
	    loginParams   : null,
	    userVariables : null,
            getSocket     : function(onmessage_cb) { return self._getSocket(onmessage_cb); }
        }, options);

        self.ws_cnt = 0;

        // Declare an instance version of the onmessage callback to wrap 'this'.
        this.wsOnMessage = function(event) { self._wsOnMessage(event); };
    };

    /// Holding the WebSocket on default getsocket.
    $.JsonRpcClient.prototype._ws_socket = null;

    /// Object <id>: { success_cb: cb, error_cb: cb }
    $.JsonRpcClient.prototype._ws_callbacks = {};

    /// The next JSON-RPC request id.
    $.JsonRpcClient.prototype._current_id = 1;

    /**
     * @fn call
     * @memberof $.JsonRpcClient
     *
     * @param method     The method to run on JSON-RPC server.
     * @param params     The params; an array or object.
     * @param success_cb A callback for successful request.
     * @param error_cb   A callback for error.
     */
    $.JsonRpcClient.prototype.call = function(method, params, success_cb, error_cb) {
    // Construct the JSON-RPC 2.0 request.

        if (!params) {
            params = {};
        }

        if (this.options.sessid) {
            params.sessid = this.options.sessid;
        }

        var request = {
            jsonrpc : '2.0',
            method  : method,
            params  : params,
            id      : this._current_id++  // Increase the id counter to match request/response
        };

        if (!success_cb) {
            success_cb = function(e){console.log("Success: ", e);};
        }

        if (!error_cb) {
            error_cb = function(e){console.log("Error: ", e);};
        }

        // Try making a WebSocket call.
        var socket = this.options.getSocket(this.wsOnMessage);
        if (socket !== null) {
            this._wsCall(socket, request, success_cb, error_cb);
            return;
        }

        // No WebSocket, and no HTTP backend?  This won't work.
        if (this.options.ajaxUrl === null) {
            throw "$.JsonRpcClient.call used with no websocket and no http endpoint.";
        }

        $.ajax({
            type     : 'POST',
            url      : this.options.ajaxUrl,
            data     : $.toJSON(request),
            dataType : 'json',
            cache    : false,

            success  : function(data) {
                if ('error' in data) error_cb(data.error, this);
                success_cb(data.result, this);
            },

            // JSON-RPC Server could return non-200 on error
            error    : function(jqXHR, textStatus, errorThrown) {
                try {
                    var response = $.parseJSON(jqXHR.responseText);

                    if ('console' in window) console.log(response);

                    error_cb(response.error, this);
                } catch (err) {
                    // Perhaps the responseText wasn't really a jsonrpc-error.
                    error_cb({ error: jqXHR.responseText }, this);
                }
            }
        });
    };

    /**
     * Notify sends a command to the server that won't need a response.  In http, there is probably
     * an empty response - that will be dropped, but in ws there should be no response at all.
     *
     * This is very similar to call, but has no id and no handling of callbacks.
     *
     * @fn notify
     * @memberof $.JsonRpcClient
     *
     * @param method     The method to run on JSON-RPC server.
     * @param params     The params; an array or object.
     */
    $.JsonRpcClient.prototype.notify = function(method, params) {
        // Construct the JSON-RPC 2.0 request.

        if (this.options.sessid) {
            params.sessid = this.options.sessid;
        }

        var request = {
            jsonrpc: '2.0',
            method:  method,
            params:  params
        };

        // Try making a WebSocket call.
        var socket = this.options.getSocket(this.wsOnMessage);
        if (socket !== null) {
            this._wsCall(socket, request);
            return;
        }

        // No WebSocket, and no HTTP backend?  This won't work.
        if (this.options.ajaxUrl === null) {
            throw "$.JsonRpcClient.notify used with no websocket and no http endpoint.";
        }

        $.ajax({
            type     : 'POST',
            url      : this.options.ajaxUrl,
            data     : $.toJSON(request),
            dataType : 'json',
            cache    : false
        });
    };

    /**
     * Make a batch-call by using a callback.
     *
     * The callback will get an object "batch" as only argument.  On batch, you can call the methods
     * "call" and "notify" just as if it was a normal $.JsonRpcClient object, and all calls will be
     * sent as a batch call then the callback is done.
     *
     * @fn batch
     * @memberof $.JsonRpcClient
     *
     * @param callback    The main function which will get a batch handler to run call and notify on.
     * @param all_done_cb A callback function to call after all results have been handled.
     * @param error_cb    A callback function to call if there is an error from the server.
     *                    Note, that batch calls should always get an overall success, and the
     *                    only error
     */
    $.JsonRpcClient.prototype.batch = function(callback, all_done_cb, error_cb) {
        var batch = new $.JsonRpcClient._batchObject(this, all_done_cb, error_cb);
        callback(batch);
        batch._execute();
    };

    /**
     * The default getSocket handler.
     *
     * @param onmessage_cb The callback to be bound to onmessage events on the socket.
     *
     * @fn _getSocket
     * @memberof $.JsonRpcClient
     */

    $.JsonRpcClient.prototype.socketReady = function() {
        if (this._ws_socket === null || this._ws_socket.readyState > 1) {
            return false;
        }

        return true;
    };

    $.JsonRpcClient.prototype.closeSocket = function() {
	var self = this;
        if (self.socketReady()) {
            self._ws_socket.onclose = function (w) {console.log("Closing Socket");};
            self._ws_socket.close();
        }
    };

    $.JsonRpcClient.prototype.loginData = function(params) {
	var self = this;
        self.options.login = params.login;
        self.options.passwd = params.passwd;
	self.options.loginParams = params.loginParams;
	self.options.userVariables = params.userVariables;
    };

    $.JsonRpcClient.prototype.connectSocket = function(onmessage_cb) {
        var self = this;

        if (self.to) {
            clearTimeout(self.to);
        }

        if (!self.socketReady()) {
            self.authing = false;

            if (self._ws_socket) {
                delete self._ws_socket;
            }

            // No socket, or dying socket, let's get a new one.
            self._ws_socket = new WebSocket(self.options.socketUrl);

            if (self._ws_socket) {
                // Set up onmessage handler.
                self._ws_socket.onmessage = onmessage_cb;
                self._ws_socket.onclose = function (w) {
                    if (!self.ws_sleep) {
                        self.ws_sleep = 1000;
                    }

                    if (self.options.onWSClose) {
                        self.options.onWSClose(self);
                    }

                    console.error("Websocket Lost " + self.ws_cnt + " sleep: " + self.ws_sleep + "msec");

                    self.to = setTimeout(function() {
                        console.log("Attempting Reconnection....");
                        self.connectSocket(onmessage_cb);
                    }, self.ws_sleep);

                    self.ws_cnt++;

                    if (self.ws_sleep < 3000 && (self.ws_cnt % 10) === 0) {
                        self.ws_sleep += 1000;
                    }
                };

                // Set up sending of message for when the socket is open.
                self._ws_socket.onopen = function() {
                    if (self.to) {
                        clearTimeout(self.to);
                    }
                    self.ws_sleep = 1000;
                    self.ws_cnt = 0;
                    if (self.options.onWSConnect) {
                        self.options.onWSConnect(self);
                    }

                    var req;
                    // Send the requests.
                    while ((req = $.JsonRpcClient.q.pop())) {
                        self._ws_socket.send(req);
                    }
                };
            }
        }

        return self._ws_socket ? true : false;
    };

    $.JsonRpcClient.prototype._getSocket = function(onmessage_cb) {
        // If there is no ws url set, we don't have a socket.
        // Likewise, if there is no window.WebSocket.
        if (this.options.socketUrl === null || !("WebSocket" in window)) return null;

        this.connectSocket(onmessage_cb);

        return this._ws_socket;
    };

    /**
     * Queue to save messages delivered when websocket is not ready
     */
    $.JsonRpcClient.q = [];

    /**
     * Internal handler to dispatch a JRON-RPC request through a websocket.
     *
     * @fn _wsCall
     * @memberof $.JsonRpcClient
     */
    $.JsonRpcClient.prototype._wsCall = function(socket, request, success_cb, error_cb) {
        var request_json = $.toJSON(request);

        if (socket.readyState < 1) {
            // The websocket is not open yet; we have to set sending of the message in onopen.
            self = this; // In closure below, this is set to the WebSocket.  Use self instead.
            $.JsonRpcClient.q.push(request_json);
        } else {
            // We have a socket and it should be ready to send on.
            socket.send(request_json);
        }

        // Setup callbacks.  If there is an id, this is a call and not a notify.
        if ('id' in request && typeof success_cb !== 'undefined') {
            this._ws_callbacks[request.id] = { request: request_json, request_obj: request, success_cb: success_cb, error_cb: error_cb };
        }
    };

    /**
     * Internal handler for the websocket messages.  It determines if the message is a JSON-RPC
     * response, and if so, tries to couple it with a given callback.  Otherwise, it falls back to
     * given external onmessage-handler, if any.
     *
     * @param event The websocket onmessage-event.
     */
    $.JsonRpcClient.prototype._wsOnMessage = function(event) {
        // Check if this could be a JSON RPC message.
        var response;
        try {
            response = $.parseJSON(event.data);

            /// @todo Make using the jsonrcp 2.0 check optional, to use this on JSON-RPC 1 backends.

            if (typeof response === 'object' &&
                'jsonrpc' in response &&
                response.jsonrpc === '2.0') {

                /// @todo Handle bad response (without id).

                // If this is an object with result, it is a response.
                if ('result' in response && this._ws_callbacks[response.id]) {
                    // Get the success callback.
                    var success_cb = this._ws_callbacks[response.id].success_cb;

    /*
                    // set the sessid if present
                    if ('sessid' in response.result && !this.options.sessid || (this.options.sessid != response.result.sessid)) {
                        this.options.sessid = response.result.sessid;
                        if (this.options.sessid) {
                            console.log("setting session UUID to: " + this.options.sessid);
                        }
                    }
    */
                    // Delete the callback from the storage.
                    delete this._ws_callbacks[response.id];

                    // Run callback with result as parameter.
                    success_cb(response.result, this);
                    return;
                } else if ('error' in response && this._ws_callbacks[response.id]) {
                // If this is an object with error, it is an error response.

                // Get the error callback.
                    var error_cb = this._ws_callbacks[response.id].error_cb;
                    var orig_req = this._ws_callbacks[response.id].request;

                    // if this is an auth request, send the credentials and resend the failed request
                    if (!self.authing && response.error.code == -32000 && self.options.login && self.options.passwd) {
                        self.authing = true;

                        this.call("login", { login: self.options.login, passwd: self.options.passwd, loginParams: self.options.loginParams,
					     userVariables: self.options.userVariables},
                            this._ws_callbacks[response.id].request_obj.method == "login" ?
                            function(e) {
                                self.authing = false;
                                console.log("logged in");
                                delete self._ws_callbacks[response.id];

                                if (self.options.onWSLogin) {
                                    self.options.onWSLogin(true, self);
                                }
                            }

                            :

                            function(e) {
                                self.authing = false;
                                console.log("logged in, resending request id: " + response.id);
                                var socket = self.options.getSocket(self.wsOnMessage);
                                if (socket !== null) {
                                    socket.send(orig_req);
                                }
                                if (self.options.onWSLogin) {
                                    self.options.onWSLogin(true, self);
                                }
                            },

                            function(e) {
                                console.log("error logging in, request id:", response.id);
                                delete self._ws_callbacks[response.id];
                                error_cb(response.error, this);
                                if (self.options.onWSLogin) {
                                self.options.onWSLogin(false, self);
                                }
                            });
                            return;
                        }

                        // Delete the callback from the storage.
                        delete this._ws_callbacks[response.id];

                        // Run callback with the error object as parameter.
                        error_cb(response.error, this);
                        return;
                    }
                }
            } catch (err) {
            // Probably an error while parsing a non json-string as json.  All real JSON-RPC cases are
            // handled above, and the fallback method is called below.
            console.log("ERROR: "+ err);
            return;
        }

        // This is not a JSON-RPC response.  Call the fallback message handler, if given.
        if (typeof this.options.onmessage === 'function') {
            event.eventData = response;
            if (!event.eventData) {
                event.eventData = {};
            }

            var reply = this.options.onmessage(event);

            if (reply && typeof reply === "object" && event.eventData.id) {
                var msg = {
                    jsonrpc: "2.0",
                    id: event.eventData.id,
                    result: reply
                };

                var socket = self.options.getSocket(self.wsOnMessage);
                if (socket !== null) {
                    socket.send($.toJSON(msg));
                }
            }
        }
    };


    /************************************************************************************************
     * Batch object with methods
     ************************************************************************************************/

    /**
     * Handling object for batch calls.
     */
    $.JsonRpcClient._batchObject = function(jsonrpcclient, all_done_cb, error_cb) {
        // Array of objects to hold the call and notify requests.  Each objects will have the request
        // object, and unless it is a notify, success_cb and error_cb.
        this._requests   = [];

        this.jsonrpcclient = jsonrpcclient;
        this.all_done_cb = all_done_cb;
        this.error_cb    = typeof error_cb === 'function' ? error_cb : function() {};

    };

    /**
     * @sa $.JsonRpcClient.prototype.call
     */
    $.JsonRpcClient._batchObject.prototype.call = function(method, params, success_cb, error_cb) {

        if (!params) {
            params = {};
        }

        if (this.options.sessid) {
            params.sessid = this.options.sessid;
        }

        if (!success_cb) {
            success_cb = function(e){console.log("Success: ", e);};
        }

        if (!error_cb) {
        error_cb = function(e){console.log("Error: ", e);};
        }

        this._requests.push({
            request    : {
            jsonrpc : '2.0',
            method  : method,
            params  : params,
            id      : this.jsonrpcclient._current_id++  // Use the client's id series.
        },
            success_cb : success_cb,
            error_cb   : error_cb
        });
    };

    /**
     * @sa $.JsonRpcClient.prototype.notify
     */
    $.JsonRpcClient._batchObject.prototype.notify = function(method, params) {
        if (this.options.sessid) {
            params.sessid = this.options.sessid;
        }

        this._requests.push({
            request    : {
                jsonrpc : '2.0',
                method  : method,
                params  : params
            }
        });
    };

    /**
     * Executes the batched up calls.
     */
    $.JsonRpcClient._batchObject.prototype._execute = function() {
        var self = this;

        if (this._requests.length === 0) return; // All done :P

        // Collect all request data and sort handlers by request id.
        var batch_request = [];
        var handlers = {};
        var i = 0;
        var call;
        var success_cb;
        var error_cb;

        // If we have a WebSocket, just send the requests individually like normal calls.
        var socket = self.jsonrpcclient.options.getSocket(self.jsonrpcclient.wsOnMessage);
        if (socket !== null) {
            for (i = 0; i < this._requests.length; i++) {
                call = this._requests[i];
                success_cb = ('success_cb' in call) ? call.success_cb : undefined;
                error_cb   = ('error_cb'   in call) ? call.error_cb   : undefined;
                self.jsonrpcclient._wsCall(socket, call.request, success_cb, error_cb);
            }

            if (typeof all_done_cb === 'function') all_done_cb(result);
            return;
        }

        for (i = 0; i < this._requests.length; i++) {
            call = this._requests[i];
            batch_request.push(call.request);

            // If the request has an id, it should handle returns (otherwise it's a notify).
            if ('id' in call.request) {
                handlers[call.request.id] = {
                    success_cb : call.success_cb,
                    error_cb   : call.error_cb
                };
            }
        }

        success_cb = function(data) { self._batchCb(data, handlers, self.all_done_cb); };

        // No WebSocket, and no HTTP backend?  This won't work.
        if (self.jsonrpcclient.options.ajaxUrl === null) {
            throw "$.JsonRpcClient.batch used with no websocket and no http endpoint.";
        }

        // Send request
        $.ajax({
            url      : self.jsonrpcclient.options.ajaxUrl,
            data     : $.toJSON(batch_request),
            dataType : 'json',
            cache    : false,
            type     : 'POST',

            // Batch-requests should always return 200
            error    : function(jqXHR, textStatus, errorThrown) {
                self.error_cb(jqXHR, textStatus, errorThrown);
            },
            success  : success_cb
        });
    };

    /**
     * Internal helper to match the result array from a batch call to their respective callbacks.
     *
     * @fn _batchCb
     * @memberof $.JsonRpcClient
     */
    $.JsonRpcClient._batchObject.prototype._batchCb = function(result, handlers, all_done_cb) {
        for (var i = 0; i < result.length; i++) {
            var response = result[i];

            // Handle error
            if ('error' in response) {
                if (response.id === null || !(response.id in handlers)) {
                    // An error on a notify?  Just log it to the console.
                    if ('console' in window) console.log(response);
                } else {
                    handlers[response.id].error_cb(response.error, this);
                }
            } else {
                // Here we should always have a correct id and no error.
                if (!(response.id in handlers) && 'console' in window) {
                    console.log(response);
                } else {
                    handlers[response.id].success_cb(response.result, this);
                }
            }
        }

        if (typeof all_done_cb === 'function') all_done_cb(result);
    };

})(jQuery);

/*
 * Verto HTML5/Javascript Telephony Signaling and Control Protocol Stack for FreeSWITCH
 * Copyright (C) 2005-2014, Anthony Minessale II <anthm@freeswitch.org>
 *
 * Version: MPL 1.1
 *
 * The contents of this file are subject to the Mozilla Public License Version
 * 1.1 (the "License"); you may not use this file except in compliance with
 * the License. You may obtain a copy of the License at
 * http://www.mozilla.org/MPL/
 *
 * Software distributed under the License is distributed on an "AS IS" basis,
 * WITHOUT WARRANTY OF ANY KIND, either express or implied. See the License
 * for the specific language governing rights and limitations under the
 * License.
 *
 * The Original Code is Verto HTML5/Javascript Telephony Signaling and Control Protocol Stack for FreeSWITCH
 *
 * The Initial Developer of the Original Code is
 * Anthony Minessale II <anthm@freeswitch.org>
 * Portions created by the Initial Developer are Copyright (C)
 * the Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * Anthony Minessale II <anthm@freeswitch.org>
 *
 * jquery.FSRTC.js - WebRTC Glue code
 *
 */

(function($) {

    // Find the line in sdpLines that starts with |prefix|, and, if specified,
    // contains |substr| (case-insensitive search).
    function findLine(sdpLines, prefix, substr) {
        return findLineInRange(sdpLines, 0, -1, prefix, substr);
    }

    // Find the line in sdpLines[startLine...endLine - 1] that starts with |prefix|
    // and, if specified, contains |substr| (case-insensitive search).
    function findLineInRange(sdpLines, startLine, endLine, prefix, substr) {
        var realEndLine = (endLine != -1) ? endLine : sdpLines.length;
        for (var i = startLine; i < realEndLine; ++i) {
            if (sdpLines[i].indexOf(prefix) === 0) {
                if (!substr || sdpLines[i].toLowerCase().indexOf(substr.toLowerCase()) !== -1) {
                    return i;
                }
            }
        }
        return null;
    }

    // Gets the codec payload type from an a=rtpmap:X line.
    function getCodecPayloadType(sdpLine) {
        var pattern = new RegExp('a=rtpmap:(\\d+) \\w+\\/\\d+');
        var result = sdpLine.match(pattern);
        return (result && result.length == 2) ? result[1] : null;
    }
    
    // Returns a new m= line with the specified codec as the first one.
    function setDefaultCodec(mLine, payload) {
        var elements = mLine.split(' ');
        var newLine = [];
        var index = 0;
        for (var i = 0; i < elements.length; i++) {
            if (index === 3) { // Format of media starts from the fourth.
                newLine[index++] = payload; // Put target payload to the first.
            }
            if (elements[i] !== payload) newLine[index++] = elements[i];
        }
        return newLine.join(' ');
    }

    $.FSRTC = function(options) {
        this.options = $.extend({
            useVideo: null,
            useStereo: false,
            userData: null,
	    localVideo: null,
	    screenShare: false,
	    useCamera: "any",
            iceServers: false,
            videoParams: {},
            audioParams: {},
            callbacks: {
                onICEComplete: function() {},
                onICE: function() {},
                onOfferSDP: function() {}
            },
        }, options);

	this.enabled = true;


        this.mediaData = {
            SDP: null,
            profile: {},
            candidateList: []
        };


	if (moz) {
            this.constraints = {
		offerToReceiveAudio: true,
		offerToReceiveVideo: this.options.useVideo ? true : false,
            };
	} else {
            this.constraints = {
		optional: [{
		    'DtlsSrtpKeyAgreement': 'true'
		}],mandatory: {
		    OfferToReceiveAudio: true,
		    OfferToReceiveVideo: this.options.useVideo ? true : false,
		}
            };
	}

        if (self.options.useVideo) {
            self.options.useVideo.style.display = 'none';
        }

        setCompat();
        checkCompat();
    };

    $.FSRTC.validRes = [];

    $.FSRTC.prototype.useVideo = function(obj, local) {
        var self = this;

        if (obj) {
            self.options.useVideo = obj;
	    self.options.localVideo = local;
	    if (moz) {
		self.constraints.offerToReceiveVideo = true;
	    } else {
		self.constraints.mandatory.OfferToReceiveVideo = true;
	    }
        } else {
            self.options.useVideo = null;
	    self.options.localVideo = null;
            if (moz) {
		self.constraints.offerToReceiveVideo = false;
	    } else {
		self.constraints.mandatory.OfferToReceiveVideo = false;
	    }
        }

        if (self.options.useVideo) {
            self.options.useVideo.style.display = 'none';
        }
    };

    $.FSRTC.prototype.useStereo = function(on) {
        var self = this;
        self.options.useStereo = on;
    };

    // Sets Opus in stereo if stereo is enabled, by adding the stereo=1 fmtp param.
    $.FSRTC.prototype.stereoHack = function(sdp) {
        var self = this;

        if (!self.options.useStereo) {
            return sdp;
        }

        var sdpLines = sdp.split('\r\n');

        // Find opus payload.
        var opusIndex = findLine(sdpLines, 'a=rtpmap', 'opus/48000'), opusPayload;

        if (!opusIndex) {
	    return sdp;
	} else {
            opusPayload = getCodecPayloadType(sdpLines[opusIndex]);
        }

        // Find the payload in fmtp line.
        var fmtpLineIndex = findLine(sdpLines, 'a=fmtp:' + opusPayload.toString());

        if (fmtpLineIndex === null) {
	    // create an fmtp line
	    sdpLines[opusIndex] = sdpLines[opusIndex] + '\r\na=fmtp:' + opusPayload.toString() + " stereo=1; sprop-stereo=1"
	} else {
            // Append stereo=1 to fmtp line.
            sdpLines[fmtpLineIndex] = sdpLines[fmtpLineIndex].concat('; stereo=1; sprop-stereo=1');
	}

        sdp = sdpLines.join('\r\n');
        return sdp;
    };

    function setCompat() {
        $.FSRTC.moz = !!navigator.mozGetUserMedia;
        //navigator.getUserMedia || (navigator.getUserMedia = navigator.mozGetUserMedia || navigator.webkitGetUserMedia || navigator.msGetUserMedia);
        if (!navigator.getUserMedia) {
            navigator.getUserMedia = navigator.mozGetUserMedia || navigator.webkitGetUserMedia || navigator.msGetUserMedia;
        }
    }

    function checkCompat() {
        if (!navigator.getUserMedia) {
            alert('This application cannot function in this browser.');
            return false;
        }
        return true;
    }

    function onStreamError(self, e) {
        console.log('There has been a problem retrieving the streams - did you allow access? Check Device Resolution', e);
        doCallback(self, "onError", e);
    }

    function onStreamSuccess(self, stream) {
        console.log("Stream Success");
        doCallback(self, "onStream", stream);
    }

    function onICE(self, candidate) {
        self.mediaData.candidate = candidate;
        self.mediaData.candidateList.push(self.mediaData.candidate);

        doCallback(self, "onICE");
    }

    function doCallback(self, func, arg) {
        if (func in self.options.callbacks) {
            self.options.callbacks[func](self, arg);
        }
    }

    function onICEComplete(self, candidate) {
        console.log("ICE Complete");
        doCallback(self, "onICEComplete");
    }

    function onChannelError(self, e) {
        console.error("Channel Error", e);
        doCallback(self, "onError", e);
    }

    function onICESDP(self, sdp) {
        self.mediaData.SDP = self.stereoHack(sdp.sdp);
        console.log("ICE SDP");
        doCallback(self, "onICESDP");
    }

    function onAnswerSDP(self, sdp) {
        self.answer.SDP = self.stereoHack(sdp.sdp);
        console.log("ICE ANSWER SDP");
        doCallback(self, "onAnswerSDP", self.answer.SDP);
    }

    function onMessage(self, msg) {
        console.log("Message");
        doCallback(self, "onICESDP", msg);
    }

    function onRemoteStream(self, stream) {
        if (self.options.useVideo) {
            self.options.useVideo.style.display = 'block';
        }

        var element = self.options.useAudio;
        console.log("REMOTE STREAM", stream, element);

        if (typeof element.srcObject !== 'undefined') {
            element.srcObject = stream;
        } else if (typeof element.mozSrcObject !== 'undefined') {
            element.mozSrcObject = stream;
        } else if (typeof element.src !== 'undefined') {
            element.src = URL.createObjectURL(stream);
        } else {
            console.error('Error attaching stream to element.');
        }

        self.options.useAudio.play();
        self.remoteStream = stream;
    }

    function onOfferSDP(self, sdp) {
        self.mediaData.SDP = self.stereoHack(sdp.sdp);
        console.log("Offer SDP");
        doCallback(self, "onOfferSDP");
    }

    $.FSRTC.prototype.answer = function(sdp, onSuccess, onError) {
        this.peer.addAnswerSDP({
            type: "answer",
            sdp: sdp
        },
        onSuccess, onError);
    };

    $.FSRTC.prototype.stopPeer = function() {
        if (self.peer) {
            console.log("stopping peer");
            self.peer.stop();
        }
    }

    $.FSRTC.prototype.stop = function() {
        var self = this;

        if (self.options.useVideo) {
            self.options.useVideo.style.display = 'none';
            if (moz) {
                self.options.useVideo['mozSrcObject'] = null;
            } else {
                self.options.useVideo['src'] = '';
            }
        }

        if (self.localStream) {
            if(typeof self.localStream.stop == 'function') {
                self.localStream.stop();
            } else {
		if (self.localStream.active){
                    var tracks = self.localStream.getTracks();
                    console.error(tracks);
		    tracks.forEach(function(track, index){
			console.log(track);
			track.stop();
		    })
                }
            }
            self.localStream = null;
        }

        if (self.options.localVideo) {
            self.options.localVideo.style.display = 'none';
            if (moz) {
                self.options.localVideo['mozSrcObject'] = null;
            } else {
                self.options.localVideo['src'] = '';
            }
        }

	if (self.options.localVideoStream) {
            if(typeof self.options.localVideoStream.stop == 'function') {
	        self.options.localVideoStream.stop();
            } else {
		if (self.localVideoStream.active){
                    var tracks = self.localVideoStream.getTracks();
                    console.error(tracks);
		    tracks.forEach(function(track, index){
			console.log(track);
			track.stop();
		    })
                }
            }
        }

        if (self.peer) {
            console.log("stopping peer");
            self.peer.stop();
        }
    };

    $.FSRTC.prototype.getMute = function() {
	var self = this;
	return self.enabled;
    }

    $.FSRTC.prototype.setMute = function(what) {
	var self = this;
	var audioTracks = self.localStream.getAudioTracks();	

	for (var i = 0, len = audioTracks.length; i < len; i++ ) {
	    switch(what) {
	    case "on":
		audioTracks[i].enabled = true;
		break;
	    case "off":
		audioTracks[i].enabled = false;
		break;
	    case "toggle":
		audioTracks[i].enabled = !audioTracks[i].enabled;
	    default:
		break;
	    }

	    self.enabled = audioTracks[i].enabled;
	}

	return !self.enabled;
    }

    $.FSRTC.prototype.createAnswer = function(params) {
        var self = this;
        self.type = "answer";
        self.remoteSDP = params.sdp;
        console.debug("inbound sdp: ", params.sdp);

        function onSuccess(stream) {
            self.localStream = stream;

            self.peer = RTCPeerConnection({
                type: self.type,
                attachStream: self.localStream,
                onICE: function(candidate) {
                    return onICE(self, candidate);
                },
                onICEComplete: function() {
                    return onICEComplete(self);
                },
                onRemoteStream: function(stream) {
                    return onRemoteStream(self, stream);
                },
                onICESDP: function(sdp) {
                    return onICESDP(self, sdp);
                },
                onChannelError: function(e) {
                    return onChannelError(self, e);
                },
                constraints: self.constraints,
                iceServers: self.options.iceServers,
                offerSDP: {
                    type: "offer",
                    sdp: self.remoteSDP
                }
            });

            onStreamSuccess(self);
        }

        function onError(e) {
            onStreamError(self, e);
        }

	var mediaParams = getMediaParams(self);

	console.log("Audio constraints", mediaParams.audio);
	console.log("Video constraints", mediaParams.video);

	if (self.options.useVideo && self.options.localVideo) {
            getUserMedia({
		constraints: {
                    audio: false,
                    video: {
			mandatory: self.options.videoParams,
			optional: []
                    },
		},
		localVideo: self.options.localVideo,
		onsuccess: function(e) {self.options.localVideoStream = e; console.log("local video ready");},
		onerror: function(e) {console.error("local video error!");}
            });
	}

        getUserMedia({
            constraints: {
		audio: mediaParams.audio,
		video: mediaParams.video
            },
            video: mediaParams.useVideo,
            onsuccess: onSuccess,
            onerror: onError
        });



    };

    function getMediaParams(obj) {

	var audio;

	if (obj.options.useMic && obj.options.useMic === "none") {
	    console.log("Microphone Disabled");
	    audio = false;
	} else if (obj.options.videoParams && obj.options.screenShare) {//obj.options.videoParams.chromeMediaSource == 'desktop') {

	    //obj.options.videoParams = {
	//	chromeMediaSource: 'screen',
	//	maxWidth:screen.width,
	//	maxHeight:screen.height
	//	chromeMediaSourceId = sourceId;
	  //  };

	    console.error("SCREEN SHARE");
	    audio = false;
	} else {
	    audio = {
		mandatory: obj.options.audioParams,
		optional: []
	    };

	    if (obj.options.useMic !== "any") {
		audio.optional = [{sourceId: obj.options.useMic}]
	    }

	}

	if (obj.options.useVideo && obj.options.localVideo) {
            getUserMedia({
		constraints: {
                    audio: false,
                    video: {
			mandatory: obj.options.videoParams,
			optional: []
                    },
		},
		localVideo: obj.options.localVideo,
		onsuccess: function(e) {self.options.localVideoStream = e; console.log("local video ready");},
		onerror: function(e) {console.error("local video error!");}
            });
	}

	var video = {};
	var bestFrameRate = obj.options.videoParams.vertoBestFrameRate;
	delete obj.options.videoParams.vertoBestFrameRate;

	if (window.moz) {
	    video = obj.options.videoParams;
	    if (!video.width) video.width = video.minWidth;
	    if (!video.height) video.height = video.minHeight;
	    if (!video.frameRate) video.frameRate = video.minFrameRate;
	} else {
	    video = {
		mandatory: obj.options.videoParams,
		optional: []
            }	    	    
	}
	
	var useVideo = obj.options.useVideo;

	if (useVideo && obj.options.useCamera && obj.options.useCamera !== "none") {
	    if (!video.optional) {
		video.optional = [];
	    }

	    if (obj.options.useCamera !== "any") {
		video.optional.push({sourceId: obj.options.useCamera});
	    }

	    if (bestFrameRate && !window.moz) {
		 video.optional.push({minFrameRate: bestFrameRate});
	    }

	} else {
	    console.log("Camera Disabled");
	    video = false;
	    useVideo = false;
	}

	return {audio: audio, video: video, useVideo: useVideo};
    }
    
    $.FSRTC.prototype.call = function(profile) {
        checkCompat();
	
        var self = this;
	var screen = false;

        self.type = "offer";

	if (self.options.videoParams && self.options.screenShare) { //self.options.videoParams.chromeMediaSource == 'desktop') {
	    screen = true;
	}

        function onSuccess(stream) {
	    self.localStream = stream;
	    
	    if (screen) {
		if (moz) {
		    self.constraints.OfferToReceiveVideo = false;
		} else {
		    self.constraints.mandatory.OfferToReceiveVideo = false;
		}
	    }
	    
            self.peer = RTCPeerConnection({
                type: self.type,
                attachStream: self.localStream,
                onICE: function(candidate) {
                    return onICE(self, candidate);
                },
                onICEComplete: function() {
                    return onICEComplete(self);
                },
                onRemoteStream: screen ? function(stream) {} : function(stream) {
                    return onRemoteStream(self, stream);
                },
                onOfferSDP: function(sdp) {
                    return onOfferSDP(self, sdp);
                },
                onICESDP: function(sdp) {
                    return onICESDP(self, sdp);
                },
                onChannelError: function(e) {
                    return onChannelError(self, e);
                },
                constraints: self.constraints,
                iceServers: self.options.iceServers,
            });

            onStreamSuccess(self, stream);
        }

        function onError(e) {
            onStreamError(self, e);
        }

	var mediaParams = getMediaParams(self);

	console.log("Audio constraints", mediaParams.audio);
	console.log("Video constraints", mediaParams.video);

	if (mediaParams.audio || mediaParams.video) {

            getUserMedia({
		constraints: {
                    audio: mediaParams.audio,
                video: mediaParams.video
		},
		video: mediaParams.useVideo,
		onsuccess: onSuccess,
		onerror: onError
            });
	} else {
	    onSuccess(null);
	}



        /*
        navigator.getUserMedia({
            video: self.options.useVideo,
            audio: true
        }, onSuccess, onError);
        */

    };

    // DERIVED from RTCPeerConnection-v1.5
    // 2013, @muazkh - github.com/muaz-khan
    // MIT License - https://www.webrtc-experiment.com/licence/
    // Documentation - https://github.com/muaz-khan/WebRTC-Experiment/tree/master/RTCPeerConnection
    window.moz = !!navigator.mozGetUserMedia;

    function RTCPeerConnection(options) {
	var gathering = false, done = false;

        var w = window,
        PeerConnection = w.mozRTCPeerConnection || w.webkitRTCPeerConnection,
        SessionDescription = w.mozRTCSessionDescription || w.RTCSessionDescription,
        IceCandidate = w.mozRTCIceCandidate || w.RTCIceCandidate;
	
        var STUN = {
            url: !moz ? 'stun:stun.l.google.com:19302' : 'stun:23.21.150.121'
        };

        var iceServers = null;

        if (options.iceServers) {
            var tmp = options.iceServers;

            if (typeof(tmp) === "boolean") {
                tmp = null;
            }

            if (tmp && !(typeof(tmp) == "object" && tmp.constructor === Array)) {
                console.warn("iceServers must be an array, reverting to default ice servers");
                tmp = null;
            }

            iceServers = {
                iceServers: tmp || [STUN]
            };

            if (!moz && !tmp) {
                iceServers.iceServers = [STUN];
            }
        }

        var optional = {
            optional: []
        };

        if (!moz) {
            optional.optional = [{
                DtlsSrtpKeyAgreement: true
            },
            {
                RtpDataChannels: options.onChannelMessage ? true : false
            }];
        }

        var peer = new PeerConnection(iceServers, optional);

        openOffererChannel();
        var x = 0;

	function ice_handler() {

	    done = true;
	    gathering = null;

            if (options.onICEComplete) {
                options.onICEComplete();
            }
	    
            if (options.type == "offer") {
                if ((!moz || (!options.sentICESDP && peer.localDescription.sdp.match(/a=candidate/)) && !x && options.onICESDP)) {
                    options.onICESDP(peer.localDescription);
                    //x = 1;
                    /*
                      x = 1;
                      peer.createOffer(function(sessionDescription) {
                      sessionDescription.sdp = serializeSdp(sessionDescription.sdp);
                      peer.setLocalDescription(sessionDescription);
                      if (options.onICESDP) {
                      options.onICESDP(sessionDescription);
                      }
                      }, onSdpError, constraints);
                    */
                }
            } else {
                if (!x && options.onICESDP) {
                    options.onICESDP(peer.localDescription);
                    //x = 1;
                    /*
                      x = 1;
                      peer.createAnswer(function(sessionDescription) {
                      sessionDescription.sdp = serializeSdp(sessionDescription.sdp);
                      peer.setLocalDescription(sessionDescription);
                      if (options.onICESDP) {
                      options.onICESDP(sessionDescription);
                      }
                      }, onSdpError, constraints);
                    */
                }
            }
        }

        peer.onicecandidate = function(event) {

	    if (done) {
		return;
	    }

	    if (!gathering) {
		gathering = setTimeout(ice_handler, 1000);
	    }
	    
	    if (event) {
		if (event.candidate) {
		    options.onICE(event.candidate);
		}
	    } else {
		done = true;

		if (gathering) {
		    clearTimeout(gathering);
		    gathering = null;
		}

		ice_handler();
	    }
        };

        // attachStream = MediaStream;
        if (options.attachStream) peer.addStream(options.attachStream);

        // attachStreams[0] = audio-stream;
        // attachStreams[1] = video-stream;
        // attachStreams[2] = screen-capturing-stream;
        if (options.attachStreams && options.attachStream.length) {
            var streams = options.attachStreams;
            for (var i = 0; i < streams.length; i++) {
                peer.addStream(streams[i]);
            }
        }

        peer.onaddstream = function(event) {
            var remoteMediaStream = event.stream;

            // onRemoteStreamEnded(MediaStream)
            remoteMediaStream.onended = function() {
                if (options.onRemoteStreamEnded) options.onRemoteStreamEnded(remoteMediaStream);
            };

            // onRemoteStream(MediaStream)
            if (options.onRemoteStream) options.onRemoteStream(remoteMediaStream);

            //console.debug('on:add:stream', remoteMediaStream);
        };

        var constraints = options.constraints || {
	    offerToReceiveAudio: true,
	    offerToReceiveVideo: true   
        };

        // onOfferSDP(RTCSessionDescription)
        function createOffer() {
            if (!options.onOfferSDP) return;

            peer.createOffer(function(sessionDescription) {
                sessionDescription.sdp = serializeSdp(sessionDescription.sdp);
                peer.setLocalDescription(sessionDescription);
                options.onOfferSDP(sessionDescription);
		/* old mozilla behaviour the SDP was already great right away */
                if (moz && options.onICESDP && sessionDescription.sdp.match(/a=candidate/)) {
                    options.onICESDP(sessionDescription);
		    options.sentICESDP = 1;
                }
            },
            onSdpError, constraints);
        }

        // onAnswerSDP(RTCSessionDescription)
        function createAnswer() {
            if (options.type != "answer") return;

            //options.offerSDP.sdp = addStereo(options.offerSDP.sdp);
            peer.setRemoteDescription(new SessionDescription(options.offerSDP), onSdpSuccess, onSdpError);
            peer.createAnswer(function(sessionDescription) {
                sessionDescription.sdp = serializeSdp(sessionDescription.sdp);
                peer.setLocalDescription(sessionDescription);
                if (options.onAnswerSDP) {
                    options.onAnswerSDP(sessionDescription);
                }
            },
            onSdpError, constraints);
        }

        // if Mozilla Firefox & DataChannel; offer/answer will be created later
        if ((options.onChannelMessage && !moz) || !options.onChannelMessage) {
            createOffer();
            createAnswer();
        }

        // DataChannel Bandwidth
        function setBandwidth(sdp) {
            // remove existing bandwidth lines
            sdp = sdp.replace(/b=AS([^\r\n]+\r\n)/g, '');
            sdp = sdp.replace(/a=mid:data\r\n/g, 'a=mid:data\r\nb=AS:1638400\r\n');

            return sdp;
        }

        // old: FF<>Chrome interoperability management
        function getInteropSDP(sdp) {
            var chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZ'.split(''),
            extractedChars = '';

            function getChars() {
                extractedChars += chars[parseInt(Math.random() * 40)] || '';
                if (extractedChars.length < 40) getChars();

                return extractedChars;
            }

            // usually audio-only streaming failure occurs out of audio-specific crypto line
            // a=crypto:1 AES_CM_128_HMAC_SHA1_32 --------- kAttributeCryptoVoice
            if (options.onAnswerSDP) sdp = sdp.replace(/(a=crypto:0 AES_CM_128_HMAC_SHA1_32)(.*?)(\r\n)/g, '');

            // video-specific crypto line i.e. SHA1_80
            // a=crypto:1 AES_CM_128_HMAC_SHA1_80 --------- kAttributeCryptoVideo
            var inline = getChars() + '\r\n' + (extractedChars = '');
            sdp = sdp.indexOf('a=crypto') == -1 ? sdp.replace(/c=IN/g, 'a=crypto:1 AES_CM_128_HMAC_SHA1_80 inline:' + inline + 'c=IN') : sdp;

            return sdp;
        }

        function serializeSdp(sdp) {
            //if (!moz) sdp = setBandwidth(sdp);
            //sdp = getInteropSDP(sdp);
            //console.debug(sdp);
            return sdp;
        }

        // DataChannel management
        var channel;

        function openOffererChannel() {
            if (!options.onChannelMessage || (moz && !options.onOfferSDP)) return;

            _openOffererChannel();

            if (!moz) return;
            navigator.mozGetUserMedia({
                audio: true,
                fake: true
            },
            function(stream) {
                peer.addStream(stream);
                createOffer();
            },
            useless);
        }

        function _openOffererChannel() {
            channel = peer.createDataChannel(options.channel || 'RTCDataChannel', moz ? {} : {
                reliable: false
            });

            if (moz) channel.binaryType = 'blob';

            setChannelEvents();
        }

        function setChannelEvents() {
            channel.onmessage = function(event) {
                if (options.onChannelMessage) options.onChannelMessage(event);
            };

            channel.onopen = function() {
                if (options.onChannelOpened) options.onChannelOpened(channel);
            };
            channel.onclose = function(event) {
                if (options.onChannelClosed) options.onChannelClosed(event);

                console.warn('WebRTC DataChannel closed', event);
            };
            channel.onerror = function(event) {
                if (options.onChannelError) options.onChannelError(event);

                console.error('WebRTC DataChannel error', event);
            };
        }

        if (options.onAnswerSDP && moz && options.onChannelMessage) openAnswererChannel();

        function openAnswererChannel() {
            peer.ondatachannel = function(event) {
                channel = event.channel;
                channel.binaryType = 'blob';
                setChannelEvents();
            };

            if (!moz) return;
            navigator.mozGetUserMedia({
                audio: true,
                fake: true
            },
            function(stream) {
                peer.addStream(stream);
                createAnswer();
            },
            useless);
        }

        // fake:true is also available on chrome under a flag!
        function useless() {
            log('Error in fake:true');
        }

        function onSdpSuccess() {}

        function onSdpError(e) {
            if (options.onChannelError) {
                options.onChannelError(e);
            }
            console.error('sdp error:', e);
        }

        return {
            addAnswerSDP: function(sdp, cbSuccess, cbError) {

                peer.setRemoteDescription(new SessionDescription(sdp), cbSuccess ? cbSuccess : onSdpSuccess, cbError ? cbError : onSdpError);
            },
            addICE: function(candidate) {
                peer.addIceCandidate(new IceCandidate({
                    sdpMLineIndex: candidate.sdpMLineIndex,
                    candidate: candidate.candidate
                }));
            },

            peer: peer,
            channel: channel,
            sendData: function(message) {
                if (channel) {
                    channel.send(message);
                }
            },

            stop: function() {
                peer.close();
                if (options.attachStream) {
                  if(typeof options.attachStream.stop == 'function') {
                    options.attachStream.stop();
                  } else {
                    options.attachStream.active = false;
                  }
                }
            }

        };
    }

    // getUserMedia
    var video_constraints = {
        mandatory: {},
        optional: []
    };

    function getUserMedia(options) {
        var n = navigator,
        media;
        n.getMedia = n.webkitGetUserMedia || n.mozGetUserMedia;
        n.getMedia(options.constraints || {
            audio: true,
            video: video_constraints
        },
        streaming, options.onerror ||
        function(e) {
            console.error(e);
        });

        function streaming(stream) {
            //var video = options.video;
            //var localVideo = options.localVideo;
            //if (video) {
              //  video[moz ? 'mozSrcObject' : 'src'] = moz ? stream : window.webkitURL.createObjectURL(stream);
                //video.play();
            //}

            if (options.localVideo) {
                options.localVideo[moz ? 'mozSrcObject' : 'src'] = moz ? stream : window.webkitURL.createObjectURL(stream);
		options.localVideo.style.display = 'block';
            }

            if (options.onsuccess) {
                options.onsuccess(stream);
            }

            media = stream;
        }

        return media;
    }

    $.FSRTC.resSupported = function(w, h) {
	for (var i in $.FSRTC.validRes) {
	    if ($.FSRTC.validRes[i][0] == w && $.FSRTC.validRes[i][1] == h) {
		return true;
	    }
	}

	return false;
    }

    $.FSRTC.bestResSupported = function() {
	var w = 0, h = 0;

	for (var i in $.FSRTC.validRes) {
	    if ($.FSRTC.validRes[i][0] > w && $.FSRTC.validRes[i][1] > h) {
		w = $.FSRTC.validRes[i][0];
		h = $.FSRTC.validRes[i][1];
	    }
	}

	return [w, h];
    }

    var resList = [[320, 180], [320, 240], [640, 360], [640, 480], [1280, 720], [1920, 1080]];
    var resI = 0;
    var ttl = 0;

    var checkRes = function (cam, func) {

	if (resI >= resList.length) {
            var res = {
               'validRes': $.FSRTC.validRes,
               'bestResSupported': $.FSRTC.bestResSupported()
            };
	    
	    localStorage.setItem("res_" + cam, $.toJSON(res));
	    
	    if (func) return func(res);
	    return;
	}

	var video = {
            mandatory: {},
            optional: []
        }	

	if (cam) {
	    video.optional = [{sourceId: cam}];
	}
	
	w = resList[resI][0];
	h = resList[resI][1];
	resI++;

	video.mandatory = {
	    "minWidth": w,
	    "minHeight": h,
	    "maxWidth": w,
	    "maxHeight": h
	};

	if (window.moz) {
	    video = video.mandatory;
	    if (!video.width) video.width = video.minWidth;
	    if (!video.height) video.height = video.minHeight;
	    if (!video.frameRate) video.frameRate = video.minFrameRate;
	}

	getUserMedia({
	    constraints: {
                audio: ttl++ == 0,
                video: video	    
	    },
	    onsuccess: function(e) {
		e.getTracks().forEach(function(track) {track.stop();});
		console.info(w + "x" + h + " supported."); $.FSRTC.validRes.push([w, h]); checkRes(cam, func);},
	    onerror: function(e) {console.error( w + "x" + h + " not supported."); checkRes(cam, func);}
        });
    }
    

    $.FSRTC.getValidRes = function (cam, func) {
	var used = [];
	var cached = localStorage.getItem("res_" + cam);
	
	if (cached) {
	    var cache = $.parseJSON(cached);

	    if (cache) {
		$.FSRTC.validRes = cache.validRes;
		console.log("CACHED RES FOR CAM " + cam, cache);
	    } else {
		console.error("INVALID CACHE");
	    }
	    return func ? func(cache) : null;
	}


	$.FSRTC.validRes = [];
	resI = 0;

	checkRes(cam, func);
    }

    $.FSRTC.checkPerms = function (runtime, check_audio, check_video) {
	getUserMedia({
	    constraints: {
		audio: check_audio,
		video: check_video,
	    },
	    onsuccess: function(e) {
		e.getTracks().forEach(function(track) {track.stop();});
		
		console.info("media perm init complete"); 
		if (runtime) {
                    setTimeout(runtime, 100, true);
		}
            },
	    onerror: function(e) {
		if (check_video && check_audio) {
		    console.error("error, retesting with audio params only");
		    return $.FSRTC.checkPerms(runtime, check_audio, false);
		}

		console.error("media perm init error");

		if (runtime) {
		    runtime(false)
		}
	    }
	});
    }

})(jQuery);


/*
 * Verto HTML5/Javascript Telephony Signaling and Control Protocol Stack for FreeSWITCH
 * Copyright (C) 2005-2014, Anthony Minessale II <anthm@freeswitch.org>
 *
 * Version: MPL 1.1
 *
 * The contents of this file are subject to the Mozilla Public License Version
 * 1.1 (the "License"); you may not use this file except in compliance with
 * the License. You may obtain a copy of the License at
 * http://www.mozilla.org/MPL/
 *
 * Software distributed under the License is distributed on an "AS IS" basis,
 * WITHOUT WARRANTY OF ANY KIND, either express or implied. See the License
 * for the specific language governing rights and limitations under the
 * License.
 *
 * The Original Code is Verto HTML5/Javascript Telephony Signaling and Control Protocol Stack for FreeSWITCH
 *
 * The Initial Developer of the Original Code is
 * Anthony Minessale II <anthm@freeswitch.org>
 * Portions created by the Initial Developer are Copyright (C)
 * the Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * Anthony Minessale II <anthm@freeswitch.org>
 *
 * jquery.verto.js - Main interface
 *
 */

(function($) {
    var sources = [];

    var generateGUID = (typeof(window.crypto) !== 'undefined' && typeof(window.crypto.getRandomValues) !== 'undefined') ?
    function() {
        // If we have a cryptographically secure PRNG, use that
        // http://stackoverflow.com/questions/6906916/collisions-when-generating-uuids-in-javascript
        var buf = new Uint16Array(8);
        window.crypto.getRandomValues(buf);
        var S4 = function(num) {
            var ret = num.toString(16);
            while (ret.length < 4) {
                ret = "0" + ret;
            }
            return ret;
        };
        return (S4(buf[0]) + S4(buf[1]) + "-" + S4(buf[2]) + "-" + S4(buf[3]) + "-" + S4(buf[4]) + "-" + S4(buf[5]) + S4(buf[6]) + S4(buf[7]));
    }

    :

    function() {
        // Otherwise, just use Math.random
        // http://stackoverflow.com/questions/105034/how-to-create-a-guid-uuid-in-javascript/2117523#2117523
        return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function(c) {
            var r = Math.random() * 16 | 0,
            v = c == 'x' ? r : (r & 0x3 | 0x8);
            return v.toString(16);
        });
    };

    /// MASTER OBJ
    $.verto = function(options, callbacks) {
        var verto = this;

        $.verto.saved.push(verto);

        verto.options = $.extend({
            login: null,
            passwd: null,
            socketUrl: null,
            tag: null,
	    localTag: null,
            videoParams: {},
            audioParams: {},
	    loginParams: {},
	    deviceParams: {onResCheck: null},
	    userVariables: {},
            iceServers: false,
            ringSleep: 6000,
	    sessid: null
        }, options);

	if (verto.options.deviceParams.useCamera) {
	    $.FSRTC.getValidRes(verto.options.deviceParams.useCamera, verto.options.deviceParams.onResCheck);
	}

	if (!verto.options.deviceParams.useMic) {
	    verto.options.deviceParams.useMic = "any";
	}

	if (!verto.options.deviceParams.useSpeak) {
	    verto.options.deviceParams.useSpeak = "any";
	}

	if (verto.options.sessid) {
	    verto.sessid = verto.options.sessid;
	} else {
            verto.sessid = localStorage.getItem("verto_session_uuid") || generateGUID();
	    localStorage.setItem("verto_session_uuid", verto.sessid);
	}

        verto.dialogs = {};
        verto.callbacks = callbacks || {};
        verto.eventSUBS = {};

        verto.rpcClient = new $.JsonRpcClient({
            login: verto.options.login,
            passwd: verto.options.passwd,
            socketUrl: verto.options.socketUrl,
	    loginParams: verto.options.loginParams,
	    userVariables: verto.options.userVariables,
            sessid: verto.sessid,
            onmessage: function(e) {
                return verto.handleMessage(e.eventData);
            },
            onWSConnect: function(o) {
                o.call('login', {});
            },
            onWSLogin: function(success) {
                if (verto.callbacks.onWSLogin) {
                    verto.callbacks.onWSLogin(verto, success);
                }
            },
            onWSClose: function(success) {
                if (verto.callbacks.onWSClose) {
                    verto.callbacks.onWSClose(verto, success);
                }
                verto.purge();
            }
        });

        if (verto.options.ringFile && verto.options.tag) {
            verto.ringer = $("#" + verto.options.tag);
        }

        verto.rpcClient.call('login', {});

    };


    $.verto.prototype.deviceParams = function(obj) {
        var verto = this;

	for (var i in obj) {
            verto.options.deviceParams[i] = obj[i];
	}

	if (obj.useCamera) {
	    $.FSRTC.getValidRes(verto.options.deviceParams.useCamera, obj ? obj.onResCheck : undefined);
	}
    };

    $.verto.prototype.videoParams = function(obj) {
        var verto = this;

	for (var i in obj) {
            verto.options.videoParams[i] = obj[i];
	}
    };

    $.verto.prototype.iceServers = function(obj) {
        var verto = this;
        verto.options.iceServers = obj;
    };

    $.verto.prototype.loginData = function(params) {
	var verto = this;
        verto.options.login = params.login;
        verto.options.passwd = params.passwd;
        verto.rpcClient.loginData(params);
    };

    $.verto.prototype.logout = function(msg) {
        var verto = this;
        verto.rpcClient.closeSocket();
        if (verto.callbacks.onWSClose) {
            verto.callbacks.onWSClose(verto, false);
        }
        verto.purge();
    };

    $.verto.prototype.login = function(msg) {
        var verto = this;
        verto.logout();
        verto.rpcClient.call('login', {});
    };

    $.verto.prototype.message = function(msg) {
        var verto = this;
        var err = 0;

        if (!msg.to) {
            console.error("Missing To");
            err++;
        }

        if (!msg.body) {
            console.error("Missing Body");
            err++;
        }

        if (err) {
            return false;
        }

        verto.sendMethod("verto.info", {
            msg: msg
        });

        return true;
    };

    $.verto.prototype.processReply = function(method, success, e) {
        var verto = this;
        var i;

        //console.log("Response: " + method, success, e);

        switch (method) {
        case "verto.subscribe":
            for (i in e.unauthorizedChannels) {
                drop_bad(verto, e.unauthorizedChannels[i]);
            }
            for (i in e.subscribedChannels) {
                mark_ready(verto, e.subscribedChannels[i]);
            }

            break;
        case "verto.unsubscribe":
            //console.error(e);
            break;
        }
    };

    $.verto.prototype.sendMethod = function(method, params) {
        var verto = this;

        verto.rpcClient.call(method, params,

        function(e) {
            /* Success */
            verto.processReply(method, true, e);
        },

        function(e) {
            /* Error */
            verto.processReply(method, false, e);
        });
    };

    function do_sub(verto, channel, obj) {

    }

    function drop_bad(verto, channel) {
        console.error("drop unauthorized channel: " + channel);
        delete verto.eventSUBS[channel];
    }

    function mark_ready(verto, channel) {
        for (var j in verto.eventSUBS[channel]) {
            verto.eventSUBS[channel][j].ready = true;
            console.log("subscribed to channel: " + channel);
            if (verto.eventSUBS[channel][j].readyHandler) {
                verto.eventSUBS[channel][j].readyHandler(verto, channel);
            }
        }
    }

    var SERNO = 1;

    function do_subscribe(verto, channel, subChannels, sparams) {
        var params = sparams || {};

        var local = params.local;

        var obj = {
            eventChannel: channel,
            userData: params.userData,
            handler: params.handler,
            ready: false,
            readyHandler: params.readyHandler,
            serno: SERNO++
        };

        var isnew = false;

        if (!verto.eventSUBS[channel]) {
            verto.eventSUBS[channel] = [];
            subChannels.push(channel);
            isnew = true;
        }

        verto.eventSUBS[channel].push(obj);

        if (local) {
            obj.ready = true;
            obj.local = true;
        }

        if (!isnew && verto.eventSUBS[channel][0].ready) {
            obj.ready = true;
            if (obj.readyHandler) {
                obj.readyHandler(verto, channel);
            }
        }

        return {
            serno: obj.serno,
            eventChannel: channel
        };

    }

    $.verto.prototype.subscribe = function(channel, sparams) {
        var verto = this;
        var r = [];
        var subChannels = [];
        var params = sparams || {};

        if (typeof(channel) === "string") {
            r.push(do_subscribe(verto, channel, subChannels, params));
        } else {
            for (var i in channel) {
                r.push(do_subscribe(verto, channel, subChannels, params));
            }
        }

        if (subChannels.length) {
            verto.sendMethod("verto.subscribe", {
                eventChannel: subChannels.length == 1 ? subChannels[0] : subChannels,
                subParams: params.subParams
            });
        }

        return r;
    };

    $.verto.prototype.unsubscribe = function(handle) {
        var verto = this;
        var i;

        if (!handle) {
            for (i in verto.eventSUBS) {
                if (verto.eventSUBS[i]) {
                    verto.unsubscribe(verto.eventSUBS[i]);
                }
            }
        } else {
            var unsubChannels = {};
            var sendChannels = [];
            var channel;

            if (typeof(handle) == "string") {
                delete verto.eventSUBS[handle];
                unsubChannels[handle]++;
            } else {
                for (i in handle) {
                    if (typeof(handle[i]) == "string") {
                        channel = handle[i];
                        delete verto.eventSUBS[channel];
                        unsubChannels[channel]++;
                    } else {
                        var repl = [];
                        channel = handle[i].eventChannel;

                        for (var j in verto.eventSUBS[channel]) {
                            if (verto.eventSUBS[channel][j].serno == handle[i].serno) {} else {
                                repl.push(verto.eventSUBS[channel][j]);
                            }
                        }

                        verto.eventSUBS[channel] = repl;

                        if (verto.eventSUBS[channel].length === 0) {
                            delete verto.eventSUBS[channel];
                            unsubChannels[channel]++;
                        }
                    }
                }
            }

            for (var u in unsubChannels) {
                console.log("Sending Unsubscribe for: ", u);
                sendChannels.push(u);
            }

            if (sendChannels.length) {
                verto.sendMethod("verto.unsubscribe", {
                    eventChannel: sendChannels.length == 1 ? sendChannels[0] : sendChannels
                });
            }
        }
    };

    $.verto.prototype.broadcast = function(channel, params) {
        var verto = this;
        var msg = {
            eventChannel: channel,
            data: {}
        };
        for (var i in params) {
            msg.data[i] = params[i];
        }
        verto.sendMethod("verto.broadcast", msg);
    };

    $.verto.prototype.purge = function(callID) {
        var verto = this;
        var x = 0;
        var i;

        for (i in verto.dialogs) {
            if (!x) {
                console.log("purging dialogs");
            }
            x++;
            verto.dialogs[i].setState($.verto.enum.state.purge);
        }

        for (i in verto.eventSUBS) {
            if (verto.eventSUBS[i]) {
                console.log("purging subscription: " + i);
                delete verto.eventSUBS[i];
            }
        }
    };

    $.verto.prototype.hangup = function(callID) {
        var verto = this;
        if (callID) {
            var dialog = verto.dialogs[callID];

            if (dialog) {
                dialog.hangup();
            }
        } else {
            for (var i in verto.dialogs) {
                verto.dialogs[i].hangup();
            }
        }
    };

    $.verto.prototype.newCall = function(args, callbacks) {
        var verto = this;

        if (!verto.rpcClient.socketReady()) {
            console.error("Not Connected...");
            return;
        }

        var dialog = new $.verto.dialog($.verto.enum.direction.outbound, this, args);

        dialog.invite();

        if (callbacks) {
            dialog.callbacks = callbacks;
        }

        return dialog;
    };

    $.verto.prototype.handleMessage = function(data) {
        var verto = this;

        if (!(data && data.method)) {
            console.error("Invalid Data", data);
            return;
        }

        if (data.params.callID) {
            var dialog = verto.dialogs[data.params.callID];
	    
	    if (data.method === "verto.attach" && dialog) {
		delete dialog.verto.dialogs[dialog.callID];
		dialog.rtc.stop();
		dialog = null;
	    }

            if (dialog) {

                switch (data.method) {
                case 'verto.bye':
                    dialog.hangup(data.params);
                    break;
                case 'verto.answer':
                    dialog.handleAnswer(data.params);
                    break;
                case 'verto.media':
                    dialog.handleMedia(data.params);
                    break;
                case 'verto.display':
                    dialog.handleDisplay(data.params);
                    break;
                case 'verto.info':
                    dialog.handleInfo(data.params);
                    break;
                default:
                    console.debug("INVALID METHOD OR NON-EXISTANT CALL REFERENCE IGNORED", dialog, data.method);
                    break;
                }
            } else {

                switch (data.method) {
                case 'verto.attach':
                    data.params.attach = true;

                    if (data.params.sdp && data.params.sdp.indexOf("m=video") > 0) {
                        data.params.useVideo = true;
                    }

                    if (data.params.sdp && data.params.sdp.indexOf("stereo=1") > 0) {
                        data.params.useStereo = true;
                    }

                    dialog = new $.verto.dialog($.verto.enum.direction.inbound, verto, data.params);
                    dialog.setState($.verto.enum.state.recovering);

                    break;
                case 'verto.invite':

                    if (data.params.sdp && data.params.sdp.indexOf("m=video") > 0) {
                        data.params.wantVideo = true;
                    }

                    if (data.params.sdp && data.params.sdp.indexOf("stereo=1") > 0) {
                        data.params.useStereo = true;
                    }

                    dialog = new $.verto.dialog($.verto.enum.direction.inbound, verto, data.params);
                    break;
                default:
                    console.debug("INVALID METHOD OR NON-EXISTANT CALL REFERENCE IGNORED");
                    break;
                }
            }

            return {
                method: data.method
            };
        } else {
            switch (data.method) {
            case 'verto.punt':
                verto.purge();
		verto.logout();
		break;
            case 'verto.event':
                var list = null;
                var key = null;

                if (data.params) {
                    key = data.params.eventChannel;
                }

                if (key) {
                    list = verto.eventSUBS[key];

                    if (!list) {
                        list = verto.eventSUBS[key.split(".")[0]];
                    }
                }

                if (!list && key && key === verto.sessid) {
                    if (verto.callbacks.onMessage) {
                        verto.callbacks.onMessage(verto, null, $.verto.enum.message.pvtEvent, data.params);
                    }
                } else if (!list && key && verto.dialogs[key]) {
                    verto.dialogs[key].sendMessage($.verto.enum.message.pvtEvent, data.params);
                } else if (!list) {
                    if (!key) {
                        key = "UNDEFINED";
                    }
                    console.error("UNSUBBED or invalid EVENT " + key + " IGNORED");
                } else {
                    for (var i in list) {
                        var sub = list[i];

                        if (!sub || !sub.ready) {
                            console.error("invalid EVENT for " + key + " IGNORED");
                        } else if (sub.handler) {
                            sub.handler(verto, data.params, sub.userData);
                        } else if (verto.callbacks.onEvent) {
                            verto.callbacks.onEvent(verto, data.params, sub.userData);
                        } else {
                            console.log("EVENT:", data.params);
                        }
                    }
                }

                break;

            case "verto.info":
                if (verto.callbacks.onMessage) {
                    verto.callbacks.onMessage(verto, null, $.verto.enum.message.info, data.params.msg);
                }
                //console.error(data);
                console.debug("MESSAGE from: " + data.params.msg.from, data.params.msg.body);

                break;

            default:
                console.error("INVALID METHOD OR NON-EXISTANT CALL REFERENCE IGNORED", data.method);
                break;
            }
        }
    };

    var del_array = function(array, name) {
        var r = [];
        var len = array.length;

        for (var i = 0; i < len; i++) {
            if (array[i] != name) {
                r.push(array[i]);
            }
        }

        return r;
    };

    var hashArray = function() {
        var vha = this;

        var hash = {};
        var array = [];

        vha.reorder = function(a) {
            array = a;
            var h = hash;
            hash = {};

            var len = array.length;

            for (var i = 0; i < len; i++) {
                var key = array[i];
                if (h[key]) {
                    hash[key] = h[key];
                    delete h[key];
                }
            }
            h = undefined;
        };

        vha.clear = function() {
            hash = undefined;
            array = undefined;
            hash = {};
            array = [];
        };

        vha.add = function(name, val, insertAt) {
            var redraw = false;

            if (!hash[name]) {
                if (insertAt === undefined || insertAt < 0 || insertAt >= array.length) {
                    array.push(name);
                } else {
                    var x = 0;
                    var n = [];
                    var len = array.length;

                    for (var i = 0; i < len; i++) {
                        if (x++==insertAt) {
                            n.push(name);
                        }
                        n.push(array[i]);
                    }

                    array = undefined;
                    array = n;
                    n = undefined;
                    redraw = true;
                }
            }

            hash[name] = val;

            return redraw;
        };

        vha.del = function(name) {
            var r = false;

            if (hash[name]) {
                array = del_array(array, name);
                delete hash[name];
                r = true;
            } else {
                console.error("can't del nonexistant key " + name);
            }

            return r;
        };

        vha.get = function(name) {
            return hash[name];
        };

        vha.order = function() {
            return array;
        };

        vha.hash = function() {
            return hash;
        };

        vha.indexOf = function(name) {
            var len = array.length;

            for (var i = 0; i < len; i++) {
                if (array[i] == name) {
                    return i;
                }
            }
        };

        vha.arrayLen = function() {
            return array.length;
        };

        vha.asArray = function() {
            var r = [];

            var len = array.length;

            for (var i = 0; i < len; i++) {
                var key = array[i];
                r.push(hash[key]);
            }

            return r;
        };

        vha.each = function(cb) {
            var len = array.length;

            for (var i = 0; i < len; i++) {
                cb(array[i], hash[array[i]]);
            }
        };

        vha.dump = function(html) {
            var str = "";

            vha.each(function(name, val) {
                str += "name: " + name + " val: " + JSON.stringify(val) + (html ? "<br>" : "\n");
            });

            return str;
        };

    };

    $.verto.liveArray = function(verto, context, name, config) {
        var la = this;
        var lastSerno = 0;
        var binding = null;
        var user_obj = config.userObj;
        var local = false;

        // Inherit methods of hashArray
        hashArray.call(la);

        // Save the hashArray add, del, reorder, clear methods so we can make our own.
        la._add = la.add;
        la._del = la.del;
        la._reorder = la.reorder;
        la._clear = la.clear;

        la.context = context;
        la.name = name;
        la.user_obj = user_obj;

        la.verto = verto;
        la.broadcast = function(channel, obj) {
            verto.broadcast(channel, obj);
        };
        la.errs = 0;

        la.clear = function() {
            la._clear();
            lastSerno = 0;

            if (la.onChange) {
                la.onChange(la, {
                    action: "clear"
                });
            }
        };

        la.checkSerno = function(serno) {
            if (serno < 0) {
                return true;
            }

            if (lastSerno > 0 && serno != (lastSerno + 1)) {
                if (la.onErr) {
                    la.onErr(la, {
                        lastSerno: lastSerno,
                        serno: serno
                    });
                }
                la.errs++;
                console.debug(la.errs);
                if (la.errs < 3) {
                    la.bootstrap(la.user_obj);
                }
                return false;
            } else {
                lastSerno = serno;
                return true;
            }
        };

        la.reorder = function(serno, a) {
            if (la.checkSerno(serno)) {
                la._reorder(a);
                if (la.onChange) {
                    la.onChange(la, {
                        serno: serno,
                        action: "reorder"
                    });
                }
            }
        };

        la.init = function(serno, val, key, index) {
            if (key === null || key === undefined) {
                key = serno;
            }
            if (la.checkSerno(serno)) {
                if (la.onChange) {
                    la.onChange(la, {
                        serno: serno,
                        action: "init",
                        index: index,
                        key: key,
                        data: val
                    });
                }
            }
        };

        la.bootObj = function(serno, val) {
            if (la.checkSerno(serno)) {

                //la.clear();
                for (var i in val) {
                    la._add(val[i][0], val[i][1]);
                }

                if (la.onChange) {
                    la.onChange(la, {
                        serno: serno,
                        action: "bootObj",
                        data: val,
                        redraw: true
                    });
                }
            }
        };

        // @param serno  La is the serial number for la particular request.
        // @param key    If looking at it as a hash table, la represents the key in the hashArray object where you want to store the val object.
        // @param index  If looking at it as an array, la represents the position in the array where you want to store the val object.
        // @param val    La is the object you want to store at the key or index location in the hash table / array.
        la.add = function(serno, val, key, index) {
            if (key === null || key === undefined) {
                key = serno;
            }
            if (la.checkSerno(serno)) {
                var redraw = la._add(key, val, index);
                if (la.onChange) {
                    la.onChange(la, {
                        serno: serno,
                        action: "add",
                        index: index,
                        key: key,
                        data: val,
                        redraw: redraw
                    });
                }
            }
        };

        la.modify = function(serno, val, key, index) {
            if (key === null || key === undefined) {
                key = serno;
            }
            if (la.checkSerno(serno)) {
                la._add(key, val, index);
                if (la.onChange) {
                    la.onChange(la, {
                        serno: serno,
                        action: "modify",
                        key: key,
                        data: val,
                        index: index
                    });
                }
            }
        };

        la.del = function(serno, key, index) {
            if (key === null || key === undefined) {
                key = serno;
            }
            if (la.checkSerno(serno)) {
                if (index === null || index < 0 || index === undefined) {
                    index = la.indexOf(key);
                }
                var ok = la._del(key);

                if (ok && la.onChange) {
                    la.onChange(la, {
                        serno: serno,
                        action: "del",
                        key: key,
                        index: index
                    });
                }
            }
        };

        var eventHandler = function(v, e, la) {
            var packet = e.data;

            //console.error("READ:", packet);

            if (packet.name != la.name) {
                return;
            }

            switch (packet.action) {

            case "init":
                la.init(packet.wireSerno, packet.data, packet.hashKey, packet.arrIndex);
                break;

            case "bootObj":
                la.bootObj(packet.wireSerno, packet.data);
                break;
            case "add":
                la.add(packet.wireSerno, packet.data, packet.hashKey, packet.arrIndex);
                break;

            case "modify":
                if (! (packet.arrIndex || packet.hashKey)) {
                    console.error("Invalid Packet", packet);
                } else {
                    la.modify(packet.wireSerno, packet.data, packet.hashKey, packet.arrIndex);
                }
                break;
            case "del":
                if (! (packet.arrIndex || packet.hashKey)) {
                    console.error("Invalid Packet", packet);
                } else {
                    la.del(packet.wireSerno, packet.hashKey, packet.arrIndex);
                }
                break;

            case "clear":
                la.clear();
                break;

            case "reorder":
                la.reorder(packet.wireSerno, packet.order);
                break;

            default:
                if (la.checkSerno(packet.wireSerno)) {
                    if (la.onChange) {
                        la.onChange(la, {
                            serno: packet.wireSerno,
                            action: packet.action,
                            data: packet.data
                        });
                    }
                }
                break;
            }
        };

        if (la.context) {
            binding = la.verto.subscribe(la.context, {
                handler: eventHandler,
                userData: la,
                subParams: config.subParams
            });
        }

        la.destroy = function() {
            la._clear();
            la.verto.unsubscribe(binding);
        };

        la.sendCommand = function(cmd, obj) {
            var self = la;
            self.broadcast(self.context, {
                liveArray: {
                    command: cmd,
                    context: self.context,
                    name: self.name,
                    obj: obj
                }
            });
        };

        la.bootstrap = function(obj) {
            var self = la;
            la.sendCommand("bootstrap", obj);
            //self.heartbeat();
        };

        la.changepage = function(obj) {
            var self = la;
            self.clear();
            self.broadcast(self.context, {
                liveArray: {
                    command: "changepage",
                    context: la.context,
                    name: la.name,
                    obj: obj
                }
            });
        };

        la.heartbeat = function(obj) {
            var self = la;

            var callback = function() {
                self.heartbeat.call(self, obj);
            };
            self.broadcast(self.context, {
                liveArray: {
                    command: "heartbeat",
                    context: self.context,
                    name: self.name,
                    obj: obj
                }
            });
            self.hb_pid = setTimeout(callback, 30000);
        };

        la.bootstrap(la.user_obj);
    };

    $.verto.liveTable = function(verto, context, name, jq, config) {
        var dt;
        var la = new $.verto.liveArray(verto, context, name, {
            subParams: config.subParams
        });
        var lt = this;

        lt.liveArray = la;
        lt.dataTable = dt;
        lt.verto = verto;

        lt.destroy = function() {
            if (dt) {
                dt.fnDestroy();
            }
            if (la) {
                la.destroy();
            }

            dt = null;
            la = null;
        };

        la.onErr = function(obj, args) {
            console.error("Error: ", obj, args);
        };

        la.onChange = function(obj, args) {
            var index = 0;
            var iserr = 0;

            if (!dt) {
                if (!config.aoColumns) {
                    if (args.action != "init") {
                        return;
                    }

                    config.aoColumns = [];

                    for (var i in args.data) {
                        config.aoColumns.push({
                            "sTitle": args.data[i]
                        });
                    }
                }

                dt = jq.dataTable(config);
            }

            if (dt && (args.action == "del" || args.action == "modify")) {
                index = args.index;

                if (index === undefined && args.key) {
                    index = la.indexOf(args.key);
                }

                if (index === undefined) {
                    console.error("INVALID PACKET Missing INDEX\n", args);
                    return;
                }
            }

            if (config.onChange) {
                config.onChange(obj, args);
            }

            try {
                switch (args.action) {
                case "bootObj":
                    if (!args.data) {
                        console.error("missing data");
                        return;
                    }
                    dt.fnClearTable();
                    dt.fnAddData(obj.asArray());
                    dt.fnAdjustColumnSizing();
                    break;
                case "add":
                    if (!args.data) {
                        console.error("missing data");
                        return;
                    }
                    if (args.redraw > -1) {
                        // specific position, more costly
                        dt.fnClearTable();
                        dt.fnAddData(obj.asArray());
                    } else {
                        dt.fnAddData(args.data);
                    }
                    dt.fnAdjustColumnSizing();
                    break;
                case "modify":
                    if (!args.data) {
                        return;
                    }
                    //console.debug(args, index);
                    dt.fnUpdate(args.data, index);
                    dt.fnAdjustColumnSizing();
                    break;
                case "del":
                    dt.fnDeleteRow(index);
                    dt.fnAdjustColumnSizing();
                    break;
                case "clear":
                    dt.fnClearTable();
                    break;
                case "reorder":
                    // specific position, more costly
                    dt.fnClearTable();
                    dt.fnAddData(obj.asArray());
                    break;
                case "hide":
                    jq.hide();
                    break;

                case "show":
                    jq.show();
                    break;

                }
            } catch(err) {
                console.error("ERROR: " + err);
                iserr++;
            }

            if (iserr) {
                obj.errs++;
                if (obj.errs < 3) {
                    obj.bootstrap(obj.user_obj);
                }
            } else {
                obj.errs = 0;
            }

        };

        la.onChange(la, {
            action: "init"
        });

    };

    var CONFMAN_SERNO = 1;

    /*
        Conference Manager without jQuery table.
     */

    $.verto.conf = function(verto, params) {
        var conf = this;

        conf.params = $.extend({
            dialog: null,
            hasVid: false,
            laData: null,
            onBroadcast: null,
            onLaChange: null,
            onLaRow: null
        }, params);

        conf.verto = verto;
        conf.serno = CONFMAN_SERNO++;

        createMainModeratorMethods();

        verto.subscribe(conf.params.laData.modChannel, {
            handler: function(v, e) {
                if (conf.params.onBroadcast) {
                    conf.params.onBroadcast(verto, conf, e.data);
                }
            }
        });

        verto.subscribe(conf.params.laData.chatChannel, {
            handler: function(v, e) {
                if (typeof(conf.params.chatCallback) === "function") {
                    conf.params.chatCallback(v,e);
                }
            }
        });
    };

    $.verto.conf.prototype.modCommand = function(cmd, id, value) {
        var conf = this;

        conf.verto.rpcClient.call("verto.broadcast", {
            "eventChannel": conf.params.laData.modChannel,
            "data": {
                "application": "conf-control",
                "command": cmd,
                "id": id,
                "value": value
            }
        });
    };

    $.verto.conf.prototype.destroy = function() {
        var conf = this;

        conf.destroyed = true;
        conf.params.onBroadcast(conf.verto, conf, 'destroy');

        if (conf.params.laData.modChannel) {
            conf.verto.unsubscribe(conf.params.laData.modChannel);
        }

        if (conf.params.laData.chatChannel) {
            conf.verto.unsubscribe(conf.params.laData.chatChannel);
        }
    };

    function createMainModeratorMethods() {
        $.verto.conf.prototype.listVideoLayouts = function() {
            this.modCommand("list-videoLayouts", null, null);
        };

        $.verto.conf.prototype.play = function(file) {
            this.modCommand("play", null, file);
        };

        $.verto.conf.prototype.stop = function() {
            this.modCommand("stop", null, "all");
        };

        $.verto.conf.prototype.record = function(file) {
            this.modCommand("recording", null, ["start", file]);
        };

        $.verto.conf.prototype.stopRecord = function() {
            this.modCommand("recording", null, ["stop", "all"]);
        };

        $.verto.conf.prototype.snapshot = function(file) {
            if (!this.params.hasVid) {
                throw 'Conference has no video';
            }
            this.modCommand("vid-write-png", null, file);
        };

        $.verto.conf.prototype.setVideoLayout = function(layout) {
            if (!this.params.hasVid) {
                throw 'Conference has no video';
            }
            this.modCommand("vid-layout", null, layout);
        };

        $.verto.conf.prototype.kick = function(memberID) {
            this.modCommand("kick", parseInt(memberID));
        };

        $.verto.conf.prototype.muteMic = function(memberID) {
            this.modCommand("tmute", parseInt(memberID));
        };

        $.verto.conf.prototype.muteVideo = function(memberID) {
            if (!this.params.hasVid) {
                throw 'Conference has no video';
            }
            this.modCommand("tvmute", parseInt(memberID));
        };

        $.verto.conf.prototype.presenter = function(memberID) {
            if (!this.params.hasVid) {
                throw 'Conference has no video';
            }
            this.modCommand("vid-res-id", parseInt(memberID), "presenter");
        };

        $.verto.conf.prototype.videoFloor = function(memberID) {
            if (!this.params.hasVid) {
                throw 'Conference has no video';
            }
            this.modCommand("vid-floor", parseInt(memberID), "force");
        };

        $.verto.conf.prototype.banner = function(memberID, text) {
            if (!this.params.hasVid) {
                throw 'Conference has no video';
            }
            this.modCommand("vid-banner", parseInt(memberID), escape(text));
        };

        $.verto.conf.prototype.volumeDown = function(memberID) {
            if (!this.params.hasVid) {
                throw 'Conference has no video';
            }
            this.modCommand("volume_in", parseInt(memberID), "down");
        };

        $.verto.conf.prototype.volumeUp = function(memberID) {
            if (!this.params.hasVid) {
                throw 'Conference has no video';
            }
            this.modCommand("volume_in", parseInt(memberID), "up");
        };

        $.verto.conf.prototype.transfer = function(memberID, exten) {
            if (!this.params.hasVid) {
                throw 'Conference has no video';
            }
            this.modCommand("transfer", parseInt(memberID), exten);
        };

        $.verto.conf.prototype.sendChat = function(message, type) {
            var conf = this;
            conf.verto.rpcClient.call("verto.broadcast", {
                "eventChannel": conf.params.laData.chatChannel,
                "data": {
                    "action": "send",
                    "message": message,
                    "type": type
                }
            });
        };
            
    }

    $.verto.modfuncs = {};

    $.verto.confMan = function(verto, params) {
        var confMan = this;

        confMan.params = $.extend({
            tableID: null,
            statusID: null,
            mainModID: null,
            dialog: null,
            hasVid: false,
            laData: null,
            onBroadcast: null,
            onLaChange: null,
            onLaRow: null
        }, params);

        confMan.verto = verto;
        confMan.serno = CONFMAN_SERNO++;
	confMan.canvasCount = confMan.params.laData.canvasCount;
	
        function genMainMod(jq) {
            var play_id = "play_" + confMan.serno;
            var stop_id = "stop_" + confMan.serno;
            var recording_id = "recording_" + confMan.serno;
            var snapshot_id = "snapshot_" + confMan.serno;
            var rec_stop_id = "recording_stop" + confMan.serno;
            var div_id = "confman_" + confMan.serno;

            var html =  "<div id='" + div_id + "'><br>" +
		"<button class='ctlbtn' id='" + play_id + "'>Play</button>" +
		"<button class='ctlbtn' id='" + stop_id + "'>Stop</button>" +
		"<button class='ctlbtn' id='" + recording_id + "'>Record</button>" +
		"<button class='ctlbtn' id='" + rec_stop_id + "'>Record Stop</button>" +
		(confMan.params.hasVid ? "<button class='ctlbtn' id='" + snapshot_id + "'>PNG Snapshot</button>" : "") +
		"<br><br></div>";

            jq.html(html);

	    $.verto.modfuncs.change_video_layout = function(id, canvas_id) {	    
		var val = $("#" + id + " option:selected").text();
		if (val !== "none") {
                    confMan.modCommand("vid-layout", null, [val, canvas_id]);
		}
	    };

	    if (confMan.params.hasVid) {
		for (var j = 0; j < confMan.canvasCount; j++) {
		    var vlayout_id = "confman_vid_layout_" + j + "_" + confMan.serno;
		    var vlselect_id = "confman_vl_select_" + j + "_" + confMan.serno;
		

		    var vlhtml =  "<div id='" + vlayout_id + "'><br>" +
			"<b>Video Layout Canvas " + (j+1) + 
			"</b> <select onChange='$.verto.modfuncs.change_video_layout(\"" + vlayout_id + "\", \"" + j + "\")' id='" + vlselect_id + "'></select> " +
			"<br><br></div>";
		    jq.append(vlhtml);
		}

		$("#" + snapshot_id).click(function() {
                    var file = prompt("Please enter file name", "");
		    if (file) {
			confMan.modCommand("vid-write-png", null, file);
		    }
		});
	    }

            $("#" + play_id).click(function() {
                var file = prompt("Please enter file name", "");
		if (file) {
                    confMan.modCommand("play", null, file);
		}
            });

            $("#" + stop_id).click(function() {
                confMan.modCommand("stop", null, "all");
            });

            $("#" + recording_id).click(function() {
                var file = prompt("Please enter file name", "");
		if (file) {
                    confMan.modCommand("recording", null, ["start", file]);
		}
            });

            $("#" + rec_stop_id).click(function() {
                confMan.modCommand("recording", null, ["stop", "all"]);
            });

        }

        function genControls(jq, rowid) {
            var x = parseInt(rowid);
            var kick_id = "kick_" + x;
            var canvas_in_next_id = "canvas_in_next_" + x;
            var canvas_in_prev_id = "canvas_in_prev_" + x;
            var canvas_out_next_id = "canvas_out_next_" + x;
            var canvas_out_prev_id = "canvas_out_prev_" + x;

            var canvas_in_set_id = "canvas_in_set_" + x;
            var canvas_out_set_id = "canvas_out_set_" + x;

            var layer_set_id = "layer_set_" + x;
            var layer_next_id = "layer_next_" + x;
            var layer_prev_id = "layer_prev_" + x;
	    
            var tmute_id = "tmute_" + x;
            var tvmute_id = "tvmute_" + x;
            var vbanner_id = "vbanner_" + x;
            var tvpresenter_id = "tvpresenter_" + x;
            var tvfloor_id = "tvfloor_" + x;
            var box_id = "box_" + x;
            var volup_id = "volume_in_up" + x;
            var voldn_id = "volume_in_dn" + x;
            var transfer_id = "transfer" + x;
	    

            var html = "<div id='" + box_id + "'>";

	    html += "<b>General Controls</b><hr noshade>";

            html += "<button class='ctlbtn' id='" + kick_id + "'>Kick</button>" +
                "<button class='ctlbtn' id='" + tmute_id + "'>Mute</button>" +
                "<button class='ctlbtn' id='" + voldn_id + "'>Vol -</button>" +
                "<button class='ctlbtn' id='" + volup_id + "'>Vol +</button>" +
                "<button class='ctlbtn' id='" + transfer_id + "'>Transfer</button>";
		
	    if (confMan.params.hasVid) {
		html += "<br><br><b>Video Controls</b><hr noshade>";


                html += "<button class='ctlbtn' id='" + tvmute_id + "'>VMute</button>" +
                    "<button class='ctlbtn' id='" + tvpresenter_id + "'>Presenter</button>" +
                    "<button class='ctlbtn' id='" + tvfloor_id + "'>Vid Floor</button>" +
                    "<button class='ctlbtn' id='" + vbanner_id + "'>Banner</button>";

		if (confMan.canvasCount > 1) {
                    html += "<br><br><b>Canvas Controls</b><hr noshade>" +
			"<button class='ctlbtn' id='" + canvas_in_set_id + "'>Set Input Canvas</button>" +
			"<button class='ctlbtn' id='" + canvas_in_prev_id + "'>Prev Input Canvas</button>" +
			"<button class='ctlbtn' id='" + canvas_in_next_id + "'>Next Input Canvas</button>" +
			
		    "<br>" +
			
		    "<button class='ctlbtn' id='" + canvas_out_set_id + "'>Set Watching Canvas</button>" +
			"<button class='ctlbtn' id='" + canvas_out_prev_id + "'>Prev Watching Canvas</button>" +
			"<button class='ctlbtn' id='" + canvas_out_next_id + "'>Next Watching Canvas</button>";
		}
		
		html += "<br>" +

                "<button class='ctlbtn' id='" + layer_set_id + "'>Set Layer</button>" +
                    "<button class='ctlbtn' id='" + layer_prev_id + "'>Prev Layer</button>" +
                    "<button class='ctlbtn' id='" + layer_next_id + "'>Next Layer</button>" +



                    "</div>";
            }

            jq.html(html);


            if (!jq.data("mouse")) {
                $("#" + box_id).hide();
            }

            jq.mouseover(function(e) {
                jq.data({"mouse": true});
                $("#" + box_id).show();
            });

            jq.mouseout(function(e) {
                jq.data({"mouse": false});
                $("#" + box_id).hide();
            });

            $("#" + transfer_id).click(function() {
                var xten = prompt("Enter Extension");
		if (xten) {
                    confMan.modCommand("transfer", x, xten);
		}
            });

            $("#" + kick_id).click(function() {
                confMan.modCommand("kick", x);
            });


            $("#" + layer_set_id).click(function() {
                var cid = prompt("Please enter layer ID", "");
		if (cid) {
                    confMan.modCommand("vid-layer", x, cid);
		}
            });

            $("#" + layer_next_id).click(function() {
                confMan.modCommand("vid-layer", x, "next");
            });
            $("#" + layer_prev_id).click(function() {
                confMan.modCommand("vid-layer", x, "prev");
            });

            $("#" + canvas_in_set_id).click(function() {
                var cid = prompt("Please enter canvas ID", "");
		if (cid) {
                    confMan.modCommand("vid-canvas", x, cid);
		}
            });

            $("#" + canvas_out_set_id).click(function() {
                var cid = prompt("Please enter canvas ID", "");
		if (cid) {
                    confMan.modCommand("vid-watching-canvas", x, cid);
		}
            });

            $("#" + canvas_in_next_id).click(function() {
                confMan.modCommand("vid-canvas", x, "next");
            });
            $("#" + canvas_in_prev_id).click(function() {
                confMan.modCommand("vid-canvas", x, "prev");
            });


            $("#" + canvas_out_next_id).click(function() {
                confMan.modCommand("vid-watching-canvas", x, "next");
            });
            $("#" + canvas_out_prev_id).click(function() {
                confMan.modCommand("vid-watching-canvas", x, "prev");
            });
	    
            $("#" + tmute_id).click(function() {
                confMan.modCommand("tmute", x);
            });

	    if (confMan.params.hasVid) {
		$("#" + tvmute_id).click(function() {
                    confMan.modCommand("tvmute", x);
		});
		$("#" + tvpresenter_id).click(function() {
                    confMan.modCommand("vid-res-id", x, "presenter");
		});
		$("#" + tvfloor_id).click(function() {
                    confMan.modCommand("vid-floor", x, "force");
		});
		$("#" + vbanner_id).click(function() {
                    var text = prompt("Please enter text", "");
		    if (text) {
			confMan.modCommand("vid-banner", x, escape(text));
		    }
		});
	    }

            $("#" + volup_id).click(function() {
                confMan.modCommand("volume_in", x, "up");
            });

            $("#" + voldn_id).click(function() {
                confMan.modCommand("volume_in", x, "down");
            });

            return html;
        }

        var atitle = "";
        var awidth = 0;

        //$(".jsDataTable").width(confMan.params.hasVid ? "900px" : "800px");

	verto.subscribe(confMan.params.laData.chatChannel, {
	    handler: function(v, e) {
		if (typeof(confMan.params.chatCallback) === "function") {
		    confMan.params.chatCallback(v,e);
		}
	    }
	});

        if (confMan.params.laData.role === "moderator") {
            atitle = "Action";
            awidth = 600;

            if (confMan.params.mainModID) {
                genMainMod($(confMan.params.mainModID));
                $(confMan.params.displayID).html("Moderator Controls Ready<br><br>");
            } else {
                $(confMan.params.mainModID).html("");
            }

            verto.subscribe(confMan.params.laData.modChannel, {
                handler: function(v, e) {
                    //console.error("MODDATA:", e.data);
                    if (confMan.params.onBroadcast) {
                        confMan.params.onBroadcast(verto, confMan, e.data);
                    }

		    if (e.data["conf-command"] === "list-videoLayouts") {
			for (var j = 0; j < confMan.canvasCount; j++) {
			    var vlselect_id = "#confman_vl_select_" + j + "_" + confMan.serno;
			    var vlayout_id = "#confman_vid_layout_" + j + "_" + confMan.serno;
			    
			    var x = 0;
			    var options;
			    
			    $(vlselect_id).selectmenu({});
			    $(vlselect_id).selectmenu("enable");
			    $(vlselect_id).empty();
			    
			    $(vlselect_id).append(new Option("Choose a Layout", "none"));

			    if (e.data.responseData) {
				options = e.data.responseData.sort();

				for (var i in options) {
				    $(vlselect_id).append(new Option(options[i], options[i]));
				    x++;
				}
			    }

			    if (x) {
				$(vlselect_id).selectmenu('refresh', true);
			    } else {
				$(vlayout_id).hide();
			    }
			}
		    } else {

			if (!confMan.destroyed && confMan.params.displayID) {
                            $(confMan.params.displayID).html(e.data.response + "<br><br>");
                            if (confMan.lastTimeout) {
				clearTimeout(confMan.lastTimeout);
				confMan.lastTimeout = 0;
                            }
                            confMan.lastTimeout = setTimeout(function() { $(confMan.params.displayID).html(confMan.destroyed ? "" : "Moderator Controls Ready<br><br>");}, 4000);
			}
		    }
                }
            });


	    if (confMan.params.hasVid) {
		confMan.modCommand("list-videoLayouts", null, null);
	    }
        }

        var row_callback = null;

        if (confMan.params.laData.role === "moderator") {
            row_callback = function(nRow, aData, iDisplayIndex, iDisplayIndexFull) {
                if (!aData[5]) {
                    var $row = $('td:eq(5)', nRow);
                    genControls($row, aData);

                    if (confMan.params.onLaRow) {
                        confMan.params.onLaRow(verto, confMan, $row, aData);
                    }
                }
            };
        }

        confMan.lt = new $.verto.liveTable(verto, confMan.params.laData.laChannel, confMan.params.laData.laName, $(confMan.params.tableID), {
            subParams: {
                callID: confMan.params.dialog ? confMan.params.dialog.callID : null
            },

            "onChange": function(obj, args) {
                $(confMan.params.statusID).text("Conference Members: " + " (" + obj.arrayLen() + " Total)");
                if (confMan.params.onLaChange) {
                    confMan.params.onLaChange(verto, confMan, $.verto.enum.confEvent.laChange, obj, args);
                }
            },

            "aaData": [],
            "aoColumns": [
                {
                    "sTitle": "ID",
                    "sWidth": "50"
                },
                {
                    "sTitle": "Number",
		    "sWidth": "250"
                },
                {
                    "sTitle": "Name",
		    "sWidth": "250"
                },
                {
                    "sTitle": "Codec",
                    "sWidth": "100"
                },
                {
                    "sTitle": "Status",
                    "sWidth": confMan.params.hasVid ? "200px" : "150px"
                },
                {
                    "sTitle": atitle,
                    "sWidth": awidth,
                }
            ],
            "bAutoWidth": true,
            "bDestroy": true,
            "bSort": false,
            "bInfo": false,
            "bFilter": false,
            "bLengthChange": false,
            "bPaginate": false,
            "iDisplayLength": 1400,

            "oLanguage": {
                "sEmptyTable": "The Conference is Empty....."
            },

            "fnRowCallback": row_callback

        });
    };

    $.verto.confMan.prototype.modCommand = function(cmd, id, value) {
        var confMan = this;

        confMan.verto.rpcClient.call("verto.broadcast", {
            "eventChannel": confMan.params.laData.modChannel,
            "data": {
		"application": "conf-control",
		"command": cmd,
		"id": id,
		"value": value
            }
	});
    };

    $.verto.confMan.prototype.sendChat = function(message, type) {
        var confMan = this;
        confMan.verto.rpcClient.call("verto.broadcast", {
            "eventChannel": confMan.params.laData.chatChannel,
            "data": {
		"action": "send",
		"message": message,
		"type": type
            }
	});
    };


    $.verto.confMan.prototype.destroy = function() {
        var confMan = this;

        confMan.destroyed = true;

        if (confMan.lt) {
            confMan.lt.destroy();
        }

        if (confMan.params.laData.chatChannel) {
            confMan.verto.unsubscribe(confMan.params.laData.chatChannel);
        }

        if (confMan.params.laData.modChannel) {
            confMan.verto.unsubscribe(confMan.params.laData.modChannel);
        }

        if (confMan.params.mainModID) {
            $(confMan.params.mainModID).html("");
        }
    };

    $.verto.dialog = function(direction, verto, params) {
        var dialog = this;

        dialog.params = $.extend({
            useVideo: verto.options.useVideo,
            useStereo: verto.options.useStereo,
	    screenShare: false,
	    useCamera: verto.options.deviceParams.useCamera,
	    useMic: verto.options.deviceParams.useMic,
	    useSpeak: verto.options.deviceParams.useSpeak,
            tag: verto.options.tag,
            localTag: verto.options.localTag,
            login: verto.options.login,
	    videoParams: verto.options.videoParams
        }, params);
	
        dialog.verto = verto;
        dialog.direction = direction;
        dialog.lastState = null;
        dialog.state = dialog.lastState = $.verto.enum.state.new;
        dialog.callbacks = verto.callbacks;
        dialog.answered = false;
        dialog.attach = params.attach || false;
	dialog.screenShare = params.screenShare || false;
	dialog.useCamera = dialog.params.useCamera;
	dialog.useMic = dialog.params.useMic;
	dialog.useSpeak = dialog.params.useSpeak;
	
        if (dialog.params.callID) {
            dialog.callID = dialog.params.callID;
        } else {
            dialog.callID = dialog.params.callID = generateGUID();
        }
	
        if (dialog.params.tag) {
            dialog.audioStream = document.getElementById(dialog.params.tag);

            if (dialog.params.useVideo) {
                dialog.videoStream = dialog.audioStream;
            }
        } //else conjure one TBD

        if (dialog.params.localTag) {
	    dialog.localVideo = document.getElementById(dialog.params.localTag);
	}

        dialog.verto.dialogs[dialog.callID] = dialog;

        var RTCcallbacks = {};

        if (dialog.direction == $.verto.enum.direction.inbound) {
            if (dialog.params.display_direction === "outbound") {
                dialog.params.remote_caller_id_name = dialog.params.caller_id_name;
                dialog.params.remote_caller_id_number = dialog.params.caller_id_number;
            } else {
                dialog.params.remote_caller_id_name = dialog.params.callee_id_name;
                dialog.params.remote_caller_id_number = dialog.params.callee_id_number;
            }

            if (!dialog.params.remote_caller_id_name) {
                dialog.params.remote_caller_id_name = "Nobody";
            }

            if (!dialog.params.remote_caller_id_number) {
                dialog.params.remote_caller_id_number = "UNKNOWN";
            }

            RTCcallbacks.onMessage = function(rtc, msg) {
                console.debug(msg);
            };

            RTCcallbacks.onAnswerSDP = function(rtc, sdp) {
                console.error("answer sdp", sdp);
            };
        } else {
            dialog.params.remote_caller_id_name = "Outbound Call";
            dialog.params.remote_caller_id_number = dialog.params.destination_number;
        }

        RTCcallbacks.onICESDP = function(rtc) {
            console.log("RECV " + rtc.type + " SDP", rtc.mediaData.SDP);

	    if (dialog.state == $.verto.enum.state.requesting || dialog.state == $.verto.enum.state.answering || dialog.state == $.verto.enum.state.active) {
		location.reload();
		return;
	    }

            if (rtc.type == "offer") {
		if (dialog.state == $.verto.enum.state.active) {
                    dialog.setState($.verto.enum.state.requesting);
		    dialog.sendMethod("verto.attach", {
			sdp: rtc.mediaData.SDP
                    });
		} else {
                    dialog.setState($.verto.enum.state.requesting);
		    
                    dialog.sendMethod("verto.invite", {
			sdp: rtc.mediaData.SDP
                    });
		}
            } else { //answer
                dialog.setState($.verto.enum.state.answering);

                dialog.sendMethod(dialog.attach ? "verto.attach" : "verto.answer", {
                    sdp: dialog.rtc.mediaData.SDP
                });
            }
        };

        RTCcallbacks.onICE = function(rtc) {
            //console.log("cand", rtc.mediaData.candidate);
            if (rtc.type == "offer") {
                console.log("offer", rtc.mediaData.candidate);
                return;
            }
        };

        RTCcallbacks.onStream = function(rtc, stream) {
            console.log("stream started");
        };

        RTCcallbacks.onError = function(e) {
            console.error("ERROR:", e);
            dialog.hangup({cause: "Device or Permission Error"});
        };

        dialog.rtc = new $.FSRTC({
            callbacks: RTCcallbacks,
	    localVideo: dialog.screenShare ? null : dialog.localVideo,
            useVideo: dialog.params.useVideo ? dialog.videoStream : null,
            useAudio: dialog.audioStream,
            useStereo: dialog.params.useStereo,
            videoParams: dialog.params.videoParams,
            audioParams: verto.options.audioParams,
            iceServers: verto.options.iceServers,
	    screenShare: dialog.screenShare,
	    useCamera: dialog.useCamera,
	    useMic: dialog.useMic,
	    useSpeak: dialog.useSpeak
        });

        dialog.rtc.verto = dialog.verto;

        if (dialog.direction == $.verto.enum.direction.inbound) {
            if (dialog.attach) {
                dialog.answer();
            } else {
                dialog.ring();
            }
        }
    };

    $.verto.dialog.prototype.invite = function() {
        var dialog = this;
        dialog.rtc.call();
    };

    $.verto.dialog.prototype.sendMethod = function(method, obj) {
        var dialog = this;
        obj.dialogParams = {};

        for (var i in dialog.params) {
            if (i == "sdp" && method != "verto.invite" && method != "verto.attach") {
                continue;
            }

            obj.dialogParams[i] = dialog.params[i];
        }

        dialog.verto.rpcClient.call(method, obj,

        function(e) {
            /* Success */
            dialog.processReply(method, true, e);
        },

        function(e) {
            /* Error */
            dialog.processReply(method, false, e);
        });
    };

    function checkStateChange(oldS, newS) {

        if (newS == $.verto.enum.state.purge || $.verto.enum.states[oldS.name][newS.name]) {
            return true;
        }

        return false;
    }

    $.verto.dialog.prototype.setState = function(state) {
        var dialog = this;

        if (dialog.state == $.verto.enum.state.ringing) {
            dialog.stopRinging();
        }

        if (dialog.state == state || !checkStateChange(dialog.state, state)) {
            console.error("Dialog " + dialog.callID + ": INVALID state change from " + dialog.state.name + " to " + state.name);
            dialog.hangup();
            return false;
        }

        console.log("Dialog " + dialog.callID + ": state change from " + dialog.state.name + " to " + state.name);

        dialog.lastState = dialog.state;
        dialog.state = state;

        if (!dialog.causeCode) {
            dialog.causeCode = 16;
        }

        if (!dialog.cause) {
            dialog.cause = "NORMAL CLEARING";
        }

        if (dialog.callbacks.onDialogState) {
            dialog.callbacks.onDialogState(this);
        }

        switch (dialog.state) {

        case $.verto.enum.state.early:
        case $.verto.enum.state.active:

	    var speaker = dialog.useSpeak;
	    console.info("Using Speaker: ", speaker);

	    if (speaker && speaker !== "any") {
		var videoElement = dialog.audioStream;

		setTimeout(function() {
		    console.info("Setting speaker:", videoElement, speaker);
		    attachSinkId(videoElement, speaker);}, 500);
	    }

	    break;

        case $.verto.enum.state.trying:
            setTimeout(function() {
                if (dialog.state == $.verto.enum.state.trying) {
                    dialog.setState($.verto.enum.state.hangup);
                }
            }, 30000);
            break;
        case $.verto.enum.state.purge:
            dialog.setState($.verto.enum.state.destroy);
            break;
        case $.verto.enum.state.hangup:

            if (dialog.lastState.val > $.verto.enum.state.requesting.val && dialog.lastState.val < $.verto.enum.state.hangup.val) {
                dialog.sendMethod("verto.bye", {});
            }

            dialog.setState($.verto.enum.state.destroy);
            break;
        case $.verto.enum.state.destroy:
            delete dialog.verto.dialogs[dialog.callID];
	    if (dialog.params.screenShare) {
		dialog.rtc.stopPeer();
	    } else {
		dialog.rtc.stop();
	    }
            break;
        }

        return true;
    };

    $.verto.dialog.prototype.processReply = function(method, success, e) {
        var dialog = this;

        //console.log("Response: " + method + " State:" + dialog.state.name, success, e);

        switch (method) {

        case "verto.answer":
        case "verto.attach":
            if (success) {
                dialog.setState($.verto.enum.state.active);
            } else {
                dialog.hangup();
            }
            break;
        case "verto.invite":
            if (success) {
                dialog.setState($.verto.enum.state.trying);
            } else {
                dialog.setState($.verto.enum.state.destroy);
            }
            break;

        case "verto.bye":
            dialog.hangup();
            break;

        case "verto.modify":
            if (e.holdState) {
                if (e.holdState == "held") {
                    if (dialog.state != $.verto.enum.state.held) {
                        dialog.setState($.verto.enum.state.held);
                    }
                } else if (e.holdState == "active") {
                    if (dialog.state != $.verto.enum.state.active) {
                        dialog.setState($.verto.enum.state.active);
                    }
                }
            }

            if (success) {}

            break;

        default:
            break;
        }

    };

    $.verto.dialog.prototype.hangup = function(params) {
        var dialog = this;

        if (params) {
            if (params.causeCode) {
		dialog.causeCode = params.causeCode;
            }

            if (params.cause) {
		dialog.cause = params.cause;
            }
        }

        if (dialog.state.val >= $.verto.enum.state.new.val && dialog.state.val < $.verto.enum.state.hangup.val) {
            dialog.setState($.verto.enum.state.hangup);
        } else if (dialog.state.val < $.verto.enum.state.destroy) {
            dialog.setState($.verto.enum.state.destroy);
        }
    };

    $.verto.dialog.prototype.stopRinging = function() {
        var dialog = this;
        if (dialog.verto.ringer) {
            dialog.verto.ringer.stop();
        }
    };

    $.verto.dialog.prototype.indicateRing = function() {
        var dialog = this;

        if (dialog.verto.ringer) {
            dialog.verto.ringer.attr("src", dialog.verto.options.ringFile)[0].play();

            setTimeout(function() {
                dialog.stopRinging();
                if (dialog.state == $.verto.enum.state.ringing) {
                    dialog.indicateRing();
                }
            },
            dialog.verto.options.ringSleep);
        }
    };

    $.verto.dialog.prototype.ring = function() {
        var dialog = this;

        dialog.setState($.verto.enum.state.ringing);
        dialog.indicateRing();
    };

    $.verto.dialog.prototype.useVideo = function(on) {
        var dialog = this;

        dialog.params.useVideo = on;

        if (on) {
            dialog.videoStream = dialog.audioStream;
        } else {
            dialog.videoStream = null;
        }

        dialog.rtc.useVideo(dialog.videoStream, dialog.localVideo);

    };

    $.verto.dialog.prototype.setMute = function(what) {
	var dialog = this;
	return dialog.rtc.setMute(what);
    };

    $.verto.dialog.prototype.getMute = function() {
	var dialog = this; 
	return dialog.rtc.getMute();
    };

    $.verto.dialog.prototype.useStereo = function(on) {
        var dialog = this;

        dialog.params.useStereo = on;
        dialog.rtc.useStereo(on);
    };

    $.verto.dialog.prototype.dtmf = function(digits) {
        var dialog = this;
        if (digits) {
            dialog.sendMethod("verto.info", {
                dtmf: digits
            });
        }
    };

    $.verto.dialog.prototype.transfer = function(dest, params) {
        var dialog = this;
        if (dest) {
            dialog.sendMethod("verto.modify", {
                action: "transfer",
                destination: dest,
                params: params
            });
        }
    };

    $.verto.dialog.prototype.hold = function(params) {
        var dialog = this;

        dialog.sendMethod("verto.modify", {
            action: "hold",
            params: params
        });
    };

    $.verto.dialog.prototype.unhold = function(params) {
        var dialog = this;

        dialog.sendMethod("verto.modify", {
            action: "unhold",
            params: params
        });
    };

    $.verto.dialog.prototype.toggleHold = function(params) {
        var dialog = this;

        dialog.sendMethod("verto.modify", {
            action: "toggleHold",
            params: params
        });
    };

    $.verto.dialog.prototype.message = function(msg) {
        var dialog = this;
        var err = 0;

        msg.from = dialog.params.login;

        if (!msg.to) {
            console.error("Missing To");
            err++;
        }

        if (!msg.body) {
            console.error("Missing Body");
            err++;
        }

        if (err) {
            return false;
        }

        dialog.sendMethod("verto.info", {
            msg: msg
        });

        return true;
    };

    $.verto.dialog.prototype.answer = function(params) {
        var dialog = this;
	
        if (!dialog.answered) {
	    if (!params) {
		params = {};
	    }

	    params.sdp = dialog.params.sdp;

            if (params) {
                if (params.useVideo) {
                    dialog.useVideo(true);
                }
		dialog.params.callee_id_name = params.callee_id_name;
		dialog.params.callee_id_number = params.callee_id_number;

		if (params.useCamera) {
		    dialog.useCamera = params.useCamera;
		}

		if (params.useMic) {
		    dialog.useMic = params.useMic;
		}

		if (params.useSpeak) {
		    dialog.useSpeak = params.useSpeak;
		}
            }
	    
            dialog.rtc.createAnswer(params);
            dialog.answered = true;
        }
    };

    $.verto.dialog.prototype.handleAnswer = function(params) {
        var dialog = this;

        dialog.gotAnswer = true;

        if (dialog.state.val >= $.verto.enum.state.active.val) {
            return;
        }

        if (dialog.state.val >= $.verto.enum.state.early.val) {
            dialog.setState($.verto.enum.state.active);
        } else {
            if (dialog.gotEarly) {
                console.log("Dialog " + dialog.callID + " Got answer while still establishing early media, delaying...");
            } else {
                console.log("Dialog " + dialog.callID + " Answering Channel");
                dialog.rtc.answer(params.sdp, function() {
                    dialog.setState($.verto.enum.state.active);
                }, function(e) {
                    console.error(e);
                    dialog.hangup();
                });
                console.log("Dialog " + dialog.callID + "ANSWER SDP", params.sdp);
            }
        }


    };

    $.verto.dialog.prototype.cidString = function(enc) {
        var dialog = this;
        var party = dialog.params.remote_caller_id_name + (enc ? " &lt;" : " <") + dialog.params.remote_caller_id_number + (enc ? "&gt;" : ">");
        return party;
    };

    $.verto.dialog.prototype.sendMessage = function(msg, params) {
        var dialog = this;

        if (dialog.callbacks.onMessage) {
            dialog.callbacks.onMessage(dialog.verto, dialog, msg, params);
        }
    };

    $.verto.dialog.prototype.handleInfo = function(params) {
        var dialog = this;

        dialog.sendMessage($.verto.enum.message.info, params.msg);
    };

    $.verto.dialog.prototype.handleDisplay = function(params) {
        var dialog = this;

        if (params.display_name) {
            dialog.params.remote_caller_id_name = params.display_name;
        }
        if (params.display_number) {
            dialog.params.remote_caller_id_number = params.display_number;
        }

        dialog.sendMessage($.verto.enum.message.display, {});
    };

    $.verto.dialog.prototype.handleMedia = function(params) {
        var dialog = this;

        if (dialog.state.val >= $.verto.enum.state.early.val) {
            return;
        }

        dialog.gotEarly = true;

        dialog.rtc.answer(params.sdp, function() {
            console.log("Dialog " + dialog.callID + "Establishing early media");
            dialog.setState($.verto.enum.state.early);

            if (dialog.gotAnswer) {
                console.log("Dialog " + dialog.callID + "Answering Channel");
                dialog.setState($.verto.enum.state.active);
            }
        }, function(e) {
            console.error(e);
            dialog.hangup();
        });
        console.log("Dialog " + dialog.callID + "EARLY SDP", params.sdp);
    };

    $.verto.ENUM = function(s) {
        var i = 0,
        o = {};
        s.split(" ").map(function(x) {
            o[x] = {
                name: x,
                val: i++
            };
        });
        return Object.freeze(o);
    };

    $.verto.enum = {};

    $.verto.enum.states = Object.freeze({
        new: {
            requesting: 1,
            recovering: 1,
            ringing: 1,
            destroy: 1,
            answering: 1,
	    hangup: 1
        },
        requesting: {
            trying: 1,
            hangup: 1,
	    active: 1
        },
        recovering: {
            answering: 1,
            hangup: 1
        },
        trying: {
            active: 1,
            early: 1,
            hangup: 1
        },
        ringing: {
            answering: 1,
            hangup: 1
        },
        answering: {
            active: 1,
            hangup: 1
        },
        active: {
            answering: 1,
            requesting: 1,
            hangup: 1,
            held: 1
        },
        held: {
            hangup: 1,
            active: 1
        },
        early: {
            hangup: 1,
            active: 1
        },
        hangup: {
            destroy: 1
        },
        destroy: {},
        purge: {
            destroy: 1
        }
    });

    $.verto.enum.state = $.verto.ENUM("new requesting trying recovering ringing answering early active held hangup destroy purge");
    $.verto.enum.direction = $.verto.ENUM("inbound outbound");
    $.verto.enum.message = $.verto.ENUM("display info pvtEvent");

    $.verto.enum = Object.freeze($.verto.enum);

    $.verto.saved = [];
    
    $.verto.unloadJobs = [];

    $(window).bind('beforeunload', function() {
	for (var f in $.verto.unloadJobs) {
	    $.verto.unloadJobs[f]();
	}

        for (var i in $.verto.saved) {
            var verto = $.verto.saved[i];
            if (verto) {
                verto.purge();
                verto.logout();
            }
        }

        return $.verto.warnOnUnload;
    });

    $.verto.videoDevices = [];
    $.verto.audioInDevices = [];
    $.verto.audioOutDevices = [];

    var checkDevices = function(runtime) {
	console.info("enumerating devices");
	var aud_in = [], aud_out = [], vid = [];	

	if ((!navigator.mediaDevices || !navigator.mediaDevices.enumerateDevices) && MediaStreamTrack.getSources) {
	    MediaStreamTrack.getSources(function (media_sources) {
		for (var i = 0; i < media_sources.length; i++) {

		    if (media_sources[i].kind == 'video') {
			vid.push(media_sources[i]);
		    } else {
			aud_in.push(media_sources[i]);
		    }
		}
		
		$.verto.videoDevices = vid;
		$.verto.audioInDevices = aud_in;
		
		console.info("Audio Devices", $.verto.audioInDevices);
		console.info("Video Devices", $.verto.videoDevices);
		runtime(true);
	    });
	} else {
	    /* of course it's a totally different API CALL with different element names for the same exact thing */
	    
	    if (!navigator.mediaDevices || !navigator.mediaDevices.enumerateDevices) {
		console.log("enumerateDevices() not supported.");
		return;
	    }

	    // List cameras and microphones.

	    navigator.mediaDevices.enumerateDevices()
		.then(function(devices) {
		    devices.forEach(function(device) {
			console.log(device);

			console.log(device.kind + ": " + device.label +
				    " id = " + device.deviceId);
			
			if (device.kind === "videoinput") {
			    vid.push({id: device.deviceId, kind: "video", label: device.label});
			} else if (device.kind === "audioinput") {
			    aud_in.push({id: device.deviceId, kind: "audio_in", label: device.label});
			} else if (device.kind === "audiooutput") {
			    aud_out.push({id: device.deviceId, kind: "audio_out", label: device.label});
			}
		    });
		    

		    $.verto.videoDevices = vid;
		    $.verto.audioInDevices = aud_in;
		    $.verto.audioOutDevices = aud_out;
		    
		    console.info("Audio IN Devices", $.verto.audioInDevices);
		    console.info("Audio Out Devices", $.verto.audioOutDevices);
		    console.info("Video Devices", $.verto.videoDevices);
		    runtime(true);
		    
		})
		.catch(function(err) {
		    console.log(" Device Enumeration ERROR: " + err.name + ": " + err.message);
		    runtime(false);
		});
	}

    };

    $.verto.refreshDevices = function(runtime) {
	checkDevices(runtime);
    }

    $.verto.init = function(obj, runtime) {
	if (!obj) {
	    obj = {};
	}

	if (!obj.skipPermCheck && !obj.skipDeviceCheck) {
	    $.FSRTC.checkPerms(function(status) {
		checkDevices(runtime);
	    }, true, true);
	} else if (obj.skipPermCheck && !obj.skipDeviceCheck) {
	    checkDevices(runtime);
	} else if (!obj.skipPermCheck && obj.skipDeviceCheck) {
	    $.FSRTC.checkPerms(function(status) {
		runtime(status);
	    }, true, true);
	} else {
	    runtime(null);
	}

    }

    $.verto.genUUID = function () {
	return generateGUID();
    }


})(jQuery);

// Last time updated at Sep 07, 2014, 08:32:23

// Latest file can be found here: https://cdn.webrtc-experiment.com/getScreenId.js

// Muaz Khan         - www.MuazKhan.com
// MIT License       - www.WebRTC-Experiment.com/licence
// Documentation     - https://github.com/muaz-khan/WebRTC-Experiment/tree/master/getScreenId.js

// ______________
// getScreenId.js

/*
getScreenId(function (error, sourceId, screen_constraints) {
    // error    == null || 'permission-denied' || 'not-installed' || 'installed-disabled' || 'not-chrome'
    // sourceId == null || 'string' || 'firefox'
    
    if(sourceId == 'firefox') {
        navigator.mozGetUserMedia(screen_constraints, onSuccess, onFailure);
    }
    else navigator.webkitGetUserMedia(screen_constraints, onSuccess, onFailure);
});
*/

(function() {
    window.getScreenId = function(callback) {
        // for Firefox:
        // sourceId == 'firefox'
        // screen_constraints = {...}
        if (!!navigator.mozGetUserMedia) {
            callback(null, 'firefox', {
                video: {
                    mozMediaSource: 'window',
                    mediaSource: 'window'
                }
            });
            return;
        }

        postMessage();

        window.addEventListener('message', onIFrameCallback);

        function onIFrameCallback(event) {
            if (!event.data) return;

            if (event.data.chromeMediaSourceId) {
                if (event.data.chromeMediaSourceId === 'PermissionDeniedError') {
                    callback('permission-denied');
                } else callback(null, event.data.chromeMediaSourceId, getScreenConstraints(null, event.data.chromeMediaSourceId));
            }

            if (event.data.chromeExtensionStatus) {
                callback(event.data.chromeExtensionStatus, null, getScreenConstraints(event.data.chromeExtensionStatus));
            }

            // this event listener is no more needed
            window.removeEventListener('message', onIFrameCallback);
        }
    };

    function getScreenConstraints(error, sourceId) {
        var screen_constraints = {
            audio: false,
            video: {
                mandatory: {
                    chromeMediaSource: error ? 'screen' : 'desktop',
                    maxWidth: window.screen.width > 1920 ? window.screen.width : 1920,
                    maxHeight: window.screen.height > 1080 ? window.screen.height : 1080
                },
                optional: []
            }
        };

        if (sourceId) {
            screen_constraints.video.mandatory.chromeMediaSourceId = sourceId;
        }

        return screen_constraints;
    }

    function postMessage() {
        if (!iframe.isLoaded) {
            setTimeout(postMessage, 100);
            return;
        }

        iframe.contentWindow.postMessage({
            captureSourceId: true
        }, '*');
    }

    var iframe = document.createElement('iframe');
    iframe.onload = function() {
        iframe.isLoaded = true;
    };
    iframe.src = 'https://www.webrtc-experiment.com/getSourceId/';
    iframe.style.display = 'none';
    (document.body || document.documentElement).appendChild(iframe);
})();

/*
 * JavaScript MD5 1.0.1
 * https://github.com/blueimp/JavaScript-MD5
 *
 * Copyright 2011, Sebastian Tschan
 * https://blueimp.net
 *
 * Licensed under the MIT license:
 * http://www.opensource.org/licenses/MIT
 *
 * Based on
 * A JavaScript implementation of the RSA Data Security, Inc. MD5 Message
 * Digest Algorithm, as defined in RFC 1321.
 * Version 2.2 Copyright (C) Paul Johnston 1999 - 2009
 * Other contributors: Greg Holt, Andrew Kepert, Ydnar, Lostinet
 * Distributed under the BSD License
 * See http://pajhome.org.uk/crypt/md5 for more info.
 */
!function(a){"use strict";function b(a,b){var c=(65535&a)+(65535&b),d=(a>>16)+(b>>16)+(c>>16);return d<<16|65535&c}function c(a,b){return a<<b|a>>>32-b}function d(a,d,e,f,g,h){return b(c(b(b(d,a),b(f,h)),g),e)}function e(a,b,c,e,f,g,h){return d(b&c|~b&e,a,b,f,g,h)}function f(a,b,c,e,f,g,h){return d(b&e|c&~e,a,b,f,g,h)}function g(a,b,c,e,f,g,h){return d(b^c^e,a,b,f,g,h)}function h(a,b,c,e,f,g,h){return d(c^(b|~e),a,b,f,g,h)}function i(a,c){a[c>>5]|=128<<c%32,a[(c+64>>>9<<4)+14]=c;var d,i,j,k,l,m=1732584193,n=-271733879,o=-1732584194,p=271733878;for(d=0;d<a.length;d+=16)i=m,j=n,k=o,l=p,m=e(m,n,o,p,a[d],7,-680876936),p=e(p,m,n,o,a[d+1],12,-389564586),o=e(o,p,m,n,a[d+2],17,606105819),n=e(n,o,p,m,a[d+3],22,-1044525330),m=e(m,n,o,p,a[d+4],7,-176418897),p=e(p,m,n,o,a[d+5],12,1200080426),o=e(o,p,m,n,a[d+6],17,-1473231341),n=e(n,o,p,m,a[d+7],22,-45705983),m=e(m,n,o,p,a[d+8],7,1770035416),p=e(p,m,n,o,a[d+9],12,-1958414417),o=e(o,p,m,n,a[d+10],17,-42063),n=e(n,o,p,m,a[d+11],22,-1990404162),m=e(m,n,o,p,a[d+12],7,1804603682),p=e(p,m,n,o,a[d+13],12,-40341101),o=e(o,p,m,n,a[d+14],17,-1502002290),n=e(n,o,p,m,a[d+15],22,1236535329),m=f(m,n,o,p,a[d+1],5,-165796510),p=f(p,m,n,o,a[d+6],9,-1069501632),o=f(o,p,m,n,a[d+11],14,643717713),n=f(n,o,p,m,a[d],20,-373897302),m=f(m,n,o,p,a[d+5],5,-701558691),p=f(p,m,n,o,a[d+10],9,38016083),o=f(o,p,m,n,a[d+15],14,-660478335),n=f(n,o,p,m,a[d+4],20,-405537848),m=f(m,n,o,p,a[d+9],5,568446438),p=f(p,m,n,o,a[d+14],9,-1019803690),o=f(o,p,m,n,a[d+3],14,-187363961),n=f(n,o,p,m,a[d+8],20,1163531501),m=f(m,n,o,p,a[d+13],5,-1444681467),p=f(p,m,n,o,a[d+2],9,-51403784),o=f(o,p,m,n,a[d+7],14,1735328473),n=f(n,o,p,m,a[d+12],20,-1926607734),m=g(m,n,o,p,a[d+5],4,-378558),p=g(p,m,n,o,a[d+8],11,-2022574463),o=g(o,p,m,n,a[d+11],16,1839030562),n=g(n,o,p,m,a[d+14],23,-35309556),m=g(m,n,o,p,a[d+1],4,-1530992060),p=g(p,m,n,o,a[d+4],11,1272893353),o=g(o,p,m,n,a[d+7],16,-155497632),n=g(n,o,p,m,a[d+10],23,-1094730640),m=g(m,n,o,p,a[d+13],4,681279174),p=g(p,m,n,o,a[d],11,-358537222),o=g(o,p,m,n,a[d+3],16,-722521979),n=g(n,o,p,m,a[d+6],23,76029189),m=g(m,n,o,p,a[d+9],4,-640364487),p=g(p,m,n,o,a[d+12],11,-421815835),o=g(o,p,m,n,a[d+15],16,530742520),n=g(n,o,p,m,a[d+2],23,-995338651),m=h(m,n,o,p,a[d],6,-198630844),p=h(p,m,n,o,a[d+7],10,1126891415),o=h(o,p,m,n,a[d+14],15,-1416354905),n=h(n,o,p,m,a[d+5],21,-57434055),m=h(m,n,o,p,a[d+12],6,1700485571),p=h(p,m,n,o,a[d+3],10,-1894986606),o=h(o,p,m,n,a[d+10],15,-1051523),n=h(n,o,p,m,a[d+1],21,-2054922799),m=h(m,n,o,p,a[d+8],6,1873313359),p=h(p,m,n,o,a[d+15],10,-30611744),o=h(o,p,m,n,a[d+6],15,-1560198380),n=h(n,o,p,m,a[d+13],21,1309151649),m=h(m,n,o,p,a[d+4],6,-145523070),p=h(p,m,n,o,a[d+11],10,-1120210379),o=h(o,p,m,n,a[d+2],15,718787259),n=h(n,o,p,m,a[d+9],21,-343485551),m=b(m,i),n=b(n,j),o=b(o,k),p=b(p,l);return[m,n,o,p]}function j(a){var b,c="";for(b=0;b<32*a.length;b+=8)c+=String.fromCharCode(a[b>>5]>>>b%32&255);return c}function k(a){var b,c=[];for(c[(a.length>>2)-1]=void 0,b=0;b<c.length;b+=1)c[b]=0;for(b=0;b<8*a.length;b+=8)c[b>>5]|=(255&a.charCodeAt(b/8))<<b%32;return c}function l(a){return j(i(k(a),8*a.length))}function m(a,b){var c,d,e=k(a),f=[],g=[];for(f[15]=g[15]=void 0,e.length>16&&(e=i(e,8*a.length)),c=0;16>c;c+=1)f[c]=909522486^e[c],g[c]=1549556828^e[c];return d=i(f.concat(k(b)),512+8*b.length),j(i(g.concat(d),640))}function n(a){var b,c,d="0123456789abcdef",e="";for(c=0;c<a.length;c+=1)b=a.charCodeAt(c),e+=d.charAt(b>>>4&15)+d.charAt(15&b);return e}function o(a){return unescape(encodeURIComponent(a))}function p(a){return l(o(a))}function q(a){return n(p(a))}function r(a,b){return m(o(a),o(b))}function s(a,b){return n(r(a,b))}function t(a,b,c){return b?c?r(b,a):s(b,a):c?p(a):q(a)}"function"==typeof define&&define.amd?define(function(){return t}):a.md5=t}(this);
(function() {
  'use strict';

  var vertoApp = angular.module('vertoApp', [
    'timer',
    'ngRoute',
    'vertoControllers',
    'vertoDirectives',
    'ngStorage',
    'ngAnimate',
    'toastr',
    'FBAngular',
    'cgPrompt',
    '720kb.tooltips',
    'ui.gravatar',
    'directive.g+signin'
  ]);

  vertoApp.config(['$routeProvider', 'gravatarServiceProvider',
    function($routeProvider, gravatarServiceProvider) {
      $routeProvider.
      when('/', {
        title: 'Loading',
        templateUrl: 'partials/splash_screen.html',
        controller: 'SplashScreenController'
      }).
      when('/login', {
        title: 'Login',
        templateUrl: 'partials/login.html',
        controller: 'LoginController'
      }).
      when('/dialpad', {
        title: 'Dialpad',
        templateUrl: 'partials/dialpad.html',
        controller: 'DialPadController'
      }).
      when('/incall', {
          title: 'In a Call',
          templateUrl: 'partials/incall.html',
          controller: 'InCallController'
        }).
      when('/browser-upgrade', {
        title: '',
        templateUrl: 'partials/browser_upgrade.html',
        controller: 'BrowserUpgradeController'
      }).
      otherwise({
        redirectTo: '/'
      });

      gravatarServiceProvider.defaults = {
        default: 'mm'  // Mystery man as default for missing avatars
      };
    }
  ]);

  vertoApp.run(['$rootScope', '$location', 'toastr', 'prompt', 'verto',
    function($rootScope, $location, toastr, prompt, verto) {
      
      $rootScope.$on( "$routeChangeStart", function(event, next, current) {
        if (!verto.data.connected) {
          if ( next.templateUrl === "partials/login.html") {
            // pass 
          } else {
            $location.path("/");
          }
        }
      });
      
      $rootScope.$on('$routeChangeSuccess', function(event, current, previous) {
        $rootScope.title = current.$$route.title;
      });

      $rootScope.safeProtocol = false;

      if (window.location.protocol == 'https:') {
        $rootScope.safeProtocol = true;
      }

      
      $rootScope.promptInput = function(title, message, label, callback) {
        var ret = prompt({
          title: title,
          message: message,
          input: true,
          label: label
        }).then(function(ret) {
          if (angular.isFunction(callback)) {
            callback(ret);
          }
        }, function() {

        });

      };

    }
  ]);

})();

(function() {
  'use strict';

  var vertoControllers = angular.module('vertoControllers', [
    'ui.bootstrap',
    'vertoService',
    'storageService',
    'ui.gravatar'
  ]);

})();

(function() {
  'use strict';

  angular
  .module('vertoControllers')
  .controller('SplashScreenController', ['$scope', '$rootScope', '$location', '$timeout', 'splashscreen', 'prompt', 'verto',
    function($scope, $rootScope, $location, $timeout, splashscreen, prompt, verto) {
      console.debug('Executing SplashScreenController.');
      
      $scope.progress_percentage = splashscreen.progress_percentage;
      $scope.message = '';
      $scope.interrupt_next = false;
      $scope.errors = [];

      var redirectTo = function(link, activity) {
        if(activity) {
          if(activity == 'browser-upgrade') {
            link = activity;
          }
        }
        
        $location.path(link);
      }

      var checkProgressState = function(current_progress, status, promise, activity, soft, interrupt, message) {
        $scope.progress_percentage = splashscreen.calculate(current_progress); 
        $scope.message = message;

        if(interrupt && status == 'error') {
          $scope.errors.push(message);
          if(!soft) {
            redirectTo('', activity); 
            return;
          } else {
            message = message + '. Continue?'; 
          };

          if(!confirm(message)) {
            $scope.interrupt_next = true;  
          }; 
        };

        if($scope.interrupt_next) {
          return;
        };

        $scope.message = splashscreen.getProgressMessage(current_progress+1);

        return true;
      };
      
      $rootScope.$on('progress.next', function(ev, current_progress, status, promise, activity, soft, interrupt, message) {
        $timeout(function() {
          if(promise) {
            promise.then(function(response) {
              message = response['message'];
              status = response['status'];
              if(checkProgressState(current_progress, status, promise, activity, soft, interrupt, message)) {
                splashscreen.next();
              };
            });

            return;
          }
          
          if(!checkProgressState(current_progress, status, promise, activity, soft, interrupt, message)) {
            return;
          }
          
          splashscreen.next();
        }, 400);
      });

      $rootScope.$on('progress.complete', function(ev, current_progress) {
        $scope.message = 'Complete';
        if(verto.data.connected) {
          redirectTo('/dialpad');
        } else {
          redirectTo('/login');
          $location.path('/login');
        }
      });

      splashscreen.next();

    }]);

})();

(function() {
  'use strict';

  angular
    .module('vertoControllers')
    .controller('BrowserUpgradeController', ['$scope', '$http',
      '$location', 'verto', 'storage', 'Fullscreen',
      function($scope, $http, $location, verto, storage, Fullscreen) {
        console.debug('Executing BrowserUpgradeController.');

      }
    ]);

})();
(function() {
  'use strict';

  angular
  .module('vertoControllers')
  .controller('ChatController', ['$scope', '$rootScope', '$http',
    '$location', '$anchorScroll', '$timeout', 'verto', 'prompt',
    function($scope, $rootScope, $http, $location, $anchorScroll, $timeout,
      verto, prompt) {
      console.debug('Executing ChatController.');

      function scrollToChatBottom() {
        // Going to the bottom of chat messages.
        var obj = document.querySelector('.chat-messages');
        obj.scrollTop = obj.scrollHeight;
        //var chat_messages_top = jQuery('.chat-messages').scrollTop();
        //var marker_position = jQuery('#chat-message-bottom').position().top;
        //jQuery('.chat-messages').scrollTop(chat_messages_top + marker_position);
      }

      var CLEAN_MESSAGE = '';

      function clearConferenceChat() {
        $scope.members = [];
        $scope.messages = [];
        $scope.message = CLEAN_MESSAGE;
      }
      clearConferenceChat();

      $scope.$watch('activePane', function() {
        if ($scope.activePane == 'chat') {
          $rootScope.chat_counter = 0;
        }
        $rootScope.activePane = $scope.activePane;
      });

      $rootScope.$on('chat.newMessage', function(event, data) {
        data.created_at = new Date();
        console.log('chat.newMessage', data);
        $scope.$apply(function() {
          $scope.messages.push(data);
          if (data.from != verto.data.name && (!$scope.chatStatus ||
              $scope.activePane != 'chat')) {
            ++$rootScope.chat_counter;
          }
          $timeout(function() {
            scrollToChatBottom();
          }, 300);
        });
      });

      function findMemberByUUID(uuid) {
        var found = false;
        for (var idx in $scope.members) {
          var member = $scope.members[idx];
          if (member.uuid == uuid) {
            found = true;
            break;
          }
        }
        if (found) {
          return idx;
        } else {
          return -1;
        }
      }

      function translateMember(member) {
        return {
          'uuid': member[0],
          'id': member[1][0],
          'number': member[1][1],
          'name': member[1][2],
          'codec': member[1][3],
          'status': JSON.parse(member[1][4]),
          'email': member[1][5].email
        };
      }

      function addMember(member) {
        $scope.members.push(translateMember(member));
      }

      $rootScope.$on('members.boot', function(event, members) {
        $scope.$apply(function() {
          clearConferenceChat();
          for (var idx in members) {
            var member = members[idx];
            addMember(member);
          }
        })
      });

      $rootScope.$on('members.add', function(event, member) {
        $scope.$apply(function() {
          addMember(member);
        });
      });

      $rootScope.$on('members.del', function(event, uuid) {
        $scope.$apply(function() {
          var memberIdx = findMemberByUUID(uuid);
          if (memberIdx != -1) {
            // Removing the member.
            $scope.members.splice(memberIdx, 1);
          }
        });
      });

      $rootScope.$on('members.update', function(event, member) {
        member = translateMember(member);
        var memberIdx = findMemberByUUID(member.uuid);
        if (memberIdx < 0) {
          console.log('Didn\'t find the member uuid ' + member.uuid);
        } else {
          $scope.$apply(function() {
            // console.log('Updating', memberIdx, ' <', $scope.members[memberIdx],
              // '> with <', member, '>');
            angular.extend($scope.members[memberIdx], member);
          });
        }
      });

      $rootScope.$on('members.clear', function(event) {
        $scope.$applyAsync(function() {
          clearConferenceChat();
          $scope.closeChat();
        });
      });

      /**
       * Public methods.
       */
      $scope.send = function() {
        // Only conferencing chat is supported for now
        // but still calling method with the conference prefix
        // so we know that explicitly.
        verto.sendConferenceChat($scope.message);
        $scope.message = CLEAN_MESSAGE;
      };

      // Participants moderation.
      $scope.confKick = function(memberID) {
        console.log('$scope.confKick');
        verto.data.conf.kick(memberID);
      };

      $scope.confMuteMic = function(memberID) {
        if(verto.data.confRole == 'moderator') {
          console.log('$scope.confMuteMic');
          verto.data.conf.muteMic(memberID);
        }
      };

      $scope.confMuteVideo = function(memberID) {
        if(verto.data.confRole == 'moderator') {
          console.log('$scope.confMuteVideo');
          verto.data.conf.muteVideo(memberID);
        }
      };

      $scope.confPresenter = function(memberID) {
        console.log('$scope.confPresenter');
        verto.data.conf.presenter(memberID);
      };

      $scope.confVideoFloor = function(memberID) {
        console.log('$scope.confVideoFloor');
        verto.data.conf.videoFloor(memberID);
      };

      $scope.confBanner = function(memberID) {
        console.log('$scope.confBanner');
        
        prompt({
          title: 'Please insert the banner text',
          input: true,
          label: '',
          value: '',
        }).then(function(text) {
          if (text) {
            verto.data.conf.banner(memberID, text);
          }
        });
      };

      //$scope.confResetBanner = function(memberID) {
      //  console.log('$scope.confResetBanner');
      //  var text = 'reset';
      //  verto.data.conf.banner(memberID, text);
      //};

      $scope.confVolumeDown = function(memberID) {
        console.log('$scope.confVolumeDown');
        verto.data.conf.volumeDown(memberID);
      };

      $scope.confVolumeUp = function(memberID) {
        console.log('$scope.confVolumeUp');
        verto.data.conf.volumeUp(memberID);
      };

      $scope.confTransfer = function(memberID) {
        console.log('$scope.confTransfer');
        prompt({
          title: 'Transfer party?',
          message: 'To what destination would you like to transfer this call?',
          input: true,
          label: 'Destination',
          value: '',
        }).then(function(exten) {
          if (exten) {
            verto.data.conf.transfer(memberID, exten);
          }
        });
      };
    }
  ]);

})();

(function() {
  'use strict';

  angular
    .module('vertoControllers')
    .controller('ContributorsController', ['$scope', '$http',
      'toastr',
      function($scope, $http, toastr) {
        var url = window.location.origin + window.location.pathname;
        $http.get(url + 'contributors.txt')
          .success(function(data) {

            var contributors = [];

            angular.forEach(data, function(value, key) {
              var re = /(.*) <(.*)>/;
              var name = value.replace(re, "$1");
              var email = value.replace(re, "$2");

              this.push({
                'name': name,
                'email': email
              });
            }, contributors);

            $scope.contributors = contributors;
          })
          .error(function() {
            toastr.error('contributors not found.');
          });
      }
    ]);
})();

(function() {
  'use strict';

  angular
    .module('vertoControllers')
    .controller('AboutController', ['$scope', '$http',
      'toastr',
      function($scope, $http, toastr) {
	    var githash = '7f85faf' || 'something is not right';
            $scope.githash = githash;

	/* leave this here for later, but its not needed right now
        $http.get(window.location.pathname + '/contributors.txt')
          .success(function(data) {

          })
          .error(function() {
            toastr.error('contributors not found.');
          });
	 */
      }
    ]);
})();

(function() {
  'use strict';

  angular
    .module('vertoControllers')
    .controller('DialPadController', ['$rootScope', '$scope',
      '$http', '$location', 'toastr', 'verto', 'storage', 'CallHistory', 'eventQueue',
      function($rootScope, $scope, $http, $location, toastr, verto, storage, CallHistory, eventQueue) {
        console.debug('Executing DialPadController.');
        
        eventQueue.process();
        
        $scope.call_history = CallHistory.all();
        $scope.history_control = CallHistory.all_control();
        $scope.has_history = Object.keys($scope.call_history).length;
        storage.data.videoCall = false;
        storage.data.userStatus = 'connecting';
        storage.data.calling = false;

        $scope.clearCallHistory = function() {
          CallHistory.clear();
          $scope.call_history = CallHistory.all();
          $scope.history_control = CallHistory.all_control();
          $scope.has_history = Object.keys($scope.call_history).length;
          return $scope.history_control;
        };

        $scope.viewCallsList = function(calls) {
          return $scope.call_list = calls;
        };

        /**
         * fill dialpad via querystring [?autocall=\d+]
         */
        if ($location.search().autocall) {
            $rootScope.dialpadNumber = $location.search().autocall;
	    delete $location.search().autocall;
            call($rootScope.dialpadNumber);
        }

	/**
	 * fill in dialpad via config.json
	 */
        if ('autocall' in verto.data) {
          $rootScope.dialpadNumber = verto.data.autocall;
	  delete verto.data.autocall;
            setTimeout(function () {
                if (!verto.data.call)
                    call($rootScope.dialpadNumber);
            }, 1500);
        }

        /**
         * used to bind click on number in call history to fill dialpad
         * 'cause inside a ng-repeat the angular isnt in ctrl scope
         */
        $scope.fillDialpadNumber = function(number) {
          $rootScope.dialpadNumber = number;
        };

        $rootScope.transfer = function() {
          if (!$rootScope.dialpadNumber) {
            return false;
          }
          verto.data.call.transfer($rootScope.dialpadNumber);
        };

        function call(extension) {
          storage.data.onHold = false;
          storage.data.cur_call = 0;
          $rootScope.dialpadNumber = extension;
          if (!$rootScope.dialpadNumber && storage.data.called_number) {
            $rootScope.dialpadNumber = storage.data.called_number;
            return false;
          } else if (!$rootScope.dialpadNumber && !storage.data.called_number) {
            toastr.warning('Enter an extension, please.');
            return false;
          }

          if (verto.data.call) {
            console.debug('A call is already in progress.');
            return false;
          }

          storage.data.mutedVideo = false;
          storage.data.mutedMic = false;

          storage.data.videoCall = false;
          verto.call($rootScope.dialpadNumber);

          storage.data.called_number = $rootScope.dialpadNumber;
          CallHistory.add($rootScope.dialpadNumber, 'outbound');
          $location.path('/incall');
        }

        /**
         * Call to the number in the $rootScope.dialpadNumber.
         */
        $rootScope.call = function(extension) {
          return call(extension);
        }
      }
    ]);

})();

(function() {
  'use strict';

  angular
    .module('vertoControllers')
    .controller('InCallController', ['$rootScope', '$scope',
      '$http', '$location', '$modal', '$timeout', 'toastr', 'verto', 'storage', 'prompt', 'Fullscreen',
      function($rootScope, $scope, $http, $location, $modal, $timeout, toatr,
        verto, storage, prompt, Fullscreen) {

        console.debug('Executing InCallController.');
        $scope.layout = null;
        $rootScope.dialpadNumber = '';
        $scope.callTemplate = 'partials/phone_call.html';
        $scope.dialpadTemplate = '';
        $scope.incall = true;


        if (storage.data.videoCall) {
          $scope.callTemplate = 'partials/video_call.html';
        }
        
        $rootScope.$on('call.conference', function(event, data) {
          $timeout(function() {
            if($scope.chatStatus) {
              $scope.openChat();
            }
          });
        });

        $rootScope.$on('call.video', function(event, data) {
          $timeout(function() {
            $scope.callTemplate = 'partials/video_call.html';
          });
        });

        /**
         * toggle dialpad in incall page
         */
        $scope.toggleDialpad = function() {
          $scope.openModal('partials/dialpad_widget.html',
            'ModalDialpadController');

          /*
          if(!$scope.dialpadTemplate) {
            $scope.dialpadTemplate = 'partials/dialpad_widget.html';
          } else {
            $scope.dialpadTemplate = '';
          }
          */
        }

        /**
         * TODO: useless?
         */
        $scope.videoCall = function() {
          prompt({
            title: 'Would you like to activate video for this call?',
            message: 'Video will be active during the next calls.'
          }).then(function() {
            storage.data.videoCall = true;
            $scope.callTemplate = 'partials/video_call.html';
          });
        };

        $scope.cbMuteVideo = function(event, data) {
          storage.data.mutedVideo = !storage.data.mutedVideo;
        }

        $scope.cbMuteMic = function(event, data) {
          storage.data.mutedMic = !storage.data.mutedMic;
        }

        $scope.confChangeVideoLayout = function(layout) {
          verto.data.conf.setVideoLayout(layout);
        };

        $scope.muteMic = verto.muteMic;
        $scope.muteVideo = verto.muteVideo;

        $timeout(function() {
          console.log('broadcast time-start incall');
          $scope.$broadcast('timer-start');
        }, 1000);

      }
    ]);
})();

(function() {
  'use strict';

  angular
    .module('vertoControllers')
    .controller('LoginController', ['$scope', '$http', '$location', 'verto', 
      function($scope, $http, $location, verto) {
        var preRoute = function() {
          if(verto.data.connected) {
            $location.path('/dialpad');
          }
        }
        preRoute();
        
        verto.data.name = $scope.storage.data.name;
        verto.data.email = $scope.storage.data.email;

        console.debug('Executing LoginController.');
      }
    ]);

})();


(function() {
  'use strict';

  angular
    .module('vertoControllers')
    .controller('MainController',
      ["$scope", "$rootScope", "$location", "$modal", "$timeout", "$q", "verto", "storage", "CallHistory", "toastr", "Fullscreen", "prompt", "eventQueue", "config", function($scope, $rootScope, $location, $modal, $timeout, $q, verto, storage, CallHistory, toastr, Fullscreen, prompt, eventQueue, config) {

      console.debug('Executing MainController.');

      var myVideo = document.getElementById("webcam");
      $scope.verto = verto;
      $scope.storage = storage;
      $scope.call_history = angular.element("#call_history").hasClass('active');
      $rootScope.chatStatus = angular.element('#wrapper').hasClass('toggled');

      /**
       * (explanation) scope in another controller extends rootScope (singleton)
       */
      $rootScope.chat_counter = 0;
      $rootScope.activePane = 'members';
      /**
       * The number that will be called.
       * @type {string}
       */
      $rootScope.dialpadNumber = '';
      
      // If verto is not connected, redirects to login page.
      if (!verto.data.connected) {
        console.debug('MainController: WebSocket not connected. Redirecting to login.');
        $location.path('/');
      }
 
      $rootScope.$on('config.http.success', function(ev) {
        $scope.login(true, true);
      });
      /**
       * Login the user to verto server and
       * redirects him to dialpad page.
       */
      $scope.login = function(redirect, httpSuccess) {
          var _connect = function () {
              if(redirect == undefined) {
                  redirect = true;
              }
              var connectCallback = function(v, connected) {
                  $scope.$apply(function() {
                      verto.data.connecting = false;
                      if (connected) {
                          storage.data.ui_connected = verto.data.connected;
                          storage.data.ws_connected = verto.data.connected;
                          storage.data.name = verto.data.name;
                          storage.data.email = verto.data.email;
                          storage.data.login = verto.data.login;
                          storage.data.password = verto.data.password;
                          if (redirect) {
                              $location.path('/dialpad');
                          }
                      }
                  });
              };

              verto.data.connecting = true;
              verto.connect(connectCallback);
          };
          if (httpSuccess) {
              _connect();
          } else {
              var configResponse = config.configure();
              configResponse.then(function (response) {
                  if (response.status == 200 && !response.data.info) {
                      //_connect();
                      if (response.data.validateEmail) {
                          if (verto.data.email) {
                              storage.data.email = verto.data.email;
                          };
                          if (verto.data.name) {
                              storage.data.name = verto.data.name;
                          };
                          if (verto.data.password) {
                              storage.data.password = verto.data.password;
                          };
                          prompt({
                              "title": "Info",
                              "message": "Please verify your email.",
                              "buttons": [
                                  {
                                      "label": "OK",
                                      "cancel": false,
                                      "primary": true
                                  }
                              ]
                          }).then(function(result){
                              console.log(result);
                          });
                      }
                  } else {
                      prompt({
                          "title": "Error",
                          "message": response.data.info,
                          "buttons": [
                              {
                                  "label": "OK",
                                  "cancel": false,
                                  "primary": true
                              }
                          ]
                      }).then(function(result){
                          console.log(result);
                      });
                  };
              });
          };
      };

      /**
       * Logout the user from verto server and
       * redirects him to login page.
       */
      $rootScope.logout = function() {
        var disconnect = function() {
          verto.data.logoutProcess = true;
          var disconnectCallback = function(v, connected) {
            console.debug('Redirecting to login page.');
            storage.data.login = '';
            storage.data.password = '';
            verto.data.password = '';
            storage.reset();
			if (typeof gapi !== 'undefined'){
				console.debug(gapi);
				gapi.auth.signOut();
			}
            $location.path('/login');
          };

          if (verto.data.call) {
            verto.hangup();
          }

          $scope.closeChat();
          verto.disconnect(disconnectCallback);

          verto.hangup();
        };

        if (verto.data.call) {
          prompt({
            title: 'Oops, Active Call in Course.',
            message: 'It seems that you are in a call. Do you want to hang up?'
          }).then(function() {
            disconnect();
          });
        } else {
          disconnect();
        }

      };

      /**
       * Shows a modal with the settings.
       */
      $scope.openModalSettings = function() {
        var modalInstance = $modal.open({
          animation: $scope.animationsEnabled,
          templateUrl: 'partials/modal_settings.html',
          controller: 'ModalSettingsController',
        });

        modalInstance.result.then(
          function(result) {
            console.log(result);
          },
          function() {
            console.info('Modal dismissed at: ' + new Date());
          }
        );

        modalInstance.rendered.then(
          function() {
            jQuery.material.init();
          }
        );
      };

      $rootScope.openModal = function(templateUrl, controller, _options) {
        var options = {
          animation: $scope.animationsEnabled,
          templateUrl: templateUrl,
          controller: controller,
        };
        
        angular.extend(options, _options);

        var modalInstance = $modal.open(options);

        modalInstance.result.then(
          function(result) {
            console.log(result);
          },
          function() {
            console.info('Modal dismissed at: ' + new Date());
          }
        );

        modalInstance.rendered.then(
          function() {
            jQuery.material.init();
          }
        );
        
        return modalInstance;
      };

      $rootScope.$on('ws.close', onWSClose);
      $rootScope.$on('ws.login', onWSLogin);

      var ws_modalInstance;

      function onWSClose(ev, data) {
        if(ws_modalInstance || verto.data.logoutProcess) {
          return;
        };
        var options = {
          backdrop: 'static',
          keyboard: false
        };
        ws_modalInstance = $scope.openModal('partials/ws_reconnect.html', 'ModalWsReconnectController', options);
      };

      function onWSLogin(ev, data) {
        if(!ws_modalInstance) {
          return;
        };

        ws_modalInstance.close();
        ws_modalInstance = null;
      };

      $scope.showAbout = function() {
        $scope.openModal('partials/about.html', 'AboutController');
      };

      $scope.showContributors = function() {
        $scope.openModal('partials/contributors.html', 'ContributorsController');
      };

      /**
       * Updates the display adding the new number touched.
       *
       * @param {String} number - New touched number.
       */
      $rootScope.dtmf = function(number) {
        $rootScope.dialpadNumber = $scope.dialpadNumber + number;
        if (verto.data.call) {
          verto.dtmf(number);
        }
      };

      /**
       * Removes the last character from the number.
       */
      $rootScope.backspace = function() {
        var number = $rootScope.dialpadNumber;
        var len = number.length;
        $rootScope.dialpadNumber = number.substring(0, len - 1);
      };


      $scope.toggleCallHistory = function() {
        if (!$scope.call_history) {
          angular.element("#call_history").addClass('active');
          angular.element("#call-history-wrapper").addClass('active');
        } else {
          angular.element("#call_history").removeClass('active');
          angular.element("#call-history-wrapper").removeClass('active');
        }
        $scope.call_history = angular.element("#call_history").hasClass('active');
      };

      $scope.toggleChat = function() {
        if ($rootScope.chatStatus && $rootScope.activePane === 'chat') {
          $rootScope.chat_counter = 0;
        }
        angular.element('#wrapper').toggleClass('toggled');
        $rootScope.chatStatus = angular.element('#wrapper').hasClass('toggled');
      };
          
      $scope.copyLink = function () {
          var text = window.location.origin + '/#?join=' + verto.data.meetingId;
          var copyText = $( "#shareLinkInput" );
          if (copyText.css('display') !== 'none')
            return;

          copyText.val(text);
          copyText.animate({
              opacity: 1,
              left: "+=50",

              width: "toggle"
          }, 1000, function() {
              // Animation complete.
              setTimeout(function () {
                  copyText.hide()
              }, 10000);
          });

          copyText.select();
          try {
              var successful = document.execCommand('copy');
              var msg = successful ? 'successful' : 'unsuccessful';
              console.log('Copying text command was ' + msg);
          } catch (err) {
              console.log('Oops, unable to copy');
          };
      }
          
      $rootScope.openChat = function() {
        $rootScope.chatStatus = false;
        angular.element('#wrapper').removeClass('toggled');
      };

      $scope.closeChat = function() {
        $rootScope.chatStatus = true;
        angular.element('#wrapper').addClass('toggled');
      };

      $scope.goFullscreen = function() {
        if (storage.data.userStatus !== 'connected') {
          return;
        }
        $rootScope.fullscreenEnabled = !Fullscreen.isEnabled();
        if (Fullscreen.isEnabled()) {
          Fullscreen.cancel();
        } else {
          Fullscreen.enable(document.getElementsByTagName('body')[0]);
        }
      };

      $rootScope.$on('call.video', function(event) {
        storage.data.videoCall = true;
      });

      $rootScope.$on('call.hangup', function(event, data) {
        if (Fullscreen.isEnabled()) {
          Fullscreen.cancel();
        }

        if (!$rootScope.chatStatus) {
          angular.element('#wrapper').toggleClass('toggled');
          $rootScope.chatStatus = angular.element('#wrapper').hasClass('toggled');
        }

        $rootScope.dialpadNumber = '';
        console.debug('Redirecting to dialpad page.');
        $location.path('/dialpad');

        try {
          $rootScope.$digest();
        } catch (e) {
          console.log('not digest');
        }
      });

      $rootScope.$on('page.incall', function(event, data) {
        var page_incall = function() {
          return $q(function(resolve, reject) {
            if (storage.data.askRecoverCall) {
              prompt({
                title: 'Oops, Active Call in Course.',
                message: 'It seems you were in a call before leaving the last time. Wanna go back to that?'
              }).then(function() {
                console.log('redirect to incall page');
                $location.path('/incall');
              }, function() {
                storage.data.userStatus = 'connecting';
                verto.hangup();
              });
            } else {
              console.log('redirect to incall page');
              $location.path('/incall');
            }
            resolve();
          });
        };
        eventQueue.events.push(page_incall);
      });

      $scope.$on('event:google-plus-signin-success', function (event,authResult) {
        // Send login to server or save into cookie
        console.log('Google+ Login Success');
        console.log(authResult);
        gapi.client.load('plus', 'v1', gapiClientLoaded);
      });

      function gapiClientLoaded() {
	    gapi.client.plus.people.get({userId: 'me'}).execute(handleEmailResponse);
      }

      function handleEmailResponse(resp){
        var primaryEmail;
	for (var i=0; i < resp.emails.length; i++) {
	  if (resp.emails[i].type === 'account') primaryEmail = resp.emails[i].value;
        }
	console.debug("Primary Email: " + primaryEmail );
	console.debug("display name: " + resp.displayName);
	console.debug("imageurl: " + resp.image.url);
	console.debug(resp);
	console.debug(verto.data);
	verto.data.email = primaryEmail;
	verto.data.name = resp.displayName;
	storage.data.name = verto.data.name;
	storage.data.email = verto.data.email;

	$scope.login();
      }

      $scope.$on('event:google-plus-signin-failure', function (event,authResult) {
        // Auth failure or signout detected
        console.log('Google+ Login Failure');
      });

      $rootScope.callActive = function(data, params) {
        verto.data.mutedMic = storage.data.mutedMic;
        verto.data.mutedVideo = storage.data.mutedVideo;

        if (!storage.data.cur_call) {
          storage.data.call_start = new Date();
        }
        storage.data.userStatus = 'connected';
        var call_start = new Date(storage.data.call_start);
        $rootScope.start_time = call_start;

        $timeout(function() {
          $scope.$broadcast('timer-start');
        });
        myVideo.play();
        storage.data.calling = false;

        storage.data.cur_call = 1;

        $location.path('/incall');

        if(params.useVideo) {
          $rootScope.$emit('call.video', 'video');
        }
      };

      $rootScope.$on('call.active', function(event, data, params) {
        $rootScope.callActive(data, params);
      });

      $rootScope.$on('call.calling', function(event, data) {
        storage.data.calling = true;
      });

      $rootScope.$on('call.incoming', function(event, data) {
        console.log('Incoming call from: ' + data);

        storage.data.cur_call = 0;
        $scope.incomingCall = true;
        storage.data.videoCall = false;
        storage.data.mutedVideo = false;
        storage.data.mutedMic = false;

        prompt({
          title: 'Incoming Call',
          message: 'from ' + data
        }).then(function() {
          var call_start = new Date(storage.data.call_start);
          $rootScope.start_time = call_start;
          console.log($rootScope.start_time);

          $scope.answerCall();
          storage.data.called_number = data;
          CallHistory.add(data, 'inbound', true);
          $location.path('/incall');
        }, function() {
          $scope.declineCall();
          CallHistory.add(data, 'inbound', false);
        });
      });

      $scope.hold = function() {
        storage.data.onHold = !storage.data.onHold;
        verto.data.call.toggleHold();
      };

      /**
       * Hangup the current call.
       */
      $scope.hangup = function() {
        if (!verto.data.call) {
          toastr.warning('There is no call to hangup.');
          $location.path('/dialpad');
          return;
        }

        //var hangupCallback = function(v, hangup) {
        //  if (hangup) {
        //    $location.path('/dialpad');
        //  } else {
        //    console.debug('The call could not be hangup.');
        //  }
        //};
        //
        //verto.hangup(hangupCallback);
        if (verto.data.shareCall) {
          verto.screenshareHangup();
        }

        verto.hangup();

        $location.path('/dialpad');
      };

      $scope.answerCall = function() {
        storage.data.onHold = false;

        verto.data.call.answer({
          useStereo: storage.data.useStereo,
          useCamera: storage.data.selectedVideo,
          useVideo: storage.data.useVideo,
          useMic: storage.data.useMic,
          callee_id_name: verto.data.name,
          callee_id_number: verto.data.login
        });


        $location.path('/incall');
      };

      $scope.declineCall = function() {
        $scope.hangup();
        $scope.incomingCall = false;
      };

      $scope.screenshare = function() {
        if (verto.data.shareCall) {
          verto.screenshareHangup();
          return false;
        }
        verto.screenshare(storage.data.called_number);
      };

      $scope.play = function() {
        var file = $scope.promptInput('Please, enter filename', '', 'File',
          function(file) {
            verto.data.conf.play(file);
            console.log('play file :', file);
          });

      };

      $scope.stop = function() {
        verto.data.conf.stop();
      };

      $scope.record = function() {
        var file = $scope.promptInput('Please, enter filename', '', 'File',
          function(file) {
            verto.data.conf.record(file);
            console.log('recording file :', file);
          });
      };

      $scope.stopRecord = function() {
        verto.data.conf.stopRecord();
      };

      $scope.snapshot = function() {
        var file = $scope.promptInput('Please, enter filename', '', 'File',
          function(file) {
            verto.data.conf.snapshot(file);
            console.log('snapshot file :', file);
          });
      };

          }]
  );

})();

(function() {
  'use strict';

  angular
    .module('vertoControllers')
	.controller('MenuController', ['$scope', '$http', '$location',
	  'verto', 'storage',
	  function($scope, $http, $location, verto, storage) {
	    console.debug('Executing MenuController.');
	  }
	]);

})();
(function() {
  'use strict';

  angular
    .module('vertoControllers')
    .controller('ModalDialpadController', ['$scope',
      '$modalInstance',
      function($scope, $modalInstance) {

        $scope.ok = function() {
          $modalInstance.close('Ok.');
        };

        $scope.cancel = function() {
          $modalInstance.dismiss('cancel');
        };

      }
    ]);
})();
(function() {
  'use strict';

  angular
    .module('vertoControllers')
    .controller('ModalWsReconnectController', ModalWsReconnectController);

    ModalWsReconnectController.$inject = ['$scope', 'storage', 'verto'];

    function ModalWsReconnectController($scope, storage, verto) {
      console.debug('Executing ModalWsReconnectController'); 
    };
                
                
})();

(function() {
  'use strict';

  angular
    .module('vertoControllers')
    .controller('ModalLoginInformationController', ['$scope',
      '$http', '$location', '$modalInstance', 'verto', 'storage',
      function($scope, $http, $location, $modalInstance, verto, storage) {
        console.debug('Executing ModalLoginInformationController.');

        $scope.verto = verto;
        $scope.storage = storage;

        $scope.ok = function() {
          $modalInstance.close('Ok.');
        };

        $scope.cancel = function() {
          $modalInstance.dismiss('cancel');
        };

      }
    ]);
    
})();
(function() {
  'use strict';

  angular
    .module('vertoControllers')
    .controller('ModalSettingsController', ['$scope', '$http',
      '$location', '$modalInstance', 'storage', 'verto',
      function($scope, $http, $location, $modalInstance, storage, verto) {
        console.debug('Executing ModalSettingsController.');

        $scope.storage = storage;
        $scope.verto = verto;
        $scope.mydata = angular.copy(storage.data);

        $scope.ok = function() {
          storage.changeData($scope.mydata);
          verto.data.instance.iceServers(storage.data.useSTUN);
          $modalInstance.close('Ok.');
        };

        $scope.cancel = function() {
          $modalInstance.dismiss('cancel');
        };

        $scope.refreshDeviceList = function() {
          return verto.refreshDevices();
        };

        $scope.resetSettings = function() {
	  if (confirm('Factory Reset Settings?')) {
            storage.factoryReset();
            $scope.logout();
            $scope.ok();
	    window.location.reload();
	  };
        };

        $scope.checkUseDedRemoteEncoder = function(option) {
          if ($scope.mydata.incomingBandwidth != 'default' || $scope.mydata.outgoingBandwidth != 'default') {
            $scope.mydata.useDedenc = true;
          } else {
            $scope.mydata.useDedenc = false;
          }
        };
      }
    ]);

})();

(function() {
  'use strict';
  var vertoDirectives = angular.module('vertoDirectives', []);
})();

/*
Sometimes autofocus HTML5 directive just isn't enough with SPAs.
This directive will force autofocus to work properly under those circumstances.
*/
(function () {
  'use strict';

  angular
  .module('vertoDirectives')
  .directive('autofocus', ['$timeout',
    function ($timeout) {
      return {
        restrict: 'A',
        link: function ($scope, $element) {
          $timeout(function () {
            $element[0].focus();
          });
        }
      };
    }
  ]);
})();
(function () {
  'use strict';

  angular
  .module('vertoDirectives')
  .directive('showControls',
  ["Fullscreen", function(Fullscreen) {
    var link = function(scope, element, attrs) {
      var i = null;
      jQuery('.video-footer').fadeIn('slow');
      jQuery('.video-hover-buttons').fadeIn('slow');
      element.parent().bind('mousemove', function() {
        if (Fullscreen.isEnabled()) {
          clearTimeout(i);
          jQuery('.video-footer').fadeIn('slow');
          jQuery('.video-hover-buttons').fadeIn(500);
          i = setTimeout(function() {
            if (Fullscreen.isEnabled()) {
              jQuery('.video-footer').fadeOut('slow');
              jQuery('.video-hover-buttons').fadeOut(500);
            }
          }, 3000);
        }
      });
      element.parent().bind('mouseleave', function() {
        jQuery('.video-footer').fadeIn();
        jQuery('.video-hover-buttons').fadeIn();
      });
    }


    return {
      link: link
    };
  }]);

})();
(function () {
  'use strict';

  angular
  .module('vertoDirectives').directive('userStatus',
  function() {
    var link = function(scope, element, attrs) {
      scope.$watch('condition', function(condition) {
        element.removeClass('connected');
        element.removeClass('disconnected');
        element.removeClass('connecting');

        element.addClass(condition);

      });
    }

    return {
      scope: {
        'condition': '='
      },
      link: link
    };

  });

})();
/**
 * To RTC work properly we need to give a <video> tag as soon as possible
 * because it needs to attach the video and audio stream to the tag.
 *
 * This directive is responsible for moving the video tag from the body to
 * the right place when a call start and move back to the body when the
 * call ends. It also hides and display the tag when its convenient.
 */
(function () {
  'use strict';

  angular
  .module('vertoDirectives')
  .directive('videoTag',
  function() {
    function link(scope, element, attrs) {
      // Moving the video tag to the new place inside the incall page.
      console.log('Moving the video to element.');
      jQuery('video').removeClass('hide').appendTo(element);
      jQuery('video').css('display', 'block');
      scope.callActive("", {useVideo: true});

      element.on('$destroy', function() {
        // Move the video back to the body.
        console.log('Moving the video back to body.');
        jQuery('video').addClass('hide').appendTo(jQuery('body'));
      });
    }

    return {
      link: link
    }
  });

})();
(function() {
  'use strict';
  var vertoService = angular.module('vertoService', []);
})();

'use strict';

/* Controllers */
var videoQuality = [];
var videoQualitySource = [{
  id: 'qvga',
  label: 'QVGA 320x240',
  width: 320,
  height: 240
}, {
  id: 'vga',
  label: 'VGA 640x480',
  width: 640,
  height: 480
}, {
  id: 'qvga_wide',
  label: 'QVGA WIDE 320x180',
  width: 320,
  height: 180
}, {
  id: 'vga_wide',
  label: 'VGA WIDE 640x360',
  width: 640,
  height: 360
}, {
  id: 'hd',
  label: 'HD 1280x720',
  width: 1280,
  height: 720
}, {
  id: 'hhd',
  label: 'HHD 1920x1080',
  width: 1920,
  height: 1080
}, ];

var videoResolution = {
  qvga: {
    width: 320,
    height: 240
  },
  vga: {
    width: 640,
    height: 480
  },
  qvga_wide: {
    width: 320,
    height: 180
  },
  vga_wide: {
    width: 640,
    height: 360
  },
  hd: {
    width: 1280,
    height: 720
  },
  hhd: {
    width: 1920,
    height: 1080
  },
};

var bandwidth = [{
  id: '250',
  label: '250kb'
}, {
  id: '500',
  label: '500kb'
}, {
  id: '1024',
  label: '1mb'
}, {
  id: '1536',
  label: '1.5mb'
}, {
  id: '2048',
  label: '2mb'
}, {
  id: '5120',
  label: '5mb'
}, {
  id: '0',
  label: 'No Limit'
}, {
  id: 'default',
  label: 'Server Default'
}, ];

var vertoService = angular.module('vertoService', ['ngCookies']);

vertoService.service('verto', ['$rootScope', '$cookieStore', '$location', 'storage',
  function($rootScope, $cookieStore, $location, storage) {
    var data = {
      // Connection data.
      instance: null,
      connected: false,

      // Call data.
      call: null,
      shareCall: null,
      callState: null,
      conf: null,
      confLayouts: [],
      confRole: null,
      chattingWith: null,
      liveArray: null,

      // Settings data.
      videoDevices: [],
      audioDevices: [],
      shareDevices: [],
      videoQuality: [],
      extension: $cookieStore.get('verto_demo_ext'),
      name: $cookieStore.get('verto_demo_name'),
      email: $cookieStore.get('verto_demo_email'),
      cid: $cookieStore.get('verto_demo_cid'),
      textTo: $cookieStore.get('verto_demo_textto') || "1000",
      login: $cookieStore.get('verto_demo_login'),
      password: '', //$cookieStore.get('verto_demo_passwd'),
      hostname: window.location.hostname,
      wsURL: ("wss://" + window.location.hostname + ":8082")
    };

    function cleanShareCall(that) {
      data.shareCall = null;
      data.callState = 'active';
      that.refreshDevices();
    }

    function cleanCall() {
      data.call = null;
      data.callState = null;
      data.conf = null;
      data.confLayouts = [];
      data.confRole = null;
      data.chattingWith = null;

      $rootScope.$emit('call.hangup', 'hangup');
    }

    function inCall() {
      $rootScope.$emit('page.incall', 'call');
    }

    function callActive(last_state, params) {
      $rootScope.$emit('call.active', last_state, params);
    }

    function calling() {
      $rootScope.$emit('call.calling', 'calling');
    }

    function incomingCall(number) {
      $rootScope.$emit('call.incoming', number);
    }

    function updateResolutions(supportedResolutions) {
      console.debug('Attempting to sync supported and available resolutions');

      //var removed = 0;

      console.debug("VQ length: " + videoQualitySource.length);
      console.debug(supportedResolutions);

      angular.forEach(videoQualitySource, function(resolution, id) {
        angular.forEach(supportedResolutions, function(res) {
          var width = res[0];
          var height = res[1];

          if(resolution.width == width && resolution.height == height) {
		videoQuality.push(resolution);
          }
        });
      });

      // videoQuality.length = videoQuality.length - removed;
      console.debug("VQ length 2: " + videoQuality.length);
      data.videoQuality = videoQuality;
      console.debug(videoQuality);
      data.vidQual = (videoQuality.length > 0) ? videoQuality[videoQuality.length - 1].id : null;
      console.debug(data.vidQual);

      return videoQuality;
    };

    var callState = {
      muteMic: false,
      muteVideo: false
    };

    return {
      data: data,
      callState: callState,

      // Options to compose the interface.
      videoQuality: videoQuality,
      videoResolution: videoResolution,
      bandwidth: bandwidth,

      refreshDevicesCallback : function refreshDevicesCallback(callback) {
        data.videoDevices = [{
	  id: 'none',
	  label: 'No Camera'
	}];
        data.shareDevices = [{
          id: 'screen',
          label: 'Screen'
        }];
        data.audioDevices = [];

        if(!storage.data.selectedShare) {
          storage.data.selectedShare = data.shareDevices[0]['id'];
        }

        for (var i in jQuery.verto.videoDevices) {
          var device = jQuery.verto.videoDevices[i];
          if (!device.label) {
            data.videoDevices.push({
              id: 'Camera ' + i,
              label: 'Camera ' + i
            });
          } else {
            data.videoDevices.push({
              id: device.id,
              label: device.label || device.id
            });
          }

          // Selecting the first source.
          if (i == 0 && !storage.data.selectedVideo) {
            storage.data.selectedVideo = device.id;
          }

          if (!device.label) {
            data.shareDevices.push({
              id: 'Share Device ' + i,
              label: 'Share Device ' + i
            });
            continue;
          }

          data.shareDevices.push({
            id: device.id,
            label: device.label || device.id
          });
        }

        for (var i in jQuery.verto.audioInDevices) {
          var device = jQuery.verto.audioInDevices[i];
          // Selecting the first source.
          if (i == 0 && !storage.data.selectedAudio) {
            storage.data.selectedAudio = device.id;
          }

          if (!device.label) {
            data.audioDevices.push({
              id: 'Microphone ' + i,
              label: 'Microphone ' + i
            });
            continue;
          }
          data.audioDevices.push({
            id: device.id,
            label: device.label || device.id
          });
        }
        console.debug('Devices were refreshed, checking that we have cameras.');

        // This means that we cannot use video!
        if (data.videoDevices.length === 0) {
          console.log('No camera, disabling video.');
          data.canVideo = false;
          data.videoDevices.push({
            id: 'none',
            label: 'No camera'
          });
        } else {
          data.canVideo = true;
        }

        if(angular.isFunction(callback)) {
          callback();
        }
      },

      refreshDevices: function(callback) {
        console.debug('Attempting to refresh the devices.');
        if(callback) {
          jQuery.verto.refreshDevices(callback);
        } else {
          jQuery.verto.refreshDevices(this.refreshDevicesCallback);
        }
      },

      /**
       * Updates the video resolutions based on settings.
       */
      refreshVideoResolution: function(resolutions) {
        console.debug('Attempting to refresh video resolutions.');

        if (data.instance) {
          var w = resolutions['bestResSupported'][0];
          var h = resolutions['bestResSupported'][1];

          if (h === 1080) {
            w = 1280;
            h = 720;
          }

          updateResolutions(resolutions['validRes']);
          data.instance.videoParams({
            minWidth: w,
            minHeight: h,
            maxWidth: w,
            maxHeight: h,
            minFrameRate: 15,
            vertoBestFrameRate: 30
          });
          videoQuality.forEach(function(qual){
            if (w === qual.width && h === qual.height) {
              if (storage.data.vidQual !== qual.id || storage.data.vidQual === undefined) {
                storage.data.vidQual = qual.id;
              }
            }

          });

        } else {
          console.debug('There is no instance of verto.');
        }
      },

      /**
       * Connects to the verto server. Automatically calls `onWSLogin`
       * callback set in the verto object.
       *
       * @param callback
       */
      connect: function(callback) {
        console.debug('Attempting to connect to verto.');
        var that = this;

        function startConference(v, dialog, pvtData) {
          $rootScope.$emit('call.video', 'video');
          $rootScope.$emit('call.conference', 'conference');
          data.chattingWith = pvtData.chatID;
          data.confRole = pvtData.role;

          var conf = new $.verto.conf(v, {
            dialog: dialog,
            hasVid: storage.data.useVideo,
            laData: pvtData,
            chatCallback: function(v, e) {
              var from = e.data.fromDisplay || e.data.from || "Unknown";
              var message = e.data.message || "";
              $rootScope.$emit('chat.newMessage', {
                from: from,
                body: message
              });
            },
            onBroadcast: function(v, conf, message) {
              console.log('>>> conf.onBroadcast:', arguments);
              if (message.action == 'response') {
                // This is a response with the video layouts list.
                if (message['conf-command'] == 'list-videoLayouts') {
                  data.confLayouts = message.responseData.sort();
                } else {
                  $rootScope.$emit('conference.broadcast', message);
                }
              }
            }
          });

          if (data.confRole == "moderator") {
            console.log('>>> conf.listVideoLayouts();');
            conf.listVideoLayouts();
          }

          data.conf = conf;

          data.liveArray = new $.verto.liveArray(
            data.instance, pvtData.laChannel,
            pvtData.laName, {
              subParams: {
                callID: dialog ? dialog.callID : null
              }
            });

          data.liveArray.onErr = function(obj, args) {
            console.log('liveArray.onErr', obj, args);
          };

          data.liveArray.onChange = function(obj, args) {
            // console.log('liveArray.onChange', obj, args);

            switch (args.action) {
              case 'bootObj':
                $rootScope.$emit('members.boot', args.data);
                args.data.forEach(function(member){
                  var callId = member[0];
                  var status = angular.fromJson(member[1][4]);
                  if (callId === data.call.callID) {
                    $rootScope.$apply(function(){
                      data.mutedMic = status.audio.muted;
                      data.mutedVideo = status.video.muted;
                    });
                  }
                });
                break;
              case 'add':
                var member = [args.key, args.data];
                $rootScope.$emit('members.add', member);
                break;
              case 'del':
                var uuid = args.key;
                $rootScope.$emit('members.del', uuid);
                break;
              case 'clear':
                $rootScope.$emit('members.clear');
                break;
              case 'modify':
                var member = [args.key, args.data];
                $rootScope.$emit('members.update', member);
                break;
              default:
                console.log('NotImplemented', args.action);
            }
          };
        }

        function stopConference() {
          console.log('stopConference()');
          if (data.liveArray) {
            data.liveArray.destroy();
            console.log('Has data.liveArray.');
            $rootScope.$emit('members.clear');
            data.liveArray = null;
          } else {
            console.log('Doesn\'t found data.liveArray.');
          }

          if (data.conf) {
            data.conf.destroy();
            data.conf = null;
          }
        }

        var callbacks = {
          onWSLogin: function(v, success) {
            data.connected = success;
            $rootScope.$emit('ws.login', success);
            console.debug('Connected to verto server:', success);

            if (angular.isFunction(callback)) {
              callback(v, success);
            }
          },

          onMessage: function(v, dialog, msg, params) {
            console.debug('onMessage:', v, dialog, msg, params);

            switch (msg) {
              case $.verto.enum.message.pvtEvent:
                if (params.pvtData) {
                  switch (params.pvtData.action) {
                    case "conference-liveArray-join":
                      console.log("conference-liveArray-join");
                      stopConference();
                      startConference(v, dialog, params.pvtData);
                      break;
                    case "conference-liveArray-part":
                      console.log("conference-liveArray-part");
                      stopConference();
                      break;
                  }
                }
                break;
              /**
                * This is not being used for conferencing chat
                * anymore (see conf.chatCallback for that).
                */
              case $.verto.enum.message.info:
                var body = params.body;
                var from = params.from_msg_name || params.from;
                $rootScope.$emit('chat.newMessage', {
                  from: from,
                  body: body
                });
                break;
              default:
                console.warn('Got a not implemented message:', msg, dialog, params);
                break;
            }
          },

          onDialogState: function(d) {
            if (!data.call) {
              data.call = d;

            }

            console.debug('onDialogState:', d);
            switch (d.state.name) {
              case "ringing":
                incomingCall(d.params.caller_id_number);
                break;
              case "trying":
                console.debug('Calling:', d.cidString());
                data.callState = 'trying';
                break;
              case "early":
                console.debug('Talking to:', d.cidString());
                data.callState = 'active';
                calling();
                break;
              case "active":
                console.debug('Talking to:', d.cidString());
                data.callState = 'active';
                callActive(d.lastState.name, d.params);
                break;
              case "hangup":
                console.debug('Call ended with cause: ' + d.cause);
                data.callState = 'hangup';
                break;
              case "destroy":
                console.debug('Destroying: ' + d.cause);
                if (d.params.screenShare) {
                  cleanShareCall(that);
                } else {
                  stopConference();
                  if (!that.reloaded) {
                    cleanCall();
                  }
                }
                break;
              default:
                console.warn('Got a not implemented state:', d);
                break;
            }
          },

          onWSClose: function(v, success) {
            console.debug('onWSClose:', success);

            $rootScope.$emit('ws.close', success);
          },

          onEvent: function(v, e) {
            console.debug('onEvent:', e);
          }
        };

        var that = this;
        function ourBootstrap() {
          // Checking if we have a failed connection attempt before
          // connecting again.
          if (data.instance && !data.instance.rpcClient.socketReady()) {
              clearTimeout(data.instance.rpcClient.to);
              data.instance.logout();
	      data.instance.login();
	      return;
          };
          data.instance = new jQuery.verto({
            login: data.login + '@' + data.hostname,
            passwd: data.password,
            socketUrl: data.wsURL,
            tag: "webcam",
            ringFile: "sounds/bell_ring2.wav",
            // TODO: Add options for this.
            audioParams: {
                googEchoCancellation: storage.data.googEchoCancellation || true,
                googNoiseSuppression: storage.data.googNoiseSuppression || true,
                googHighpassFilter: storage.data.googHighpassFilter || true
            },
            iceServers: storage.data.useSTUN
          }, callbacks);

          // We need to know when user reloaded page and not react to
          // verto events in order to not stop the reload and redirect user back
          // to the dialpad.
          that.reloaded = false;
          jQuery.verto.unloadJobs.push(function() {
            that.reloaded = true;
          });
	    data.instance.deviceParams({
		useCamera: storage.data.selectedVideo,
		useMic: storage.data.selectedAudio,
		onResCheck: that.refreshVideoResolution
	    });
        }

        if (data.mediaPerm) {
          ourBootstrap();
        } else {
	    $.FSRTC.checkPerms(ourBootstrap, true, true);
        }
      },

      mediaPerm: function(callback) {
	  $.FSRTC.checkPerms(callback, true, true);
      },

      /**
       * Login the client.
       *
       * @param callback
       */
      login: function(callback) {
        data.instance.loginData({
          login: data.login + '@' + data.hostname,
          passwd: data.password
        });
        data.instance.login();

        if (angular.isFunction(callback)) {
          callback(data.instance, true);
        }
      },

      /**
       * Disconnects from the verto server. Automatically calls `onWSClose`
       * callback set in the verto object.
       *
       * @param callback
       */
      disconnect: function(callback) {
        console.debug('Attempting to disconnect to verto.');

        data.instance.logout();
        data.connected = false;

        console.debug('Disconnected from verto server.');

        if (angular.isFunction(callback)) {
          callback(data.instance, data.connected);
        }
      },

      /**
       * Make a call.
       *
       * @param callback
       */
      call: function(destination, callback) {
        console.debug('Attempting to call destination ' + destination + '.');

        var call = data.instance.newCall({
          destination_number: destination,
          caller_id_name: data.name,
          caller_id_number: data.callerid ? data.callerid : data.email,
          outgoingBandwidth: storage.data.outgoingBandwidth,
          incomingBandwidth: storage.data.incomingBandwidth,
          useVideo: storage.data.useVideo,
          useStereo: storage.data.useStereo,
          useCamera: storage.data.selectedVideo,
          useMic: storage.data.selectedAudio,
          dedEnc: storage.data.useDedenc,
          mirrorInput: storage.data.mirrorInput,
          userVariables: {
            email : storage.data.email,
            meetingId :  this.data.meetingId,
            avatar: "http://gravatar.com/avatar/" + md5(storage.data.email) + ".png?s=600"
          }
        });

        data.call = call;

        data.mutedMic = false;
        data.mutedVideo = false;

        this.refreshDevices();

        if (angular.isFunction(callback)) {
          callback(data.instance, call);
        }
      },

      screenshare: function(destination, callback) {
        console.log('share screen video');

        var that = this;

        getScreenId(function(error, sourceId, screen_constraints) {
          var call = data.instance.newCall({
            destination_number: destination + '-screen',
            caller_id_name: data.name + ' (Screen)',
            caller_id_number: data.login + ' (Screen)',
            outgoingBandwidth: storage.data.outgoingBandwidth,
            incomingBandwidth: storage.data.incomingBandwidth,
            videoParams: screen_constraints.video.mandatory,
            useVideo: storage.data.useVideo,
            screenShare: true,
            dedEnc: storage.data.useDedenc,
            mirrorInput: storage.data.mirrorInput,
            userVariables: {
              email : storage.data.email,
              avatar: "http://gravatar.com/avatar/" + md5(storage.data.email) + ".png?s=600"
            }
          });

          data.shareCall = call;

          console.log('shareCall', data);

          data.mutedMic = false;
          data.mutedVideo = false;

          that.refreshDevices();

        });

      },

      screenshareHangup: function() {
        if (!data.shareCall) {
          console.debug('There is no call to hangup.');
          return false;
        }

        console.log('shareCall End', data.shareCall);
        data.shareCall.hangup();

        console.debug('The screencall was hangup.');

      },

      /**
       * Hangup the current call.
       *
       * @param callback
       */
      hangup: function(callback) {
        console.debug('Attempting to hangup the current call.');

        if (!data.call) {
          console.debug('There is no call to hangup.');
          return false;
        }

        data.call.hangup();

        if (data.conf) {
          data.conf.destroy();
          data.conf = null;
        }

        console.debug('The call was hangup.');

        if (angular.isFunction(callback)) {
          callback(data.instance, true);
        }
      },

      /**
       * Send a DTMF to the current call.
       *
       * @param {string|integer} number
       * @param callback
       */
      dtmf: function(number, callback) {
        console.debug('Attempting to send DTMF "' + number + '".');

        if (!data.call) {
          console.debug('There is no call to send DTMF.');
          return false;
        }

        data.call.dtmf(number);
        console.debug('The DTMF was sent for the call.');

        if (angular.isFunction(callback)) {
          callback(data.instance, true);
        }
      },

      /**
       * Mute the microphone for the current call.
       *
       * @param callback
       */
      muteMic: function(callback) {
        console.debug('Attempting to mute mic for the current call.');

        if (!data.call) {
          console.debug('There is no call to mute.');
          return false;
        }

        data.call.dtmf('0');
        data.mutedMic = !data.mutedMic;
        console.debug('The mic was muted for the call.');

        if (angular.isFunction(callback)) {
          callback(data.instance, true);
        }
      },

      /**
       * Mute the video for the current call.
       *
       * @param callback
       */
      muteVideo: function(callback) {
        console.debug('Attempting to mute video for the current call.');

        if (!data.call) {
          console.debug('There is no call to mute.');
          return false;
        }

        data.call.dtmf('*0');
        data.mutedVideo = !data.mutedVideo;
        console.debug('The video was muted for the call.');

        if (angular.isFunction(callback)) {
          callback(data.instance, true);
        }
      },
      /*
      * Method is used to send conference chats ONLY.
      */
      sendConferenceChat: function(message) {
        data.conf.sendChat(message, "message");
      },
      /*
      * Method is used to send user2user chats.
      * VC does not yet support that.
      */
      sendMessage: function(body, callback) {
        data.call.message({
          to: data.chattingWith,
          body: body,
          from_msg_name: data.name,
          from_msg_number: data.cid
        });

        if (angular.isFunction(callback)) {
          callback(data.instance, true);
        }
      }
    };
  }
]);

'use strict';

var vertoService = angular.module('vertoService');

vertoService.service('config', ['$rootScope', '$http', '$location', 'storage', 'verto',
  function($rootScope, $http, $location, storage, verto) {
    var configure = function() {
      /**
       * Load stored user info into verto service
       */
      if(storage.data.name) {
        verto.data.name = verto.data.name || storage.data.name;
      }
      if(storage.data.email) {
        verto.data.email = verto.data.email || storage.data.email;
      }
      //if(storage.data.login) {
      //  verto.data.login = storage.data.login;
      //}
      if(storage.data.password) {
        verto.data.password = verto.data.password || storage.data.password;
      }

      /*
       * Load the Configs before logging in
       * with cache buster
       */
      var url = window.location.origin + window.location.pathname;
      url += 'config.json?cachebuster=' + Math.floor((Math.random() * 1000000) + 1);
      if (verto.data.email) {
          //url += '&email=' + verto.data.email;
      };

      if (verto.data.name) {
          //url += '&name=' + verto.data.name;
      };
      if (verto.data.password) {
          //url += '&password=' + verto.data.password;
      };

      var _requestData = {
          "email": verto.data.email,
          "join": $location.search().join || verto.data.join,
          "name": verto.data.name,
          "password": verto.data.password
      }

      var httpRequest = $http.post(url, _requestData);

      var httpReturn = httpRequest.then(function(response) {
        var data = response.data;

        /* save these for later as we're about to possibly over write them */
        var name = verto.data.name;
        var email = verto.data.email;

        console.debug("googlelogin: " + data.googlelogin);
        if (data.googlelogin){
          verto.data.googlelogin = data.googlelogin;
          verto.data.googleclientid = data.googleclientid;
        }
        
        angular.extend(verto.data, data);

        /**
         * use stored data (localStorage) for login, allow config.json to take precedence
         */

        if (name != '' && data.name == '') {
          verto.data.name = name;
        }
        if (email != '' && data.email == '') {
          verto.data.email = email;
        }
        //if (verto.data.login == '' && verto.data.password == '' && storage.data.login != '' && storage.data.password != '') {
        //  verto.data.login = storage.data.login;
        //  verto.data.password = storage.data.password;
        //}

        if (verto.data.autologin == "true" && !verto.data.autologin_done) {
          console.debug("auto login per config.json");
          verto.data.autologin_done = true;
        }
        
        if(data.valid && verto.data.autologin && verto.data.name.length && verto.data.email.length && verto.data.login.length && verto.data.password.length) {
          $rootScope.$emit('config.http.success', data);
        };
        return response;
      }, function(response) {
        $rootScope.$emit('config.http.error', response);        
        return response;
      });

      return httpReturn;
    };

    return {
      'configure': configure
    };
  }]);


'use strict';

  angular
    .module('vertoService')
    .service('eventQueue', ['$rootScope', '$q', 'storage', 'verto',
      function($rootScope, $q, storage, verto) {
        
        var events = [];
        
        var next = function() {
          var fn, fn_return;
          
          fn = events.shift();
          
          if (fn == undefined) {
            $rootScope.$emit('eventqueue.complete');
            return;
          }
          fn_return = fn();

          var emitNextProgress = function() {
            $rootScope.$emit('eventqueue.next');
          };

          fn_return.then(
            function() {
              emitNextProgress();
            }, 
            function() {
              emitNextProgress();
            }
          );
        };

        var process = function() {
          $rootScope.$on('eventqueue.next', function (ev){
            next();
          });
          
          next(); 
        };

        return {
          'next': next,
          'process': process,
          'events': events
        };

      }]);
 

(function() {
  'use strict';
  var vertoService = angular.module('storageService', ['ngStorage']);
})();

'use strict';

  angular
  .module('storageService')
  .service('storage', ['$rootScope', '$localStorage',
  function($rootScope, $localStorage) {
    var data = $localStorage,
	defaultSettings = {
	  ui_connected: false,
          ws_connected: false,
          cur_call: 0,
          called_number: '',
          useVideo: true,
          call_history: {},
          history_control: [],
          call_start: false,
          name: '',
          email: '',
          login: '',
          password: '',
          userStatus: 'disconnected',
          mutedVideo: false,
          mutedMic: false,
          selectedVideo: null,
          selectedAudio: null,
          selectedShare: null,
          useStereo: true,
          useSTUN: true,
          useDedenc: false,
          mirrorInput: false,
          outgoingBandwidth: 'default',
          incomingBandwidth: 'default',
          vidQual: undefined,
          askRecoverCall: false,
          googNoiseSuppression: true,
          googHighpassFilter: true,
          googEchoCancellation: true
       };

    data.$default(defaultSettings);

    function changeData(verto_data) {
      jQuery.extend(true, data, verto_data);
    }

    return {
      data: data,
      changeData: changeData,
      reset: function() {
        data.ui_connected = false;
        data.ws_connected = false;
        data.cur_call = 0;
        data.userStatus = 'disconnected';
      },
      factoryReset: function() {
        localStorage.clear();
        // set defaultSettings again
        data.$reset(defaultSettings);
      },
    };
  }
]);

'use strict';

  angular
  .module('storageService')
  .factory('CallHistory', ["storage", function(storage) {

    var history = storage.data.call_history;
    var history_control = storage.data.history_control;

    var addCallToHistory = function(number, direction, status) {
      if(history[number] == undefined) {
        history[number] = [];
      }

      history[number].unshift({
        'number': number,
        'direction': direction,
        'status': status,
        'call_start': Date()
      });

      var index = history_control.indexOf(number);
      console.log(index);
      if(index > -1) {
        history_control.splice(index, 1);
      }

      history_control.unshift(number);

    };

    var getCallsFromHistory = function(number) {
      return history[number];
    };

    return {
      all: function() {
        return history;
      },
      all_control: function() {
        return history_control;
      },
      get: function(number) {
        return getCallsFromHistory(number);
      },
      add: function(number, direction, status) {
        return addCallToHistory(number, direction, status);
      },
      clear: function() {
        storage.data.call_history = {};
        storage.data.history_control = [];
        history = storage.data.call_history;
        history_control = storage.data.history_control;
        return history_control;
      }
    }
}]);

'use strict';

  angular
    .module('storageService')
    .service('splashscreen', ['$rootScope', '$q', 'storage', 'config', 'verto',
      function($rootScope, $q, storage, config, verto) {
        
        var checkBrowser = function() {
          return $q(function(resolve, reject) {
            var activity = 'browser-upgrade';
            var result = {
              'activity': activity,
              'soft': false,
              'status': 'success',
              'message': 'Checking browser compability.'
            };

            navigator.getUserMedia = navigator.getUserMedia ||
              navigator.webkitGetUserMedia ||
              navigator.mozGetUserMedia;

            if (!navigator.getUserMedia) {
              result['status'] = 'error';
              result['message'] = 'Error: browser doesn\'t support WebRTC.';
              reject(result); 
            }

            resolve(result); 

          });
        };

        var checkMediaPerm = function() {
          return $q(function(resolve, reject) {
            var activity = 'media-perm';
            var result = {
              'activity': activity,
              'soft': false,
              'status': 'success',
              'message': 'Checking media permissions'
            };

            verto.mediaPerm(function(status) {
              if(!status) {
                result['status'] = 'error';
                result['message'] = 'Error: Media Permission Denied';
                verto.data.mediaPerm = false;
                reject(result);
              }
              verto.data.mediaPerm = true;
              resolve(result); 
            });
          });
        };

        var refreshMediaDevices = function() {
          return $q(function(resolve, reject) {
            var activity = 'refresh-devices';
            var result = {
              'status': 'success',
              'soft': true,
              'activity': activity,
              'message': 'Refresh Media Devices.'
            };
            
            verto.refreshDevices(function(status) {
              verto.refreshDevicesCallback(function() {
                resolve(result);
              });
            });

          });
        };

        var provisionConfig = function() {
          return $q(function(resolve, reject) {
            var activity = 'provision-config';
            var result = {
              'status': 'promise',
              'soft': true,
              'activity': activity,
              'message': 'Provisioning configuration.'
            };

            var configResponse = config.configure();

            var configPromise = configResponse.then(
              function(response) {
                /**
                 * from angular docs:
                 * A response status code between 200 and 299 is considered a success status and will result in the success callback being called
                 */
                if(response.status >= 200 && response.status <= 299) {
                  return result;
                } else {
                  result['status'] = 'error';
                  result['message'] = 'Error: Provision failed.';
                  return result;
                }
              });

              result['promise'] = configPromise;
              
              resolve(result);
          });
        };

        var checkLogin = function() {
          return $q(function(resolve, reject) {
            var activity = 'check-login';
            var result = {
              'status': 'success',
              'soft': true,
              'activity': activity,
              'message': 'Checking login.'
            };

            if(verto.data.connecting || verto.data.connected) {
              resolve(result);
              return;
            };

            var checkUserStored = function() {
              /**
               * if user data saved, use stored data for logon and not connecting
               * not connecting prevent two connects
               */
              if (storage.data.ui_connected && storage.data.ws_connected && !verto.data.connecting) {
                verto.data.name = storage.data.name;
                verto.data.email = storage.data.email;
                verto.data.login = storage.data.login;
                verto.data.password = storage.data.password;

                verto.data.connecting = true;
                verto.connect(function(v, connected) {
                  verto.data.connecting = false;
                  resolve(result);
                });
              }; 
            };
              resolve(result);
              return;
              if(storage.data.ui_connected && storage.data.ws_connected) {
              checkUserStored(); 
            } else {

            };
          });
        };

        var progress = [
          checkBrowser,
          checkMediaPerm,
          refreshMediaDevices,
          provisionConfig,
          checkLogin
        ];

        var progress_message = [
          'Checking browser compability.',
          'Checking media permissions',
          'Refresh Media Devices.',
          'Provisioning configuration.',
          'Checking login.'
        ];
        
        var getProgressMessage = function(current_progress) {
          if(progress_message[current_progress] != undefined) {
            return progress_message[current_progress]; 
          } else {
            return 'Please wait...';
          }
        };

        var current_progress = -1;
        var progress_percentage = 0;

        var calculateProgress = function(index) {
          var _progress;
          
          _progress = index + 1;
          progress_percentage = (_progress / progress.length) * 100;
          return progress_percentage;
        };

        var nextProgress = function() {
          var fn, fn_return, status, interrupt, activity, soft, message, promise;
          interrupt = false;
          current_progress++;
          
          if(current_progress >= progress.length) {
            $rootScope.$emit('progress.complete', current_progress);
            return;
          }
          
          fn = progress[current_progress];
          fn_return = fn();

          var emitNextProgress = function(fn_return) {
            if(fn_return['promise'] != undefined) {
              promise = fn_return['promise'];
            }

            status = fn_return['status'];
            soft = fn_return['soft'];
            activity = fn_return['activity'];
            message = fn_return['message'];

            if(status != 'success') {
              interrupt = true;
            }

            $rootScope.$emit('progress.next', current_progress, status, promise, activity, soft, interrupt, message);

          };

          fn_return.then(
            function(fn_return) {
              emitNextProgress(fn_return);
            },
            function(fn_return) {
              emitNextProgress(fn_return);
            }
          );
          
        };

        return {
          'next': nextProgress,
          'getProgressMessage': getProgressMessage,
          'progress_percentage': progress_percentage,
          'calculate': calculateProgress
        };

      }]);
 
