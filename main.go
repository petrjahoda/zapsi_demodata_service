package main

import (
	"database/sql"
	"github.com/kardianos/service"
	"github.com/petrjahoda/database"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"math/rand"
	"strconv"
	"sync"
	"time"
)

const version = "2020.4.2.17"
const programName = "Zapsi Demodata Service"
const programDescription = "Created demodata life it comes from Zapsi devices"
const downloadInSeconds = 10
const config = "user=postgres password=Zps05..... dbname=version3 host=localhost port=5432 sslmode=disable"
const numberOfDevicesToCreate = 20

var serviceRunning = false

var (
	activeDevices  []database.Device
	runningDevices []database.Device
	deviceSync     sync.Mutex
)

type program struct{}

func (p *program) Start(s service.Service) error {
	logInfo("MAIN", "Starting "+programName+" on "+s.Platform())
	go p.run()
	serviceRunning = true
	return nil
}

func (p *program) run() {
	time.Sleep(2 * time.Second)
	logInfo("MAIN", "Program version "+version+" started")
	for {
		start := time.Now()
		logInfo("MAIN", "Program running")
		writeProgramVersionIntoSettings()
		createDevicesAndWorkplaces("MAIN")
		createTerminals("MAIN")
		createWorkshiftsForWorkplaces("MAIN")
		logInfo("MAIN", "Active devices: "+strconv.Itoa(len(activeDevices))+", running devices: "+strconv.Itoa(len(runningDevices)))
		for _, activeDevice := range activeDevices {
			activeDeviceIsRunning := checkDevice(activeDevice)
			if !activeDeviceIsRunning {
				go runDevice(activeDevice)
			}
		}
		addAdditionalDowntimes()
		addAdditionalProducts()
		addAdditionalOrders()
		addAdditionalUsers()
		go updateDowntimeRecords()
		go updateOrderRecords()
		go updateUserRecords()
		if time.Since(start) < (downloadInSeconds * time.Second) {
			sleeptime := downloadInSeconds*time.Second - time.Since(start)
			logInfo("MAIN", "Sleeping for "+sleeptime.String())
			time.Sleep(sleeptime)
		}
	}

}

func updateUserRecords() {
	for serviceRunning {
		db, err := gorm.Open(postgres.Open(config), &gorm.Config{})
		sqlDB, _ := db.DB()

		if err != nil {
			logError("MAIN", "Problem opening  database: "+err.Error())
			return
		}
		var userRecords []database.UserRecord
		db.Where("user_id = ?", 1).Find(&userRecords)
		rand.Seed(time.Now().UnixNano())
		min := 2
		max := 4
		random := rand.Intn(max-min+1) + min
		for _, record := range userRecords {
			record.UserID = random
			db.Save(&record)
		}
		sqlDB.Close()
		time.Sleep(10 * time.Second)
	}
}

func addAdditionalUsers() {
	db, err := gorm.Open(postgres.Open(config), &gorm.Config{})
	sqlDB, _ := db.DB()
	defer sqlDB.Close()
	if err != nil {
		logError("MAIN", "Problem opening  database: "+err.Error())
		return
	}
	var users []database.User
	db.Find(&users)
	if len(users) == 1 {
		password := hashPasswordFromString([]byte("pj"))
		var jahoda database.User
		jahoda.FirstName = "Petr"
		jahoda.SecondName = "Jahoda"
		jahoda.Email = "petr@jahoda.cz"
		jahoda.Password = password
		jahoda.UserRoleID = 3
		jahoda.UserTypeID = 1
		jahoda.Locale = "EnUS"
		db.Save(&jahoda)
		password = hashPasswordFromString([]byte("rm"))
		var malina database.User
		malina.FirstName = "Radek"
		malina.SecondName = "Malina"
		malina.Email = "radek@malina.cz"
		malina.Password = password
		malina.UserRoleID = 3
		malina.UserTypeID = 1
		malina.Locale = "DeDE"
		db.Save(&malina)
	}
}

func hashPasswordFromString(pwd []byte) string {
	logInfo("MAIN", "Hashing password")
	timer := time.Now()
	hash, err := bcrypt.GenerateFromPassword(pwd, bcrypt.MinCost)
	if err != nil {
		logError("MAIN", "Cannot hash password: "+err.Error())
		return ""
	}
	logInfo("MAIN", "Password hashed in  "+time.Since(timer).String())
	return string(hash)
}

func updateOrderRecords() {
	for serviceRunning {
		db, err := gorm.Open(postgres.Open(config), &gorm.Config{})
		sqlDB, _ := db.DB()

		if err != nil {
			logError("MAIN", "Problem opening  database: "+err.Error())
			return
		}
		var orderRecords []database.OrderRecord
		db.Where("date_time_end is null").Where("order_id = ?", 1).Find(&orderRecords)
		rand.Seed(time.Now().UnixNano())
		min := 2
		max := 4
		random := rand.Intn(max-min+1) + min
		for _, record := range orderRecords {
			record.OrderID = random
			db.Save(&record)
		}
		sqlDB.Close()
		time.Sleep(10 * time.Second)
	}
}

func addAdditionalOrders() {
	db, err := gorm.Open(postgres.Open(config), &gorm.Config{})
	sqlDB, _ := db.DB()
	defer sqlDB.Close()
	if err != nil {
		logError("MAIN", "Problem opening  database: "+err.Error())
		return
	}
	var orders []database.Order
	db.Find(&orders)
	if len(orders) == 1 {
		var chairs database.Order
		chairs.Name = "Chairs"
		chairs.Barcode = 0
		chairs.ProductID = sql.NullInt32{
			Int32: 2,
			Valid: true,
		}
		chairs.Cavity = 0
		chairs.CountRequest = 99999
		db.Save(&chairs)
		var tables database.Order
		tables.Name = "Tables"
		tables.Barcode = 0
		tables.ProductID = sql.NullInt32{
			Int32: 3,
			Valid: true,
		}
		tables.Cavity = 0
		tables.CountRequest = 99999
		db.Save(&tables)
		var ceramics database.Order
		ceramics.Name = "Ceramics"
		ceramics.Barcode = 0
		ceramics.ProductID = sql.NullInt32{
			Int32: 4,
			Valid: true,
		}
		ceramics.Cavity = 0
		ceramics.CountRequest = 99999
		db.Save(&ceramics)
	}
}

func addAdditionalProducts() {
	db, err := gorm.Open(postgres.Open(config), &gorm.Config{})
	sqlDB, _ := db.DB()
	defer sqlDB.Close()
	if err != nil {
		logError("MAIN", "Problem opening  database: "+err.Error())
		return
	}
	var products []database.Product
	db.Find(&products)
	if len(products) == 1 {
		var chair database.Product
		chair.Name = "chair"
		chair.Barcode = 0
		chair.CycleTime = 0
		chair.DownTimeDuration = 0
		db.Save(&chair)
		var table database.Product
		table.Name = "table"
		table.Barcode = 0
		table.CycleTime = 0
		table.DownTimeDuration = 0
		db.Save(&table)
		var vase database.Product
		vase.Name = "vase"
		vase.Barcode = 0
		vase.CycleTime = 0
		vase.DownTimeDuration = 0
		db.Save(&vase)
	}
}

func addAdditionalDowntimes() {
	db, err := gorm.Open(postgres.Open(config), &gorm.Config{})
	sqlDB, _ := db.DB()
	defer sqlDB.Close()
	if err != nil {
		logError("MAIN", "Problem opening  database: "+err.Error())
		return
	}
	var downtimeRecords []database.DowntimeRecord
	db.Find(&downtimeRecords)
	if len(downtimeRecords) == 1 {
		var smoking database.Downtime
		smoking.Name = "Smoking"
		smoking.DowntimeTypeID = 1
		db.Save(&smoking)
		var cleaning database.Downtime
		cleaning.Name = "Cleaning"
		cleaning.DowntimeTypeID = 1
		db.Save(&cleaning)
		var toolChange database.Downtime
		toolChange.Name = "Change of tools"
		toolChange.DowntimeTypeID = 1
		db.Save(&toolChange)
	}
}

func updateDowntimeRecords() {
	for serviceRunning {
		db, err := gorm.Open(postgres.Open(config), &gorm.Config{})
		sqlDB, _ := db.DB()

		if err != nil {
			logError("MAIN", "Problem opening  database: "+err.Error())
			return
		}
		var downtimeRecords []database.DowntimeRecord
		db.Where("date_time_end is null").Where("downtime_id = ?", 1).Find(&downtimeRecords)
		rand.Seed(time.Now().UnixNano())
		min := 2
		max := 4
		random := rand.Intn(max-min+1) + min
		for _, record := range downtimeRecords {
			record.DowntimeID = random
			db.Save(&record)
		}
		sqlDB.Close()
		time.Sleep(10 * time.Second)
	}
}

func createWorkshiftsForWorkplaces(reference string) {
	db, err := gorm.Open(postgres.Open(config), &gorm.Config{})
	sqlDB, _ := db.DB()
	defer sqlDB.Close()
	if err != nil {
		logError(reference, "Problem opening database: "+err.Error())
		return
	}
	var workplaceWorkShifts []database.WorkplaceWorkshift
	db.Find(&workplaceWorkShifts)
	if len(workplaceWorkShifts) == 0 {
		logInfo("MAIN", "Creating workplace workshifts")
		for i := 1; i <= numberOfDevicesToCreate; i++ {
			createWorkshiftsForWorkplace(reference, i)
		}
	}
}

func createWorkshiftsForWorkplace(reference string, workplaceId int) {
	db, err := gorm.Open(postgres.Open(config), &gorm.Config{})
	sqlDB, _ := db.DB()
	defer sqlDB.Close()
	if err != nil {
		logError(reference, "Problem opening database: "+err.Error())
		activeDevices = nil
		return
	}
	for i := 1; i <= 3; i++ {
		newWorkplaceWorkshift := database.WorkplaceWorkshift{
			WorkplaceID: workplaceId,
			WorkshiftID: i,
		}
		db.Create(&newWorkplaceWorkshift)
	}
}

func createTerminals(reference string) {
	db, err := gorm.Open(postgres.Open(config), &gorm.Config{})
	sqlDB, _ := db.DB()
	defer sqlDB.Close()
	if err != nil {
		logError(reference, "Problem opening database: "+err.Error())
		return
	}
	var deviceType database.DeviceType
	db.Where("name=?", "Zapsi Touch").Find(&deviceType)
	var activeTerminals []database.Device
	db.Where("device_type_id=?", deviceType.ID).Where("activated = ?", "1").Find(&activeTerminals)
	if len(activeTerminals) == 0 {
		logInfo("MAIN", "Creating terminals")
		for i := 1; i <= numberOfDevicesToCreate; i++ {
			addTerminalWithWorkplace("MAIN", "CNC Terminal "+strconv.Itoa(i), "192.168.1."+strconv.Itoa(i), i)
		}
	}
}

func addTerminalWithWorkplace(reference string, workplaceName string, ipAddress string, i int) {
	db, err := gorm.Open(postgres.Open(config), &gorm.Config{})
	sqlDB, _ := db.DB()
	defer sqlDB.Close()
	if err != nil {
		logError(reference, "Problem opening database: "+err.Error())
		return
	}
	var deviceType database.DeviceType
	db.Where("name=?", "Zapsi Touch").Find(&deviceType)
	newTerminal := database.Device{Name: workplaceName, DeviceTypeID: int(deviceType.ID), IpAddress: ipAddress, TypeName: "Zapsi Touch", Activated: true}
	db.Create(&newTerminal)
	newRecord := database.DeviceWorkplaceRecord{
		DeviceID:    int(newTerminal.ID),
		WorkplaceID: i,
	}
	db.Create(&newRecord)

}
func (p *program) Stop(s service.Service) error {
	serviceRunning = false
	for len(runningDevices) != 0 {
		logInfo("MAIN", "Stopping, still running devices: "+strconv.Itoa(len(runningDevices)))
		time.Sleep(1 * time.Second)
	}
	logInfo("MAIN", "Stopped on platform "+s.Platform())
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
		logError("MAIN", err.Error())
	}
	err = s.Run()
	if err != nil {
		logError("MAIN", "Problem starting "+serviceConfig.Name)
	}
}

func createDevicesAndWorkplaces(reference string) {
	db, err := gorm.Open(postgres.Open(config), &gorm.Config{})
	sqlDB, _ := db.DB()
	defer sqlDB.Close()
	if err != nil {
		logError(reference, "Problem opening database: "+err.Error())
		activeDevices = nil
		return
	}
	var deviceType database.DeviceType
	db.Where("name=?", "Zapsi").Find(&deviceType)
	db.Where("device_type_id=?", deviceType.ID).Where("activated = ?", "1").Find(&activeDevices)
	defer sqlDB.Close()
	if len(activeDevices) == 0 {
		logInfo("MAIN", "Creating devices")
		for i := 0; i < numberOfDevicesToCreate; i++ {
			addDeviceWithWorkplace("MAIN", "CNC "+strconv.Itoa(i), "192.168.0."+strconv.Itoa(i))
		}
	}
}

func addDeviceWithWorkplace(reference string, workplaceName string, ipAddress string) {
	db, err := gorm.Open(postgres.Open(config), &gorm.Config{})
	sqlDB, _ := db.DB()
	defer sqlDB.Close()
	if err != nil {
		logError(reference, "Problem opening database: "+err.Error())
		return
	}
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

func checkDevice(device database.Device) bool {
	for _, runningDevice := range runningDevices {
		if runningDevice.Name == device.Name {
			return true
		}
	}
	return false
}

func runDevice(device database.Device) {
	logInfo(device.Name, "Device started running")
	deviceSync.Lock()
	runningDevices = append(runningDevices, device)
	deviceSync.Unlock()
	deviceIsActive := true
	actualCycle := 0
	totalCycles := 0
	actualState := "poweroff"
	for deviceIsActive && serviceRunning {
		start := time.Now()
		if actualCycle >= totalCycles {
			actualCycle, actualState, totalCycles = generateNewState()
		}
		switch actualState {
		case "production":
			logInfo(device.Name, "Production -> "+strconv.Itoa(actualCycle)+" of "+strconv.Itoa(totalCycles))
			generateProductionData(device)
		case "downtime":
			logInfo(device.Name, "Downtime -> "+strconv.Itoa(actualCycle)+" of "+strconv.Itoa(totalCycles))
			generateDowntimeData(device)
		case "poweroff":
			logInfo(device.Name, "Poweroff -> "+strconv.Itoa(actualCycle)+" of "+strconv.Itoa(totalCycles))
		}
		logInfo(device.Name, "Processing takes "+time.Since(start).String())
		sleep(device, start)
		deviceIsActive = checkActive(device)
		actualCycle++
	}
	removeDeviceFromRunningDevices(device)
	logInfo(device.Name, "Device not active, stopped running")

}

func sleep(device database.Device, start time.Time) {
	if time.Since(start) < (downloadInSeconds * time.Second) {
		sleepTime := downloadInSeconds*time.Second - time.Since(start)
		logInfo(device.Name, "Sleeping for "+sleepTime.String())
		time.Sleep(sleepTime)
	}
}

func checkActive(device database.Device) bool {
	for _, activeDevice := range activeDevices {
		if activeDevice.Name == device.Name {
			logInfo(device.Name, "Device still active")
			return true
		}
	}
	logInfo(device.Name, "Device not active")
	return false
}

func removeDeviceFromRunningDevices(device database.Device) {
	deviceSync.Lock()
	for idx, runningDevice := range runningDevices {
		if device.Name == runningDevice.Name {
			runningDevices = append(runningDevices[0:idx], runningDevices[idx+1:]...)
		}
	}
	deviceSync.Unlock()
}

func writeProgramVersionIntoSettings() {
	db, err := gorm.Open(postgres.Open(config), &gorm.Config{})
	sqlDB, _ := db.DB()
	defer sqlDB.Close()
	if err != nil {
		logError("MAIN", "Problem opening  database: "+err.Error())
		return
	}
	var settings database.Setting
	db.Where("name=?", programName).Find(&settings)
	settings.Name = programName
	settings.Value = version
	db.Save(&settings)
	logInfo("MAIN", "Updated version in database for "+programName)
}
