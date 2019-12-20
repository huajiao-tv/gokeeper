package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

var In string
var Out string

func main() {
	flag.StringVar(&In, "in", "", "Input ini file or path if if be a path, the out param will ignore ")
	flag.StringVar(&Out, "out", "", "output ini file, if empty will overwrite the input ini file")
	flag.Parse()
	fi, err := os.Stat(In)
	if err != nil {
		log.Fatalf("invalid in params: %s error: %s\n", In, err)
	}
	if fi.IsDir() {
		if err := fmtDir(In); err != nil {
			log.Fatalf("format path: %s fail error: %s\n", fi.Name(), err)
		}
		fmt.Println("success")
		return
	}
	if Out == "" {
		Out = In
	}
	if err := fmtFile(In, Out); err != nil {
		log.Fatalf("fomate file fail: %s", err.Error())
	}
	log.Println("success")
}
