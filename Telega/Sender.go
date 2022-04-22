package Telega

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
)

const telegramApiKey = "452751479:AAHpwefrwewecwecwecwecwecwecE"

func Send(chat_id int64, msg string) bool {
	if chat_id == 0 {
		//log.Println("Не отправлен, chatId = 0")
		return false
	}

	form := &url.Values{}

	form.Add("chat_id", strconv.FormatInt(chat_id, 10))
	form.Add("text", msg)

	buffer := new(bytes.Buffer)
	buffer.WriteString(form.Encode())

	req, err := http.NewRequest("POST", fmt.Sprintf("https://api.telegram.org/bot%v/sendMessage", telegramApiKey), buffer)

	if err != nil {
		log.Println("Error: SendTelegram:", err)
		return false
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "Bearer b7d03a6947b217efb6f3ec3bd3504582")

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		log.Println("Error: SendTelegram:", err)
		return false
	} else {
		//log.Println("SendTelegram: Answer:", resp.Body)

		if err := resp.Body.Close(); err != nil {
			return false
		}
	}

	return true
}
