package main

import (
	"context"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/CloudDetail/node-agent/metric"
	"github.com/CloudDetail/node-agent/netanaly"
	"github.com/CloudDetail/node-agent/pinger"
)

func main() {
	go func() {
		log.Println(http.ListenAndServe(":6061", nil))
	}()
	ctx, cancel := context.WithCancel(context.Background())
	go netanaly.UpdateNetConnectInfo(ctx)
	go pinger.Ping(ctx)

	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM)
	<-stopChan
	cancel()
	time.Sleep(2 * time.Second)
}
