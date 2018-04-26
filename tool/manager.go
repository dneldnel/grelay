package tool

import (
	"bytes"
	"fmt"

	"github.com/tidwall/gjson"
)

//Manager Processing unit
type Manager struct {
	logger       *Logger
	subscribeid  int64
	authid       int64
	extranonceid int64
	user         string
	pass         string
}

//NewManager constructor
// func NewManager(level int) *Manager {

// 	m := new(Manager)
// 	m.logger = &Logger{Level: level}
// 	return m
// }

//InitiateLogger Initiate logger
func (this *Manager) InitiateLogger(level int) {
	//this.loglevel = level
	this.logger = &Logger{Level: level}
}

//ProcessMinerMsg Process Miner-Proxy json
func (proc *Manager) ProcessMinerMsg(ret []byte) {

	hasMethod := bytes.Contains(ret, []byte("method"))
	//hasResult := bytes.Contains(ret, []byte("result"))
	hasParams := bytes.Contains(ret, []byte("params"))
	//hasExtranonce := bytes.Contains(ret, []byte("mining.extranonce.subscribe"))
	//hasAuthorize := bytes.Contains(ret, []byte("mining.authorize"))

	//{"id": 3, "method": "mining.extranonce.subscribe", "params": []}
	//{"method": "mining.submit", "params": ["bFQrErYrzHjgLyFcjeCCpg4GJwMcov3Te7.aaa", "1417", "00000000", "5ae14ad9", "49b81d16"], "id":10}
	//{"id": 3, "method": "mining.extranonce.subscribe", "params": []}
	if hasMethod && hasParams {
		// logger.Debug("got method: " + gjson.GetBytes(ret, "method").String())

		method := gjson.GetBytes(ret, "method")
		param := gjson.GetBytes(ret, "params")
		id := gjson.GetBytes(ret, "id").Int()

		if method.String() == "mining.subscribe" { //usually id=1 from miner

			proc.logger.Info("Miner --> Initiate subscription with params:" + param.String())
			proc.subscribeid = id
		} else if method.String() == "mining.authorize" && param.Exists() {
			proc.user = param.Array()[0].String()
			proc.pass = param.Array()[1].String()
			proc.authid = id
			proc.logger.Info(fmt.Sprintf("Miner --> User: %s/%s id=%d", proc.user, proc.pass, proc.authid))
		} else if method.String() == "mining.extranonce.subscribe" {
			proc.extranonceid = id
			proc.logger.Debug("Miner --> " + string(ret))
			proc.logger.Info("Miner --> Subscribe extranonce")
		} else if method.String() == "mining.submit" && param.Exists() {
			proc.logger.Debug("Miner --> " + string(ret))
			proc.logger.Info(fmt.Sprintf("Miner --> Submit result id %d by %s", id, proc.user))
		} else {
			proc.logger.Info(string(ret))
		}
	} else {
		proc.logger.Info(string(ret))
	}
}
