package main

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"
	"weewe-socket/Configs"
	"wewewe-socket/TCP"
)

type Worker struct {
	container *Container
	id        int64
	level     chan *Level
	stop      chan struct{}
}


func (w *Worker) Run() {
	log.Printf("Worker %v start.\n", w.id)

	for {
		select {
		case level := <-w.level:
			//price := w.container.GetPrice(level.SecCodeTrans)  // берем котировку полученную из сокета (старая версия)
			price := App.GetPrice(level.SecCodeTrans) // берем котировку полученную из своего массива от терминалов напрямую

			if price == nil {
				w.container.WaitGroup.Done()
				//log.Printf("\t Worker %v: Запись по %v не нашлась", w.id, level.SecCodeTrans)
				continue
			}

			// Запускаем обработку
			w.container.Calculate(level, price, w.id)

			// Сообщаем что одного посчитали
			w.container.WaitGroup.Done()

		case <-w.stop:
			fmt.Printf("Worker %v stop.\n", w.id)
			break
		}
	}
}

var App *Application

func main() {
	App = NewApplication()

	container := NewContainer()

	//запускаем SocketServer получающий котировки от терминлов
	server := TCP.NewServer(Configs.TerminalSocketServer)

	server.OnClose(onClose)
	server.OnConnect(onConnect)
	server.OnMessage(onMessage)

	r := mux.NewRouter()

	r.HandleFunc("/", showAll)
	r.HandleFunc("/{time}/{code}", showCode)
	r.HandleFunc("/{time}/{code}/{provider}", showProvider)

	//запускаем сервер для отдачи котировок по запросам http в формате json
	go http.ListenAndServe(Configs.SocketServer, r)

	/* временно отключил*/
	// Врубаем HTTP сервер, который принимает измененя от пользователей по уровням
	go HTTPServer(container)

	// Врубим воркеры
	container.InitWorkers()

	// Подргузи уровни
	container.LoadLevels()

	// Запустим проверку
	container.Check()

	// Таймер на подгрузку увроней
	//container.LoadLevelsTimer(30 * time.Second)
	container.LoadLevelsTimer(300 * time.Second) // каждые 5 минут синхронизируемся с базой

	// Таймер на проверки - подгружает цены и запускает расчеты для проверки уровней
	container.CheckTimer(5 * time.Second)

	/**/

	// Запускаем бесконечный таймер чтоб приложение жило вечно
	container.Infinity()
}

func HTTPServer(container *Container) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		//log.Println("HTTP: Income DO:", r.FormValue("DO"))

		switch r.FormValue("DO") {
		case "Add":
			level := &Level{}

			if v, e := strconv.ParseInt(r.FormValue("UserID"), 10, 64); e != nil {
				log.Println("HTTP: Error parse: UserID")
				fmt.Fprintf(w, "Error: UserID") // Раскопируй меня :)
				return
			} else {
				level.UserID = v
			}

			if v, e := strconv.ParseFloat(r.FormValue("Price"), 64); e != nil {
				log.Println("HTTP: Error parse: Price")
				fmt.Fprintf(w, "Error: Price")
				return
			} else {
				level.Price = v
			}

			level.SecCodeTrans = r.FormValue("SecCodeTrans")

			if v, e := strconv.ParseInt(r.FormValue("UpDown"), 10, 64); e != nil {
				log.Println("HTTP: Error parse: UpDown")
				fmt.Fprintf(w, "Error: UpDown")
				return
			} else {
				level.UpDown = v
			}

			if v, e := strconv.ParseInt(r.FormValue("IDX"), 10, 64); e != nil {
				log.Println("HTTP: Error parse: IDX")
				fmt.Fprintf(w, "Error: IDX")
				return
			} else {
				level.IDX = v
			}

			if v, e := strconv.ParseInt(r.FormValue("Sent"), 10, 64); e != nil {
				log.Println("HTTP: Error parse: Sent")
				fmt.Fprintf(w, "Error: Sent")
				return
			} else {
				level.Sent = v
			}

			if v, e := strconv.ParseInt(r.FormValue("SMS"), 10, 64); e != nil {
				log.Println("HTTP: Error parse: SMS")
				fmt.Fprintf(w, "Error: SMS")
				return
			} else {
				level.SMS = v
			}

			if v, e := strconv.ParseInt(r.FormValue("DateTime"), 10, 64); e != nil {
				log.Println("HTTP: Error parse: DateTime")
				fmt.Fprintf(w, "Error: DateTime")
				return
			} else {
				level.DateTime = v
			}

			if v, err := url.QueryUnescape(r.FormValue("Comment")); err != nil {
				log.Println("HTTP: Error parse: Comment")
				fmt.Fprintf(w, "Error: Comment")
				return
			} else {
				level.Comment = v
			}

			if v, e := strconv.ParseInt(r.FormValue("Age"), 10, 64); e != nil {
				log.Println("HTTP: Error parse: Age")
				fmt.Fprintf(w, "Error: Age")
				return
			} else {
				level.Age = v
			}

			if v, e := strconv.ParseInt(r.FormValue("Delta"), 10, 64); e != nil {
				log.Println("HTTP: Error parse: Delta")
				fmt.Fprintf(w, "Error: Delta")
				return
			} else {
				level.Delta = v
			}

			if v, e := strconv.ParseInt(r.FormValue("Pause"), 10, 64); e != nil {
				log.Println("HTTP: Error parse: Pause")
				fmt.Fprintf(w, "Error: Pause")
				return
			} else {
				level.Pause = v
			}

			if v, e := strconv.ParseInt(r.FormValue("Disable"), 10, 64); e != nil {
				log.Println("HTTP: Error parse: Disable")
				fmt.Fprintf(w, "Error: Disable")
				return
			} else {
				level.Disable = v
			}

			if v, err := url.QueryUnescape(r.FormValue("Email")); err != nil {
				log.Println("HTTP: Error parse: Email")
				fmt.Fprintf(w, "Error: Email")
				return
			} else {
				level.Email = v
			}

			if v, e := strconv.ParseInt(r.FormValue("ICQ"), 10, 64); e != nil {
				log.Println("HTTP: Error parse: ICQ")
				fmt.Fprintf(w, "Error: ICQ")
				return
			} else {
				level.ICQ = v
			}

			if v, err := url.QueryUnescape(r.FormValue("Phone")); err != nil {
				log.Println("HTTP: Error parse: Phone")
				fmt.Fprintf(w, "Error: Phone")
				return
			} else {
				level.Phone = v
			}

			level.Forex = r.FormValue("Forex")
			level.Locale = r.FormValue("Locale")

			if v, err := url.QueryUnescape(r.FormValue("FullNameRu")); err != nil {
				log.Println("HTTP: Error parse: FullNameRu")
				fmt.Fprintf(w, "Error: FullNameRu")
				return
			} else {
				level.FullNameRu = v
			}

			if v, err := url.QueryUnescape(r.FormValue("FullNameEn")); err != nil {
				log.Println("HTTP: Error parse: FullNameEn")
				fmt.Fprintf(w, "Error: FullNameEn")
				return
			} else {
				level.FullNameEn = v
			}

			if v, e := strconv.ParseInt(r.FormValue("CodeID"), 10, 64); e != nil {
				log.Println("HTTP: Error parse: CodeID")
				fmt.Fprintf(w, "Error: ICQ")
				return
			} else {
				level.CodeID = v
			}

			level.CurTarif = r.FormValue("CurTarif")
			level.Google = r.FormValue("Google")
			//level.Google = r.FormValue("Google")

			if v, e := strconv.ParseInt(r.FormValue("SignUpDate"), 10, 64); e != nil {
				log.Println("HTTP: Error parse: SignUpDate")
				fmt.Fprintf(w, "Error: SignUpDate")
				return
			} else {
				level.SignUpDate = v
			}

			if v, e := strconv.ParseInt(r.FormValue("Premium"), 10, 64); e != nil {
				log.Println("HTTP: Error parse: Premium")
				fmt.Fprintf(w, "Error: Premium")
				return
			} else {
				level.Premium = v
			}

			if v, e := strconv.ParseInt(r.FormValue("Bonus"), 10, 64); e != nil {
				log.Println("HTTP: Error parse: Bonus")
				fmt.Fprintf(w, "Error: Bonus")
				return
			} else {
				level.Bonus = v
			}

			if v, e := strconv.ParseInt(r.FormValue("TelegramChatID"), 10, 64); e != nil {
				log.Println("HTTP: Error parse: TelegramChatID")
				fmt.Fprintf(w, "Error: TelegramChatID")
				return
			} else {
				level.TelegramChatID = v
			}

			if v, e := strconv.ParseInt(r.FormValue("DoubledTelegram"), 10, 64); e != nil {
				log.Println("HTTP: Error parse: DoubledTelegram")
				fmt.Fprintf(w, "Error: DoubledTelegram")
				return
			} else {
				level.DoubledTelegram = v
			}

			container.LockerLevels.Lock()
			// Добавляем в список
			container.Levels[level.IDX] = level
			//PrintLevels(container, level.IDX)
			container.LockerLevels.Unlock()

			//log.Println("HTTP: Added IDX:", level.IDX, "Count:", len(container.Levels))

			fmt.Fprintf(w, "OK")

		case "UpdateDisable":
			var idx int64
			var disable int64
			var e error

			if disable, e = strconv.ParseInt(r.FormValue("Disable"), 10, 64); e != nil {
				log.Println("HTTP: Error parse: Disable")
				fmt.Fprintf(w, "Error: Disable")
				return
			}

			if idx, e = strconv.ParseInt(r.FormValue("IDX"), 10, 64); e != nil {
				log.Println("HTTP: Error parse: IDX")
				fmt.Fprintf(w, "Error: IDX")
				return
			}

			container.LockerLevels.Lock()
			// Добавляем в список
			if level, ok := container.Levels[idx]; ok {
				level.Disable = disable
			}
			//PrintLevelsUser(container, 25)
			container.LockerLevels.Unlock()

			//log.Println("HTTP: Updated IDX:", idx, "Count:", len(container.Levels))
			fmt.Fprintf(w, "OK")

		case "Del":

			idx, e := strconv.ParseInt(r.FormValue("IDX"), 10, 64)
			//log.Printf("Пришел POST запрос на удаление уровня IDX:%v", r.FormValue("IDX"))
			if e != nil {
				//log.Println("HTTP Del: Error parse: IDX")
				fmt.Fprintf(w, "Error: IDX")
				return
			}

			container.LockerLevels.Lock()
			delete(container.Levels, idx)
			//log.Printf("Удален уровень IDX:%v", idx)
			container.LockerLevels.Unlock()

			//log.Println("HTTP: Deleted IDX:", idx, "Count:", len(container.Levels))

			fmt.Fprintf(w, "OK")

		}
	})

	http.ListenAndServe(Configs.HttpPort, nil)
}
