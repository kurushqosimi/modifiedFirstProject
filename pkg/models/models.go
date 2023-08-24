package models

import "time"

type Config struct {
	Host string `json:"host"`
	Port string `json:"port"`
}
type Notes struct {
	ID        int       `json:"id"`
	Content   string    `json:"content"`
	Date      time.Time `json:"date"`
	Active    bool      `json:"active"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt time.Time `json:"deleted_at"`
}
