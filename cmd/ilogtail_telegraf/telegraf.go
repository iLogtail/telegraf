package main

import (
	"context"
	"flag"
	"github.com/influxdata/telegraf/agent"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/logger"
	_ "github.com/influxdata/telegraf/plugins/aggregators/all"
	_ "github.com/influxdata/telegraf/plugins/inputs/all"
	_ "github.com/influxdata/telegraf/plugins/outputs/all"
	_ "github.com/influxdata/telegraf/plugins/parsers/all"
	_ "github.com/influxdata/telegraf/plugins/processors/all"
	_ "github.com/influxdata/telegraf/plugins/secretstores/all"
	_ "github.com/influxdata/telegraf/plugins/serializers/all"
	"log"
	"os"
	"sync"
)

var flagConfig = flag.String("config", "telegraf.conf", "")
var flagConfigDir = flag.String("config-directory", "conf.d", "")

func main() {
	flag.Parse()
	// load global config, such as logger
	if err := setupLogger(); err != nil {
		log.Printf("E! [agent] Starting err when read config %s: %v", *flagConfig, err)
		os.Exit(1)
		return
	}
	files, err := config.WalkDirectory(*flagConfigDir)
	if err != nil {
		log.Printf("E! [agent] Starting err when read dir config %s: %v", *flagConfigDir, err)
		os.Exit(1)
		return
	}
	ctx, cancelFunc := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	for _, file := range files {
		c := config.NewConfig()
		if err := c.LoadConfig(file); err != nil {
			log.Printf("E! [agent] error when load conifg %s: %v", file, err)
			continue
		}
		na := agent.NewAgent(c)
		go func() {
			wg.Add(1)
			defer wg.Done()
			err := na.Run(ctx)
			if err != nil {
				log.Printf("E! [agent] error when start conifg %s: %v", file, err)
			}
		}()
	}
	<-SetupSignalHandler()
	log.Print("I! [agent] stopping")
	cancelFunc()
	wg.Wait()
	log.Print("I! [agent] stopped")
}

func setupLogger() error {
	c := config.NewConfig()
	if err := c.LoadConfig(*flagConfig); err != nil {
		return err
	}
	logConfig := logger.LogConfig{
		Debug:               c.Agent.Debug,
		Quiet:               c.Agent.Quiet,
		LogTarget:           c.Agent.LogTarget,
		Logfile:             c.Agent.Logfile,
		RotationInterval:    c.Agent.LogfileRotationInterval,
		RotationMaxSize:     c.Agent.LogfileRotationMaxSize,
		RotationMaxArchives: c.Agent.LogfileRotationMaxArchives,
		LogWithTimezone:     c.Agent.LogWithTimezone,
	}
	if err := logger.SetupLogging(logConfig); err != nil {
		return err
	}
	log.Printf("I! Starting Telegraf %s%s brought to you by InfluxData the makers of InfluxDB", internal.Version, internal.Customized)
	return nil
}
