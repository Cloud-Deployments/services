package job

import (
	"crypto"
	"sync"
	"time"
)

type Job struct {
	Id             string        `json:"id"`
	OrganizationId string        `json:"organization_id"`
	Connection     SSHConnection `json:"connection"`
	Command        string        `json:"command"`
	Status         string        `json:"status"`
	ClaimedAt      *time.Time    `json:"claimed_at"`
	StartedAt      *time.Time    `json:"started_at"`
	FinishedAt     *time.Time    `json:"finished_at"`
	CreatedAt      *time.Time    `json:"created_at"`
	RepeatsLeft    int           `json:"repeats_left"`

	mutex *sync.RWMutex
}

type SSHConnection struct {
	Host       string             `json:"host"`
	Port       int                `json:"port"`
	User       string             `json:"user"`
	Password   string             `json:"password"`
	PrivateKey *crypto.PrivateKey `json:"private_key"`
}
