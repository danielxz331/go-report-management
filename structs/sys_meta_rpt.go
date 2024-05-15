package structs

import (
	"time"
)

type SysMetaRpt struct {
	ID         uint      `gorm:"primaryKey;autoIncrement"`
	SiteID     *uint     `gorm:"index"`
	CreatedBy  *uint     `gorm:"index"`
	DateCreate time.Time `gorm:"autoUpdateTime;column:datecreate"`
	Module     string    `gorm:"size:100"`
	Name       string    `gorm:"size:50"`
	Query      string    `gorm:"type:longtext"`
	Graph      string
	Status     int
	Where      string `gorm:"column:_where"`
	Headers    string
}

func (SysMetaRpt) TableName() string {
	return "sys_meta_rpt"
}
