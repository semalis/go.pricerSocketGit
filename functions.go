package main

import "time"

func GetDayStart() int64 {
	now := time.Now()

	y, m, d := now.Date()


	return time.Date(y, m, d, 0, 0, 0, 0, now.Location()).Unix()
}
