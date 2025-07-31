package model

type Corporation struct {
	Entity
	Factorys []Factory `gorm:"foreignKey:CorporationID;references:ID"` // 关联的 Factory 列表
}

func (t *Corporation) TableName() string {
	return "t_org_corporation"
}

type Factory struct {
	Entity
	CorporationID string `gorm:"size:36;not null"` // 关联的 Corporation ID
	Corporation   Corporation `gorm:"foreignKey:CorporationID;references:ID"` // 关联的 Corporation 实体
	Workshops     []Workshop `gorm:"foreignKey:FactoryID;references:ID"` // 关联的 Workshop 列表
}

func (t *Factory) TableName() string {
	return "t_org_factory"
}

type Workshop struct {
	Entity
	FactoryID string `gorm:"size:36;not null"` // 关联的 Factory ID
	Factory   Factory `gorm:"foreignKey:FactoryID;references:ID"` // 关联的 Factory 实体
}

func (t *Workshop) TableName() string {
	return "t_org_workshop"
}
