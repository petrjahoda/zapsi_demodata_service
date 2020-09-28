package main

import (
	"github.com/petrjahoda/database"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"math/rand"
	"time"
)

func generateDowntimeData(device database.Device) {
	db, err := gorm.Open(postgres.Open(config), &gorm.Config{})
	if err != nil {
		logError("MAIN", "Problem opening  database: "+err.Error())
		return
	}
	sqlDB, err := db.DB()
	defer sqlDB.Close()
	var deviceToReturn database.Device
	db.Where("name=?", device.Name).Find(&deviceToReturn)
	var analogPort database.DevicePort
	db.Where("device_id=?", device.ID).Where("name=?", "Amperage").Find(&analogPort)

	timeToInsert := time.Now()
	min := 80
	max := 100
	randomNumber := rand.Intn(max-min) + min
	recordToInsert := database.DevicePortAnalogRecord{DateTime: timeToInsert, Data: float32(randomNumber), DevicePortID: int(analogPort.ID)}
	db.Create(&recordToInsert)
}

func generateProductionData(device database.Device) {
	db, err := gorm.Open(postgres.Open(config), &gorm.Config{})
	if err != nil {
		logError("MAIN", "Problem opening database: "+err.Error())
		return
	}
	sqlDB, err := db.DB()
	defer sqlDB.Close()
	var deviceToReturn database.Device
	db.Where("name=?", device.Name).Find(&deviceToReturn)
	var digitalPort database.DevicePort
	db.Where("device_id=?", device.ID).Where("name=?", "Production").Find(&digitalPort)
	var analogPort database.DevicePort
	db.Where("device_id=?", device.ID).Where("name=?", "Amperage").Find(&analogPort)
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
