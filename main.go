package main
 
import (
	"flag"
	"fmt"
	"net"
)
 
func main() {
	domainPtr := flag.String("domain", "","Domain to be resolved")
	repeatPtr := flag.Int("r", 1,"Number of run times")
    result := make(map[string]int)
	flag.Parse()
    fmt.Println("domain:", *domainPtr)
	for rep := 0; rep < *repeatPtr ; rep++ {
		iprecords, _ := net.LookupIP(*domainPtr)
		for _, ip := range iprecords {
				result[ip.String()]++
		    }
	    }
		for key,value := range result{
			fmt.Println(key,"=",value)
	    }
}
