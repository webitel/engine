for (var i = 99950; i <= 99989; i++) {
    var a = {
        "login": "" + i,
        "role": "user",
        "password": "" + i,
        "parameters": [
            "password=" + i,
            "cc-agent=true",
            "cc-agent-busy-delay-time=3",
            `cc-agent-contact='{originate_timeout=15,presence_id=${i}@10.10.10.144}{webitel_call_uuid=` + "${create_uuid()},sip_invite_domain=10.10.10.144}${sofia_contact(*/"
                + `${i}@10.10.10.144)},` + "${verto_contact( " + `${i}@10.10.10.144)}'`,
            "cc-agent-max-no-answer=3",
            "cc-agent-no-answer-delay-time=10",
            "cc-agent-reject-delay-time=3",
            "cc-agent-wrap-up-time=3",
            "vm-enabled=false",
            "webitel-extensions="
        ],
        "variables": [
            "account_role=user"
        ]
    };
    webitel.httpApi('POST', '/api/v2/accounts?domain=10.10.10.144',  a, (e)=> {if (e) console.error(e)})
}