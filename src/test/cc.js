/**
 * Created by igor on 01.11.16.
 */

"use strict";
    
for (let i = 99950; i < 99999; i++) {
    webitel.httpApi('POST', `/api/v2/callcenter/agent/${i}/status?domain=10.10.10.144`, {status: 'Available'}, e => {
        if (e) console.error(e)
    });
    webitel.httpApi('POST', `/api/v2/callcenter/agent/${i}/state?domain=10.10.10.144`, {status: 'Waiting'}, e => {
        if (e) console.error(e)
    });
}