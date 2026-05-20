package handlers

import "gorm.io/gorm"

type Server struct {
	DB *gorm.DB
	JWTSecret []byte
}







