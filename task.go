package updater

type TaskStatus int

const (
	TaskStatusCreated TaskStatus = iota
	TaskStatusRunning
	TaskStatusCompleted
	TaskStatusFailed
)

type Task interface {
	GetTaskID() string
	GetType() string
	GetStatus() TaskStatus
	GetContent() []byte
	SetStatus(status TaskStatus)
	Run() error
	Stop() error
}
