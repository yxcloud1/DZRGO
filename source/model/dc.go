package model

import (
	"time"

	"gorm.io/gorm"
)

type DCDriver struct {
	Entity
	Url     string `gorm:"size:255;not null"` // 关联的 URL
	Enabled bool   `gorm:"default 1"`
}

type DCItem struct {
	IntID       int            `gorm:"column:int_id;autoIncrement;not null;<-:create"` // 自增 ID
	ID          string         `gorm:"size:36;"`                            // 字符串主键，建议用 UUID
	Name        string         `gorm:"size:100;primaryKey"`                            // 名称字段
	Description string         `gorm:"size:255"`                                       // 可选描述信息
	DriverID    string         `gorm:"size:36;primaryKey"`                             // 关联的 Driver ID
	Address     string         `gorm:"size:255"`
	Writable    bool           `gorm:"column:writable"`
	Value       string         `gorm:"size:500"` // 数据值
	Quality     string         `gorm:"size:50"`  // 数据质量，例如 "good", "bad", "unknown"
	DataType    string         `gorm:"size:50"`  // 数据类型，例如 "string", "int", "float", "bool" 等
	Timestamp   time.Time      `gorm:"type:DateTime"`
	CreatedAt   time.Time      `gorm:"type:DateTime"` // 自动维护
	UpdatedAt   time.Time      `gorm:"type:DateTime"` // 自动维护
	DeletedAt   gorm.DeletedAt // 软删除支持（可选）
}
