package main

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

var updgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Client struct {
	conn *websocket.Conn
	hub  *Hub
	send chan []byte
}

type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (c *Client) ReadPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	for {
		_, messagage, err := c.conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}
		c.hub.broadcast <- messagage
	}
}

func (c *Client) WritePump() {
	for message := range c.send {
		err := c.conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			log.Println(err)
			return
		}
	}
}

func WsHandler(hub *Hub, w http.ResponseWriter, r *http.Request) {
	connection, _ := updgrader.Upgrade(w, r, nil)
	client := &Client{conn: connection, hub: hub, send: make(chan []byte, 256)}
	hub.register <- client
	go client.WritePump()
	go client.ReadPump()
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case message := <-h.broadcast:
			for client := range h.clients {
				client.send <- message
			}
		}
	}
}
