package models

type Task struct {
	ID          int64      `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Status      TaskStatus `json:"status"`
	CreatedAt   string     `json:"created_at"`
	FormedAt    *string    `json:"formed_at"`
	CompletedAt *string    `json:"completed_at"`
	Case        *Case      `json:"case"`
	Cluster     *Cluster   `json:"cluster"`
	User        *User      `json:"user"`
}

type TaskStatus int32

const (
	TaskStatusOpen TaskStatus = iota
	TaskStatusInProgress
	TaskStatusClosed
)

type User struct {
	ID    int64  `json:"id"`
	Email string `json:"email"`
}
