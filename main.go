package main

import (
	"context"
	"time"

	"github.com/rhettcp/go-common/env"
	"github.com/rhettcp/go-common/logging"
	"github.com/rhettcp/go-common/signals"
	"github.com/rhettcp/go-common/status"
)

var (
	log = logging.Log
)

func main() {
	logging.SetLogLevel(env.GetLogLevel())
	log.Info("Starting Skimmer server...")

	can := status.ServeStatusEndpoint(context.Background())

	odtSkimmer := OdtSkimmer{}
	/*gbSkimmer := GBSkimmer{config: GBConfig{
		filters: []string{"tavor 7.62", "tavor 308"},
	}}*/

	skimmerEngine, _ := NewSkimmerEngine(SkimmerConfig{SkimmerInterval: 3 * time.Minute})
	skimmerEngine.AddSkimmer(&odtSkimmer)
	//skimmerEngine.AddSkimmer(&gbSkimmer)

	go skimmerEngine.RunSkimmerEngine()

	log.Info("Skimmer running...")
	err := signals.SignalHandlerLoop(skimmerEngine)
	log.Info("Skimmer closing...")
	can()
	log.Fatal(err)
}
