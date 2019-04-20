package main

import (
	"flag"

	"github.com/q3k/bugless/lib/micro"

	log "github.com/inconshreveable/log15"
)

func main() {
	flag.Parse()
	m := micro.New()
	l := log.New()

	if err := m.Listen(); err != nil {
		l.Crit("could not listen", "err", err)
		return
	}

	if err := m.Serve(); err != nil {
		l.Crit("could not serve", "err", err)
		return
	}

	<-m.Done()
}
