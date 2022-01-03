package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"sort"
	"strings"
	"sync"
)

// Global variables
var wg sync.WaitGroup
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

// LookupIP is a replacement for default LookupIP so we can use go as dns resolver instead of the OS
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

func getHealthcheck(ch chan<- string) {
	defer wg.Done()
	client := &http.Client{}
	req, _ := http.NewRequest("GET", "https://s3.glbimg.com/healthcheck", nil)
	req.Header.Set("Connection", "close")
	res, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}
	defer res.Body.Close()
	ch <- string(res.StatusCode)
}

// dnsResolver resolves a dns add the result to a map and has concurrency control
func dnsResolver(domain string, ch chan<- string) {
	defer wg.Done()
	iprecords, _ := LookupIP(domain)
	for _, ip := range iprecords {
		ch <- ip.String()
	}
}

func main() {
	var (
		ch = make(chan string)
		result = make(map[string]int)
		setWait bool
	)

	domainPtr := flag.String("domain", "", "`Domain` to be resolved")
	repeatPtr := flag.Int("r", 1, "`Number` of run times")
	multithread := flag.Bool("d", false, "Enables multithread. Default is false.")
	action := flag.String("action", "", "dns or get")

	flag.Parse()
	fmt.Println("domain:", *domainPtr)

	for rep := 1; rep <= *repeatPtr; rep++ {
		printProgressBar(rep, *repeatPtr, "Progress", "Complete", 25, "=")
		if *multithread {
			setWait = true
			wg.Add(1)
			if *action == "dns" {
				go dnsResolver(*domainPtr, ch)
			} else {
				go getHealthcheck(ch)
			}
		} else {
			if *action == "dns" {
				dnsResolver(*domainPtr, ch)
			} else {
				getHealthcheck(ch)
			}
		}
	}

	go func() {
		if setWait == true {
			wg.Wait()
		}
		close(ch)
	}()

	for res := range ch {
		result[res]++
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
