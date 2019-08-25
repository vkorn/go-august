//go:generate enumer -type=LockStatuses -transform=snake -trimprefix=Lock -json -text -yaml

package august

import "time"

const (
	lockOperator = "superuser"
)

// LockStatuses describes possible lock statuses.
type LockStatuses int

const (
	LockUnknown LockStatuses = iota
	LockUnlocked
	LockLocked
)

// Lock defines August Lock.
type Lock struct {
	ID        string `json:"LockID"`
	Name      string `json:"LockName"`
	UserType  string `json:"UserType"`
	HouseID   string `json:"HouseID"`
	HouseName string `json:"HouseName"`
}

// CanOperate checks whether user can operate this lock.
func (l *Lock) CanOperate() bool {
	return l.UserType == lockOperator
}

// LockStatus describes lock's current status.
type LockStatus struct {
	Status    LockStatuses `json:"status"`
	ChangedAt string       `json:"dateTime"`
	IsChanged bool         `json:"isLockStatusChanged"`
	IsValid   bool         `json:"valid"`
	DoorState string       `json:"doorState"`
}

// SecondsSinceLastChange calculates number of seconds since the last status change.
func (s *LockStatus) SecondsSinceLastChange() float64 {
	t, err := time.Parse(time.RFC3339, s.ChangedAt)
	if err != nil {
		return -1
	}

	return time.Now().UTC().Sub(t).Seconds()
}

// LockUser describes an authorized user.
type LockUser struct {
	UserType  string   `json:"UserType"`
	FirstName string   `json:"FirstName"`
	LastName  string   `json:"LastName"`
	IDs       []string `json:"identifiers"`
}

// CanOperate checks whether user can operate this lock.
func (l *LockUser) CanOperate() bool {
	return l.UserType == lockOperator
}

// LockDetails extends Lock information with more details.
// Note: APIs are returning more extensive info.
type LockDetails struct {
	Lock

	Type       int     `json:"Type"`
	Created    string  `json:"Created"`
	Updated    string  `json:"Updated"`
	Calibrated bool    `json:"Calibrated"`
	Battery    float64 `json:"battery"`

	Status *LockStatus `json:"LockStatus"`

	Users map[string]*LockUser `json:"users"`

	SkuNumber       string `json:"skuNumber"`
	TimeZone        string `json:"timeZone"`
	ZWaveDSK        string `json:"zWaveDSK"`
	EntryCodes      bool   `json:"supportsEntryCodes"`
	SerialNumber    string `json:"SerialNumber"`
	FirmwareVersion string `json:"currentFirmwareVersion"`
	HomeKitEnabled  bool   `json:"homeKitEnabled"`
	ZWaveEnabled    bool   `json:"zWaveEnabled"`
}
