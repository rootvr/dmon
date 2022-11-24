package structure

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"time"

	logger "dmon/mod/logger"
	response "dmon/mod/response"
	storage "dmon/mod/storage"
	payload "dmon/mod/storage/payload"
	utils "dmon/mod/utils"

	docker_types "github.com/docker/docker/api/types"
	docker_client "github.com/docker/docker/client"
)

var _modname = "structure"
var _channel_name = "dmon_structure_out"

var _ctx = context.Background()
var _cli *docker_client.Client

var timezone string
var locale string

func getHostTimezone() string {
	timezone := time.Now().Location().String()
	if timezone == "Local" {
		cmd := "ls -lah /etc/localtime | cut -d \">\" -f2 | tr -d \" \" | rev | cut -d \"/\" -f1,2 | rev | tr -d \"\n\""
		stdout, err := exec.Command("sh", "-c", cmd).Output()
		utils.Panic(_modname, err, "docker::err", "unable to retrieve machine timezone")
		return string(stdout)
	} else {
		return timezone
	}
}

func getHostLocale() string {
	locale, ok := os.LookupEnv("LANG")
	if locale == "" || ok == false {
		cmd := "cat /etc/locale.conf | grep LANG | cut -d \"=\" -f2"
		stdout, err := exec.Command("sh", "-c", cmd).Output()
		utils.Panic(_modname, err, "docker::err", "unable to retrieve machine locale")
		return string(stdout)
	} else {
		return locale
	}
}

func clientInit(network_args []string, tick time.Duration) {
	cli, err := docker_client.NewClientWithOpts(docker_client.FromEnv, docker_client.WithAPIVersionNegotiation())
	utils.Panic(_modname, err, "docker::err", "docker client initialization error")
	_cli = cli
	getHostInfo()
	getNetworkList(network_args)
	for {
		<-time.Tick(tick)
		getContainerList()
	}
}

func getHostInfo() {
	hostInfo, err := _cli.Info(_ctx)
	utils.Panic(_modname, err, "docker::err", "docker error during host info retrieve")
	structureHostPayload := structureHostPayloadInit(hostInfo)
	storageHostHSetSupervisor(structureHostPayload)
	storageDisHostSupervisor(structureHostPayload)
}

func structureHostPayloadInit(hostInfo docker_types.Info) *payload.RedisStructHostPayload {
	structureHostPayload := payload.RedisStructHostPayload{}
	structureHostPayload.Timestamp = time.Now().Format("2006-01-02 15:04:05.000000000")
	structureHostPayload.Type = "structure"
	structureHostPayload.SubType = "host"
	structureHostPayload.OperatingSystem = hostInfo.OperatingSystem
	structureHostPayload.OSType = hostInfo.OSType
	structureHostPayload.Architecture = hostInfo.Architecture
	structureHostPayload.Name = hostInfo.Name
	structureHostPayload.NCPU = hostInfo.NCPU
	structureHostPayload.MemTotal = (hostInfo.MemTotal) / (1000 * 1000)
	structureHostPayload.KernelVersion = hostInfo.KernelVersion
	return &structureHostPayload
}

func storageHostHSetSupervisor(structureHostPayload *payload.RedisStructHostPayload) {
	storage.RedisStructureHostHashSet(*structureHostPayload)
	storage.RedisStructureHostHashGet(*structureHostPayload)
}

func storageDisHostSupervisor(structureHostPayload *payload.RedisStructHostPayload) {
	jsonPayload := response.JsonPayloadInit(structureHostPayload)
	storage.RedisDisPublish(jsonPayload, _channel_name)
}

func getNetworkList(network_args []string) {
	networks, err := _cli.NetworkList(_ctx, docker_types.NetworkListOptions{})
	utils.Panic(_modname, err, "docker::err", "docker error during network list retrieve")
	for _, network := range networks {
		for _, network_arg := range network_args {
			if network.Name == network_arg {
				networkInspect(network)
			}
		}
	}
}

func networkInspect(network docker_types.NetworkResource) {
	network, err := _cli.NetworkInspect(_ctx, network.ID, docker_types.NetworkInspectOptions{})
	utils.Panic(_modname, err, "docker::err", "docker error during network inspect")
	structureNetworkPayload := structureNetworkPayloadInit(network)
	storageNetworkHSetSupervisor(structureNetworkPayload)
	storageDisNetworkSupervisor(structureNetworkPayload)
}

func structureNetworkPayloadInit(network docker_types.NetworkResource) *payload.RedisStructNetPayload {
	structureNetworkPayload := payload.RedisStructNetPayload{}
	structureNetworkPayload.Timestamp = time.Now().Format("2006-01-02 15:04:05.000000000")
	structureNetworkPayload.Type = "structure"
	structureNetworkPayload.SubType = "network"
	structureNetworkPayload.ID = network.ID
	structureNetworkPayload.Name = network.Name
	return &structureNetworkPayload
}

func storageNetworkHSetSupervisor(structureNetworkPayload *payload.RedisStructNetPayload) {
	storage.RedisStructureNetworkHashSet(*structureNetworkPayload)
	storage.RedisStructureNetworkHashGet(*structureNetworkPayload)
}

func storageDisNetworkSupervisor(structureNetworkPayload *payload.RedisStructNetPayload) {
	jsonPayload := response.JsonPayloadInit(structureNetworkPayload)
	storage.RedisDisPublish(jsonPayload, _channel_name)
}

func getContainerList() {
	containers, err := _cli.ContainerList(_ctx, docker_types.ContainerListOptions{})
	utils.Panic(_modname, err, "docker::err", "docker error during container list retrieve")
	for _, container := range containers {
		containerJsonBase := containerInspect(container)
		cpuPerc := containerStats(container)

		structureContainerPayload := structureContainerPayloadInit(containerJsonBase, cpuPerc)
		structureContainerPayload.IPAddresses = make(map[string]string)
		structureContainerPayload.Ports = make(map[string][]payload.RedisStructContNetPortPayload)

		structureContainerPayload.Locale = locale
		structureContainerPayload.Timezone = timezone

		for network, info := range container.NetworkSettings.Networks {
			structureContainerPayload.IPAddresses[network] = info.IPAddress
		}
		for _, v := range container.Ports {
			structureContainerPayload.Ports[v.IP] = append(structureContainerPayload.Ports[v.IP], payload.RedisStructContNetPortPayload{PrivatePort: v.PrivatePort, PublicPort: v.PublicPort, Type: v.Type})
		}

		storageContainerHSetSupervisor(structureContainerPayload)
		storageDisContainerSupervisor(structureContainerPayload)
	}
}

func containerInspect(container docker_types.Container) *docker_types.ContainerJSONBase {
	containerjson, err := _cli.ContainerInspect(_ctx, container.ID)
	utils.Panic(_modname, err, "docker::err", "docker error during container inspect")
	containerJsonBase := containerjson.ContainerJSONBase
	return containerJsonBase
}

func containerStats(container docker_types.Container) float64 {
	stats, err := _cli.ContainerStats(context.Background(), container.ID, false)
	utils.Panic(_modname, err, "docker::err", "docker error during container stats retrieve")
	var statsData map[string]interface{}
	jerr := json.NewDecoder(stats.Body).Decode(&statsData)
	utils.Panic(_modname, jerr, "json::error", "json error during container stats body buffer decode")
	cpuPercRaw := computeCpuUsage(statsData)
	cpuPerc := math.Round(cpuPercRaw*100) / 100
	return cpuPerc
}

func computeCpuUsage(statsData map[string]interface{}) float64 {
	cpuu_total_usage :=
		statsData["cpu_stats"].(map[string]interface{})["cpu_usage"].(map[string]interface{})["total_usage"].(float64)
	cpuu_system_cpu_usage :=
		statsData["cpu_stats"].(map[string]interface{})["system_cpu_usage"].(float64)

	pcpus_online_cpus :=
		statsData["precpu_stats"].(map[string]interface{})["online_cpus"].(float64)

	pcpus_total_usage :=
		statsData["precpu_stats"].(map[string]interface{})["cpu_usage"].(map[string]interface{})["total_usage"].(float64)
	pcpus_system_cpu_usage :=
		statsData["precpu_stats"].(map[string]interface{})["system_cpu_usage"].(float64)

	cpuUsage := 0.0
	cpuDelta := cpuu_total_usage - pcpus_total_usage
	systemDelta := cpuu_system_cpu_usage - pcpus_system_cpu_usage

	if systemDelta > 0.0 {
		cpuUsage = (cpuDelta / systemDelta) * 100.0 * pcpus_online_cpus
	}

	return cpuUsage
}

func structureContainerPayloadInit(container *docker_types.ContainerJSONBase, cpuPerc float64) *payload.RedisStructContPayload {
	structureContainerPayload := payload.RedisStructContPayload{}
	structureContainerPayload.Timestamp = time.Now().Format("2006-01-02 15:04:05.000000000")
	structureContainerPayload.Type = "structure"
	structureContainerPayload.SubType = "container"
	structureContainerPayload.ID = container.ID
	structureContainerPayload.Name = container.Name
	structureContainerPayload.Image = container.Image
	structureContainerPayload.CPUPerc = fmt.Sprintf("%f%%", cpuPerc)
	return &structureContainerPayload
}

func storageContainerHSetSupervisor(structureContainerPayload *payload.RedisStructContPayload) {
	storage.RedisStructureContainerHashSet(*structureContainerPayload)
	storage.RedisStructureContainerHashGet(*structureContainerPayload)
}

func storageDisContainerSupervisor(structureContainerPayload *payload.RedisStructContPayload) {
	jsonPayload := response.JsonPayloadInit(structureContainerPayload)
	storage.RedisDisPublish(jsonPayload, _channel_name)
}

func InitSpawn(waitGroup *sync.WaitGroup, network_args []string, timer_arg string) {
	defer waitGroup.Done()
	logger.Info(_modname, "proc::status", "thread spawn")
	tickArg, err := strconv.ParseInt(timer_arg, 10, 64)
	utils.Panic(_modname, err, "timer::error", "error in strconv and ParseInt of string network tick timer")
	timezone = getHostTimezone()
	locale = getHostLocale()
	clientInit(network_args, time.Duration(tickArg)*time.Second)
	logger.Info(_modname, "proc::status", "thread exit")
}
