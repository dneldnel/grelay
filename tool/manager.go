package tool

import (
	"bytes"
	"fmt"
	"strconv"

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
	diff         *string
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

//ProcessMinerMessage Process msg from miner
func (proc *Manager) ProcessMinerMessage(ret []byte, diff2 *string) []byte {

	proc.diff = diff2
	hasMethod := bytes.Contains(ret, []byte("method"))
	//hasResult := bytes.Contains(ret, []byte("result"))
	hasParams := bytes.Contains(ret, []byte("params"))

	//{"id": 3, "method": "mining.extranonce.subscribe", "params": []}
	//{"method": "mining.submit", "params": ["bFQrErYrzHjgLyFcjeCCpg4GJwMcov3Te7.aaa", "1417", "00000000", "5ae14ad9", "49b81d16"], "id":10}
	//{"id": 3, "method": "mining.extranonce.subscribe", "params": []}
	if hasMethod && hasParams {
		// logger.Debug("got method: " + gjson.GetBytes(ret, "method").String())

		method := gjson.GetBytes(ret, "method")
		param := gjson.GetBytes(ret, "params")
		id := gjson.GetBytes(ret, "id").Int()

		if method.String() == "mining.subscribe" { //usually id=1 from miner
			proc.subscribeid = id
			proc.logger.Debug("Miner Raw--> " + string(ret))
			proc.logger.Debug("Miner subscribeid = " + string(proc.subscribeid))
			proc.logger.Info("Miner --> Initiate subscription with params:" + param.String())

		} else if method.String() == "mining.authorize" && param.Exists() {
			proc.logger.Debug("Miner Raw--> " + string(ret))

			proc.authid = id

			proc.user = param.Array()[0].String()
			proc.pass = param.Array()[1].String() //"d=20" / "sd=20" /"x"

			_, err := strconv.ParseFloat(*(proc.diff), 32)

			if err == nil {
				//Analyze pass
				b := bytes.Split([]byte(proc.pass), []byte("="))

				//could be x in the pass field
				if len(b) != 2 {

				}

				passNew := []byte("d=" + *proc.diff)
				ret = bytes.Replace(ret, []byte(proc.pass), passNew, -1)

				proc.logger.Debug("Miner --> Diff adjusted:" + string(ret))

				proc.logger.Info(fmt.Sprintf("Miner --> User: %s/%s id=%d", proc.user, passNew, proc.authid))
			} else {
				//fmt.Printf("%q\n", b)

				proc.logger.Info(fmt.Sprintf("Miner --> User: %s/%s id=%d", proc.user, proc.pass, proc.authid))
			}
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
	return ret
}

//ProcessPoolMsg Process Miner-Proxy json
func (proc *Manager) ProcessPoolMessage(ret []byte) []byte {

	//println(str)
	hasSetDiff := bytes.Contains(ret, []byte("mining.set_difficulty"))
	hasMiningNotify := bytes.Contains(ret, []byte("mining.notify"))
	hasMethod := bytes.Contains(ret, []byte("method"))
	hasResult := bytes.Contains(ret, []byte("result"))
	hasParams := bytes.Contains(ret, []byte("params"))
	hasExtranonce := bytes.Contains(ret, []byte("mining.extranonce.subscribe"))

	id := gjson.GetBytes(ret, "id").Int()
	result := gjson.GetBytes(ret, "result")

	//{"id":1,"result":[[["mining.set_difficulty","1"],["mining.notify","7d7806ce0fe1d84d46bc7c164a169f2a"]],"8100f103",4],"error":null}
	if id == proc.subscribeid && hasResult && hasSetDiff && hasMiningNotify {
		proc.logger.Debug("Pool Raw: " + string(ret))
		proc.logger.Info("Pool  --> Subscribe OK ")
		proc.subscribeid = 0
	} else if hasMethod && hasParams && hasSetDiff {
		proc.logger.Debug("Pool Raw: " + string(ret))
		// newdiff := []byte("[" + *diff + "]")

		diffOrig := gjson.GetBytes(ret, "params").String()

		// ret = bytes.Replace(ret, []byte(diffOrig), newdiff, 1)

		proc.logger.Info("Pool  --> Setting diff" + string(diffOrig))

	} else if hasMethod && hasParams && hasMiningNotify {
		//python版本在此增加clean_job
		proc.logger.Debug("Pool Raw: " + string(ret))
		proc.logger.Info("Pool  --> Mining Notify ")
	} else if hasMethod && hasExtranonce {
		proc.logger.Debug("Pool Raw: " + string(ret))
		proc.logger.Info("Pool  --> extranonce subscribe ok")
	} else if result.Bool() {
		error := gjson.GetBytes(ret, "error").String()

		if id == proc.authid && error == "" {
			proc.logger.Debug("Pool Raw: " + string(ret))
			proc.logger.Info("Pool  --> Worker authorised")
			proc.authid = 0
		} else if id == proc.extranonceid && error == "" {
			proc.logger.Debug("Pool Raw: " + string(ret))
			proc.logger.Info("Pool  --> Extranonce subscription OK.")
			proc.extranonceid = 0
		} else if error == "" {
			proc.logger.Debug("Pool Raw: " + string(ret))
			proc.logger.Info("Pool  --> Share accepted")
		} else {
			proc.logger.Warning("Pool WTF result=true Raw: " + string(ret))
		}
	} else if !result.Bool() {
		error := gjson.GetBytes(ret, "error").String()
		proc.logger.Debug("Pool Raw: " + string(ret))
		proc.logger.Info("Pool  --> Share rejected with error: " + error)
	} else {
		proc.logger.Warning("Pool WTF Raw: " + string(ret))
	}

	return ret

}
