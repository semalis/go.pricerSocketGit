package main

import (
	"sync"
	"time"
)

type Price struct {
	SecCode   string
	DateTime  int64
	Close     float64
	Percent   float64
	Go        float64
	Open      float64
	High      float64
	Low       float64
	Vol       float64
	H5        float64
	L5        float64
	PriceStep float64
	StepPrice float64
	Provider  string
}

type PriceTerminal struct {
	SecCode   string  `json:"seccode"`
	DateTime  int64   `json:"datetime"`
	Close     float64 `json:"close"`
	Percent   float64 `json:"percent"`
	StepPrice float64 `json:"step_price"`
	Provider  string  `json:"provider_"`
}

type Level struct {
	sync.Mutex

	UserID          int64   `db:"user_id"`
	Price           float64 `db:"price"`
	TelegramChatID  int64   `db:"telegram_chatId"`
	DoubledTelegram int64   `db:"doubled_telegram"`
	App             string  `db:"app"`
	Ignore          bool
}

type TerminalStatus struct {
	Name     string
	Datetime time.Time
}

type Expression struct {
	ASeccode string `db:"seccodeA"`

	Operation string `db:"operation"`
}
