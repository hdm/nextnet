package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"runtime"
	"sync"

	"golang.org/x/time/rate"
)

type scanResult struct {
	Host  string            `json:"host"`
	Port  string            `json:"port,omitempty"`
	Proto string            `json:"proto,omitempty"`
	Probe string            `json:"probe,omitempty"`
	Name  string            `json:"name,omitempty"`
	Nets  []string          `json:"nets,omitempty"`
	Info  map[string]string `json:"info"`
}

var (
	limiter *rate.Limiter
	ppsrate *int
	probes  []Prober
	wgIn    sync.WaitGroup
	wgOut   sync.WaitGroup
)

func usage() {
	fmt.Println("Usage: " + os.Args[0] + " [cidr] ... [cidr]")
	fmt.Println("")
	fmt.Println("Probes a list of networks for potential pivot points.")
	fmt.Println("")
	fmt.Println("Options:")
	flag.PrintDefaults()
}

func outputWriter(out <-chan scanResult) {

	for found := range out {
		j, err := json.Marshal(found)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error marshaling result: '%v' : %s\n", found, err)
			continue
		}
		if _, err := os.Stdout.Write(j); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing %q to stdout: %s", j, err)
			continue
		}

		if _, err := os.Stdout.Write([]byte("\n")); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing newline to stdout: %s", err)
			continue
		}
	}
	wgOut.Done()
}

func initializeProbes(out chan<- scanResult) {
	for _, probe := range probes {
		probe.Initialize()
		probe.SetOutput(out)
		probe.SetLimiter(limiter)
	}
}

func waitProbes() {
	for _, probe := range probes {
		probe.Wait()
	}
}

func processAddress(in <-chan string) {
	for addr := range in {
		for _, probe := range probes {
			probe.AddTarget(addr)
		}
	}

	for _, probe := range probes {
		probe.CloseInput()
	}

	wgIn.Done()
}

func main() {

	runtime.GOMAXPROCS(runtime.NumCPU())

	flag.Usage = func() { usage() }
	version := flag.Bool("version", false, "Show the application version")
	ppsrate = flag.Int("rate", 1000, "Set the maximum packets per second rate")

	flag.Parse()

	if *version {
		PrintVersion("nextnet")
		os.Exit(0)
	}

	limiter = rate.NewLimiter(rate.Limit(*ppsrate), *ppsrate*3)

	// Input addresses
	addrChan := make(chan string)

	// Output structs
	outputChan := make(chan scanResult)

	// Configure the probes
	initializeProbes(outputChan)

	// Launch a single input address processor
	wgIn.Add(1)
	go processAddress(addrChan)

	// Launch a single output writer
	wgOut.Add(1)
	go outputWriter(outputChan)

	// Parse CIDRs and feed IPs to the input channel
	for _, cidr := range flag.Args() {
		AddressesFromCIDR(cidr, addrChan)
	}

	// Close the cidr input channel
	close(addrChan)

	// Wait for the input feed to complete
	wgIn.Wait()

	// Wait for pending probes
	waitProbes()

	// Close the output handle
	close(outputChan)

	// Wait for the output goroutine
	wgOut.Wait()
}
