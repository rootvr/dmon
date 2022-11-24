package response

import (
	"encoding/json"

	utils "dmon/mod/utils"
)

var _modname = "response"

func JsonPayloadInit(res interface{}) []byte {
	jsonRes, err := json.Marshal(res)
	utils.Panic(_modname, err, "res:error", "marshalling error")
	return jsonRes
}
