package main

import (
	"strings"
	"time"
	"runtime"
	"fmt"
	// "os"
	log "github.com/Sirupsen/logrus"
	"github.com/jessevdk/go-flags"
)


var VERSION string = "0.0.1"

var opts struct {
	Version   string `short:"v" long:"version" default:"0.1.0" description:"get version"`
	LogLevel  string `short:"l" long:"log_level" default:"info" description:"log level,Valid options are: error, warn, info, debug"`
	Command   string `short:"c" long:"command" default:"main"`
	Filter string `short:"f" long:"filter" default:"ctn" description:"get ctn filter exchange ctn&cun"`
	InPut	 string `short:"i" long:"input" default:"input file"`
	OutPut	 string `short:"o" long:"output" default:"output file"`
	GeonameID string `short:"g" long:"id" default:"16779264"`
	CountryCode string `short:"d" long:"code" default:"CN"`
	CountryName string `short:"n" long:"name" default:"China"`
	Type			string `short:"t" long:"type" default:"mask" dedcription:"ipip ipmask type "`

	// Path      string `short:"p" long:"path" default:"." description:"program path"`
	// CpePath		string `short:"c" long:"cpe_path" default:"" description:"cpe file path"`
}
func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	fmt.Println("---- init cpu init.go");
	
}
func init(){
	// _, err := flags.Parse(&opts)
	// if err != nil {
	// 	if !strings.Contains(err.Error(), "Usage") {
	// 		log.Fatalf("error: %v", err)
	// 	} else {
	// 		return
	// 	}
	// }
	// fmt.Println("---- init parse init.go");
	
	parser := flags.NewParser(&opts, flags.Default|flags.IgnoreUnknown)
	parser.Parse()


	// log.Infof("opts: %+v", opts)	
	// log.Infof("version: %v",opts.Version)
	// log.Infof("loglevel: %s",opts.LogLevel)
	// // log.Info("HttpAddr:",opts.HttpAddr)
	// log.Infof("Command: %s",opts.Command)
	// log.Infof("CpePath: %s",opts.CpePath)
	// log.Infof("Routebin: %s",opts.Routebin)

}

// func init(){
// 	fmt.Println("---- init in init.go");
// }

func init() {
	// log = logrus.New()
	// log.Level = logrus.InfoLevel
	// f := new(logrus.TextFormatter)
	// f.TimestampFormat = "2006-01-02 15:04:05"
	// f.FullTimestamp = true
	// log.Formatter = f
	// fmt.Println("---- init log init.go ",opts.LogLevel);

	level,err := log.ParseLevel(strings.ToLower(opts.LogLevel))
	// level,err := log.ParseLevel("INFO")
	
	if err != nil {
		log.Fatalf("log level error: %v",err)
	}
	log.SetLevel(level)
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
		TimestampFormat: time.RFC3339,
	})

	// f, err := os.OpenFile("/tmp/upwan.log", os.O_WRONLY | os.O_CREATE | os.O_APPEND, 0755)
	// if err == nil {
	// 	log.SetOutput(f)
	// }

}