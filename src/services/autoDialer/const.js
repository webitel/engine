/**
 * Created by igor on 24.05.16.
 */

'use strict';

const MEMBER_STATE = {
    Idle: 0,
    Process: 1,
    End: 2
};

const VAR_SEPARATOR = String.fromCharCode(27);

const DIALER_STATES = {
    Idle: 0,
    Work: 1,
    Sleep: 2,
    ProcessStop: 3,
    End: 4,
    Error: 5
};

const DIALER_CAUSE = {
    Init: "INIT",
    ProcessStop: "QUEUE_STOP",
    ProcessRecovery: "QUEUE_RECOVERY",
    ProcessSleep: "QUEUE_SLEEP",
    ProcessReady: "QUEUE_HUNTING",
    ProcessNotFoundMember: "NOT_FOUND_MEMBER",
    ProcessComplete: "QUEUE_COMPLETE",
    ProcessExpire: "QUEUE_EXPIRE",
    ProcessInternalError: "QUEUE_ERROR"
};

const DIFF_CHANGE_MSEC = 2000;

const AGENT_STATUS = {
    LoggedOut: 'Logged Out',
    Available: 'Available',
    AvailableOnDemand: 'Available (On Demand)',
    OnBreak: 'On Break'
};

const AGENT_STATE = {
    // Idle: 'Reserved',
    Idle: 'Idle',
    Reserved: 'Reserved',
    Waiting: 'Waiting',
    InQueueCall: 'In a queue call'
};

const END_CAUSE = {
    NO_ROUTE: "NO_ROUTE",
    NO_COMMUNICATIONS: "NO_COMMUNICATIONS",
    MAX_TRY: "MAX_TRY_COUNT",
    PROCESS_CRASH: "PROCESS_CRASH",
    ACCEPT: "ACCEPT",
    MEMBER_EXPIRED: "MEMBER_EXPIRED",
    ABANDONED: "ABANDONED",
    MANAGER_REQUEST: "MANAGER_REQUEST"
};

const CODE_RESPONSE_ERRORS = ["UNALLOCATED_NUMBER", END_CAUSE.NO_ROUTE, END_CAUSE.MEMBER_EXPIRED, "INVALID_NUMBER_FORMAT", "NETWORK_OUT_OF_ORDER", "OUTGOING_CALL_BARRED", "SERVICE_UNAVAILABLE", "CHAN_NOT_IMPLEMENTED", "SERVICE_NOT_IMPLEMENTED", "INCOMPATIBLE_DESTINATION", "MANDATORY_IE_MISSING", "PROGRESS_TIMEOUT", "GATEWAY_DOWN"];
const CODE_RESPONSE_RETRY = ["NO_ROUTE_DESTINATION", "DESTINATION_OUT_OF_ORDER", "USER_BUSY", "CALL_REJECTED", "NO_USER_RESPONSE", "NO_ANSWER", "SUBSCRIBER_ABSENT", "NUMBER_CHANGED", "NORMAL_UNSPECIFIED", "NORMAL_CIRCUIT_CONGESTION", "ORIGINATOR_CANCEL", "LOSE_RACE", "USER_NOT_REGISTERED"];
const CODE_RESPONSE_OK = ["NORMAL_CLEARING"];
const CODE_RESPONSE_MINUS_PROBE = ["RECOVERY_ON_TIMER_EXPIRE", "NORMAL_TEMPORARY_FAILURE"];

const MAX_MEMBER_RETRY = 999;

const MEMBER_VARIABLE_CALLER_NUMBER = "outbound_caller_id_number";

const DIALER_TYPES = {
    VoiceBroadcasting: "Voice Broadcasting",
    ProgressiveDialer: "Progressive Dialer",
    PredictiveDialer: "Predictive Dialer"
};

const AGENT_STRATEGY = {
    RANDOM: "random", //?? ?????????????????? ??????????????.
    WITH_FEWEST_CALLS: "with_fewest_calls", //?????????????? ???? ?????????????????? ?? ???????????????????? ?????????????????????? ??????????????.
    WITH_LEAST_TALK_TIME: "with_least_talk_time", //?????????????? ???? ?????????????????? ?? ???????????????????? ???????????????? ?? ??????????????????.
    LONGEST_IDLE_AGENT: "longest_idle_agent", //?????????????? ???? ?????????????????? ?? ???????????????????? ???????????????? ?? ????????????????.
    TOP_DOWN: "top-down", //???????????? ????????????-????????.
    WITH_LEAST_UTILIZATION: "with_least_utilization",
    WITH_HIGHEST_WAITING_TIME: "with_highest_waiting_time",
};

const NUMBER_STRATEGY = {
    TOP_DOWN: "top-down",
    BY_PRIORITY: "by-priority"
};

const VARIABLES = {
    COMMUNICATION_TYPE_NAME : "dlr_communication_type_name",
    COMMUNICATION_TYPE_CODE : "dlr_communication_type_code"
};

const ROUTE_RESOURCES_STRATEGY = {
    TOP_DOWN: "top_down",
    RANDOM: "random"
};

const HANGUP_CODES = {
    "UNSPECIFIED":	0,
    "UNALLOCATED_NUMBER": 1,
    "NO_ROUTE_TRANSIT_NET":	2,
    "NO_ROUTE_DESTINATION":	3,
    "CHANNEL_UNACCEPTABLE": 6,
    "CALL_AWARDED_DELIVERED": 7,
    "NORMAL_CLEARING": 16,
    "USER_BUSY": 17,
    "NO_USER_RESPONSE": 18,
    "NO_ANSWER": 19,
    "SUBSCRIBER_ABSENT": 20,
    "CALL_REJECTED": 21,
    "NUMBER_CHANGED": 22,
    "REDIRECTION_TO_NEW_DESTINATION": 23,
    "EXCHANGE_ROUTING_ERROR": 25,
    "DESTINATION_OUT_OF_ORDER": 27,
    "INVALID_NUMBER_FORMAT": 28,
    "FACILITY_REJECTED": 29,
    "RESPONSE_TO_STATUS_ENQUIRY": 30,
    "NORMAL_UNSPECIFIED": 31,
    "NORMAL_CIRCUIT_CONGESTION": 34,
    "NETWORK_OUT_OF_ORDER": 38,
    "NORMAL_TEMPORARY_FAILURE": 41,
    "SWITCH_CONGESTION": 42,
    "ACCESS_INFO_DISCARDED": 43,
    "REQUESTED_CHAN_UNAVAIL": 44,
    "PRE_EMPTED": 45,
    "FACILITY_NOT_SUBSCRIBED": 50,
    "OUTGOING_CALL_BARRED": 52,
    "INCOMING_CALL_BARRED": 54,
    "BEARERCAPABILITY_NOTAUTH": 57,
    "BEARERCAPABILITY_NOTAVAIL": 58,
    "SERVICE_UNAVAILABLE": 63,
    "BEARERCAPABILITY_NOTIMPL": 65,
    "CHAN_NOT_IMPLEMENTED": 66,
    "FACILITY_NOT_IMPLEMENTED": 69,
    "SERVICE_NOT_IMPLEMENTED": 79,
    "INVALID_CALL_REFERENCE": 81,
    "INCOMPATIBLE_DESTINATION": 88,
    "INVALID_MSG_UNSPECIFIED": 95,
    "MANDATORY_IE_MISSING": 96,
    "MESSAGE_TYPE_NONEXIST": 97,
    "WRONG_MESSAGE": 98,
    "IE_NONEXIST": 99,
    "INVALID_IE_CONTENTS": 100,
    "WRONG_CALL_STATE": 101,
    "RECOVERY_ON_TIMER_EXPIRE": 102,
    "MANDATORY_IE_LENGTH_ERROR": 103,
    "PROTOCOL_ERROR": 111,
    "INTERWORKING": 127,
    "ORIGINATOR_CANCEL": 487,
    "CRASH": 500,
    "SYSTEM_SHUTDOWN": 501,
    "LOSE_RACE": 502,
    "MANAGER_REQUEST": 503,
    "BLIND_TRANSFER": 600,
    "ATTENDED_TRANSFER": 601,
    "ALLOTTED_TIMEOUT": 602,
    "USER_CHALLENGE": 603,
    "MEDIA_TIMEOUT": 604,
    "PICKED_OFF": 605,
    "USER_NOT_REGISTERED": 606,
    "PROGRESS_TIMEOUT": 607,
    "GATEWAY_DOWN": 609,
    [END_CAUSE.ABANDONED] : 687
};

const mapCodes = new Map();
for (let key in HANGUP_CODES) {
    mapCodes.set(key, HANGUP_CODES[key])
}

function getHangupCode(codeName) {
    if (mapCodes.has(codeName))
        return mapCodes.get(codeName);
    return 0;
}

module.exports = {
    DIALER_STATES: DIALER_STATES,
    DIALER_CAUSE: DIALER_CAUSE,
    DIFF_CHANGE_MSEC: DIFF_CHANGE_MSEC,
    AGENT_STATUS: AGENT_STATUS,
    AGENT_STATE: AGENT_STATE,
    END_CAUSE: END_CAUSE,
    CODE_RESPONSE_ERRORS: CODE_RESPONSE_ERRORS,
    CODE_RESPONSE_RETRY: CODE_RESPONSE_RETRY,
    CODE_RESPONSE_OK: CODE_RESPONSE_OK,
    MAX_MEMBER_RETRY: MAX_MEMBER_RETRY,
    CODE_RESPONSE_MINUS_PROBE: CODE_RESPONSE_MINUS_PROBE,
    DIALER_TYPES: DIALER_TYPES,
    MEMBER_STATE: MEMBER_STATE,
    AGENT_STRATEGY: AGENT_STRATEGY,
    NUMBER_STRATEGY: NUMBER_STRATEGY,
    VAR_SEPARATOR,
    MEMBER_VARIABLE_CALLER_NUMBER,
    getHangupCode: getHangupCode,
    ROUTE_RESOURCES_STRATEGY,
    VARIABLES,
};