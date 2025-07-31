package model

import "time"

// 纺丝机
type SpinMachine struct {
	Entity
	WorkshopID    string   `gorm:"size:36;not null"`                    // 车间 ID，建议使用 UUID
	Workshop      Workshop `gorm:"foreignKey:WorkshopID;references:ID"` // 关联的车间实体
	PositionCount int      `gorm:"not null;default:0"`                  // 位置数量，表示该纺丝机的甬道数量
	LineOffset    int      `gorm:"default:0"`                           // 线偏移量，表示该纺丝机的线偏移量
}

// 摆丝机
type SpinTraverse struct {
	Entity
	LineID           int       `gorm:"default:0"`                           // 生产线 ID，表示该摆丝机所在的生产线
	WorkshopID       string    `gorm:"size:36;"`                    // 机台 ID，建议使用 UUID
	Workshop         Workshop  `gorm:"foreignKey:WorkshopID;references:ID"` // 关联的纺丝机实体
	BeginTime        time.Time `gorm:"type:DateTime"`                       // 机台开始时间
	EstEndTime       time.Time `gorm:"type:DateTime"`                       // 机台预计结束时间
	ActualEndTime    time.Time `gorm:"type:DateTime"`                       // 机台实际结束时间
	SetTime          int       `gorm:""`                            // 机台设置时间，单位可以是分钟或秒
	CompensationTime int       `gorm:""`                            // 机台添加时间，单位可以是分钟或秒
	CompletedTime    int       `gorm:""`                            // 机台完成时间，单位可以是分钟或秒
	State            int       `gorm:""`                            // 机台状态，例如 0: 停机, 1: 运行, 2: 故障
	IsBroken         bool      `gorm:""`                            // 是否断裂，true 表示断裂，false 表示正常
}

// 线
type SpinLine struct {
	Entity
	SpinMachineID  string      `gorm:"size:36;not null"`                       // 机台 ID，建议使用 UUID
	IndexOfMachine int         `gorm:"type:int;default:1"`                     // 机台索引，表示该线在机台中的位置
	SpinMachine    SpinMachine `gorm:"foreignKey:SpinMachineID;references:ID"` // 关联的纺丝机实体
	OrderID        string      `gorm:"size:36;not null"`                       // 订单 ID，建议使用 UUID
	Spec           string      `gorm:"size:100;"`                              // 规格，例如 "cotton", "polyester", "blended" 等
	PacketID       string      `gorm:"size:36;not null"`                       // 包 ID，建议使用 UUID
	MonthSEQ       int         `gorm:"type:int;default:1"`                     // 月份序列，表示该线的月份序列
	PackageID      string      `gorm:"size:36;not null;default:1"`             // 包 ID，建议使用 UUID
	EndBrkCount    int         `gorm:"default:0"`                              // 结束断裂次数
	TowBrkCount    int         `gorm:"default:0"`                              // 纱线断裂次数
	LabelPrintTime *time.Time  `gorm:"type:DateTime"`                          // 标签打印时间
}

type SpinBreakLog struct {
	Entity
	LineID     string    `gorm:"size:36;not null"` // 生产线 ID，建议使用 UUID
	PacketID   string    `gorm:"size:36;not null"` // 包 ID，建议使用 UUID
	StartTime  time.Time `gorm:"type:DateTime"`    // 断裂开始时间
	EndTime    time.Time `gorm:"type:DateTime"`    // 断裂结束时间
	BreakType  string    `gorm:"size:50;not null"` // 断裂类型，例如 "partial", "complete"
	Reason     string    `gorm:"size:255"`         // 断裂原因，描述断裂的原因
	Shift      string    `gorm:"size:50"`          // 班次，例如 "morning", "afternoon", "night"
	Class      string    `gorm:"size:50"`          // 机台班次，例如 "A", "B", "C"
	OperatorID string    `gorm:"size:36"`          // 操作员 ID，建议使用 UUID
}
