package biz

import (
	"database/sql"

	"gorm.io/gorm"
)

// gorm的模型
type Customer struct {
	// 嵌入4个基础字段
	gorm.Model
	// 业务逻辑
	CustomerWork
	// token部分
	CustomerToken
}

// 业务逻辑部分
type CustomerWork struct {
	PhoneNum string `gorm:"type: varchar(15);uniqueIndex" json:"phone_num,omitempty"`
	Name     string `gorm:"type: varchar(15);uniqueIndex" json:"name,omitempty"`
	Email    string `gorm:"type: varchar(255);uniqueIndex" json:"email,omitempty"`
	Wechat   string `gorm:"type: varchar(255);uniqueIndex" json:"wechat,omitempty"`
	CityID   uint   `gorm:"index;" json:"cityid,omitempty"`
}

type CustomerToken struct {
	Token          string       `gorm:"type: varchar(4095);" json:"token,omitempty"`
	TokenCreatedAt sql.NullTime `gorm:"" json:"token_created_at,omitempty"`
}
