package models

import "time"

type TaskListResponse struct {
	ID          int64      `json:"id"`
	Title       string     `json:"title`
	Description string     `json:"description`
	Status      TaskStatus `json:"status`
	CreatedAt   time.Time  `json:"created_at`
	FormedAt    *time.Time `json:"formed_at`
	CompletedAt *time.Time `json:"completed_at`

	Case *Case `gorm:"foreignKey:CaseID" json:"case`

	Cluster *Cluster `gorm:"foreignKey:ClusterID" json:"cluster`

	User *User `gorm:"foreignKey:UserID" json:"user`
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

type Case struct {
	ID       int64    `json:"id"`
	Title    string   `json:"title"`
	Solution string   `json:"solution"`
	Cluster  *Cluster `json:"cluster`
}

type Cluster struct {
	ID int64 `json:"id"`
	//ClusterIndex int64  `json:"cluster_index"`
	Name      string `json:"name"`
	Frequency int64  `json:"frequency"`
}
