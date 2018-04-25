package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"log"
	"net"
	"strings"
	"tool"

	"github.com/tidwall/gjson"
)

var listeningOn *string
var listeningOnDefault = "33333"
var pool *string
var poolDefault = "asia1.bsod.pw:3833"

var logger = &tool.Logger{Level: 4}

var j interface{}

func main() {
	///Configure logger
	//logger = &Log{level: 1} //1:debug info only

	pool = flag.String("pool", poolDefault, "Input pool address:port")
	listeningOn = flag.String("listen", listeningOnDefault, "Input listen port")
	//runtime.GOMAXPROCS(1)
	logger.Info(fmt.Sprintf("Pool: %s, listen on %s", *pool, *listeningOn))

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

			//processMinerMessage(b)
			processMessage(b)

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

		//processpoolMessage(b)
		processMessage(b)

		minerWriter.Write(b)
		minerWriter.Write([]byte("\n"))
		minerWriter.Flush()
		//conn.Write(r)
		//conn.Write([]byte("\n"))
	}

}

func processpoolMessage(ret []byte) {
	str := string(ret)
	//println(str)
	// id := gjson.GetBytes(ret, "id")
	setDiff := strings.Contains(str, "mining.set_difficulty")
	miningNotify := strings.Contains(str, "mining.notify")

	if setDiff {
		log.Printf("pool  --> Set Diff")
		return
	} else if miningNotify {
		log.Printf("pool  --> Mining notify")
		return
	}

	result := gjson.GetBytes(ret, "result")

	if result.Exists() {
		rets := result.String()

		error := gjson.GetBytes(ret, "error").String()
		if rets == "true" && error == "" {
			log.Printf("pool  --> Share Accepted")
		} else if rets == "false" && error != "" {
			log.Printf("pool  --> Share Rejected with %s", error)
		} else {
			println("pool  -->%s", str)
		}

	} else {
		log.Printf("pool WTF --> %s \n", string(ret))
	}
}

func processMinerMessage(ret []byte) {
	id := gjson.GetBytes(ret, "id")
	str := string(ret)

	if id.String() == "" {
		log.Printf("Unknown miner command: %s", string(ret))
		return
	}

	if id.Int() == 1 { // mining.subscribe
		log.Print("MINER --> sending subscribe request")
	} else if id.Int() == 2 { //mining.authorize

		if strings.Contains(str, "params") && strings.Contains(str, "=") {
			t := gjson.GetBytes(ret, "params.0").String()

			s := strings.Split(t, "=")
			u := strings.Split(s[0], ".")
			if len(u) > 1 {
				user := u[0]
				worker := u[1]
				log.Printf("Found user %s, worker %s", user, worker)
			} else {
				log.Printf("Found user %s, no worker found", u[0])
			}

		}

	} else if id.Int() > 2 {
		if strings.Contains(str, "mining.submit") { // mining.submit
			g := gjson.GetBytes(ret, "params.0").String()
			log.Printf("Miner --> submit share id %d by %s", id.Int(), g)
		}

	} else {
		log.Printf("Miner --> %s \n", string(ret))
	}

}

func processMessage(ret []byte) {
	//var output string = ""
	var authid int64 = 0

	if bytes.Contains(ret, []byte("method")) {
		// logger.Debug("got method: " + gjson.GetBytes(ret, "method").String())

		method := gjson.GetBytes(ret, "method")
		param := gjson.GetBytes(ret, "params")

		if method.String() == "mining.subscribe" { //usually id=1 from miner
			mv := gjson.GetBytes(ret, "params").String()
			logger.Info("Miner subscribe with params:" + mv)
		} else if method.String() == "mining.authorize" && param.Exists() {
			user := param.Array()[0]
			pass := param.Array()[1]

			authid = gjson.GetBytes(ret, "id").Int()
			logger.Info(fmt.Sprintf("Got user: %s/%s id=%d", user.String(), pass.String(), authid))
		} else {
			logger.Debug(string(ret))
		}
	} else {
		logger.Debug(string(ret))
	}
}
