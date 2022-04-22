package main

import (
	"M-socket/Configs"
	"dsfd-socket/Telega"
	"fmt"
	"github.com/jmoiron/sqlx"
	"log"
	"sync"
	"time"
)


type Application struct {
	LockerTerminals sync.RWMutex
	LockerPrices    sync.RWMutex
	LockerMath      sync.RWMutex

	PricesOrder []string

	mysql     *sqlx.DB
	Terminals map[string]*TerminalStatus
	Prices    map[string]map[string]*PriceTerminal
	Math      map[string]*Expression
}

func NewApplication() *Application {
	app := &Application{
		Terminals:   make(map[string]*TerminalStatus),
		Prices:      make(map[string]map[string]*PriceTerminal),
		Math:        make(map[string]*Expression),
		PricesOrder: []string{"JBl", "JHGJ", "JH", "KLJHK", "UIGj", "LKJN", "LKJ", "UYGJ", "KJHl"},
	}

	// Прикорячимся к БД
	app.mysqlInit()

	// Таймер проверки терминалов
	app.timerTerminalChecker(Configs.TimerTerminalChecker)

	// Таймер загрузки кастомных котировок пользователей
	app.timerLoadExpressions(Configs.TimerLoadExpressions)

	// Грузим выражения (кастомные котировки пользователей)
	go app.loadExpressions()

	log.Println("Started!")

	return app
}

// GetMySQL Возвращает коннект к БД
func (a *Application) GetMySQL() *sqlx.DB {
	return a.mysql
}

// SetTerminalAlive Выставляет терминал в живое состояние :)
func (a *Application) SetTerminalAlive(name string) {
	a.LockerTerminals.Lock()

	if terminal, ok := a.Terminals[name]; !ok {
		newTerminal := &TerminalStatus{
			Name:     name,
			Datetime: time.Now(),
		}

		a.Terminals[name] = newTerminal
		log.Printf("Терминал %s подключился.", name)
		Telega.Send(Configs.MyTelega, fmt.Sprintf("Терминал %s подключился.", name))

	} else {
		terminal.Datetime = time.Now()
	}

	a.LockerTerminals.Unlock()
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// Конектится к БД или шваркнется в панику
func (a *Application) mysqlInit() {
	str := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
		Configs.MySQLUser,
		Configs.MySQLPassword,
		Configs.MySQLAddress,
		Configs.MySQLPort,
		Configs.MySQLBase,
	)

	db, err := sqlx.Connect("mysql", str)

	if err != nil {
		log.Fatalln("Нет коннекта к БД")
	}

	db.SetMaxOpenConns(25)
	db.SetConnMaxLifetime(time.Second * 30)

	a.mysql = db

	log.Println("MySQL Connected")
}

// Таймер проверки терминалов
// Проверяем, шлет ли терминал котировки, если нет, сигналим в телеграм
func (a *Application) timerTerminalChecker(d time.Duration) {
	time.AfterFunc(d, func() {
		now := time.Now()

		a.LockerTerminals.Lock()

		for _, term := range a.Terminals {
			if now.Sub(term.Datetime) > time.Minute*3 {
				log.Printf("Терминал %s не работает", term.Name)
				Telega.Send(Configs.MyTelega, fmt.Sprintf("Терминал %s не работает.", term.Name))
			}
		}

		a.LockerTerminals.Unlock()

		// Самовызов
		a.timerTerminalChecker(d)
	})
}

// Таймер загрузки выражений
func (a *Application) timerLoadExpressions(d time.Duration) {
	time.AfterFunc(d, func() {
		a.loadExpressions()

		// Самовызов
		a.timerLoadExpressions(d)
	})
}

// Загрузка пользовательских выражений
func (a *Application) loadExpressions() {
	res, err := App.GetMySQL().Queryx("SELECT seccodeOut, seccodeA, seccodeB, operation FROM `user_quotes`")

	if err != nil {
		log.Println(err.Error())
		return
	}

	App.LockerMath.Lock()

	App.Math = make(map[string]*Expression)

	for res.Next() {
		exp := &Expression{}

		if err := res.StructScan(exp); err != nil {
			log.Println("ROW INIT ERROR:", err.Error())
		} else {
			App.Math[exp.Out] = exp
		}
	}

	App.LockerMath.Unlock()
}

// GetPrice Получение цены
func (a *Application) GetPrice(key string) *PriceTerminal {
	a.LockerPrices.RLock()

	for _, k := range a.PricesOrder {
		if price, ok := a.Prices[k][key]; ok {
			a.LockerPrices.RUnlock()
			return price
		}
	}

	a.LockerPrices.RUnlock()

	return nil
}
