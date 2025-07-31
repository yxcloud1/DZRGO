package model

import (
	"fmt"
	"log"
	"time"

	"github.com/yxcloud1/go-comm/db"
	"gorm.io/gorm"
)

type Entity struct {
	IntID       int            `gorm:"column:int_id;autoIncrement;not null;<-:create"` // 自增 ID
	ID          string         `gorm:"primaryKey;size:36;not null"`          // 字符串主键，建议用 UUID
	Name        string         `gorm:"size:100;not null"`                    // 名称字段
	Description string         `gorm:"size:255"`                             // 可选描述信息
	CreatedAt   time.Time      `gorm:"type:DateTime"` // 自动维护
	UpdatedAt   time.Time      `gorm:"type:DateTime"`// 自动维护
	DeletedAt   gorm.DeletedAt  // 软删除支持（可选）
}

type SysLog struct {
	Entity
	Level      string `gorm:"size:50;not null"`  // 日志级别，例如 "info", "warn", "error"
	Station    string `gorm:"size:100;not null"` // 站点或模块名称
	OperatorID string `gorm:"size:36;not null"`  // 操作员 ID，建议使用 UUID
	Message    string `gorm:"size:2000;not null"` // 日志消息内容
}

var (
	entitys []interface{}
)

func init() {
	entitys = []interface{}{
		&SysLog{},

		&YarnBreakLog{},
		&YarnPosition{},

		&ProdPacketHistory{},
		&ProdPacket{},
		&ProdPacketChangeLog{},

		&POLineCurrentOrder{},
		&POLineHistoryOrder{},

		&SpinBreakLog{},
		&SpinLine{},
		&SpinMachine{},
		&SpinTraverse{},

		&Corporation{},
		&Factory{},
		&Workshop{},

		&DCDriver{},
		&DCItem{},

		&LIMSCustomSample{},
		&LimsDcRequestLog{},
		&LimsDcLog{},

		&View{},
		&ViewParam{},
		&ViewColumn{},
	}
}

func GetEntitys() []interface{} {
	return entitys
}

func InitializeBaseData(){
	corp := []Corporation{}
	if tx := db.DB().Conn().Find(&corp); tx.Error != nil || len(corp) == 0{
		db.DB().Conn().Create(&Corporation{
			Entity: Entity{
				ID: "DZR",
				Name: "山东大自然",
			},
		})
		db.DB().Conn().Create(&Factory{
			Entity: Entity{
				ID: "DZR-F1",
				Name: "大自然工厂1",
			},
			CorporationID: "DZR",
		})
		db.DB().Conn().Create(&Factory{
			Entity: Entity{
				ID: "DZR-F2",
				Name: "大自然工厂2",
			},
			CorporationID: "DZR",
		})
	}else{
		log.Println(tx.Error)
	}
}

func BuildTraverse(){

}


func BuildPosition(){
	var m []SpinMachine
	if tx := db.DB().Conn().Find(&m); tx.Error != nil {
		log.Println("BuildPosition error:", tx.Error)
		return
	}else{
		for _, machine := range m {
			if machine.PositionCount <= 0 {
				log.Printf("SpinMachine %s has no positions defined, skipping...\n", machine.ID)
				continue
			}
			for i := 1; i <= machine.PositionCount; i++ {
				position := YarnPosition{
					Entity: Entity{
						ID:  fmt.Sprintf("%0*s-%03d",3, machine.ID, i),
						Name: fmt.Sprintf("%0*s-%03d",3, machine.ID, i),
					},
					SpinMachineID: machine.ID,
				}

				if db.DB().Conn().Where("id = ?", position.ID).First(&position).RowsAffected ==0 {
				if err := db.DB().Conn().Create(&position).Error; err != nil {
					log.Printf("Failed to create position for SpinMachine %s: %v\n", machine.ID, err)
				}
			}
			}
		}
		log.Println("Positions built successfully.")
	}
}