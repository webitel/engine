module.exports.WebitelCommandTypes = {
    Call: {
        name: 'call',
        perm: 0
    }, //+
    Hangup: {
        name: 'hangup',
        perm: 0
    }, //+
    Park: {
        name: '//api uuid_park',
        perm: 0
    },
    ToggleHold: {
        name: 'toggle_hold',
        perm: 0
    }, //+
    Hold: {
        name: 'hold',
        perm: 0
    }, //+
    UnHold: {
        name: 'unhold',
        perm: 0
    }, //+
    Transfer: {
        name: 'transfer',
        perm: 0
    }, //+
    AttendedTransfer: {
        name: 'attended_transfer',
        perm: 0
    }, //+
    Bridge: {
        name: 'bridge',
        perm: 0
    }, //+
    Dtmf: {
        name: 'dtmf',
        perm: 0
    }, //+
    Broadcast: {
        name: 'broadcast',
        perm: 0
    }, //+
    AttXfer: {
        name: 'att_xfer',
        perm: 0
    }, //+
    AttXfer2: {
        name: 'att_xfer2',
        perm: 0
    },
    AttXferBridge: {
        name: 'att_xfer_bridge',
        perm: 0
    },
    AttXferCancel: {
        name: 'att_xfer_cancel',
        perm: 0
    },
    VideoRefresh: {
        name: 'video_refresh',
        perm: 0
    },
    Auth: {
        name: 'auth',
        perm: 0
    }, // +-
    GetVar: {
        name: 'getvar',
        perm: 0
    }, // +-
    SetVar: {
        name: 'setvar',
        perm: 0
    }, // +-

    // SYS Api
    Domain: {
        List: {
            name: 'api domain list',
            perm: 2
        }, //+
        Create: {
            name: 'api domain create',
            perm: 2
        }, //+
        Remove: {
            name: 'api domain remove',
            perm: 2
        }, //+
        Item: {
            name: 'api domain',
            perm: 1
        }, //+
        Update: {
            name: 'api domain change',
            perm: 1
        } //+
    },
    Account: {
        List: {
            name: 'api account list',
            perm: 0
        }, //+
        Create: {
            name: 'api account create',
            perm: 1
        }, //+
        Change: {
            name: 'api account change',
            perm: 0
        }, // +
        Remove: {
            name: 'api account remove',
            perm: 1
        }, // +
        Item: {
            name: 'api account',
            perm: 0
        }
    },
    Device: {
        List: {
            name: 'api device list',
            perm: 0
        }, //+
        Create: {
            name: 'api device create',
            perm: 1
        }, //+
        Change: {
            name: 'api device change',
            perm: 0
        }, //+
        Remove: {
            name: 'api device remove',
            perm: 0
        } //+
    },
    ListUsers: {
        name: 'api list_users',
        perm: 0
    },
    SendCommandWebitel: {
        name: 'sendCommandWebitel',
        perm: 0
    },

    // Users
    Login: {
        name: 'login',
        perm: 0
    },
    Logout: {
        name: 'logout',
        perm: 0
    },
    ReloadAgents: {
        name: 'reloadAgents',
        perm: 1
    },
    Rawapi: {
        name: 'rawapi',
        perm: 2
    },
    Eavesdrop: {
        name: 'eavesdrop',
        perm: 0
    },
    Displace: {
        name: 'displace',
        perm: 0
    },
    Dump: {
        name: 'channel_dump',
        perm: 0
    },

    SipProfile: {
        List: {
            name: 'sip_profile_list',
            perm: 2
        },
        Rescan: {
            name: 'sip_gateway_rescan',
            perm: 2
        }
    },

    Gateway: {
        List: {
            name: 'sip_gateway_list',
            perm: 1
        },
        Create: {
            name: 'sip_gateway_create',
            perm: 1
        },
        Change: {
            name: 'sip_gateway_change',
            perm: 1
        },
        Remove: {
            name: 'sip_gateway_remove',
            perm: 1
        },
        Up: {
            name: 'sip_gateway_up',
            perm: 1
        },
        Down: {
            name: 'sip_gateway_down',
            perm: 1
        },
        Kill: {
            name: 'sip_gateway_kill',
            perm: 1
        }
    },

    Show: {
        Channel: {
            name: 'show_channel',
            perm: 0
        }
    },
    Token: {
        Generate: {
            name: 'token_generate',
            perm: 0
        }
    },

    Event: {
        On: {
            name: "subscribe",
            perm: 0
        },
        Off: {
            name: "unsubscribe",
            perm: 0
        }
    },

    CallCenter: {
        List: {
            name: "",
            perm: 1
        },

        Create: {
            name: "",
            perm: 1
        },

        Delete: {
            name: "",
            perm: 1
        },

        State: {
            name: "",
            perm: 1
        },

        Ready: {
            name: "cc ready",
            perm: 0
        },

        Busy: {
            name: "cc busy",
            perm: 0
        },

        Login: {
            name: 'cc login',
            perm: 0
        },

        Logout: {
            name: "cc logout",
            perm: 0
        },

        Tier: {
            List: {
                name: 'cc tier from user',
                perm: 0
            },
            Create: {
                name: "cc tier create",
                perm: 1
            },

            SetLvl: {
                name: "cc tier set lvl",
                perm: 1
            },

            SetPos: {
                name: "cc tier set pos",
                perm: 1
            },

            Delete: {
                name: "cc tier del",
                perm: 1
            }
        }
    },

    CDR: {
        RecordCall: {
            name: "cdr_get_record_link",
            perm: 0
        }
    },

    Chat: {
        Send: {
            name: "chat_send",
            perm: 0
        }
    },

    Sys: {
        Message: {
            name: "sys_msg",
            perm: 0
        }
    },

    WhoAmI: {
        "name": "api whoami"
    },

    BlackList: {
        List: {
            name: "blacklist_list"
        },
        GetNames: {
            name: "blacklist_names"
        },
        GetFromName: {
            name: "blacklist_get_name"
        },
        RemoveNumber: {
            name: "blacklist_remove_number"
        },
        RemoveName: {
            name: "blacklist_remove_name"
        },
        Create: {
            name: "blacklist_create"
        },
        Search: {
            name: "blacklist_search"
        }
    },

    License: {
        List: {
            name: 'license_list'
        },
        Item: {
            name: 'license_item'
        },
        Upload: {
            name: 'license_upload'
        },
        Remove: {
            name: 'license_remove'
        }
    }
};

module.exports.RootName = 'root';

module.exports.ACCOUNT_EVENTS = {
    ONLINE: 'ACCOUNT_ONLINE',
    OFFLINE: 'ACCOUNT_OFFLINE'
};

var ACCOUNT_ROLE = module.exports.ACCOUNT_ROLE = {
    ROOT: {
        name: 'root',
        val: 2
    },
    ADMIN: {
        name: 'admin',
        val: 1
    },
    USER: {
        name: 'user',
        val: 0
    }
};

ACCOUNT_ROLE.getRoleFromName = function (name) {
    switch (('' + name).toLowerCase()) {
        case this.USER.name:
            return this.USER;
        case this.ADMIN.name:
            return this.ADMIN;
        case this.ROOT.name:
            return this.ROOT;
        default:
            // Custom
            return {
                "name": name,
                "val": -1
            };
    }
};

module.exports.ACCOUNT_SATUS_TYPE = {
    READY: "ONHOOK",
    BUSY: "ISBUSY",
    UNREGISTER: "NONREG"
};

module.exports.ACCOUNT_STATE_TYPE = {
    NONE: "NONE",
    DND: "DND"
};

module.exports.WEBITEL_EVENT_NAME_TYPE = {
    USER_STATE: "CC_AGENT_STATE",
    USER_STATUS: "CC_AGENT_STATUS"
};