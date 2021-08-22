package database

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"time"
)

type Feed struct {
	gorm.Model

	// properties
	Title       string
	Description string
	Link        string

	// relationships
	Items []Item
}

type Item struct {
	gorm.Model

	// properties
	Title       string
	Description string
	Link        string
	GUID        string `gorm:"index:idx_item"`
	PublishDate *time.Time

	// relationships
	FeedID uint
	Feed Feed
}

type Database struct {
	db *gorm.DB
}

func Connect(dsn string) (*Database, error) {
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// run the migrations
	if err := db.AutoMigrate(&Feed{}, &Item{}); err != nil {
		return nil, err
	}

	return &Database{db: db}, nil
}

func (d *Database) GetFeeds() ([]Feed, error) {
	var feeds []Feed
	tx := d.db.Preload("Items").Find(&feeds)

	return feeds, tx.Error
}

func (d *Database) CreateItem(feed *Feed, item *Item) error {
	return d.db.Model(feed).Association("Items").Append(item)
}

func (d *Database) GetItems() ([]Item, error) {
	var items []Item
	tx := d.db.Preload("Feed").Find(&items)

	return items, tx.Error
}
