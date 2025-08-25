package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gorilla/websocket"
	"github.com/spf13/cobra"
)

var server, room, token string

type Message struct {
	Room      string `json:"room"`
	User      string `json:"user"`
	Device    string `json:"device"`
	Text      string `json:"text"`
	Timestamp int64  `json:"timestamp"`
}

func main() {
	root := &cobra.Command{
		Use:   "chat-client",
		Short: "CLI chat client",
		Run:   run,
	}
	root.Flags().StringVarP(&server, "server", "s", "ws://localhost:8080/ws", "WS server")
	root.Flags().StringVarP(&room, "room", "r", "", "Room UUID")
	root.Flags().StringVarP(&token, "token", "t", "", "JWT token")
	root.MarkFlagRequired("room")
	root.MarkFlagRequired("token")
	root.Execute()
}

func run(cmd *cobra.Command, args []string) {
	url := fmt.Sprintf("%s/%s", server, room)
	header := map[string][]string{"Authorization": {"Bearer " + token}}

	conn, _, err := websocket.DefaultDialer.Dial(url, header)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	fmt.Println("Connected to room:", room)
	go func() {
		for {
			_, msg, _ := conn.ReadMessage()
			fmt.Println(string(msg))
		}
	}()

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		text := scanner.Text()
		if text == "/exit" {
			break
		}
		m := Message{Text: text}
		data, _ := json.Marshal(m)
		_ = conn.WriteMessage(websocket.TextMessage, data)
		time.Sleep(10 * time.Millisecond)
	}
}
