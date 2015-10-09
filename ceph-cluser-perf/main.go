package main

import (
	"flag"
	"fmt"
	"time"
)

var sleepTime int
var debug bool


func init() {
	flag.BoolVar(&debug, "debug", false, "turn on debugging")
	flag.IntVar(&sleepTime, "sleep-time", 10, "Number of seconds between runs")
	flag.Parse()
}


func main() {
	ticker := time.NewTicker(time.Second * time.Duration(sleepTime))

	go func() {
		for t := range ticker.C {
			if debug {
				fmt.Println("DEBUG", time.Now(), " - ", t)
			}
			node.GetCPUPercent()
			node.Get_disk_stats()
			Get_io_status()
		}
	}()
	// run for a year - as collectd will restart it
	time.Sleep(time.Second * 86400 * 365 * 100)
	ticker.Stop()
	fmt.Println("Ticker stopped")
}
