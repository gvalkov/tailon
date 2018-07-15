package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestRelativeRoot(t *testing.T) {
	r1 := SetupRoutes("/tailon/")
	assertHttpCode(t, r1, "/tailon/", 200)
	assertHttpCode(t, r1, "/tailon/ws/", 200)
	assertHttpCode(t, r1, "/tailon/static/favicon.ico", 200)

	r2 := SetupRoutes("/")
	assertHttpCode(t, r2, "/", 200)
	assertHttpCode(t, r2, "/ws/", 200)
	assertHttpCode(t, r2, "/static/favicon.ico", 200)
}

func TestSockjsList(t *testing.T) {
	setupConfig()
	ts := httptest.NewServer(SetupRoutes("/tailon/"))
	conn := sockjsConnect(t, ts)
	defer ts.Close()

	if err := conn.WriteJSON([]string{"list"}); err != nil {
		t.Fatal(err)
	}

	v := map[string][]ListEntry{}
	conn.ReadMessage()
	sockjsReadJSON(t, conn, &v)
	if len(v["__default__"]) != 3 {
		t.Fatal("len() != 3")
	}
}

func TestFrontendMessage(t *testing.T) {
	setupConfig()
	ts := httptest.NewServer(SetupRoutes("/tailon/"))
	conn := sockjsConnect(t, ts)
	defer ts.Close()

	msg_tail := `{"command":"tail","script":null,"entry":{"path":"testdata/ex1/var/log/1.log","alias":"/tmp/t1","size":14342,"mtime":"2018-07-14T15:07:33.524768369+02:00","exists":true},"nlines":10}`

	msg_grep := `{"command":"grep","script":".*","entry":{"path":"testdata/ex1/var/log/1.log","alias":"/tmp/t1","size":14342,"mtime":"2018-07-14T15:07:33.524768369+02:00","exists":true},"nlines":10}`

	// Run tail on file - there should be only 1 tail proc.
	if err := conn.WriteJSON([]string{msg_tail}); err != nil {
		t.Fatal(err)
	}

	procs := getChildProcs()
	if len(procs) != 1 {
		t.Fatal(procs)
	}

	// Run grep on file - there should be 1 tail and 1 grep procs.
	if err := conn.WriteJSON([]string{msg_grep}); err != nil {
		t.Fatal(err)
	}

	procs = getChildProcs()
	if len(procs) != 2 {
		t.Fatal(procs)
	}

	// Run tail again - the grep proc should have been killed and there should be 1 tail proc.
	if err := conn.WriteJSON([]string{msg_tail}); err != nil {
		t.Fatal(err)
	}

	procs = getChildProcs()
	if len(procs) != 1 {
		t.Fatal(procs)
	}
}

func sockjsReadJSON(t *testing.T, conn *websocket.Conn, v interface{}) {
	_, msg, err := conn.ReadMessage()
	if err != nil {
		t.Fatal(err)
	}

	a := []string{}
	msg = msg[1:len(msg)]
	if err := json.Unmarshal(msg, &a); err != nil {
		t.Fatal(err)
	}

	if err := json.Unmarshal([]byte(a[0]), &v); err != nil {
		t.Fatal(err)
	}
}

func sockjsConnect(t *testing.T, ts *httptest.Server) *websocket.Conn {
	url := "ws" + ts.URL[4:] + "/tailon/ws/0/0/websocket"
	conn, _, err := websocket.DefaultDialer.Dial(url, map[string][]string{"Origin": []string{ts.URL}})

	if err != nil {
		t.Fatal(err)
	}

	return conn
}

func assertHttpCode(t *testing.T, routes *http.ServeMux, path string, code int) {
	ts := httptest.NewServer(routes)
	defer ts.Close()

	res, err := http.Get(ts.URL + path)

	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != code {
		t.Fatalf("GET %s: %d != %d", path, res.StatusCode, code)
	}
}

func setupConfig() {
	config = makeConfig()
	config.FileSpecs = []FileSpec{
		FileSpec{"testdata/ex1/var/log/1.log", "file", "", ""},
		FileSpec{"testdata/ex1/var/log/2.log", "file", "", ""},
		FileSpec{"testdata/ex1/var/log/na.log", "file", "", ""},
	}
}

func getChildProcs() map[string]string {
	procs := make(map[string]string)
	out, _ := exec.Command("sh", "-c", fmt.Sprintf("pgrep -l -P %d", os.Getpid())).Output()

	for _, line := range strings.Split(strings.Trim(string(out), "\n"), "\n") {
		parts := strings.Split(line, " ")
		procs[parts[1]] = parts[0]
	}

	return procs
}
