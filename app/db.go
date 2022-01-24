package app

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
)

type SQLiteDB struct {
	DB *sql.DB
}

func NewSqliteDB(db *sql.DB) *SQLiteDB {
	return &SQLiteDB{
		DB: db,
	}
}

func (db *SQLiteDB) GetActivityDefinition(activityId int32) ActivityDefinition {
	row := db.DB.QueryRow("SELECT * FROM DestinyActivityDefinition WHERE id = ?", activityId)
	var id int32
	var jsonContent string
	err := row.Scan(&id, &jsonContent)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
	}

	activity := ActivityDefinition{}
	jsonErr := json.Unmarshal([]byte(jsonContent), &activity)
	if jsonErr != nil {
		log.Fatal(jsonErr)
	}

	return activity
}

func (db *SQLiteDB) GetInventoryItemDefinition(itemId int32) InventoryItemDefinition {
	row := db.DB.QueryRow("SELECT * FROM DestinyInventoryItemDefinition WHERE id = ?", itemId)
	var id int32
	var jsonContent string
	err := row.Scan(&id, &jsonContent)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
	}

	item := InventoryItemDefinition{}
	jsonErr := json.Unmarshal([]byte(jsonContent), &item)
	if jsonErr != nil {
		log.Fatal(jsonErr)
	}

	return item
}
