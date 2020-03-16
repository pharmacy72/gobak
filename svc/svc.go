package svc

import (
	"github.com/kardianos/service"
	"github.com/pharmacy72/gobak/config"
	"github.com/pharmacy72/gobak/snap"
	"go.uber.org/zap"
)

type program struct {
	exit        chan struct{}
	internalRun func() error
	log         *zap.Logger
}

func (p *program) Start(s service.Service) error {
	snap.Incr(config.Current().NameBase, "counters", snap.CounterStart, 1)
	if service.Interactive() {
		p.log.Info("gobak is running in terminal.")
	} else {
		p.log.Info("gobak is running under service manager.")
	}
	p.exit = make(chan struct{})
	go p.Run()
	return nil
}

func (p *program) Run() {

	if err := p.internalRun(); err != nil {
		p.log.Error(err.Error())
	}
}
func (p *program) Stop(s service.Service) error {
	snap.Incr(config.Current().NameBase, "counters", snap.CounterStop, 1)
	p.log.Info("gobak service is stopping!")
	close(p.exit)
	return nil
}

//New create instance of *service.Service using config,
//internalrun it a function which will be runned
func New(config *service.Config, internalRun func() error) (service.Service, error) {
	prg := &program{
		internalRun: internalRun,
	}
	serviceInctance, err := service.New(prg, config)
	if err != nil {
		return nil, err
	}
	return serviceInctance, nil
}
