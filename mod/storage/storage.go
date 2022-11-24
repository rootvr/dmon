package storage

import (
	"context"
	"encoding/json"
	"strings"

	logger "dmon/mod/logger"
	payload "dmon/mod/storage/payload"
	utils "dmon/mod/utils"

	redis "github.com/go-redis/redis/v8"
)

var _modname = "storage"

var _ctx = context.Background()
var _rdb *redis.Client

func clientInit(redis_args []string) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     strings.Join(redis_args, ":"),
		Password: "",
		DB:       0,
	})
	_rdb = rdb
	pingPongTest()
}

func pingPongTest() {
	pong, err := _rdb.Ping(_ctx).Result()
	utils.Panic(_modname, err, "redis:error", "ping-pong test error")
	logger.Info(_modname, "redis::test", pong)
}

func RedisNetworkHashSet(rnetp payload.RedisNetPayload) {
	_rdb.HSet(_ctx, rnetp.Timestamp,
		"type", rnetp.Type,
		"sub_type", rnetp.SubType,
		"send_ipv4", rnetp.SendIP,
		"recv_ipv4", rnetp.RecvIP,
		"protocol", rnetp.Protocol,
		"time_delta", rnetp.TimeDelta)
}

func RedisNetworkHashGet(rnetp payload.RedisNetPayload) {
	res, err := _rdb.HGetAll(_ctx, rnetp.Timestamp).Result()
	utils.Panic(_modname, err, "redis:error", "network general hget error: not found")
	logger.Info(_modname, "redis::hget",
		rnetp.Timestamp,
		res["type"],
		res["sub_type"],
		res["send_ipv4"],
		res["recv_ipv4"],
		res["protocol"],
		res["time_delta"])
}

func RedisStructureHostHashSet(rstrhp payload.RedisStructHostPayload) {
	_rdb.HSet(_ctx, rstrhp.Timestamp,
		"type", rstrhp.Type,
		"sub_type", rstrhp.SubType,
		"os", rstrhp.OperatingSystem,
		"os_type", rstrhp.OSType,
		"arch", rstrhp.Architecture,
		"name", rstrhp.Name,
		"ncpu", rstrhp.NCPU,
		"memtot", rstrhp.MemTotal,
		"kernel_version", rstrhp.KernelVersion)
}

func RedisStructureHostHashGet(rstrhp payload.RedisStructHostPayload) {
	res, err := _rdb.HGetAll(_ctx, rstrhp.Timestamp).Result()
	utils.Panic(_modname, err, "redis:error", "structure host hget error: not found")
	logger.Info(_modname, "redis::hget",
		rstrhp.Timestamp,
		res["type"],
		res["sub_type"],
		res["os"],
		res["os_type"],
		res["arch"],
		res["kernel_version"])
}

func RedisStructureNetworkHashSet(rstrnp payload.RedisStructNetPayload) {
	_rdb.HSet(_ctx, rstrnp.Timestamp,
		"type", rstrnp.Type,
		"sub_type", rstrnp.SubType,
		"id", rstrnp.ID,
		"name", rstrnp.Name)
}

func RedisStructureNetworkHashGet(rstrnp payload.RedisStructNetPayload) {
	res, err := _rdb.HGetAll(_ctx, rstrnp.Timestamp).Result()
	utils.Panic(_modname, err, "redis:error", "structure network hget error: not found")
	logger.Info(_modname, "redis::hget",
		rstrnp.Timestamp,
		res["type"],
		res["sub_type"],
		res["id"],
		res["name"])
}

func RedisStructureContainerHashSet(rstrcp payload.RedisStructContPayload) {
	ipAddressed, _ := json.Marshal(rstrcp.IPAddresses)
	ports, _ := json.Marshal(rstrcp.Ports)

	_rdb.HSet(_ctx, rstrcp.Timestamp,
		"type", rstrcp.Type,
		"sub_type", rstrcp.SubType,
		"id", rstrcp.ID,
		"name", rstrcp.Name,
		"image", rstrcp.Image,
		"locale", rstrcp.Locale,
		"timezone", rstrcp.Timezone,
		"ip_addresses", string(ipAddressed),
		"ports", string(ports))
}

func RedisStructureContainerHashGet(rstrcp payload.RedisStructContPayload) {
	res, err := _rdb.HGetAll(_ctx, rstrcp.Timestamp).Result()
	utils.Panic(_modname, err, "redis:error", "structure container hget error: not found")
	logger.Info(_modname, "redis::hget",
		rstrcp.Timestamp,
		res["type"],
		res["sub_type"],
		res["id"][:12],
		res["name"])
}

func RedisDisPublish(jsonPayload []byte, channelName string) {
	_rdb.Publish(_ctx, channelName, jsonPayload)
	_rdb.Publish(_ctx, "dmon_merged_out", jsonPayload)
}

func Init(redis_args []string) {
	logger.Info(_modname, "proc::status", "start")
	clientInit(redis_args)
	logger.Info(_modname, "proc::status", "exit")
}
