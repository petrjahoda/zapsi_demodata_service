package main

import (
	"github.com/jinzhu/gorm"
	"github.com/petrjahoda/zapsi_database"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

func CreateDirectoryIfNotExists(device zapsi_database.Device) {
	deviceDirectory := filepath.Join(".", strconv.Itoa(device.ID)+"-"+device.Name)

	if _, checkPathError := os.Stat(deviceDirectory); checkPathError == nil {
		LogInfo(device.Name, "Device directory exists")
	} else if os.IsNotExist(checkPathError) {
		LogWarning(device.Name, "Device directory not exist, creating")
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

func GenerateDowntimeData(device zapsi_database.Device) {
	connectionString, dialect := zapsi_database.CheckDatabaseType(DatabaseType, DatabaseIpAddress, DatabasePort, DatabaseLogin, DatabaseName, DatabasePassword)
	db, err := gorm.Open(dialect, connectionString)
	if err != nil {
		LogError("MAIN", "Problem opening "+DatabaseName+" database: "+err.Error())
		return
	}
	defer db.Close()
	var deviceToReturn zapsi_database.Device
	db.Where("name=?", device.Name).Find(&deviceToReturn)
	var analogPort zapsi_database.DevicePort
	db.Where("device_id=?", device.ID).Where("name=?", "Amperage").Find(&analogPort)

	timeToInsert := time.Now()
	min := 80
	max := 100
	randomNumber := rand.Intn(max-min) + min
	recordToInsert := zapsi_database.DevicePortAnalogRecord{DateTime: timeToInsert, Data: float32(randomNumber), DevicePortId: analogPort.ID, Duration: time.Second * 10}
	db.NewRecord(recordToInsert)
	db.Create(&recordToInsert)
	analogPort.ActualData = strconv.Itoa(randomNumber)
	analogPort.ActualDataDateTime = timeToInsert
	db.Save(&analogPort)
}

func GenerateProductionData(device zapsi_database.Device) {
	connectionString, dialect := zapsi_database.CheckDatabaseType(DatabaseType, DatabaseIpAddress, DatabasePort, DatabaseLogin, DatabaseName, DatabasePassword)
	db, err := gorm.Open(dialect, connectionString)
	if err != nil {
		LogError("MAIN", "Problem opening "+DatabaseName+" database: "+err.Error())
		return
	}
	defer db.Close()
	var deviceToReturn zapsi_database.Device
	db.Where("name=?", device.Name).Find(&deviceToReturn)
	var digitalPort zapsi_database.DevicePort
	db.Where("device_id=?", device.ID).Where("name=?", "Production").Find(&digitalPort)
	var analogPort zapsi_database.DevicePort
	db.Where("device_id=?", device.ID).Where("name=?", "Amperage").Find(&analogPort)

	timeToInsert := time.Now()
	timeToInsertForZero := timeToInsert.Add(1 * time.Second)
	recordToInsertOne := zapsi_database.DevicePortDigitalRecord{DateTime: timeToInsert, Data: 1, DevicePortId: digitalPort.ID, Duration: time.Second * 9}
	db.NewRecord(recordToInsertOne)
	db.Create(&recordToInsertOne)
	recordToInsertZero := zapsi_database.DevicePortDigitalRecord{DateTime: timeToInsertForZero, Data: 0, DevicePortId: digitalPort.ID, Duration: time.Second * 1}
	db.NewRecord(recordToInsertZero)
	db.Create(&recordToInsertZero)
	digitalPort.ActualData = "0"
	digitalPort.ActualDataDateTime = timeToInsert
	db.Save(&digitalPort)

	min := 80
	max := 100
	randomNumber := rand.Intn(max-min) + min
	recordToInsert := zapsi_database.DevicePortAnalogRecord{DateTime: timeToInsert, Data: float32(randomNumber), DevicePortId: analogPort.ID, Duration: time.Second * 10}
	db.NewRecord(recordToInsert)
	db.Create(&recordToInsert)
	analogPort.ActualData = strconv.Itoa(randomNumber)
	analogPort.ActualDataDateTime = timeToInsert
	db.Save(&analogPort)
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
