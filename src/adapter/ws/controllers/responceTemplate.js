/**
 * Created by Igor Navrotskyj on 06.08.2015.
 */

'use strict';

var handleSocketError = require(__appRoot + '/middleware/handleWebSocketError'),
    log = require(__appRoot + '/lib/log')(module)
    ;

module.exports = {
    getCommandResponseJSON: function (_ws, id, res) {
        try {
            var complete,
                response;

            if (res && typeof res['body'] == 'string') {
                complete = (res['body'].indexOf('-ERR') == 0 || res['body'].indexOf('-USAGE') == 0) ? "-ERR" : "+OK";
                response = res['body']
            } else {
                complete = res ? '+OK' : '-ERR: bad response';
                response = res;
            }

            _ws.send(JSON.stringify({
                'exec-uuid': id,
                'exec-complete': complete,
                'exec-response': response
            }));
        } catch (e) {
            handleSocketError(_ws);
            log.warn('Error send response');
        };
    },
    
    getCommandResponseJSONError: function (_ws, id, err) {
        try {
            _ws.send(JSON.stringify({
                'exec-uuid': id,
                'exec-complete': "-ERR",
                'exec-response': err && err.message
            }));
        } catch (e) {
            handleSocketError(_ws);
            log.warn('Error send response');
        };
    }
};