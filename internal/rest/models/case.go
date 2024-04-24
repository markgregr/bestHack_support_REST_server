package models

type CaseTask struct {
	ID       int64  `json:"id"`
	Title    string `json:"title"`
	Solution string `json:"solution"`
}

type Case struct {
	ID       int64    `json:"id"`
	Title    string   `json:"title"`
	Solution string   `json:"solution"`
	Cluster  *Cluster `json:"cluster"`
}
