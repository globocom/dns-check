package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
)

//Global varialbes
var result = make(map[string]int)
var wg sync.WaitGroup
var mutex = &sync.RWMutex{}
var DefaultResolver = net.Resolver{PreferGo: true}

// printProgressBar prints a friendly progress bar so the user can be sure if the program is running or waiting
func printProgressBar(iteration, total int, prefix, suffix string, length int, fill string) {
	percent := float64(iteration) / float64(total) * 100
	filledLength := int(length * iteration / total)
	end := ">"

	if iteration == total {
		end = "="
	}
	bar := strings.Repeat(fill, filledLength) + end + strings.Repeat("-", (length-filledLength))
	fmt.Printf("\r%s [%s] %.0f%% %s", prefix, bar, percent, suffix)
	if iteration == total {
		fmt.Println()
	}
}

//LookupIP is a replacemente for default LookupIP so we can use go as dns resolver instead of the OS
func LookupIP(host string) ([]net.IP, error) {
	addrs, err := DefaultResolver.LookupIPAddr(context.Background(), host)
	if err != nil {
		return nil, err
	}
	ips := make([]net.IP, len(addrs))
	for i, ia := range addrs {
		ips[i] = ia.IP
	}
	return ips, nil
}

func getHealthcheck() {
	//defer wg.Done()
	resp, err := http.Get("https://s3.glbimg.com/healthcheck")
	statusCode := strconv.Itoa(resp.StatusCode)
	mutex.Lock()
	result[statusCode]++
	mutex.Unlock()

	if err != nil {
		log.Fatalln(err)
	}

}

// dnsResolver resolves a dns add the result to a map and has concurrency control
func dnsResolver(domain string) {
	//defer wg.Done()
	iprecords, _ := LookupIP(domain)
	for _, ip := range iprecords {
		mutex.Lock()
		result[ip.String()]++
		mutex.Unlock()
	}
}
func main() {
	domainPtr := flag.String("domain", "", "`Domain` to be resolved")
	repeatPtr := flag.Int("r", 1, "`Number` of run times")
	multithread := flag.Bool("d", false, "Enables multithread. Default is false.")
	action := flag.String("action", "", "dns or get")

	var setWait bool
	flag.Parse()
	fmt.Println("domain:", *domainPtr)
	for rep := 1; rep <= *repeatPtr; rep++ {
		printProgressBar(rep, *repeatPtr, "Progress", "Complete", 25, "=")
		if *multithread {
			setWait = true
			//go func() {
			defer wg.Done()
			wg.Add(1)
			if *action == "dns" {
				go dnsResolver(*domainPtr)
			} else {
				go getHealthcheck()
			}
		} else {
			if *action == "dns" {
				dnsResolver(*domainPtr)
			} else {
				getHealthcheck()
			}
		}
	}
	if setWait == true {
		wg.Wait()
	}
	allkeys := make([]string, 0, len(result))
	for key := range result {
		allkeys = append(allkeys, key)
	}
	sort.Strings(allkeys)

	for _, key := range allkeys {
		fmt.Println(key, "=", result[key])
	}
}
