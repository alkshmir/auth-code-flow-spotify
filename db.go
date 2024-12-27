package main

import (
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	ID        uint   `gorm:"primaryKey"`
	Username  string `gorm:"unique;not null"`
	Password  string `gorm:"not null"` // hashed password by bcrypt
	CreatedAt time.Time
}

type SpotifyToken struct {
	gorm.Model
	UserID        uint
	User          User
	ID            uint   `gorm:"primaryKey"`
	SpotifyUserId string `gorm:"not null"`
	AccessToken   string `gorm:"not null"`
	RefreshToken  string `gorm:"not null"`
}

// Use defer db.Close() after calling this function
func initDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	db.AutoMigrate(&User{})
	return db
}
