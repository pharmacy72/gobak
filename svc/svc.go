package svc

import (
	"log"

	"github.com/kardianos/service"
	"github.com/pharmacy72/gobak/config"
	"github.com/pharmacy72/gobak/snap"
)

type program struct {
	exit        chan struct{}
	internalrun func() error
}

func (p *program) Start(s service.Service) error {
	snap.Incr(config.Current().NameBase, "counters", snap.CounterStart, 1)
	if service.Interactive() {
		log.Println("Gobak is running in terminal.")
	} else {
		log.Println("Gobak is running under service manager.")
	}
	p.exit = make(chan struct{})
	go p.Run()
	return nil
}

func (p *program) Run() {

	if e := p.internalrun(); e != nil {
		log.Println(e)
	}
}
func (p *program) Stop(s service.Service) error {
	snap.Incr(config.Current().NameBase, "counters", snap.CounterStop, 1)
	log.Println("Gobak service is stopping!")
	close(p.exit)
	return nil
}

//New create instance of *service.Service using config,
//internalrun it a function which will be runned
func New(config *service.Config, internalrun func() error) (service.Service, error) {
	prg := &program{}
	prg.internalrun = internalrun
	serviceInctance, err := service.New(prg, config)
	if err != nil {
		return nil, err
	}
	return serviceInctance, nil
}
