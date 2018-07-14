package main

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"net/http"
	"net/http/httptest"
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
	ts := httptest.NewServer(SetupRoutes("/tailon/"))
	conn := sockjsConnect(t, ts)

	config.FileSpecs = []FileSpec{
		FileSpec{"testdata/ex1/var/log/1.log", "file", "", ""},
		FileSpec{"testdata/ex1/var/log/2.log", "file", "", ""},
		FileSpec{"testdata/ex1/var/log/na.log", "file", "", ""},
	}

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
