//go:generate go run frontend_gen.go

package main

import (
	"encoding/json"
	"github.com/gorilla/handlers"
	"github.com/gvalkov/tailon/cmd"
	"github.com/shurcooL/httpfs/html/vfstemplate"
	"github.com/shurcooL/httpgzip"
	"github.com/igm/sockjs-go/v3/sockjs"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"time"
)

func setupRoutes(relativeroot string) *http.ServeMux {
	router := http.NewServeMux()

	// Use either "frontend_dev.go" or "frontend_bin.go", depending on the "dev" build tag.
	staticHandler := noCacheControl(httpgzip.FileServer(FrontendAssets, httpgzip.FileServerOptions{IndexHTML: true}))

	sockjsHandler := sockjs.NewHandler(relativeroot+"ws", sockjs.DefaultOptions, wsHandler)
	staticHandler = http.StripPrefix(relativeroot+"vfs/", staticHandler)

	router.Handle(relativeroot+"vfs/", staticHandler)
	router.Handle(relativeroot+"ws/", sockjsHandler)
	router.HandleFunc(relativeroot+"files/", downloadHandler)
	router.HandleFunc(relativeroot+"", indexHandler)

	return router
}

func setupServer(config *Config, addr string, logger *log.Logger) *http.Server {
	router := setupRoutes(config.RelativeRoot)
	loggingRouter := handlers.LoggingHandler(os.Stderr, router)

	server := http.Server{
		Addr:         addr,
		Handler:      loggingRouter,
		ErrorLog:     logger,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	return &server
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	t := template.Must(vfstemplate.ParseFiles(FrontendAssets, nil, "/templates/base.html", "/templates/tailon.html"))
	t.Execute(w, config)
}

func downloadHandler(w http.ResponseWriter, r *http.Request) {
	if !config.AllowDownload {
		http.Error(w, "downloads forbidden by server", http.StatusForbidden)
		return
	}

	path := r.URL.Query()["path"][0]
	if !fileAllowed(path) {
		log.Printf("warn: attempt to access unknown file: %s", path)
		http.Error(w, "unknown file", http.StatusNotFound)
		return
	}
	http.ServeFile(w, r, path)
}

func noCacheControl(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
		h.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}

// FrontendCommand instances are the messages that the client sends to the server when the file, tool or script change.
type FrontendCommand struct {
	Command string
	Script  string
	Entry   ListEntry
	Nlines  int
}

// The main sockjs handler.
func wsHandler(session sockjs.Session) {
	messages := make(chan string)
	done := make(chan struct{})
	defer close(done)

	go wsWriter(session, messages, done)

	for {
		if msg, err := session.Recv(); err == nil {
			messages <- msg
			continue
		} else {
			log.Print(err)
		}
		break
	}
}

// Goroutine handling received messages and streaming of file contents.
func wsWriter(session sockjs.Session, messages chan string, done <-chan struct{}) {
	// The processes that make up the pipeline. The stdout of procA is connected to the stdin of procB.
	var procA *exec.Cmd
	var procB *cmd.Cmd

	cmdOptions := cmd.Options{Buffered: false, Streaming: true}

	for {
		select {
		case msg := <-messages:
			if msg == "list" {
				lst := createListing(config.FileSpecs)
				b, err := json.Marshal(lst)
				if err != nil {
					log.Println("error: ", err)
				}
				session.Send(string(b))
			} else if msg[0] == '{' {
				msgJSON := FrontendCommand{}
				json.Unmarshal([]byte(msg), &msgJSON)

				if !fileAllowed(msgJSON.Entry.Path) {
					log.Print("Unknown file: ", msgJSON.Entry.Path)
					continue
				}

				killProcs(procA, procB)

				// Check if the command is using another command for stdin.
				stdinSource := config.CommandSpecs[msgJSON.Command].Stdin
				if stdinSource != "" {
					actionA := config.CommandSpecs[stdinSource].Action
					actionA = expandCommandArgs(actionA, msgJSON)
					procA = exec.Command(actionA[0], actionA[1:]...)
					log.Print("Running command: ", actionA)
				}

				actionB := config.CommandSpecs[msgJSON.Command].Action
				actionB = expandCommandArgs(actionB, msgJSON)
				procB = cmd.NewCmdOptions(cmdOptions, actionB[0], actionB[1:]...)
				log.Print("Running command: ", actionB)

				// Start streaming procB's stdout and stderr to the client.
				go streamOutput(procA, procB, session)
			}
		case <-done:
			killProcs(procA, procB)
			return
		}
	}
}

// Expands the variables in main.CommandSpec.Action with the values in the
// frontend command. For example:
//    ["tail", "-n", "$lines", "-F", "$path"] -> ["tail", "-n", "10", "-F", "f1.txt"]
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

// Goroutine that streams command stdout and stderr to the client.
func streamOutput(procA *exec.Cmd, procB *cmd.Cmd, session sockjs.Session) {
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
