package model

import "sync"

type User struct {
	Uuid     string `json:"uuid"`
	Username string `json:"username"`
	Score    int    `json:"score"`
}

var Users map[string]User

var once sync.Once

func Init() {
	once.Do(func() {
		Users = make(map[string]User)
	})
}
