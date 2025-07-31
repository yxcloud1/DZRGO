package model

import "time"

type LIMSCustomSample struct {
	Entity
	PacketID     string `gorm:"size:36;"` // 关联的包 ID，建议使用 UUID
	ProdOrderId  string
	SerialNumber int `gorm:"defaukt:1"` // 样品序列号，唯一标识

	Spec      string `gorm:"size:100;"`        // 样品规格
	LineId    string `gorm:"size:36;not null"` // 关联的生产线 ID，建议使用 UUID
	ItemCodes string `gorm:"size:500"`         // 样品项，JSON 格式存储
}

type LimsDcRequestLog struct {
	ID         int       `gorm:"column:id;autoIncrement;not null;<-:create"` // 自增 ID
	ClientIP   string    `gorm:"column:client_ip;size:50"`
	RequestUrl string    `gorm:"column:request_url;size:500"`
	DeviceType string    `gorm:"size:100;not null"` // 名称字段
	DeviceID   string    `gorm:"size:255"`          // 可选描述信息
	Request    string    `gorm:"column:request"`
	Response   string    `gorm:"column:response"`
	CreatedAt  time.Time `gorm:"type:DateTime"` // 自动维护
}

type LimsDcLog struct {
	ID          int       `gorm:"column:id;autoIncrement;not null;<-:create"` // 自增 ID
	DeviceType  string    `gorm:"size:100;not null"`                          // 名称字段
	DeviceID    string    `gorm:"size:255"`                                   // 可选描述信息
	SampleID    string    `gorm:"column:sample_id;size:50"`
	RawID       int       `gorm:"column:raw_id"`
	RawData     string    `gorm:"column:raw_data"`
	ItemCodes   string    `gorm:"column:item_codes;size:100"`
	ItemValue1  string    `gorm:"column:item_value1;szie:255"`
	ItemValue2  string    `gorm:"column:item_value2;szie:255"`
	ItemValue3  string    `gorm:"column:item_value3;szie:255"`
	ItemValue4  string    `gorm:"column:item_value4;szie:255"`
	ItemValue5  string    `gorm:"column:item_value5;szie:255"`
	ItemValue6  string    `gorm:"column:item_value6;szie:255"`
	ItemValue7  string    `gorm:"column:item_value7;szie:255"`
	ItemValue8  string    `gorm:"column:item_value8;szie:255"`
	ItemValue9  string    `gorm:"column:item_value9;szie:255"`
	ItemValue10 string    `gorm:"column:item_value10;szie:255"`
	ItemValue11 string    `gorm:"column:item_value11;szie:255"`
	ItemValue12 string    `gorm:"column:item_value12;szie:255"`
	ItemValue13 string    `gorm:"column:item_value13;szie:255"`
	ItemValue14 string    `gorm:"column:item_value14;szie:255"`
	ItemValue15 string    `gorm:"column:item_value15;szie:255"`
	CreatedAt   time.Time `gorm:"type:DateTime"` // 自动维护
}
