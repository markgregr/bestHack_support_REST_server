package models

type Case struct {
	ID       int64  `json:"id"`
	Title    string `json:"title"`
	Solution string `json:"solution"`
}
