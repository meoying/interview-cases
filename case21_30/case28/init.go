package case28

import "gorm.io/gorm"

func InitTables(db *gorm.DB) error {
	return db.AutoMigrate(&Order{},&Notification{})
}
