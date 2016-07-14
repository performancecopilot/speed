package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/performancecopilot/speed"
)

var metric *speed.PCPSingletonMetric

func main() {

	var err error
	metric, err = speed.NewPCPSingletonMetric(
		0,
		"http.requests",
		speed.Int32Type,
		speed.CounterSemantics,
		speed.OneUnit,
		"Number of Requests",
		"Counter that increments every request",
	)
	if err != nil {
		panic(err)
	}

	writer, err := speed.NewPCPWriter("example", speed.ProcessFlag)
	if err != nil {
		panic(err)
	}

	err = writer.Register(metric)
	if err != nil {
		panic(err)
	}

	err = writer.Start()
	if err != nil {
		panic(err)
	}
	defer writer.Stop()

	http.HandleFunc("/increment", handleIncrement)
	go http.ListenAndServe(":8080", nil)

	fmt.Println("To stop the server press enter")
	os.Stdin.Read(make([]byte, 1))
	os.Exit(0)
}

func handleIncrement(w http.ResponseWriter, r *http.Request) {
	v := metric.Val().(int32)
	v++
	metric.Set(v)
	fmt.Fprintf(w, "incremented\n")
}
