package slaves

type (
	SlaveMonitoring struct {
		SlavesAvailable []Slave
	}

	Slave struct {
		Address       string
		CurrentStatus *SlaveStatus
	}

	SlaveStatus int
)

const (
	WAITING_WORK SlaveStatus = iota
	WORKING
)

func InitializeNewSlaveMonitoring() *SlaveMonitoring {
	return &SlaveMonitoring{}
}
