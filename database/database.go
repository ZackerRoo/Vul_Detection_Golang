package database

import (
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type VulnerabilityRecord struct {
	ID               uint   `gorm:"primary_key"`
	ContractSource   string `gorm:"type:text"`
	HasVulnerability bool   `gorm:"type:boolean"`
	DetectedLines    string `gorm:"type:text"`
	Message          string
	CreatedAt        time.Time
}

type Database interface {
	InsertRecord(record *VulnerabilityRecord) error
	GetRecordByID(id uint) (*VulnerabilityRecord, error)
	GetAllRecords() ([]VulnerabilityRecord, error)
	UpdateRecord(record *VulnerabilityRecord) error
	DeleteRecord(id uint) error
}

type GromDataBase struct {
	DB *gorm.DB
}

// 初始化数据库连接
func NewDatabase(dsn string) (*GromDataBase, error) {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := db.AutoMigrate(&VulnerabilityRecord{}); err != nil { // 如果在mysql中没有这个表自动创建表
		return nil, fmt.Errorf("failed to migrate database schema: %w", err)
	}
	return &GromDataBase{DB: db}, nil

}

func (db *GromDataBase) InsertRecord(record *VulnerabilityRecord) error {
	return db.DB.Create(record).Error
}

func (db *GromDataBase) GetRecordByID(id uint) (*VulnerabilityRecord, error) {
	var record VulnerabilityRecord
	err := db.DB.First(&record, id).Error
	return &record, err
}

func (db *GromDataBase) GetAllRecords() ([]*VulnerabilityRecord, error) {
	var records []*VulnerabilityRecord
	err := db.DB.Find(&records).Error //查询所有的方法就是Find 不然查一条就是First
	return records, err
}

func (db *GromDataBase) DeleteRecord(id uint) error {
	return db.DB.Delete(&VulnerabilityRecord{}, id).Error
}
