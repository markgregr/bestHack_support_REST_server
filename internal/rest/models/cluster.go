package models

type Cluster struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Frequency int64  `json:"frequency"`
}

type GetClusterResponse struct {
	ID        int64   `json:"id"`
	Name      string  `json:"name"`
	Frequency int64   `json:"frequency"`
	Cases     []*Case `json:"cases"`
	Tasks     []*Task `json:"tasks"`
}
