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


var domain = '10.10.10.144';
for (var j = 99950; j <= 99999; j++) {
    var i = '' + j;
    db.collection('extension').update({destination_number: i}, {
        "destination_number": i,
        "domain": domain,
        "userRef" : i + '@' + domain,
        "name" : "ext_" + i,
        "version": 2,
        "callflow": [
            {
                "setVar" : [
                    "ringback=$${us-ring}",
                    "transfer_ringback=$${uk-ring}",
                    "hangup_after_bridge=true",
                    "continue_on_fail=true"
                ]
            },
            {
                "echo" : ""
            }
        ]

    }, {upsert: true})

}