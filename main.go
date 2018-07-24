package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/mpdroog/healthd/config"
	"github.com/mpdroog/healthd/worker"
	"net/http"
	"strings"
	"github.com/coreos/go-systemd/daemon"
)

func doc(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Documentation on <a href='https://github.com/mpdroog/healthd'>https://github.com/mpdroog/healthd</a>"))
}
func health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-HEALTH", "0.01")

	b, e := json.Marshal(worker.States)
	if e != nil {
		fmt.Println("config: " + e.Error())
		w.WriteHeader(500)
		w.Write([]byte(`{"msg": "Failed encoding JSON!"}`))
		return
	}

	if _, e := w.Write(b); e != nil {
		fmt.Println("health: " + e.Error())
		return
	}
}
func zenoss(w http.ResponseWriter, r *http.Request) {
	var err []string
	for name, state := range worker.States {
		if !state.Ok {
			err = append(err, fmt.Sprintf("[%s] %s\n", name, state.String()))
		}
	}

	var s string
	if len(err) == 0 {
		s = fmt.Sprintf("OK: %d Active nodes.\n", len(worker.States))
	} else {
		w.WriteHeader(500)
		s = fmt.Sprintf("ERR: " + strings.Join(err, ", "))
	}

	w.Header().Set("Content-Type", "text/html")
	w.Header().Set("X-HEALTH", "0.01")
	if _, e := w.Write([]byte(s)); e != nil {
		fmt.Println("zenoss: " + e.Error())
		return
	}
}

func main() {
	flag.StringVar(&config.Scriptdir, "d", "/etc/healthd/script.d", "Path to scripts to run")
	flag.BoolVar(&config.Verbose, "v", false, "Verbose-mode (log more)")
	flag.Parse()

	if e := config.Init(); e != nil {
		panic(e)
	}
	defer config.Close()

	if config.Verbose {
		fmt.Printf("%+v\n", config.C)
	}

	http.HandleFunc("/", doc)
	http.HandleFunc("/zenoss", zenoss)
	http.HandleFunc("/_mon", zenoss)
	http.HandleFunc("/health", health)

	if e := worker.Init(); e != nil {
		panic(e)
	}

	sent, e := daemon.SdNotify(false, "READY=1")
	if e != nil {
		panic(e)
	}
	if !sent {
		fmt.Printf("SystemD notify NOT sent\n")
	}

	if config.Verbose {
		fmt.Println("HTTP listening on :10515")
	}
	if e := http.ListenAndServe(":10515", nil); e != nil {
		panic(e)
	}
}
