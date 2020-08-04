package main

import (
	"github.com/petrjahoda/database"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

func CreateDirectoryIfNotExists(device database.Device) {
	deviceDirectory := filepath.Join(".", strconv.Itoa(int(device.ID))+"-"+device.Name)
	if _, checkPathError := os.Stat(deviceDirectory); checkPathError == nil {
		LogInfo(device.Name, "Device directory exists")
	} else if os.IsNotExist(checkPathError) {
		LogInfo(device.Name, "Device directory not exist, creating")
		mkdirError := os.MkdirAll(deviceDirectory, 0777)
		if mkdirError != nil {
			LogError(device.Name, "Unable to create device directory: "+mkdirError.Error())
		} else {
			LogInfo(device.Name, "Device directory created")
		}
	} else {
		LogError(device.Name, "Device directory does not exist")
	}
}

func GenerateDowntimeData(device database.Device) {
	db, err := gorm.Open(postgres.Open(config), &gorm.Config{})
	if err != nil {
		LogError("MAIN", "Problem opening  database: "+err.Error())
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

func GenerateProductionData(device database.Device) {
	db, err := gorm.Open(postgres.Open(config), &gorm.Config{})
	if err != nil {
		LogError("MAIN", "Problem opening database: "+err.Error())
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

func GenerateNewState() (actualCycle int, actualState string, totalCycles int) {
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
