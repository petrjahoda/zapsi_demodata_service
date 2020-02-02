package main

import (
	"github.com/jinzhu/gorm"
	"github.com/petrjahoda/zapsi_database"
	"strconv"
	"sync"
	"time"
)

const version = "2020.1.2.2"
const programName = "Zapsi Demodata Service"
const deleteLogsAfter = 240 * time.Hour
const downloadInSeconds = 10

var (
	activeDevices  []zapsi_database.Device
	runningDevices []zapsi_database.Device
	deviceSync     sync.Mutex
)

func main() {
	time.Sleep(2 * time.Second)
	LogDirectoryFileCheck("MAIN")
	LogInfo("MAIN", "Program version "+version+" started")
	CreateConfigIfNotExists()
	LoadSettingsFromConfigFile()
	LogDebug("MAIN", "Using ["+DatabaseType+"] on "+DatabaseIpAddress+":"+DatabasePort+" with database "+DatabaseName)
	for {
		start := time.Now()
		LogInfo("MAIN", "Program running")
		CompleteDatabaseCheck()
		UpdateActiveDevices("MAIN")
		if len(activeDevices) == 0 {
			CreateDevicesAndWorkplaces()
		}
		DeleteOldLogFiles()
		LogInfo("MAIN", "Active devices: "+strconv.Itoa(len(activeDevices))+", running devices: "+strconv.Itoa(len(runningDevices)))
		for _, activeDevice := range activeDevices {
			activeDeviceIsRunning := CheckDevice(activeDevice)
			if !activeDeviceIsRunning {
				go RunDevice(activeDevice)
			}
		}
		if time.Since(start) < (downloadInSeconds * time.Second) {
			sleeptime := downloadInSeconds*time.Second - time.Since(start)
			LogInfo("MAIN", "Sleeping for "+sleeptime.String())
			time.Sleep(sleeptime)
		}
	}
}

func CreateDevicesAndWorkplaces() {
	connectionString, dialect := zapsi_database.CheckDatabaseType(DatabaseType, DatabaseIpAddress, DatabasePort, DatabaseLogin, DatabaseName, DatabasePassword)
	db, err := gorm.Open(dialect, connectionString)
	if err != nil {
		LogError("MAIN", "Problem opening "+DatabaseName+" database: "+err.Error())
		return
	}
	defer db.Close()
	for i := 0; i < 20; i++ {
		AddTestWorkplace("MAIN", "CNC "+strconv.Itoa(i), "192.168.0."+strconv.Itoa(i))
	}
}

func AddTestWorkplace(reference string, workplaceName string, ipAddress string) {
	connectionString, dialect := zapsi_database.CheckDatabaseType(DatabaseType, DatabaseIpAddress, DatabasePort, DatabaseLogin, DatabaseName, DatabasePassword)
	db, err := gorm.Open(dialect, connectionString)
	if err != nil {
		LogError(reference, "Problem opening "+DatabaseName+" database: "+err.Error())
		return
	}
	defer db.Close()
	var deviceType zapsi_database.DeviceType
	db.Where("name=?", "Zapsi").Find(&deviceType)
	newDevice := zapsi_database.Device{Name: workplaceName, DeviceTypeId: deviceType.ID, IpAddress: ipAddress, TypeName: "Zapsi", Activated: true}
	db.Create(&newDevice)
	var device zapsi_database.Device
	db.Where("name=?", workplaceName).Find(&device)
	deviceDigitalPort := zapsi_database.DevicePort{Name: "Production", Unit: "ks", PortNumber: 1, DevicePortTypeId: 1, DeviceId: device.ID}
	deviceAnalogPort := zapsi_database.DevicePort{Name: "Amperage", Unit: "A", PortNumber: 3, DevicePortTypeId: 2, DeviceId: device.ID}
	db.Create(&deviceDigitalPort)
	db.Create(&deviceAnalogPort)
	var section zapsi_database.WorkplaceSection
	db.Where("name=?", "Machines").Find(&section)
	var state zapsi_database.State
	db.Where("name=?", "Poweroff").Find(&state)
	var mode zapsi_database.WorkplaceMode
	db.Where("name=?", "Production").Find(&mode)
	newWorkplace := zapsi_database.Workplace{Name: workplaceName, Code: workplaceName, WorkplaceSectionId: section.ID, ActualStateId: state.ID, ActualWorkplaceModeId: mode.ID}
	db.Create(&newWorkplace)
	var workplace zapsi_database.Workplace
	db.Where("name=?", workplaceName).Find(&workplace)
	var devicePortDigital zapsi_database.DevicePort
	db.Where("name=?", "Production").Where("device_id=?", device.ID).Find(&devicePortDigital)
	var productionState zapsi_database.State
	db.Where("name=?", "Production").Find(&productionState)
	digitalPort := zapsi_database.WorkplacePort{Name: "Production", DevicePortId: devicePortDigital.ID, WorkplaceId: workplace.ID, StateId: productionState.ID}
	db.Create(&digitalPort)
	var devicePortAnalog zapsi_database.DevicePort
	db.Where("name=?", "Amperage").Where("device_id=?", device.ID).Find(&devicePortAnalog)
	var poweroffState zapsi_database.State
	db.Where("name=?", "Poweroff").Find(&poweroffState)
	analogPort := zapsi_database.WorkplacePort{Name: "Amperage", DevicePortId: devicePortAnalog.ID, WorkplaceId: workplace.ID, StateId: poweroffState.ID}
	db.Create(&analogPort)

}

func CheckDevice(device zapsi_database.Device) bool {
	for _, runningDevice := range runningDevices {
		if runningDevice.Name == device.Name {
			return true
		}
	}
	return false
}

func RunDevice(device zapsi_database.Device) {
	LogInfo(device.Name, "Device started running")
	deviceSync.Lock()
	runningDevices = append(runningDevices, device)
	deviceSync.Unlock()
	deviceIsActive := true
	CreateDirectoryIfNotExists(device)
	actualCycle := 0
	totalCycles := 0
	actualState := "poweroff"
	for deviceIsActive {
		start := time.Now()
		if actualCycle >= totalCycles {
			actualCycle, actualState, totalCycles = GenerateNewState()
		}
		switch actualState {
		case "production":
			LogInfo(device.Name, "Production -> "+strconv.Itoa(actualCycle)+" of "+strconv.Itoa(totalCycles))
			GenerateProductionData(device)
		case "downtime":
			LogInfo(device.Name, "Downtime -> "+strconv.Itoa(actualCycle)+" of "+strconv.Itoa(totalCycles))
			GenerateDowntimeData(device)
		case "poweroff":
			LogInfo(device.Name, "Poweroff -> "+strconv.Itoa(actualCycle)+" of "+strconv.Itoa(totalCycles))
		}
		LogInfo(device.Name, "Processing takes "+time.Since(start).String())
		Sleep(device, start)
		deviceIsActive = CheckActive(device)
		actualCycle++
	}
	RemoveDeviceFromRunningDevices(device)
	LogInfo(device.Name, "Device not active, stopped running")

}

func Sleep(device zapsi_database.Device, start time.Time) {
	if time.Since(start) < (downloadInSeconds * time.Second) {
		sleepTime := downloadInSeconds*time.Second - time.Since(start)
		LogInfo(device.Name, "Sleeping for "+sleepTime.String())
		time.Sleep(sleepTime)
	}
}

func CheckActive(device zapsi_database.Device) bool {
	for _, activeDevice := range activeDevices {
		if activeDevice.Name == device.Name {
			LogInfo(device.Name, "Device still active")
			return true
		}
	}
	LogInfo(device.Name, "Device not active")
	return false
}

func RemoveDeviceFromRunningDevices(device zapsi_database.Device) {
	deviceSync.Lock()
	for idx, runningDevice := range runningDevices {
		if device.Name == runningDevice.Name {
			runningDevices = append(runningDevices[0:idx], runningDevices[idx+1:]...)
		}
	}
	deviceSync.Unlock()
}

func UpdateActiveDevices(reference string) {
	connectionString, dialect := zapsi_database.CheckDatabaseType(DatabaseType, DatabaseIpAddress, DatabasePort, DatabaseLogin, DatabaseName, DatabasePassword)
	db, err := gorm.Open(dialect, connectionString)
	if err != nil {
		LogError(reference, "Problem opening "+DatabaseName+" database: "+err.Error())
		activeDevices = nil
		return
	}
	defer db.Close()
	var deviceType zapsi_database.DeviceType
	db.Where("name=?", "Zapsi").Find(&deviceType)
	db.Where("device_type_id=?", deviceType.ID).Where("activated = true").Find(&activeDevices)
	LogDebug("MAIN", "Zapsi device type id is "+strconv.Itoa(int(deviceType.ID)))
}

func CompleteDatabaseCheck() {
	firstRunCheckComplete := false
	for firstRunCheckComplete == false {
		databaseOk := zapsi_database.CheckDatabase(DatabaseType, DatabaseIpAddress, DatabasePort, DatabaseLogin, DatabaseName, DatabasePassword)
		tablesOk, err := zapsi_database.CheckTables(DatabaseType, DatabaseIpAddress, DatabasePort, DatabaseLogin, DatabaseName, DatabasePassword)
		if err != nil {
			LogInfo("MAIN", "Problem creating tables: "+err.Error())
		}
		if databaseOk && tablesOk {
			WriteProgramVersionIntoSettings()
			firstRunCheckComplete = true
		}
	}
}

func WriteProgramVersionIntoSettings() {
	connectionString, dialect := zapsi_database.CheckDatabaseType(DatabaseType, DatabaseIpAddress, DatabasePort, DatabaseLogin, DatabaseName, DatabasePassword)
	db, err := gorm.Open(dialect, connectionString)
	if err != nil {
		LogError("MAIN", "Problem opening "+DatabaseName+" database: "+err.Error())
		return
	}
	defer db.Close()
	var settings zapsi_database.Setting
	db.Where("name=?", programName).Find(&settings)
	settings.Name = programName
	settings.Value = version
	db.Save(&settings)
	LogDebug("MAIN", "Updated version in database for "+programName)
}
