package models

type TaskGetResponse struct {
	ID          int64      `json:"id"`
	Title       string     `json:"title`
	Description string     `json:"description`
	Status      TaskStatus `json:"status`
	CreatedAt   string     `json:"created_at`
	FormedAt    *string    `json:"formed_at`
	CompletedAt *string    `json:"completed_at`

	Case *Case `gorm:"foreignKey:CaseID" json:"case`

	Cluster *Cluster `gorm:"foreignKey:ClusterID" json:"cluster`

	User *User `gorm:"foreignKey:UserID" json:"user`
}
