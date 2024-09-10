package handler

import (
	"log"
	"net/http"
)

func HandleWebSocketConnections(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade to WebSocket: %v", err)
		return
	}
	defer ws.Close()

	mu.Lock()
	clients[ws] = true
	mu.Unlock()

	for {
		_, _, err := ws.ReadMessage()
		if err != nil {
			mu.Lock()
			delete(clients, ws)
			mu.Unlock()
			break
		}
	}
}

func HandleMessages() {
	for {
		order := <-broadcast
		log.Printf("%v status is %v", order.AppTransId, order.Status)
		Db[order.AppTransId] = Order{AppTransId: order.AppTransId, Status: order.Status}
		mu.Lock()
		for client := range clients {
			err := client.WriteJSON(order)
			if err != nil {
				log.Printf("Failed to send message: %v", err)
				client.Close()
				delete(clients, client)
			}
		}
		mu.Unlock()
	}
}
