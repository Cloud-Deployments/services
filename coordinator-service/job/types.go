package job

import (
	"crypto"
	"sync"
	"time"
)

type Job struct {
	Id          string
	Connection  SSHConnection
	Command     string
	Status      string
	ClaimedAt   *time.Time
	StartedAt   *time.Time
	FinishedAt  *time.Time
	CreatedAt   *time.Time
	RepeatsLeft int

	mutex *sync.RWMutex
}

type SSHConnection struct {
	Host       string
	Port       int
	User       string
	Password   string
	PrivateKey *crypto.PrivateKey
}
