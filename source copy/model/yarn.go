package model

import (
	"time"
)

type YarnBreakLog struct {
	Entity
	PositionID string    `gorm:"size:36;not null"` // 位置 ID，建议使用 UUID
	LineID    string    `gorm:"size:36;not null"` // 生产线 ID，建议使用 UUID
	StartTime  time.Time `gorm:"type:DateTime"`         // 断裂开始时间
	EndTime    *time.Time `gorm:"type:DateTime"`// 断裂结束时间
	BreakType  string    `gorm:"size:50;not null"` // 断裂类型，例如 "partial", "complete"
	Reason     string    `gorm:"size:255"`         // 断裂原因，描述断裂的原因
	Shift      string    `gorm:"size:50"`          // 班次，例如 "morning", "afternoon", "night"
	Class      string    `gorm:"size:50"`          // 机台班次，例如 "A", "B", "C"
	OperatorID string    `gorm:"size:36"`          // 操作员 ID，建议使用 UUID
}

type YarnPosition struct {
	Entity
	SpinMachineID string      `gorm:"size:36;not null"`                       // 机台 ID，建议使用 UUID
	SpinMachine   SpinMachine `gorm:"foreignKey:SpinMachineID;references:ID"` // 关联的纺丝机实体
	LineID        string      `gorm:"size:36;"`                       // 生产线 ID，建议使用 UUID
	State         bool        `gorm:"default(0)"`                    // 位置状态，使用位字段表示不同状态
}
