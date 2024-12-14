package queue

import "github.com/Cloud-Deployments/services/coordinator/job"

type Driver interface {
	Enqueue(job *job.Job) error
	Dequeue() (*job.Job, error)
	Close() error
}
