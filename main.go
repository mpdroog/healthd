package main

import (
	"net/http"
	"fmt"
	"flag"
	"github.com/mpdroog/healthd/config"
	"github.com/mpdroog/healthd/worker"
	"encoding/json"
	"strings"
)

func doc(w http.ResponseWriter, r *http.Request) {
	//
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
	for _, state := range worker.States {
		if !state.Ok {
			err = append(err, state.String())
		}
	}

	var s string
	if len(err) == 0 {
		s = fmt.Sprintf("OK: %d Active nodes.\n", len(worker.States))
	} else {
		w.WriteHeader(500)
		s = fmt.Sprintf(strings.Join(err, ", "))
	}

	w.Header().Set("Content-Type", "text/html")
	w.Header().Set("X-HEALTH", "0.01")
	if _, e := w.Write([]byte(s)); e != nil {
		fmt.Println("zenoss: " + e.Error())
		return
	}
}

func main() {
	flag.StringVar(&config.Confdir, "d", "/etc/healthd/conf.d", "Path to config-files per service")
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
	http.HandleFunc("/health", health)

	if e := worker.Init(); e != nil {
		panic(e)
	}

	if config.Verbose {
		fmt.Println("HTTP listening on :10515")
	}
	if e := http.ListenAndServe(":10515", nil); e != nil {
		panic(e)
	}
}