package main

import (
	"df-socket/Configs"
	"encoding/json"
	"fdf-socket/Telega"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/appleboy/go-fcm"
	"github.com/blacked/go-zabbix"
)

func NewContainer() *Container {
	c := &Container{
		PricesOrder: []string{"LK;", "LKJHl", "LKJHl", "LJKH", "LUHKj", "OIUHGLK", "LKJBkm", "KJHk", "LJHB"},
		Levels:      make(map[int64]*Level),
		Prices:      make(map[string]map[string]*Price),
		PushTTL:     make(map[int64]int64),
	}

	c.Manager = NewManager(c)

	return c
}

type Container struct {
	PricesOrder []string

	Manager *Manager
	Levels  map[int64]*Level
	Prices  map[string]map[string]*Price

	WaitGroup sync.WaitGroup

	PushTTL map[int64]int64

	LockerPush   sync.RWMutex // Отдельный лок для PushTTL
	LockerPrice  sync.RWMutex // Отдельный лок для Price
	LockerLevels sync.RWMutex // Отдельный лок для Levels

	Messages int64
}

// LoadPrices deprecate
func (c *Container) LoadPrices() {
	//return

	timeDBFetchStart := time.Now()

	resp, err := http.Get(Configs.SocketAddress)

	// handle the error if there is one
	if err != nil {
		log.Println(err)
		return
	}

	defer resp.Body.Close()

	html, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Println(err)
		return
	}

	var prices = make(map[string]map[string]*Price)

	result := make(map[string]interface{})

	err = json.Unmarshal(html, &result)

	if err != nil {
		log.Println(err)
		return
	}

	for k, rows := range result {
		if k == "undefined" {
			continue
		}

		if _, ok := prices[k]; !ok {
			prices[k] = make(map[string]*Price)
		}

		for _, row := range rows.(map[string]interface{}) {
			if row.(map[string]interface{})["seccode"] == nil {
				log.Println("Какое то говно:", row)
				continue
			}

			priceStep := 0.0
			stepPrice := 0.0

			if row.(map[string]interface{})["price_step"] != nil {
				priceStep = row.(map[string]interface{})["price_step"].(float64)
			}

			if row.(map[string]interface{})["step_price"] != nil {
				switch row.(map[string]interface{})["step_price"].(type) {
				case string:
					stepPrice, err = strconv.ParseFloat(row.(map[string]interface{})["step_price"].(string), 64)

					if err != nil {
						stepPrice = 0
					}
				case float64:
					stepPrice = row.(map[string]interface{})["step_price"].(float64)
				}
			}

			price := &Price{
				SecCode:   row.(map[string]interface{})["seccode"].(string),
				DateTime:  int64(row.(map[string]interface{})["datetime"].(float64)),
				Close:     row.(map[string]interface{})["close"].(float64),
				Percent:   row.(map[string]interface{})["percent"].(float64),
				Go:        row.(map[string]interface{})["go"].(float64),
				Open:      row.(map[string]interface{})["open"].(float64),
				High:      row.(map[string]interface{})["high"].(float64),
				Low:       row.(map[string]interface{})["low"].(float64),
				Vol:       row.(map[string]interface{})["vol"].(float64),
				H5:        row.(map[string]interface{})["h5"].(float64),
				L5:        row.(map[string]interface{})["l5"].(float64),
				PriceStep: priceStep,
				StepPrice: stepPrice,
				Provider:  row.(map[string]interface{})["provider_"].(string),
			}

			prices[k][price.SecCode] = price
		}
	}

	timeDBFetchFinish := time.Now()
	log.Printf("Распотрашили цены за %v мсек.\n", timeDBFetchFinish.Sub(timeDBFetchStart).Milliseconds())

	c.LockerPrice.Lock()
	c.Prices = prices
	c.LockerPrice.Unlock()
}

func (c *Container) LoadLevels() {
	//lastLogin := GetDayStart()

	//timeDBFetchStart := time.Now()

	res, err := App.GetMySQL().Queryx("SELECT DISTINCT " +
		"price_lvl.user_id, " +
		"price_lvl.price, " +
		"price_lvl.seccode_trans, " +
		"price_lvl.up_dwn, " +
		"price_lvl.idx, " +
		"price_lvl.sent, " +
		"price_lvl.sms, " +
		"price_lvl.datetime, " +
		"price_lvl.comment, " +
		"price_lvl.age, " +
		"price_lvl.delta, " +
		"price_lvl.pause, " +
		"price_lvl.`disable`, " +
		"price_lvl.`app`, " +
		"users.email, " +
		"users.icq, " +
		"users.phone, " +
		"users.forex, " +
		"users.locale, " +
		"code_trans.Full_name, " +
		"code_trans.Full_name_en, " +
		"code_trans.id, " +
		"users.cur_tarif," +
		"users.google, " +
		"users.signup_date, " +
		"users.Premium, " +
		"users.bonus," +
		"users.telegram_chatId," +
		"users.doubled_telegram " +
		"FROM price_lvl " +
		"INNER JOIN users ON users.id = price_lvl.user_id " +
		"INNER JOIN code_trans ON price_lvl.seccode_trans = code_trans.seccode_trans")

	if err != nil {
		log.Println(err.Error())
		return
	}

	//timeDBForStart := time.Now()

	levels := make(map[int64]*Level)
	counts := 0
	//log.Println("Загрузка данных ...")
	for res.Next() {
		level := &Level{}

		counts++

		if err := res.StructScan(level); err != nil {
			log.Println("ROW INIT ERROR:", err.Error())
		} else {
			levels[level.IDX] = level
		}
		//log.Printf("IDX:%v , seccodetrans: %v, app %v\n", level.IDX, level.SecCodeTrans, level.App)
		//log.Printf("%v %v %v %v %v %v %v %v %v", level.IDX, level.UserID, level.SMS, level.Disable, level.Delta, level.CodeID, level.SecCodeTrans, level.Price, level.Comment)
	}

	//timeDBFetchFinish := time.Now()

	//log.Printf("БД: %v мсек. FOR: %v мсек, Кол-во: %v\n",
	//	timeDBFetchFinish.Sub(timeDBFetchStart).Milliseconds(),
	//	timeDBFetchFinish.Sub(timeDBForStart).Milliseconds(),
	//	counts,
	//)

	c.LockerLevels.Lock()
	c.Levels = levels
	c.LockerLevels.Unlock()
}

// Грузит цены и проверяет уровни, уровни не грузит
func (c *Container) Check() {

	//timeStart := time.Now()

	//c.LoadPrices()  // загрузка котировок из сокета Get запросом (старая версия)

	levels := make(map[int64]*Level)

	c.LockerLevels.Lock()
	for k, v := range c.Levels {
		levels[k] = v
	}
	c.LockerLevels.Unlock()

	c.WaitGroup.Add(len(levels))

	for _, level := range levels {
		c.Manager.Compute(level)
	}

	c.WaitGroup.Wait()

	//timeFinish := time.Now()

	//log.Printf("Логика отработала за %v мсек.", timeFinish.Sub(timeStart).Milliseconds())

	//num := rand.Int63n(1000)
	//str := time.Now().String()
	//
	//res, err := MySQL.Exec("UPDATE `test` SET num = ?, str = ? WHERE id = ?", num, str, 1)
	//
	//if err != nil {
	//	log.Println("ERROR:", err)
	//} else {
	//	rows, _ := res.RowsAffected()
	//	log.Println("Rows affected", rows)
	//}
}

func (c *Container) InitWorkers() {
	for i := 0; i < 16; i++ {
		c.Manager.AddWorker()
	}
}

func (c *Container) LoadLevelsTimer(d time.Duration) {
	time.AfterFunc(d, func() {
		c.LoadLevels()
		c.LoadLevelsTimer(d)
	})
}

func (c *Container) CheckTimer(d time.Duration) {
	time.AfterFunc(d, func() {
		c.Check()
		c.CheckTimer(d)
	})
}

// Получаем цену
func (c *Container) GetPrice(key string) *Price {
	c.LockerPrice.RLock()

	for _, k := range c.PricesOrder {
		if price, ok := c.Prices[k][key]; ok {
			c.LockerPrice.RUnlock()
			return price
		}
	}

	c.LockerPrice.RUnlock()

	return nil
}

// Бесконечный шлет кол-во сообщений в 30 сек
func (c *Container) Infinity() {
	ticker := time.NewTicker(30 * time.Second)

	for {
		select {
		case <-ticker.C:
			metrics := make([]*zabbix.Metric, 0, 1)

			msgCount := atomic.SwapInt64(&c.Messages, 0)

			metrics = append(metrics, zabbix.NewMetric("LKJkjhlk.23", "KJHLKj", strconv.FormatInt(msgCount, 10)))

			packet := zabbix.NewPacket(metrics)

			z := zabbix.NewSender(Configs.ZabbixAddress, Configs.ZabbixPort)
			_, err := z.Send(packet)

			if err != nil {
				log.Println("Zabbix:", err)
			}

			// Удаление заблокированных пушей
			c.LockerPush.Lock()

			Now := time.Now().Unix()

			for k, v := range c.PushTTL {
				if v < Now {
					delete(c.PushTTL, k)
					//log.Printf("очистил игнор IDX %v", k)
				}
			}

			c.LockerPush.Unlock()
		}
	}
}

var below = map[string]string{"ru": " ниже ", "en": " belov ", "de": " unten ", "it": " sotto ", "es": " abajo ", "fr": " ci-dessous "}
var above = map[string]string{"ru": " выше ", "en": " above ", "de": " oben ", "it": " superiore ", "es": " arriba ", "fr": " au-dessus "}
var pager_arr = map[int64]string{0: "icq", 1: "phone", 2: "email", 3: "android", 4: "telegram"}
var address_arr = make(map[int64]string)

func (c *Container) Calculate(level *Level, price *PriceTerminal, workerID int64) { // новая версия, когда получаем котировки напрямую от терминалов в сокет
	//func (c *Container) Calculate(level *Level, price *Price, workerID int64) {  // старая версия, когда брали котировки из сокета по запросу
	Now := time.Now().Unix()

	c.LockerPush.Lock()
	allowSend := true

	if t, ok := c.PushTTL[level.IDX]; ok && t > Now {
		allowSend = false
	}
	c.LockerPush.Unlock()

	if allowSend == false {
		//log.Printf("Worker %v: Проигнорили [%v] потому что уже отправляли", workerID, level.IDX)
		return
	}
	//log.Printf("Запустили расчет")
	//if level.UserID != 25 {
	//	return
	//} // заглушка - обсчитывать только пользователя с ID = 25

	level.Lock()
	defer level.Unlock()

	//log.Printf("Worker %v: SC:%v, Close:%v, H5:%v, L5:%v", workerID, price.SecCode, price.Close, price.H5, price.L5)

	if price.Close <= 0 {
		log.Println("цена 0")
		return
	}

	// Игнорируемый
	if level.Ignore {
		log.Println("игнорируем")
		return
	}

	// TODO Выключено?
	if level.Disable != 0 {
		//log.Println("Lvl disabled")
		return
	}

	dateDelta := Now - price.DateTime
	direction := "" //int64(0)
	comment := ""
	up_dwn := int64(0)

	if price.DateTime < 1 || dateDelta > 499 || dateDelta < -400 {
		//log.Printf("%v кривой датетайм DateTime: %v dateDelta: %v",level.SecCodeTrans,price.DateTime,dateDelta)
		return
	}
	// надо внимательно разобрать это условие
	// если уровень находится в паузе, и вышло время выходить из паузы, то выводим и ставим активным

	if price.Close < level.Price {
		up_dwn = 1
	} else {
		up_dwn = -1
	}

	if level.Pause == 1 && level.DateTime < Now {
		dateTime := Now
		_, err := App.GetMySQL().Exec(
			"UPDATE `price_lvl` SET `up_dwn` = ?, `sent` = 0,  `pause` = 0 WHERE `idx` = ?", up_dwn, level.IDX,
		)
		if err != nil {
			log.Printf("Error: Worker_update %v: %v", workerID, err.Error())
		} else {
			level.Pause = 0
			level.UpDown = up_dwn
			level.DateTime = dateTime
			//log.Printf("снял паузу в базе IDX [%v]", level.IDX)
		}
	}
	if level.Pause == 1 { // если уровень всетаки остался в паузе, дальше не проверяем
		return
	}
	// если уровень не на паузе
	if level.UpDown == 1 {
		if price.H5 > 0 {
			if price.H5 >= level.Price && (level.DateTime+60) < Now {
				comment = fmt.Sprintf("%v Cross Up Price  %v Cur: %v High: %v %v", level.SecCodeTrans, level.Price, price.Close, price.H5, level.Comment)
			}
		} else {
			if price.Close >= level.Price {
				comment = fmt.Sprintf("%v Cross Up Price  %v Cur: %v %v", level.SecCodeTrans, level.Price, price.Close, level.Comment)
			}
		}
		direction = above[fmt.Sprintf("%v", level.Locale)]
		if len(direction) == 0 {
			direction = above["en"]
		}
	}
	if level.UpDown == -1 {
		if price.L5 > 0 {
			if price.L5 <= level.Price && (level.DateTime+60) < Now {
				comment = fmt.Sprintf("%v Cross Down Price  %v Cur: %v Low: %v %v", level.SecCodeTrans, level.Price, price.Close, price.L5, level.Comment)
			}
		} else {
			if price.Close <= level.Price {
				comment = fmt.Sprintf("%v Cross Down Price  %v Cur: %v %v", level.SecCodeTrans, level.Price, price.Close, level.Comment)
			}
		}
		direction = below[level.Locale]
		if len(direction) == 0 {
			direction = below["en"]
		}
	}
	// TODO ХЗ что и как оно посчиталось )
	//dateTime := int64(100500)
	//age := time.Now().Add(time.Hour * 24 * 30).Unix()

	if len(comment) > 0 {

		//if level.UserID == 25 {comment = fmt.Sprintf("%v [GO]", comment)}

		address_arr = map[int64]string{0: "icq", 1: level.Phone, 2: level.Email, 3: "android", 4: strconv.FormatInt(level.TelegramChatID, 10)}

		//log.Println(comment)
		voice := ""

		if level.Locale == "ru" {
			voice = fmt.Sprintf("%v %v %v %v", level.FullNameRu, direction, level.Price, level.Comment) //strtoupper($s['Full_name']) . $direction . $s['price'] . ' ' . $s['comment']);
		} else {
			voice = fmt.Sprintf("%v %v %v %v", level.FullNameEn, direction, level.Price, level.Comment) //strtoupper($s['Full_name']) . $direction . $s['price'] . ' ' . $s['comment']);
		}

		t_ := time.Now()
		date_ := fmt.Sprintf("%d%02d%02d", t_.Year(), t_.Month(), t_.Day())
		time_ := fmt.Sprintf("%02d%02d%02d", t_.Hour(), t_.Minute(), t_.Second())
		pager := pager_arr[level.SMS]
		address := address_arr[level.SMS]

		sent := "1"
		if pager == "phone" || pager == "email" {
			sent = "0"
		}

		//mysql::Query('INSERT IGNORE INTO `send_message` (`user` ,`date` ,`time`, `pager`,  `message`, `sent`,`datetime`,`address`,`site`,`num_id`, `hash`,`seccode_id`,`voice`)VALUES'. $values_);
		_, err := App.GetMySQL().Exec("INSERT IGNORE INTO `send_message` (`user` ,`date` ,`time`, `pager`,  `message`, `sent`,`datetime`,`address`,`site`,`num_id`, `hash`,`seccode_id`,`voice`,`app`) "+
			"VALUES (?, ?, ?, ?,?, ?, ?, ?,?, ?, ?, ?,?,?)",
			level.UserID, date_, time_, pager, comment, sent, Now, address, "1", "1", "1", level.CodeID, voice, level.App)
		//log.Printf("INSERT IGNORE INTO `send_message` (`user` ,`date` ,`time`, `pager`,  `message`, `sent`,`datetime`,`address`,`site`,`num_id`, `hash`,`seccode_id`,`voice`,`app`) "+
		//	"VALUES (?, ?, ?, ?,?, ?, ?, ?,?, ?, ?, ?,?,?)",
		//	level.UserID, date_, time_, pager, comment, sent, Now, address, "1", "1", "1", level.CodeID, voice, level.App)

		if err != nil {
			log.Printf("Error: Worker0 %v: %v", workerID, err.Error())
		} else {

			if level.Delta == 0 {
				_, err := App.GetMySQL().Exec(
					"Delete from `price_lvl` WHERE `idx` = ?", level.IDX,
				)
				if err != nil {
					log.Printf("Error SQL: Del lvl WorkerID %v: %v", workerID, err.Error())
				} else {
					//log.Println("удалил из базы отработанный уровень")
					/*
						newDatetime := Now + 3600 // ставим паузу на час, что бы он больше не считался до следующего обновления из базы
						level.DateTime = newDatetime
						level.Pause = 1
						// Если нет паузы залочим глобально, на час, на всякий случай

						c.LockerPush.Lock()
						c.PushTTL[level.IDX] = newDatetime
						c.LockerPush.Unlock()
					*/

					c.LockerLevels.Lock()
					delete(c.Levels, level.IDX)
					c.LockerLevels.Unlock()
				}
			} else {
				newDatetime := Now + level.Delta*60
				_, err := App.GetMySQL().Exec(
					"UPDATE `price_lvl` SET `up_dwn` = ?, `sent` = ?, `datetime` = ?,  `pause` = 1 WHERE `idx` = ?",
					up_dwn, 0, newDatetime, level.IDX,
				)
				if err != nil {
					log.Printf("Error SQL: Update lvl WorkerID %v: %v", workerID, err.Error())
				} else {
					level.UpDown = up_dwn
					level.Pause = 1
					level.DateTime = newDatetime
					// Если есть пауза лочим уровень глобально, на время паузы
					c.LockerPush.Lock()
					c.PushTTL[level.IDX] = newDatetime
					c.LockerPush.Unlock()

					//log.Printf("ставим на паузу на %v минут, до %v", level.Delta, newDatetime)
				}
			}

			//	if level.Pause > 0 {
			//		// Если есть пауза лочим на время паузы
			//		c.PushTTL[level.IDX] = Now + level.Pause*60
			//	} else {
			//		// Если нет паузы залочим на час на всякий случай
			//		c.PushTTL[level.IDX] = Now + 3600
			//	}

			if level.SMS == 3 { // если pager = android  то шлем на android
				c.SendPush(level.UserID, "message", level.App)
				if level.DoubledTelegram == 1 {
					Telega.Send(level.TelegramChatID, comment)
				}

			}
			if level.SMS == 4 { // если pager = telegram  то шлем на telegram
				Telega.Send(level.TelegramChatID, comment)
				//c.SendTelegram()
			}
			if level.SMS == 2 { // если pager = email  то шлем на email
				//log.Println("Отправляем на email")
				// отправлялка на email
			}
			if level.SMS == 1 { // если pager = phone  то шлем SMS
				//log.Println("Отправляем СМС")
				// отправлялка SMS
			}
		}

	}

	// Что то тут обновляем )
	//if false {
	//	_, err := MySQL.Exec(
	//		"UPDATE `price_lvl` SET `up_dwn` = ?, `sent` = ?, `datetime` = ?, `age` = ?, `pause` = 0 WHERE `idx` = ?",
	//		direction, 0, dateTime, age, 0, level.IDX,
	//	)
	//
	//	if err != nil {
	//		log.Printf("Error: Worker %v: %v", workerID, err.Error())
	//	} else {
	//		level.UpDown = direction
	//		level.Sent = 0
	//		level.DateTime = dateTime
	//		//level.Age = age
	//		level.Pause = 0
	//	}
	//
	//	// TODO Тут выход или чего ? )
	//}
	//}

	///////////////////////////////////////
	// Хуячить свои дела тут :)

	//if level.Price < price.Close {
	//	c.SentPush(25, fmt.Sprintf("Worker %v: Цена из БД меньше цены из сокета", workerID))
	//}
	//
	//if level.Price > price.Close {
	//	c.SentPush(25, fmt.Sprintf("Worker %v: Цена из БД больше цены из сокета", workerID))
	//}
}

func (c *Container) SendPush(userID int64, msg string, app string) {
	//log.Println("Запускаем сендер Push")
	var err error

	t := &struct {
		Token string `db:"token"`
	}{}

	rows, err := App.GetMySQL().Queryx("SELECT `token` FROM `fcm_token` WHERE `deactivate` = 0 and `user_id` = ? and `app` = ?", userID, app)

	if err != nil {
		log.Println("Error: MySQL_1:", err)
		return
	}

	for rows.Next() {
		if err := rows.StructScan(t); err == nil {
			ttl := uint(60 * 60 * 24)
			msg := &fcm.Message{
				To: t.Token,
				Data: map[string]interface{}{
					"user": userID,
					"data": msg,
				},
				Priority:   "high",
				TimeToLive: &ttl,
			}

			var FCMKey = Configs.FCMKey_0
			if app == "1" {
				FCMKey = Configs.FCMKey_1
			}

			client, err := fcm.NewClient(FCMKey)

			if err != nil {
				log.Println("Error: FCM:", err)
				continue
			}

			res, err := client.Send(msg)

			if err != nil {
				log.Println("Error: FCM:", err)
				continue
			}

			if res.Error != nil {
				_, err = App.GetMySQL().Exec("UPDATE `fcm_token` SET deactivate = 1, error = ? WHERE token = ?", res.Error.Error(), t.Token)
				if err != nil {
					log.Println(" error update fcm table ", err)
				}
			}
			//log.Printf("FMC: Success: %v", res.Success)
		} else {
			log.Println("Error: Send Push:", err)
		}
	}

	atomic.AddInt64(&c.Messages, 1)
}
