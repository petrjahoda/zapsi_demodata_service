package main

import (
	"github.com/petrjahoda/database"
	"gorm.io/gorm"
	"math/rand"
	"time"
)

func generateDowntimeData(db *gorm.DB, analogPort database.DevicePort) {
	timeToInsert := time.Now()
	min := 80
	max := 100
	randomNumber := rand.Intn(max-min) + min
	recordToInsert := database.DevicePortAnalogRecord{DateTime: timeToInsert, Data: float32(randomNumber), DevicePortID: int(analogPort.ID)}
	db.Create(&recordToInsert)
}

func generateProductionData(db *gorm.DB, digitalPort database.DevicePort, analogPort database.DevicePort) {
	timeToInsert := time.Now()
	timeToInsertForZero := timeToInsert.Add(1 * time.Second)
	recordToInsertOne := database.DevicePortDigitalRecord{DateTime: timeToInsert, Data: 1, DevicePortID: int(digitalPort.ID)}
	db.Create(&recordToInsertOne)
	recordToInsertZero := database.DevicePortDigitalRecord{DateTime: timeToInsertForZero, Data: 0, DevicePortID: int(digitalPort.ID)}
	db.Create(&recordToInsertZero)
	min := 80
	max := 100
	randomNumber := rand.Intn(max-min) + min
	recordToInsert := database.DevicePortAnalogRecord{DateTime: timeToInsert, Data: float32(randomNumber), DevicePortID: int(analogPort.ID)}
	db.Create(&recordToInsert)
}

func generateNewState() (actualCycle int, actualState string, totalCycles int) {
	min := 1
	max := 4
	randomNumber := rand.Intn(max-min) + min
	switch randomNumber {
	case 1:
		return 0, "poweroff", 150
	case 2:
		return 0, "downtime", 150
	default:
		return 0, "production", 300
	}
}
