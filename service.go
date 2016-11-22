/*! \file crow.go
    \brief Main file for the service.
    Written in GO
    Created 2016-11-14 By Nathan Thomas
    
    The goal is to create a free service for monitoring the things i need/care about
    Terminology
    Crow = Watcher of the application
    Egg = Webiste/resource to monitor
    Squawk = Alert to be sent
    Crew = People to receive the alerts that get sent

_ = "breakpoint"
*/

package main

import (
	"fmt"
	"os"
	"os/signal"
	"time"
	"sync"
	"flag"
	
	"github.com/NathanRThomas/CrowsNest/crow"
)

//-------------------------------------------------------------------------------------------------------------------------//
//----- PRIVATE FUNCTIONS -------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//


//-------------------------------------------------------------------------------------------------------------------------//
//----- MAIN --------------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

func main() {
	
	//handle any passed in flags
	squawkTest := flag.String("testsquawk", "", "Alias of the crew member to send a squawk to")
	
	flag.Parse()
	
	crowService := crow.Crow_c {}
	err := crowService.Init()
	
	if err != nil {	//see if we initalized correctly
		fmt.Println(err)
		os.Exit(0)
	}
	
	//check the flags
	if len(*squawkTest) > 0 {
		if found := crowService.SendCrewMemberSquawk(*squawkTest, "This is a test squawk sent from Crow's Nest"); !found {
			fmt.Println("Crew member not found!")
		}
		os.Exit(0)
	}
	
	wg := new(sync.WaitGroup)
	wg.Add(1)
	
	//this handles killing the service gracefully
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func(wg *sync.WaitGroup){
		<-c
		//for sig := range c {
			// sig is a ^C, handle it
			fmt.Println("Crow's Nest service exiting gracefully")
			//os.Exit(0)
			defer wg.Done()
		//}
	}(wg)
	
	//this is our polling interval
	ticker := time.NewTicker(time.Minute * time.Duration(1))	//check every minute, that's our min
	go func() {
		for range ticker.C {
			crowService.CheckAllEggs()	//tells crow to check all the eggs
		}
	} ()
	
	crowService.CheckAllEggs()	//tells crow to check all the eggs
	
	wg.Wait()	//wait for the slave and possible master to finish
}