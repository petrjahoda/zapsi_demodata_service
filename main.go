package main

import (
	"github.com/kardianos/service"
	"github.com/petrjahoda/database"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"strconv"
	"sync"
	"time"
)

const version = "2020.3.2.22"
const programName = "Zapsi Demodata Service"
const programDescription = "Created demodata life it comes from Zapsi devices"
const downloadInSeconds = 10
const config = "user=postgres password=Zps05..... dbname=version3 host=database port=5432 sslmode=disable"

var serviceRunning = false

var (
	activeDevices  []database.Device
	runningDevices []database.Device
	deviceSync     sync.Mutex
)

type program struct{}

func (p *program) Start(s service.Service) error {
	LogInfo("MAIN", "Starting "+programName+" on "+s.Platform())
	go p.run()
	serviceRunning = true
	return nil
}

func (p *program) run() {
	time.Sleep(2 * time.Second)
	LogInfo("MAIN", "Program version "+version+" started")
	for {
		start := time.Now()
		LogInfo("MAIN", "Program running")
		WriteProgramVersionIntoSettings()
		UpdateActiveDevices("MAIN")
		if len(activeDevices) == 0 {
			CreateDevicesAndWorkplaces()
		}
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
func (p *program) Stop(s service.Service) error {
	serviceRunning = false
	for len(runningDevices) != 0 {
		LogInfo("MAIN", "Stopping, still running devices: "+strconv.Itoa(len(runningDevices)))
		time.Sleep(1 * time.Second)
	}
	LogInfo("MAIN", "Stopped on platform "+s.Platform())
	return nil
}

func main() {
	serviceConfig := &service.Config{
		Name:        programName,
		DisplayName: programName,
		Description: programDescription,
	}
	prg := &program{}
	s, err := service.New(prg, serviceConfig)
	if err != nil {
		LogError("MAIN", err.Error())
	}
	err = s.Run()
	if err != nil {
		LogError("MAIN", "Problem starting "+serviceConfig.Name)
	}
}

func CreateDevicesAndWorkplaces() {
	db, err := gorm.Open(postgres.Open(config), &gorm.Config{})
	if err != nil {
		LogError("MAIN", "Problem opening database: "+err.Error())
		return
	}
	sqlDB, err := db.DB()
	defer sqlDB.Close()
	for i := 0; i < 20; i++ {
		AddTestWorkplace("MAIN", "CNC "+strconv.Itoa(i), "192.168.0."+strconv.Itoa(i))
	}
}

func AddTestWorkplace(reference string, workplaceName string, ipAddress string) {
	db, err := gorm.Open(postgres.Open(config), &gorm.Config{})
	if err != nil {
		LogError(reference, "Problem opening database: "+err.Error())
		return
	}
	sqlDB, err := db.DB()
	defer sqlDB.Close()
	var deviceType database.DeviceType
	db.Where("name=?", "Zapsi").Find(&deviceType)
	newDevice := database.Device{Name: workplaceName, DeviceTypeID: int(deviceType.ID), IpAddress: ipAddress, TypeName: "Zapsi", Activated: true}
	db.Create(&newDevice)
	var device database.Device
	db.Where("name=?", workplaceName).Find(&device)
	deviceDigitalPort := database.DevicePort{Name: "Production", Unit: "ks", PortNumber: 1, DevicePortTypeID: 1, DeviceID: int(device.ID)}
	deviceAnalogPort := database.DevicePort{Name: "Amperage", Unit: "A", PortNumber: 3, DevicePortTypeID: 2, DeviceID: int(device.ID)}
	db.Create(&deviceDigitalPort)
	db.Create(&deviceAnalogPort)
	var section database.WorkplaceSection
	db.Where("name=?", "Machines").Find(&section)
	var state database.State
	db.Where("name=?", "Poweroff").Find(&state)
	var mode database.WorkplaceMode
	db.Where("name=?", "Production").Find(&mode)
	newWorkplace := database.Workplace{Name: workplaceName, Code: workplaceName, WorkplaceSectionID: int(section.ID), WorkplaceModeID: int(mode.ID)}
	db.Create(&newWorkplace)
	var workplace database.Workplace
	db.Where("name=?", workplaceName).Find(&workplace)
	var devicePortDigital database.DevicePort
	db.Where("name=?", "Production").Where("device_id=?", device.ID).Find(&devicePortDigital)
	var productionState database.State
	db.Where("name=?", "Production").Find(&productionState)
	digitalPort := database.WorkplacePort{Name: "Production", DevicePortID: int(devicePortDigital.ID), WorkplaceID: int(workplace.ID), StateID: int(productionState.ID), CounterOK: true}
	db.Create(&digitalPort)
	var devicePortAnalog database.DevicePort
	db.Where("name=?", "Amperage").Where("device_id=?", device.ID).Find(&devicePortAnalog)
	var poweroffState database.State
	db.Where("name=?", "Poweroff").Find(&poweroffState)
	analogPort := database.WorkplacePort{Name: "Amperage", DevicePortID: int(devicePortAnalog.ID), WorkplaceID: int(workplace.ID), StateID: int(poweroffState.ID)}
	db.Create(&analogPort)

}

func CheckDevice(device database.Device) bool {
	for _, runningDevice := range runningDevices {
		if runningDevice.Name == device.Name {
			return true
		}
	}
	return false
}

func RunDevice(device database.Device) {
	LogInfo(device.Name, "Device started running")
	deviceSync.Lock()
	runningDevices = append(runningDevices, device)
	deviceSync.Unlock()
	deviceIsActive := true
	CreateDirectoryIfNotExists(device)
	actualCycle := 0
	totalCycles := 0
	actualState := "poweroff"
	for deviceIsActive && serviceRunning {
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

func Sleep(device database.Device, start time.Time) {
	if time.Since(start) < (downloadInSeconds * time.Second) {
		sleepTime := downloadInSeconds*time.Second - time.Since(start)
		LogInfo(device.Name, "Sleeping for "+sleepTime.String())
		time.Sleep(sleepTime)
	}
}

func CheckActive(device database.Device) bool {
	for _, activeDevice := range activeDevices {
		if activeDevice.Name == device.Name {
			LogInfo(device.Name, "Device still active")
			return true
		}
	}
	LogInfo(device.Name, "Device not active")
	return false
}

func RemoveDeviceFromRunningDevices(device database.Device) {
	deviceSync.Lock()
	for idx, runningDevice := range runningDevices {
		if device.Name == runningDevice.Name {
			runningDevices = append(runningDevices[0:idx], runningDevices[idx+1:]...)
		}
	}
	deviceSync.Unlock()
}

func UpdateActiveDevices(reference string) {
	db, err := gorm.Open(postgres.Open(config), &gorm.Config{})
	if err != nil {
		LogError(reference, "Problem opening database: "+err.Error())
		activeDevices = nil
		return
	}
	sqlDB, err := db.DB()
	defer sqlDB.Close()
	var deviceType database.DeviceType
	db.Where("name=?", "Zapsi").Find(&deviceType)
	db.Where("device_type_id=?", deviceType.ID).Where("activated = ?", "1").Find(&activeDevices)
	LogInfo("MAIN", "Zapsi device type id is "+strconv.Itoa(int(deviceType.ID)))
}

func WriteProgramVersionIntoSettings() {
	db, err := gorm.Open(postgres.Open(config), &gorm.Config{})
	if err != nil {
		LogError("MAIN", "Problem opening  database: "+err.Error())
		return
	}
	sqlDB, err := db.DB()
	defer sqlDB.Close()
	var settings database.Setting
	db.Where("name=?", programName).Find(&settings)
	settings.Name = programName
	settings.Value = version
	db.Save(&settings)
	LogInfo("MAIN", "Updated version in database for "+programName)
}
