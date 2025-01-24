package models

import (
	"sync"
	"time"
)

const (
	Monitoramentos = 1
	Delay          = 20
	LogFileName    = "log.txt"
)

// Variáveis globais usadas no projeto
var (
	Ticker                *time.Ticker
	IntervaloLimpeza      time.Duration = 24 * time.Hour
	EmailNotificacao      string
	Mutex                 sync.Mutex
	EncerrarMonitoramento chan bool = make(chan bool)
)
