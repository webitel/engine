package model

import "time"

const APP_SERVICE_NAME = "engine"

const APP_SERVICE_TTL = time.Second * 30
const APP_DEREGESTER_CRITICAL_TTL = time.Second * 60
