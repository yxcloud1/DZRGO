package model

import "time"

type ProdPacket struct {
	Entity
	MachineID             string `gorm:"size:36;"`         // 机台 ID，建议使用 UUID
	LineID                string `gorm:"size:36;not null"` // 生产线 ID，建议使用 UUID
	ProdOrderID           string
	Spec                  string
	LayerShift            string    `gorm:"size:50"`       // 层次班次，例如 "morning", "afternoon", "night"
	LayerClass            string    `gorm:"size:50"`       // 层次机台班次，例如 "A", "B", "C"
	LayerBeginTime        time.Time `gorm:"type:DateTime"` // 层次开始时间
	LayerEndTime          time.Time `gorm:"type:DateTime"` // 层次结束时间
	LayerSetTime          int       `gorm:""`              // 层次设置时间，单位可以是分钟或秒
	LayerCompensationTime int       `gorm:""`              // 层次添加时间，单位可以是分钟或秒
	EndBrkCount           int       `gorm:"default:0"`     // 结束断裂次数
	TowBrkCount           int       `gorm:"default:0"`     // 纱线断裂次数
	PktTime               time.Time `gorm:"type:DateTime"` //打包时间
	PktShift              string    `gorm:"size:50"`       // 班次，例如 "morning", "afternoon", "night"
	PktClass              string    `gorm:"size:50"`       // 机台班次，例如 "A", "B", "C"
	NW                    float64   `gorm:""`              // 净重，单位可以是克或千克
	GW                    float64   `gorm:""`              // 毛重，单位可以是克或千克
	EstWeight             float64   `gorm:""`              // 预计重量，单位可以是克或千克
	PktPressure           float64   `gorm:""`              // 打包压力，单位可以是帕斯卡（Pa）或牛顿（N）
	HoldingPressure       float64   `gorm:""`              // 保持压力，单位可以是帕斯卡（Pa）或牛顿（N）
	HoldingTime           int       `gorm:""`              // 保持时间，单位可以是秒
	PktHeight1            float64   // 打包高度1，单位可以是厘米或毫米
	PktHeight2            float64   // 打包高度2，单位可以是厘米或毫米
	PktHeight3            float64   // 打包高度3，单位可以是厘米或毫米
	Moisture              float64   // 湿度，单位可以是百分比
	PktOperator           string    `gorm:"size:100"`            // 打包操作人
	LayerBrkCount         int       `gorm:"not null;default(0)"` // 层次断裂次数
	YarnBrkCount          int       `gorm:"not null;default(0)"` // 纱线断裂次数
	ProState              string    `gorm:"size:50;not null"`    // 生产状态，例如 "pending", "in_progress", "completed", "failed"
}

type ProdPacketChangeLog struct {
	Entity
	PacketId              string `gorm:"type:nvarchar(64);"` // 包号
	MachineID             string `gorm:"size:36;"`           // 机台 ID，建议使用 UUID
	LineID                string `gorm:"size:36;not null"`   // 生产线 ID，建议使用 UUID
	ProdOrderID           string
	Spec                  string
	LayerShift            string    `gorm:"size:50"`       // 层次班次，例如 "morning", "afternoon", "night"
	LayerClass            string    `gorm:"size:50"`       // 层次机台班次，例如 "A", "B", "C"
	LayerBeginTime        time.Time `gorm:"type:DateTime"` // 层次开始时间
	LayerEndTime          time.Time `gorm:"type:DateTime"` // 层次结束时间
	LayerSetTime          int       `gorm:""`              // 层次设置时间，单位可以是分钟或秒
	LayerCompensationTime int       `gorm:""`              // 层次添加时间，单位可以是分钟或秒
	EndBrkCount           int       `gorm:"default:0"`     // 结束断裂次数
	TowBrkCount           int       `gorm:"default:0"`     // 纱线断裂次数
	PktTime               time.Time `gorm:"type:DateTime"` //打包时间
	PktShift              string    `gorm:"size:50"`       // 班次，例如 "morning", "afternoon", "night"
	PktClass              string    `gorm:"size:50"`       // 机台班次，例如 "A", "B", "C"
	NW                    float64   `gorm:""`              // 净重，单位可以是克或千克
	GW                    float64   `gorm:""`              // 毛重，单位可以是克或千克
	EstWeight             float64   `gorm:""`              // 预计重量，单位可以是克或千克
	PktPressure           float64   `gorm:""`              // 打包压力，单位可以是帕斯卡（Pa）或牛顿（N）
	HoldingPressure       float64   `gorm:""`              // 保持压力，单位可以是帕斯卡（Pa）或牛顿（N）
	HoldingTime           int       `gorm:""`              // 保持时间，单位可以是秒
	PktHeight1            float64   // 打包高度1，单位可以是厘米或毫米
	PktHeight2            float64   // 打包高度2，单位可以是厘米或毫米
	PktHeight3            float64   // 打包高度3，单位可以是厘米或毫米
	Moisture              float64   // 湿度，单位可以是百分比
	PktOperator           string    `gorm:"size:100"`            // 打包操作人
	LayerBrkCount         int       `gorm:"not null;default(0)"` // 层次断裂次数
	YarnBrkCount          int       `gorm:"not null;default(0)"` // 纱线断裂次数
	ProState              string    `gorm:"size:50;not null"`    // 生产状态，例如 "pending", "in_progress", "completed", "failed"
	Operator              string    `gorm:"size:100"`            // 操作人
}

type ProdPacketHistory struct {
	Entity
	PacketID     string    `gorm:"type:nvarchar(64);not null"` // 包号
	StatusBefore string    `gorm:"type:nvarchar(32)"`          // 改变前状态
	StatusAfter  string    `gorm:"type:nvarchar(32)"`          // 改变后状态
	Location     string    `gorm:"type:nvarchar(128)"`         // 操作地点
	Operator     string    `gorm:"type:nvarchar(64)"`          // 操作人
	OperatedAt   time.Time `gorm:"type:datetime"`              // 操作时间
	Remark       string    `gorm:"type:nvarchar(255)"`         // 备注
}
