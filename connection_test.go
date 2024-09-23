package main

import (
	"github.com/gorilla/websocket"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHub(t *testing.T) {
	hub := NewHub()
	go hub.run()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		WsHandler(hub, writer, request)
	}))
	defer server.Close()
	wsURL := "ws" + server.URL[len("http"):] + "/ws"
	client1, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatal(err)
		return
	}
	defer client1.Close()

	client2, _, err := websocket.DefaultDialer.Dial(wsURL, nil)

	defer client2.Close()

	textmessage := "hello this is Kerim"

	terr := client1.WriteMessage(websocket.TextMessage, []byte(textmessage))
	if terr != nil {
		t.Fatal(terr)
		return
	}

	_, message, err := client2.ReadMessage()
	if err != nil {
		t.Fatal(err)

	}

	if string(message) == textmessage {
		t.Log("this is true")
	}
	if string(message) != textmessage {
		t.Errorf("Expected message '%s', but got '%s'", textmessage, string(message))
	}
}
