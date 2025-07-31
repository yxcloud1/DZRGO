package model

type View struct {
	Entity
	Text        string `gorm:"size:100;not null"` // 视图名称
	Command     string `gorm:"size:500;not null"` // 视图命令，存储 SQL 查询或其他命令
	CommandType string `gorm:"size:50;not null"`  // 命令类型，例如 "sql", "api", "script" 等
}

type ViewParam struct {
	Entity
	ViewID       string `gorm:"size:36;not null"`           // 关联的视图 ID，建议使用 UUID
	Name         string `gorm:"size:100;not null"`          // 参数名称
	DefaultValue string `gorm:"size:500"`                   // 默认值，表示参数的默认值
	UIType       string `gorm:"size:50;not null"`           // 参数类型，例如 "string", "int", "float", "bool" 等
	Type         string `gorm:"size:50;not null;default:P"` // 参数类型，例如 "input", "select", "checkbox" 等
	Data         string `gorm:"size:8000"`                  // 事件类型，例如 "click", "change", "submit" 等
}

type ViewColumn struct {
	Entity
	ViewID      string
	DisplayName string
	FieldName   string
	Index       int
	Format      string
	Style       string
}
