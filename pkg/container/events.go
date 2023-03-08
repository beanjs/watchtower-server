package container

import (
	"context"
	"time"

	"github.com/containrrr/watchtower/pkg/notifications"
	"github.com/containrrr/watchtower/pkg/types"
	dt "github.com/docker/docker/api/types"
	log "github.com/sirupsen/logrus"
)

type Events struct {
	wait   chan struct{}
	stop   chan struct{}
	client Client
}

func NewEvents(client Client) *Events {
	return &Events{
		wait:   make(chan struct{}),
		stop:   make(chan struct{}),
		client: client,
	}
}

func (e Events) run() {
	msg, err := e.client.Events(context.Background(), dt.EventsOptions{})
	// caser := cases.Title(language.English)

	for {
		select {
		case <-e.stop:
			e.wait <- struct{}{}
			return
		case e := <-err:
			log.Error(e)
		case m := <-msg:
			if m.Type != "container" {
				continue
			}

			notifications.LocalLog.Debugf("container event:%s", m.Action)
			con, _ := e.client.GetContainer(types.ContainerID(m.ID))
			span := time.Now().Format("2006-01-02 15:04:05 -0700 MST")

			if m.Action == "start" {
				log.Infof("Starting %s\nImage %s\n%s", con.Name(), con.ImageName(), span)
			}

			if m.Action == "stop" {
				log.Infof("Stopping %s\nImage %s\n%s", con.Name(), con.ImageName(), span)
			}
		}
	}
}

func (e Events) Start() {
	go e.run()
}

func (e Events) Stop() {
	e.stop <- struct{}{}
	<-e.wait
}
