package main

import (
	"encoding/json"
	"github.com/go-cmd/cmd"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"gopkg.in/igm/sockjs-go.v2/sockjs"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

func SetupRoutes(relativeroot string) *mux.Router {
	router := mux.NewRouter()
	prefix := router.PathPrefix(relativeroot).Subrouter()

	sockjsHandler := sockjs.NewHandler("/ws", sockjs.DefaultOptions, wsHandler)
	staticHandler := noCacheControl(http.FileServer(http.Dir("frontend/dist/")))
	staticHandler = http.StripPrefix(config.RelativeRoot+"static/", staticHandler)

	prefix.HandleFunc("/", indexHandler)
	prefix.PathPrefix("/static/").Handler(staticHandler)
	prefix.Handle("/ws/", sockjsHandler)
	prefix.Path("/ws/{any:.*}").Handler(sockjsHandler)

	return router
}

func SetupServer(config Config, logger *log.Logger) *http.Server {
	router := SetupRoutes(config.RelativeRoot)
	loggingRouter := handlers.LoggingHandler(os.Stderr, router)

	server := http.Server{
		Addr:         config.BindAddr,
		Handler:      loggingRouter,
		ErrorLog:     logger,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	return &server
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.ParseFiles("templates/base.html", "templates/tailon.html"))
	t.Execute(w, config)
}

func noCacheControl(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
		h.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}

type FrontendCommand struct {
	Command string
	Script  string
	Entry   ListEntry
	Nlines  int
}

func getCommandArgs(action []string, cmd FrontendCommand) []string {
	var res = make([]string, 0)

	for _, arg := range action {
		switch arg {
		case "$lines":
			res = append(res, strconv.Itoa(cmd.Nlines))
		case "$path":
			res = append(res, cmd.Entry.Path)
		default:
			res = append(res, arg)
		}
	}

	return res
}

func runCommand(proc *cmd.Cmd, session sockjs.Session) {
	statusChan := proc.Start()

	for {
		select {
		case line := <-proc.Stdout:
			msg := []string{"o", line}
			data, _ := json.Marshal(msg)
			session.Send(string(data))
		case line := <-proc.Stderr:
			msg := []string{"e", line}
			data, _ := json.Marshal(msg)
			session.Send(string(data))
		case <-statusChan:
		}
	}
}

func wsWriter(session sockjs.Session, c chan string, done <-chan struct{}) {
	var proc *cmd.Cmd

	cmdOptions := cmd.Options{Buffered: false, Streaming: true}

	for {
		select {
		case msg := <-c:
			if msg == "list" {
				lst := createListing(config.FileSpecs)
				b, err := json.Marshal(lst)
				if err != nil {
					log.Println("error: ", err)
				}
				session.Send(string(b))
			} else if msg[0] == '{' {
				msg_json := FrontendCommand{}
				json.Unmarshal([]byte(msg), &msg_json)

				if proc != nil {
					log.Printf("Stopping pid %d", proc.Status().PID)
					proc.Stop()
				}

				action := config.CommandSpecs[msg_json.Command].Action
				action = getCommandArgs(action, msg_json)
				proc = cmd.NewCmdOptions(cmdOptions, action[0], action[1:]...)

				log.Print("Running command: ", action)
				go runCommand(proc, session)
			}
		case <-done:
			if proc != nil {
				log.Printf("Stopping pid %d", proc.Status().PID)
				proc.Stop()
			}
			return
		}
	}
}

func wsHandler(session sockjs.Session) {
	c := make(chan string)
	done := make(chan struct{})
	defer close(done)

	go wsWriter(session, c, done)

	for {
		if msg, err := session.Recv(); err == nil {
			c <- msg
			continue
		}
		break
	}
}
