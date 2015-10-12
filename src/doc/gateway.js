/**
 * "Включить шлюз"
 * gwName{String} - название шлюза
 * profileName{String} - название профайла на каком поднимать, по дефолту external
 */
webitel.gatewayUp(
    gwName,
    profileName,
    function(res) {
        if (res.status === 0 && this.responseText === "+OK attached\n") {
            // OK
        } else {
            // Беда
        }
    }
);

/**
 * "Отключить шлюз"
 * gwName{String} - название шлюза
 */
webitel.gatewayDown(
    gwName,
    function(res) {
        if (res.status === 0 && this.responseText === "+OK detached\n") {
            // OK
        } else {
            // Беда
        }
    }
);

/**
 * "Список шлюзов"
 * domainName{String} - название домена
 **/
webitel.gatewayList(
    domainName,
    function(res) {
        if (res.status === 0) {
            // OK
        } else {
            // Беда
        }
    }
);

/**
 * "Изменить параметры шлюза"
 * gwName{String} - название шлюза
 * sendParams{Array} - масив объектов с name, value
 **/
var sendParams = [
    {
        "name": "realm",
        "value": "pre.webitel.com"
    },
    {
        "name": "foo",
        "value": "bar"
    }
]
webitel.gatewayChange(
    gwName,
    "params",
    sendParams,
    function (res) {
        if (res.status === 0) {
            // OK
        } else {
            // Беда
        }
    }
);


/**
 * Создать шлюз
 * options{Object} - параметры шлюза
 */
var options = {
    "routes": {
        "default": {
            "name": "Kyivstar",
            "destination_number": "^(\d{6,12})$",
            "callflow": [
                {
                    "recordSession": "start"
                },
                {
                    "bridge": {
                        "endpoints": [
                            {
                                "type": "sipGateway",
                                "name": "myGateway",
                                "parameters": [
                                    "origination_caller_id_number=0443907830",
                                    "absolute_codec_string=PCMA,PCMU"
                                ],
                                "dialString": "&reg0.$1"
                            }
                        ]
                    }
                }
            ]
        },
        "spublic": {
            "name": "Public",
            "destination_number": ['1232222'],
            "callflow": [
                {
                    "hangup": "USER_BUSY"
                }
            ]
        }
    },
    "name": "myGateway",
    "profile": "external",
    "domain": "myDomain",
    "username": "webitel",
    "password": "***",
    "realm": "pre.webitel.com",
    "params": [{
        "name": "foo",
        "value": "bar"
    }],
    "var": [],
    "ivar": [],
    "ovar": []
};
webitel.gatewayCreate(
    options,
    function (res) {
        if (res.status === 0) {
            // OK
        } else {
            // Беда
        }
    }
);

/**
 * Удалить шлюз
 * gwName{String} - название шлюза,
 * profile{String} - название прафайла,
 */
webitel.gatewayKill(
    profile,
    gwName,
    function(res) {
        if (res.status === 0) {
            // OK
        } else {
            // Беда
        }
    }
);

