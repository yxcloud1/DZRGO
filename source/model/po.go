package model

import "time"

type POView struct {
	Entity
	DataType        string `gorm:"size:50;not null"` // 数据类型，例如 "string", "int", "float", "bool" 等
	Format          string `gorm:"size:500"`         // 数据格式，例如 "json", "xml", "csv" 等
	Style           string `gorm:"size:500"`         // 数据样式，例如 "default", "compact", "detailed" 等
	RowIndex        int    `gorm:"not null"`         // 行索引，用于排序或定位
	ColIndex        int    `gorm:"not null"`         // 列索引，用于排序或定位
	Width           int    `gorm:"not null"`         // 列宽度，单位可以是像素或百分比
	Height          int    `gorm:"not null"`         // 行高度，单位可以是像素或百分比
	Visible         bool   `gorm:"not null"`         // 是否可见
	RowSpan         int    `gorm:"not null"`         // 行跨度，表示该视图在垂直方向上占据的行数
	ColSpan         int    `gorm:"not null"`         // 列跨度，表示该视图在水平方向上占据的列数
	Expression      string `gorm:"size:500"`         // 表达式，用于计算或转换数据
	DefaultValue    string `gorm:"size:500"`         // 默认值，用于初始化数据
	Placeholder     string `gorm:"size:500"`         // 占位符，用于输入提示
	ValidationRules string `gorm:"size:500"`         // 验证规则，用于数据校验
	Tooltip         string `gorm:"size:500"`         // 工具提示，用于提供额外信息
	Icon            string `gorm:"size:100"`         // 图标，用于视觉标识
	FixedText       string `gorm:"size:500"`         // 固定文本，用于显示静态信息
}

type POOrder struct {
	Entity
	Spec             string    `gorm:"size:100;not null"` // 规格，例如 "asc", "desc"
	Lines            string    `gorm:"size:500"`          // 订单行，表示订单中的具体项
	State            string    `gorm:"size:50"`           // 订单状态，例如 "pending", "processing", "completed", "cancelled"
	ExecuteTime      time.Time `gorm:"type:DateTime"`     // 执行时间，表示订单的执行时间
	ExecuteStatus    string    `gorm:"size:200;not null"` // 执行状态，例如 "pending", "completed", "failed"
	ExecuteResult    string    `gorm:"size:500"`          // 执行结果，表示订单执行的结果
	ExecuteMessage   string    `gorm:"size:500"`          // 执行消息，表示订单执行的详细信息
	ExecuteUser      string    `gorm:"size:100"`          // 执行用户，表示执行订单的用户
	ExecuteUserID    string    `gorm:"size:36"`           // 执行用户 ID，建议使用 UUID
	ExecuteTimeStart time.Time `gorm:"type:DateTime"`     // 执行开始时间，表示订单执行的开始时间
	ExecuteTimeEnd   time.Time `gorm:"type:DateTime"`     // 执行结束时间，表示订单执行的结束时间
}

type POOrderItem struct {
	Entity
	OrderID string `gorm:"size:36;not null"`  // 订单 ID，建议使用 UUID
	Value   string `gorm:"size:500;not null"` // 订单项的值
}

type POLineCurrentOrder struct {
	Entity
	OrderID     string    `gorm:"size:36;not null"` // 订单 ID，建议使用 UUID
	Spec        string    `gorm:"size:36;"`         // 行 ID，建议使用 UUID
	ExecuteTime time.Time `gorm:"type:DateTime"`    // 执行时间，表示订单行的执行时间
}

type POLineHistoryOrder struct {
	Entity
	OrderID     string `gorm:"size:36;not null"` // 订单 ID，建议使用 UUID
	Spec           string    `gorm:"size:100;"`        // 规格，例如 "asc", "desc"\
	MachineID   string
	LineID      string    `gorm:"size:36;not null"` // 行 ID，建议使用 UUID
	ExecuteTime time.Time `gorm:"type:DateTime"`    // 执行时间，表示订单行的执行时间
	EndTime     time.Time `gorm:"type:DateTime"`    // 结束时间，表示订单行的结束时间
}

type POAdjustmentOrder struct {
	Entity
	OrderID        string    `gorm:"size:36;not null"` // 订单 ID，建议使用 UUID
	Time           time.Time `gorm:"type:DateTime"`    // 调整时间，表示调整操作的时间
	UserID         string    `gorm:"size:36"`          // 调整用户 ID，建议使用 UUID
	Reason         string    `gorm:"size:500"`         // 调整原因，表示为什么需要进行调整
	AdjustmentType string    `gorm:"size:50;not null"` // 调整类型，例如 "increase", "decrease", "set"
}

type POAdjustmentOrderItem struct {
	Entity
	OrderID      string `gorm:"size:36;not null"`  // 订单 ID，建议使用 UUID
	AdjustmentID string `gorm:"size:36;not null"`  // 调整项 ID，建议使用 UUID
	Field        string `gorm:"size:100;not null"` // 字段名称，表示需要调整的字段
	Value        string `gorm:"size:500;not null"` // 调整后的值
}
