package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"log"
	"net"
	"tool"

	"github.com/tidwall/gjson"
)

var listeningOn *string
var listeningOnDefault = "33333"
var pool *string
var poolDefault = "asia1.bsod.pw:3833"
var diff *string
var diffDefault = "50"

var logger = &tool.Logger{Level: 4}

var authid int64 = 0
var extranonceid int64
var subscribeid int64
var user = ""
var pass = ""
var worker = ""

var j interface{}

func main() {
	///Configure logger
	//logger = &Log{level: 1} //1:debug info only

	pool = flag.String("pool", poolDefault, "Input pool address:port")
	listeningOn = flag.String("listen", listeningOnDefault, "Input listen port")
	diff = flag.String("diff", diffDefault, "Mining difficult")

	flag.Parse()
	//runtime.GOMAXPROCS(1)
	logger.Info(fmt.Sprintf("Pool: %s, diff:%s, listen on %s", *pool, *diff, *listeningOn))

	listener, err := net.Listen("tcp", "0.0.0.0:"+*listeningOn)
	checkError(err)
	for {
		conn, err := listener.Accept()
		checkError(err)
		fmt.Printf("incoming connection %s -> %s \n", conn.RemoteAddr(), conn.LocalAddr())
		go handle(conn.(*net.TCPConn))
	}
}

//检查错误
func checkError(err error) int {
	if err != nil {
		if err.Error() == "EOF" {
			fmt.Println("用户退出了")
			return 0
		}
		log.Fatal("an error!", err.Error())
		return -1
	}
	return 1
}

func handle(miner *net.TCPConn) {

	defer log.Println("Going to close conn", miner.RemoteAddr().String())

	defer miner.Close()

	logger.Info("Connectiing " + *pool)
	pool, err := net.Dial("tcp", *pool)
	checkError(err)
	defer pool.Close()
	go func() {
		defer log.Println("Going to close miner conn", miner.RemoteAddr().String())
		defer miner.Close()
		defer pool.Close()

		// buf := make([]byte, 1024)
		// //log.Println(buf)

		// io.CopyBuffer(pool, miner, buf)

		var m = new(tool.Manager)
		m.InitiateLogger(5)

		// 构建reader和writer
		poolWriter := bufio.NewWriter(pool)
		minerReader := bufio.NewReader(miner)

		for {
			// 读取一行数据, 以"\n"结尾
			b, _, err := minerReader.ReadLine()
			if err != nil {
				return
			}

			//log.Printf("Miner -> %s \n", string(b))

			b = processMinerMessage(b)
			//m.ProcessMinerMsg(b)

			poolWriter.Write(b)
			poolWriter.Write([]byte("\n"))
			poolWriter.Flush()
			//conn.Write(r)
			//conn.Write([]byte("\n"))
		}

	}()

	// 构建reader和writer
	poolReader := bufio.NewReader(pool)
	minerWriter := bufio.NewWriter(miner)

	for {
		// 读取一行数据, 以"\n"结尾
		b, _, err := poolReader.ReadLine()
		if err != nil {
			return
		}

		//log.Printf("pool -> %s \n", string(b))

		b = processPoolMessage(b)
		//processMessage(b)

		minerWriter.Write(b)
		minerWriter.Write([]byte("\n"))
		minerWriter.Flush()
		//conn.Write(r)
		//conn.Write([]byte("\n"))
	}

}

func processPoolMessage(ret []byte) []byte {

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
	if id == subscribeid && hasResult && hasSetDiff && hasMiningNotify {
		logger.Debug("Pool Raw: " + string(ret))
		logger.Info("Pool  --> Subscribe OK ")
		subscribeid = 0
	} else if hasMethod && hasParams && hasSetDiff {
		logger.Debug("Pool Raw: " + string(ret))
		// newdiff := []byte("[" + *diff + "]")

		diffOrig := gjson.GetBytes(ret, "params").String()

		// ret = bytes.Replace(ret, []byte(diffOrig), newdiff, 1)

		logger.Info("Pool  --> Setting diff" + string(diffOrig))

	} else if hasMethod && hasParams && hasMiningNotify {
		//python版本在此增加clean_job
		logger.Debug("Pool Raw: " + string(ret))
		logger.Info("Pool  --> Mining Notify ")
	} else if hasMethod && hasExtranonce {
		logger.Debug("Pool Raw: " + string(ret))
		logger.Info("Pool  --> extranonce subscribe ok")
	} else if result.Bool() {
		error := gjson.GetBytes(ret, "error").String()

		if id == authid && error == "" {
			logger.Debug("Pool Raw: " + string(ret))
			logger.Info("Pool  --> Worker authorised")
			authid = 0
		} else if id == extranonceid && error == "" {
			logger.Debug("Pool Raw: " + string(ret))
			logger.Info("Pool  --> Extranonce subscription OK.")
			extranonceid = 0
		} else if error == "" {
			logger.Debug("Pool Raw: " + string(ret))
			logger.Info("Pool  --> Share accepted")
		} else {
			logger.Warning("Pool WTF result=true Raw: " + string(ret))
		}
	} else if !result.Bool() {
		error := gjson.GetBytes(ret, "error").String()
		logger.Debug("Pool Raw: " + string(ret))
		logger.Info("Pool  --> Share rejected with error: " + error)
	} else {
		logger.Warning("Pool WTF Raw: " + string(ret))
	}

	return ret

}

func processMinerMessage(ret []byte) []byte {

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
			subscribeid = id
			logger.Debug("Miner Raw--> " + string(ret))
			logger.Debug("Miner subscribeid = " + string(subscribeid))
			logger.Info("Miner --> Initiate subscription with params:" + param.String())

		} else if method.String() == "mining.authorize" && param.Exists() {
			logger.Debug("Miner Raw--> " + string(ret))

			user = param.Array()[0].String()
			pass = param.Array()[1].String() //"d=20" / "sd=20"

			//Analyze pass
			b := bytes.Split([]byte(pass), []byte("="))

			//could be x in the pass field
			if len(b) != 2 {

			}

			passNew := []byte("d=24")
			ret = bytes.Replace(ret, []byte(pass), passNew, -1)

			logger.Debug("Miner --> Diff adjusted:" + string(ret))

			//fmt.Printf("%q\n", b)

			authid = id
			logger.Info(fmt.Sprintf("Miner --> User: %s/%s id=%d", user, pass, authid))
		} else if method.String() == "mining.extranonce.subscribe" {
			extranonceid = id
			logger.Debug("Miner --> " + string(ret))
			logger.Info("Miner --> Subscribe extranonce")
		} else if method.String() == "mining.submit" && param.Exists() {
			logger.Debug("Miner --> " + string(ret))
			logger.Info(fmt.Sprintf("Miner --> Submit result id %d by %s", id, user))
		} else {
			logger.Info(string(ret))
		}
	} else {
		logger.Info(string(ret))
	}
	return ret
}
