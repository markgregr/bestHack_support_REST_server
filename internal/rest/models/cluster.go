package models

import casesv1 "github.com/markgregr/bestHack_support_protos/gen/go/workflow/cases"

type Cluster struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Frequency int64  `json:"frequency"`
}

type GetClusterResponse struct {
	ID        int64           `json:"id"`
	Name      string          `json:"name"`
	Frequency int64           `json:"frequency"`
	Cases     []*casesv1.Case `json:"cases"`
	Tasks     []*casesv1.Task `json:"tasks"`
}
