package models

import "gorm.io/gorm"

type Setting struct {
    gorm.Model
    Name    string `json:"name"`
    Address string `json:"address"`
    Phone   string `json:"phone"`
    Email   string `json:"email"`
    OpenAt  string `json:"open_at"`
    CloseAt string `json:"close_at"`
}