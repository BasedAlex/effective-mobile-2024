package logger

import (
	log "github.com/sirupsen/logrus"
)

func New(lvl string) {
    // LOG_LEVEL not set, default to debug
    if lvl == "" {
        lvl = "debug"
    }
    ll, err := log.ParseLevel(lvl)
    if err != nil {
        ll = log.DebugLevel
    }
    log.SetLevel(ll)
}