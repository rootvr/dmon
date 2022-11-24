package network

import (
	"bufio"
	"io"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"

	logger "dmon/mod/logger"
	response "dmon/mod/response"
	storage "dmon/mod/storage"
	payload "dmon/mod/storage/payload"
	utils "dmon/mod/utils"
)

var _modname = "network"
var _channel_name = "dmon_network_out"

func retrieveSpawn(scanner *bufio.Scanner, cmdReader io.ReadCloser) {
	for scanner.Scan() {
		resTrim := strings.TrimSpace(scanner.Text())
		resNounic := strings.ReplaceAll(resTrim, "→", " ")
		reMultiWhSp := regexp.MustCompile(`[\s\p{Zs}]{2,}`)
		resRegex := reMultiWhSp.ReplaceAllString(resNounic, " ")
		resSplit := strings.Split(resRegex, " ")

		networkPayload := networkPayloadInit(resSplit[1:])
		storageHSetSupervisor(networkPayload)
		storageDisSupervisor(networkPayload)
	}
}

func networkPayloadInit(res []string) *payload.RedisNetPayload {
	networkPayload := payload.RedisNetPayload{}
	networkPayload.Timestamp = res[0] + " " + res[1]
	networkPayload.Type = "network"
	networkPayload.SubType = "general"
	networkPayload.SendIP = res[2]
	networkPayload.RecvIP = res[3]
	networkPayload.Protocol = res[4]
	dt, err := strconv.ParseFloat(res[len(res)-1], 64)
	utils.Panic(_modname, err, "delta:error", "error in strconv and ParseFloat of float64 delta time")
	networkPayload.TimeDelta = dt
	return &networkPayload
}

func storageHSetSupervisor(networkPayload *payload.RedisNetPayload) {
	storage.RedisNetworkHashSet(*networkPayload)
	storage.RedisNetworkHashGet(*networkPayload)
}

func storageDisSupervisor(networkPayload *payload.RedisNetPayload) {
	jsonPayload := response.JsonPayloadInit(networkPayload)
	storage.RedisDisPublish(jsonPayload, _channel_name)
}

func cmdInit(interface_args []string) *exec.Cmd {
	cmdName := "tshark"
	cmdTFlag := "-t"
	cmdTArg := "ad"
	cmdYFlag := "-Y"
	cmdYArg := "tcp and not ssl and not tls and tcp.time_delta > 0"
	cmdZFlag := "-z"
	cmdZArg := "proto,colinfo,tcp.time_delta,tcp.time_delta"
	cmdArgs := []string{cmdTFlag, cmdTArg, cmdYFlag, cmdYArg, cmdZFlag, cmdZArg}
	cmdArgs = append(cmdArgs, interface_args...)
	cmd := exec.Command(cmdName, cmdArgs...)
	logger.Info(_modname, "network::cmd", cmd.String())
	return cmd
}

func retrieveSupervisor(interface_args []string) {
	cmd := cmdInit(interface_args)
	cmdReader, err := cmd.StdoutPipe()
	utils.Panic(_modname, err, "cmd::error", "error creating StdoutPipe for cmd")
	scanner := bufio.NewScanner(cmdReader)
	go retrieveSpawn(scanner, cmdReader)
	err = cmd.Start()
	utils.Panic(_modname, err, "cmd::error", "error starting cmd")
	err = cmd.Wait()
	utils.Panic(_modname, err, "cmd::error", "error waiting for cmd")
}

func InitSpawn(waitGroup *sync.WaitGroup, interface_args []string) {
	defer waitGroup.Done()
	logger.Info(_modname, "proc::status", "thread spawn")
	retrieveSupervisor(interface_args)
	logger.Info(_modname, "proc::status", "thread exit")
}
