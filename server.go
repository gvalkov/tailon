package main

import (
	"encoding/json"
	"github.com/gorilla/handlers"
	"github.com/gvalkov/tailon/cmd"
	"gopkg.in/igm/sockjs-go.v2/sockjs"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"time"
)

func SetupRoutes(relativeroot string) *http.ServeMux {
	router := http.NewServeMux()

	sockjsHandler := sockjs.NewHandler(relativeroot+"ws", sockjs.DefaultOptions, wsHandler)
	staticHandler := noCacheControl(http.FileServer(http.Dir("frontend/dist/")))
	staticHandler = http.StripPrefix(relativeroot+"static/", staticHandler)

	router.HandleFunc(relativeroot+"", indexHandler)
	router.Handle(relativeroot+"static/", staticHandler)
	router.Handle(relativeroot+"ws/", sockjsHandler)

	return router
}

func SetupServer(config *Config, logger *log.Logger) *http.Server {
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

func expandCommandArgs(action []string, cmd FrontendCommand) []string {
	var res = make([]string, 0)

	for _, arg := range action {
		switch arg {
		case "$lines":
			res = append(res, strconv.Itoa(cmd.Nlines))
		case "$path":
			res = append(res, cmd.Entry.Path)
		case "$script":
			res = append(res, cmd.Script)
		default:
			res = append(res, arg)
		}
	}

	return res
}

func runCommand(procA *exec.Cmd, procB *cmd.Cmd, session sockjs.Session) {
	if procA != nil {
		procB.Stdin, _ = procA.StdoutPipe()
		procA.Start()
	}

	statusChan := procB.Start()

	for {
		select {
		case line := <-procB.Stdout:
			msg := []string{"o", line}
			data, _ := json.Marshal(msg)
			session.Send(string(data))
		case line := <-procB.Stderr:
			msg := []string{"e", line}
			data, _ := json.Marshal(msg)
			session.Send(string(data))
		case <-statusChan:
		}
	}
}

func wsWriter(session sockjs.Session, c chan string, done <-chan struct{}) {
	var procA *exec.Cmd
	var procB *cmd.Cmd

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

				killProcs(procA, procB)

				// Check if the command is using another command for stdin.
				stdinSource := config.CommandSpecs[msg_json.Command].Stdin
				if stdinSource != "" {
					actionA := config.CommandSpecs[stdinSource].Action
					actionA = expandCommandArgs(actionA, msg_json)
					procA = exec.Command(actionA[0], actionA[1:]...)
					log.Print("Running command: ", actionA)
				}

				actionB := config.CommandSpecs[msg_json.Command].Action
				actionB = expandCommandArgs(actionB, msg_json)
				procB = cmd.NewCmdOptions(cmdOptions, actionB[0], actionB[1:]...)
				log.Print("Running command: ", actionB)

				go runCommand(procA, procB, session)
			}
		case <-done:
			killProcs(procA, procB)
			return
		}
	}
}

func killProcs(procA *exec.Cmd, procB *cmd.Cmd) {
	if procA != nil {
		log.Printf("Stopping pid %d", procA.Process.Pid)
		procA.Process.Kill()
		procA.Wait()
	}

	if procB != nil {
		log.Printf("Stopping pid %d", procB.Status().PID)
		if procB.Stdin != nil {
			procB.Stdin.Close()
		}
		procB.Stop()
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
		} else {
			log.Print(err)
		}
		break
	}
}
