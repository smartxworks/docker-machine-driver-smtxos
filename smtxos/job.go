package smtxos

import (
	"fmt"
)

const (
	JobStatePending   = "pending"
	JobStateProcesing = "processing"
	JobStateDone      = "done"
	JobStateFailed    = "failed"
)

type Job struct {
	JobID       string      `json:"job_id"`
	Description string      `json:"description"`
	State       string      `json:"state"`
	Resources   interface{} `json:"resources"`
}

func (c *client) GetJob(id string) (*Job, error) {
	var data *struct {
		Job *Job `json:"job"`
	}
	if err := c.do("GET", fmt.Sprintf("/v2/jobs/%s", id), nil, nil, &data, true); err != nil {
		return nil, err
	}
	return data.Job, nil
}
