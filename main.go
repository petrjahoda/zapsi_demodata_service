package main

import (
	"github.com/jinzhu/gorm"
	"strconv"
	"sync"
	"time"
)

const version = "2019.4.3.30"
const deleteLogsAfter = 240 * time.Hour
const downloadInSeconds = 10

var (
	activeDevices  []Device
	runningDevices []Device
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
		CheckDatabase()
		CheckTables()
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
	connectionString, dialect := CheckDatabaseType()
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
	connectionString, dialect := CheckDatabaseType()
	db, err := gorm.Open(dialect, connectionString)
	if err != nil {
		LogError(reference, "Problem opening "+DatabaseName+" database: "+err.Error())
		return
	}
	defer db.Close()
	var deviceType DeviceType
	db.Where("name=?", "Zapsi").Find(&deviceType)
	newDevice := Device{Name: workplaceName, DeviceType: deviceType.ID, IpAddress: ipAddress, TypeName: "Zapsi", Activated: true}
	db.Create(&newDevice)
	var device Device
	db.Where("name=?", workplaceName).Find(&device)
	deviceDigitalPort := DevicePort{Name: "Production", Unit: "ks", PortNumber: 1, DevicePortTypeId: 1, DeviceId: device.ID}
	deviceAnalogPort := DevicePort{Name: "Amperage", Unit: "A", PortNumber: 3, DevicePortTypeId: 2, DeviceId: device.ID}
	db.Create(&deviceDigitalPort)
	db.Create(&deviceAnalogPort)
	var section WorkplaceSection
	db.Where("name=?", "Machines").Find(&section)
	var state State
	db.Where("name=?", "Offline").Find(&state)
	var mode WorkplaceMode
	db.Where("name=?", "Production").Find(&mode)
	newWorkplace := Workplace{Name: workplaceName, Code: workplaceName, WorkplaceSectionId: section.ID, ActualStateId: state.ID, ActualWorkplaceModeId: mode.ID}
	db.Create(&newWorkplace)
	var workplace Workplace
	db.Where("name=?", workplaceName).Find(&workplace)
	var devicePortDigital DevicePort
	db.Where("name=?", "Production").Where("device_id=?", device.ID).Find(&devicePortDigital)
	var productionState State
	db.Where("name=?", "Production").Find(&productionState)
	digitalPort := WorkplacePort{Name: "Production", DevicePortId: devicePortDigital.ID, WorkplaceId: workplace.ID, StateId: productionState.ID}
	db.Create(&digitalPort)
	var devicePortAnalog DevicePort
	db.Where("name=?", "Amperage").Where("device_id=?", device.ID).Find(&devicePortAnalog)
	var offlineState State
	db.Where("name=?", "Offline").Find(&offlineState)
	analogPort := WorkplacePort{Name: "Amperage", DevicePortId: devicePortAnalog.ID, WorkplaceId: workplace.ID, StateId: offlineState.ID}
	db.Create(&analogPort)

}

func CheckDevice(device Device) bool {
	for _, runningDevice := range runningDevices {
		if runningDevice.Name == device.Name {
			return true
		}
	}
	return false
}

func RunDevice(device Device) {
	LogInfo(device.Name, "Device started running")
	deviceSync.Lock()
	runningDevices = append(runningDevices, device)
	deviceSync.Unlock()
	deviceIsActive := true
	device.CreateDirectoryIfNotExists()
	actualCycle := 0
	totalCycles := 0
	actualState := "offline"
	for deviceIsActive {
		start := time.Now()
		if actualCycle >= totalCycles {
			actualCycle, actualState, totalCycles = device.GenerateNewState()
		}
		switch actualState {
		case "production":
			LogInfo(device.Name, "Production -> "+strconv.Itoa(actualCycle)+" of "+strconv.Itoa(totalCycles))
			device.GenerateProductionData()
		case "downtime":
			LogInfo(device.Name, "Downtime -> "+strconv.Itoa(actualCycle)+" of "+strconv.Itoa(totalCycles))
			device.GenerateDowntimeData()
		case "offline":
			LogInfo(device.Name, "Offline -> "+strconv.Itoa(actualCycle)+" of "+strconv.Itoa(totalCycles))
		}
		LogInfo(device.Name, "Processing takes "+time.Since(start).String())
		device.Sleep(start)
		deviceIsActive = CheckActive(device)
		actualCycle++
	}
	RemoveDeviceFromRunningDevices(device)
	LogInfo(device.Name, "Device not active, stopped running")

}

func CheckActive(device Device) bool {
	for _, activeDevice := range activeDevices {
		if activeDevice.Name == device.Name {
			LogInfo(device.Name, "Device still active")
			return true
		}
	}
	LogInfo(device.Name, "Device not active")
	return false
}

func RemoveDeviceFromRunningDevices(device Device) {
	for idx, runningDevice := range runningDevices {
		if device.Name == runningDevice.Name {
			deviceSync.Lock()
			runningDevices = append(runningDevices[0:idx], runningDevices[idx+1:]...)
			deviceSync.Unlock()
		}
	}
}

func UpdateActiveDevices(reference string) {
	connectionString, dialect := CheckDatabaseType()
	db, err := gorm.Open(dialect, connectionString)
	if err != nil {
		LogError(reference, "Problem opening "+DatabaseName+" database: "+err.Error())
		activeDevices = nil
		return
	}
	defer db.Close()
	var deviceType DeviceType
	db.Where("name=?", "Zapsi").Find(&deviceType)
	db.Where("device_type=?", deviceType.ID).Where("activated = true").Find(&activeDevices)
	LogDebug("MAIN", "Zapsi device type id is "+strconv.Itoa(int(deviceType.ID)))
}
