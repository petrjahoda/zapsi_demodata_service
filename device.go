package main

import (
	"github.com/jinzhu/gorm"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

func (device Device) CreateDirectoryIfNotExists() {
	deviceDirectory := filepath.Join(".", strconv.Itoa(int(device.ID))+"-"+device.Name)

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

func (device Device) GenerateDowntimeData() {
	connectionString, dialect := CheckDatabaseType()
	db, err := gorm.Open(dialect, connectionString)
	if err != nil {
		LogError("MAIN", "Problem opening "+DatabaseName+" database: "+err.Error())
		return
	}
	defer db.Close()
	var deviceToReturn Device
	db.Where("name=?", device.Name).Find(&deviceToReturn)
	var analogPort DevicePort
	db.Where("device_id=?", device.ID).Where("name=?", "Amperage").Find(&analogPort)

	timeToInsert := time.Now()
	min := 80
	max := 100
	randomNumber := rand.Intn(max-min) + min
	recordToInsert := DeviceAnalogRecord{DateTime: timeToInsert, Data: float32(randomNumber), DevicePortId: analogPort.ID, Interval: float32(10)}
	db.NewRecord(recordToInsert)
	db.Create(&recordToInsert)
	analogPort.ActualData = strconv.Itoa(randomNumber)
	analogPort.ActualDataDateTime = timeToInsert
	db.Save(&analogPort)
}

func (device Device) GenerateProductionData() {
	connectionString, dialect := CheckDatabaseType()
	db, err := gorm.Open(dialect, connectionString)
	if err != nil {
		LogError("MAIN", "Problem opening "+DatabaseName+" database: "+err.Error())
		return
	}
	defer db.Close()
	var deviceToReturn Device
	db.Where("name=?", device.Name).Find(&deviceToReturn)
	var digitalPort DevicePort
	db.Where("device_id=?", device.ID).Where("name=?", "Production").Find(&digitalPort)
	var analogPort DevicePort
	db.Where("device_id=?", device.ID).Where("name=?", "Amperage").Find(&analogPort)

	timeToInsert := time.Now()
	timeToInsertForZero := timeToInsert.Add(1 * time.Second)
	recordToInsertOne := DeviceDigitalRecord{DateTime: timeToInsert, Data: 1, DevicePortId: digitalPort.ID, Interval: float32(9)}
	db.NewRecord(recordToInsertOne)
	db.Create(&recordToInsertOne)
	recordToInsertZero := DeviceDigitalRecord{DateTime: timeToInsertForZero, Data: 0, DevicePortId: digitalPort.ID, Interval: float32(1)}
	db.NewRecord(recordToInsertZero)
	db.Create(&recordToInsertZero)
	digitalPort.ActualData = "0"
	digitalPort.ActualDataDateTime = timeToInsert
	db.Save(&digitalPort)

	min := 80
	max := 100
	randomNumber := rand.Intn(max-min) + min
	recordToInsert := DeviceAnalogRecord{DateTime: timeToInsert, Data: float32(randomNumber), DevicePortId: analogPort.ID, Interval: float32(10)}
	db.NewRecord(recordToInsert)
	db.Create(&recordToInsert)
	analogPort.ActualData = strconv.Itoa(randomNumber)
	analogPort.ActualDataDateTime = timeToInsert
	db.Save(&analogPort)
}

func (device Device) GenerateNewState() (actualCycle int, actualState string, totalCycles int) {
	min := 1
	max := 4
	randomNumber := rand.Intn(max-min) + min
	switch randomNumber {
	case 1:
		return 0, "offline", 150
	case 2:
		return 0, "downtime", 150
	default:
		return 0, "production", 300
	}
}

func (device Device) Sleep(start time.Time) {
	if time.Since(start) < (downloadInSeconds * time.Second) {
		sleepTime := downloadInSeconds*time.Second - time.Since(start)
		LogInfo(device.Name, "Sleeping for "+sleepTime.String())
		time.Sleep(sleepTime)
	}
}
