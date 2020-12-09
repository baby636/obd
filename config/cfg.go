package config

import (
	"flag"
	"log"
	"testing"
	"time"

	"github.com/go-ini/ini"
)

var (
	//Cfg               *ini.File
	configPath   = flag.String("configPath", "config/conf.ini", "Config file path")
	ServerPort   = 60020
	ReadTimeout  = 5 * time.Second
	WriteTimeout = 10 * time.Second

	TrackerHost = "localhost:60060"

	ChainNode_Type = "test1"
	//P2P
	P2P_hostIp     = "127.0.0.1"
	P2P_sourcePort = 4001
)

func Init() {
	testing.Init()
	flag.Parse()
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	//Cfg, err := ini.Load("config/conf.ini")
	Cfg, err := ini.Load(*configPath)
	if err != nil {
		log.Println(err)
		return
	}
	section, err := Cfg.GetSection("server")
	if err != nil {
		log.Println(err)
		return
	}
	ServerPort = section.Key("port").MustInt(60020)
	ReadTimeout = time.Duration(section.Key("readTimeout").MustInt(5)) * time.Second
	WriteTimeout = time.Duration(section.Key("writeTimeout").MustInt(5)) * time.Second

	p2pNode, err := Cfg.GetSection("p2p")
	if err != nil {
		log.Println(err)
		return
	}
	P2P_hostIp = p2pNode.Key("hostIp").String()
	P2P_sourcePort = p2pNode.Key("sourcePort").MustInt()

	//tracker
	tracker, err := Cfg.GetSection("tracker")
	if err != nil {
		log.Println(err)
		return
	}
	if len(tracker.Key("host").String()) == 0 {
		panic("empty tracker host")
	}
	TrackerHost = tracker.Key("host").MustString("localhost:60060")
}
