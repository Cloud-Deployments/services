package runner

import (
	"log"
)

type Pool struct {
	OrganizationId string
	Manager        *Manager
	runners        map[*Runner]bool
	register       chan *Runner
	unregister     chan *Runner
}

func NewPool(organizationId string, manager *Manager) *Pool {
	return &Pool{
		OrganizationId: organizationId,
		Manager:        manager,
		runners:        make(map[*Runner]bool),
		register:       make(chan *Runner),
		unregister:     make(chan *Runner),
	}
}

func (p *Pool) AddRunner(runner *Runner) {
	p.runners[runner] = true
}

func (p *Pool) RemoveRunner(runner *Runner) {
	if _, ok := p.runners[runner]; ok {
		delete(p.runners, runner)
		close(runner.Send)
	}
}

func (p *Pool) Run() {
	for {
		select {
		case client := <-p.register:
			p.AddRunner(client)
			log.Println("Runner connected:", client.Addr.String())
		case client := <-p.unregister:
			runnerId := client.Id
			clientSecret := client.Token
			organizationId := p.OrganizationId

			go func(runnerId, runnerSecret, organizationId string) {
				err := p.Manager.ApiClient.Logout(runnerId, runnerSecret, organizationId)
				if err != nil {
					log.Println("Error logging out runner:", err)
				}
			}(runnerId, clientSecret, organizationId)

			log.Println("Runner disconnected")
			p.RemoveRunner(client)
		}
	}
}
