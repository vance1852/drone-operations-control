package console

import "time"

type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	RealName string `json:"realName"`
	Phone    string `json:"phone"`
	Role     int    `json:"role"`
	Status   int    `json:"status"`
}

type Drone struct {
	ID              string  `json:"id"`
	Name            string  `json:"name"`
	ModelClass      int     `json:"modelClass"`
	CommissionedOn  *string `json:"commissionedOn"`
	SerialNumber    string  `json:"serialNumber"`
	Endpoint        string  `json:"endpoint"`
	HomeZone        string  `json:"homeZone"`
	OwnerName       string  `json:"ownerName"`
	OwnerPhone      string  `json:"ownerPhone"`
	TelemetryStatus int     `json:"telemetryStatus"`
	Status          int     `json:"status"`
}

type Operator struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Gender     int       `json:"gender"`
	Phone      string    `json:"phone"`
	Skills     string    `json:"skills"`
	Status     int       `json:"status"`
	CreateTime time.Time `json:"createTime"`
}

type CapabilityItem struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Duration    int     `json:"duration"`
	Status      int     `json:"status"`
}

type Command struct {
	ID              string     `json:"id"`
	CommandNo       string     `json:"commandNo"`
	DroneID         string     `json:"droneId"`
	DroneName       string     `json:"droneName"`
	CapabilityID    string     `json:"capabilityId"`
	CapabilityName  string     `json:"capabilityName"`
	OperatorID      *string    `json:"operatorId"`
	OperatorName    string     `json:"operatorName"`
	AppointmentTime *time.Time `json:"appointmentTime"`
	Status          int        `json:"status"`
	Remark          string     `json:"remark"`
	Version         int64      `json:"version"`
}

type TelemetryRecord struct {
	ID                string    `json:"id"`
	DroneID           string    `json:"droneId"`
	DroneName         string    `json:"droneName"`
	BatteryLevel      *float64  `json:"batteryLevel"`
	MotorTemperature  *float64  `json:"motorTemperature"`
	NetworkLatencyMS  *float64  `json:"networkLatencyMs"`
	LocalizationError *float64  `json:"localizationError"`
	JointLoad         *float64  `json:"jointLoad"`
	Remark            string    `json:"remark"`
	RecordTime        time.Time `json:"recordTime"`
}

type Log struct {
	ID         string    `json:"id"`
	Username   string    `json:"username"`
	Operation  string    `json:"operation"`
	Method     string    `json:"method"`
	IP         string    `json:"ip"`
	CreateTime time.Time `json:"createTime"`
}

type Page[T any] struct {
	Records []T `json:"records"`
	Total   int `json:"total"`
	Current int `json:"current"`
	Size    int `json:"size"`
}

type DashboardStats struct {
	DroneCount        int `json:"droneCount"`
	OperatorCount     int `json:"operatorCount"`
	PendingCommands   int `json:"pendingCommands"`
	CompletedCommands int `json:"completedCommands"`
}
