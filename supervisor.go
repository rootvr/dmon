package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"

	logger "dmon/mod/logger"
	network "dmon/mod/network"
	storage "dmon/mod/storage"
	structure "dmon/mod/structure"
)

var _modname = "main"

func ParsePort(port string) error {
	_, err := strconv.ParseUint(port, 10, 16)
	return err
}

func requiredFlags(flagName string) bool {
	res := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == flagName {
			res = true
		}
	})
	return res
}

func parseArgs() ([]string, []string, string, []string) {
	var interface_args []string
	var network_args []string
	var redis_args []string

	i := flag.String("i", "", "(Wireshark arg) Physical network interfaces (comma-separated list)")
	n := flag.String("n", "", "(Docker arg) Virtual network interfaces (comma-separated list)")
	t := flag.String("t", "", "(Docker arg) Structure polling timer (in seconds)")
	r := flag.String("r", "", "(Redis arg) Server IP and PORT")

	flag.Parse()
	if !requiredFlags("i") {
		fmt.Println("CALL ERROR: -i flag w/ args are required")
		flag.PrintDefaults()
		os.Exit(0)
	}
	if !requiredFlags("n") {
		fmt.Println("CALL ERROR: -n flag w/ args are required")
		flag.PrintDefaults()
		os.Exit(0)
	}
	if !requiredFlags("t") {
		fmt.Println("CALL ERROR: -t flag w/ args are required")
		flag.PrintDefaults()
		os.Exit(0)
	}
	if !requiredFlags("r") {
		fmt.Println("CALL ERROR: -r flag w/ args are required")
		flag.PrintDefaults()
		os.Exit(0)
	}

	if strings.Contains(*i, ",") {
		interface_args = strings.Split(*i, ",")
		for a := range interface_args {
			interface_args[a] = "-i" + interface_args[a]
		}
	} else {
		interface_args = append(interface_args, *i)
		interface_args[0] = "-i" + interface_args[0]
	}

	if strings.Contains(*n, ",") {
		network_args = strings.Split(*n, ",")
	} else {
		network_args = append(network_args, *n)
	}

	if strings.Contains(*r, ":") {
		redis_args = strings.Split(*r, ":")
	} else {
		fmt.Println("CALL ERROR: -r flag w/ args must be `-r IP:PORT`, given", *r)
		flag.PrintDefaults()
		os.Exit(0)
	}

	if redis_args[0] != "localhost" && net.ParseIP(redis_args[0]) == nil {
		fmt.Println("CALL ERROR: -r flag w/ args invalid IP string, given", redis_args[0])
		flag.PrintDefaults()
		os.Exit(0)
	}
	if err := ParsePort(redis_args[1]); err != nil {
		fmt.Println("CALL ERROR: -r flag w/ args invalid PORT string, given", redis_args[1])
		flag.PrintDefaults()
		os.Exit(0)
	}

	return interface_args, network_args, *t, redis_args
}

func main() {
	interface_args, network_args, timer_arg, redis_args := parseArgs()

	logger.Info(_modname, "proc::status", "start")
	logger.Info(_modname, "args::iface", interface_args)
	logger.Info(_modname, "args::netws", network_args)

	storage.Init(redis_args)

	var waitGroup sync.WaitGroup

	waitGroup.Add(1)
	go network.InitSpawn(&waitGroup, interface_args)

	waitGroup.Add(1)
	go structure.InitSpawn(&waitGroup, network_args, timer_arg)

	waitGroup.Wait()

	logger.Info(_modname, "proc::status", "exit")
}
