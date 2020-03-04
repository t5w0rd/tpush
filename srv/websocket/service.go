package websocket

import (
	"github.com/gorilla/websocket"
	"net/http"
	log "github.com/micro/go-micro/v2/logger"
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

type Service struct {
}

func NewService() *Service {
	s := &Service{

	}
	return s
}

func (s *Service) handleStream(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
		return
	}
	client := &Client{
		conn: conn,
	}
	go client.readPump()
	go client.writePump()
}

func (s *Service) Run() error {
	http.HandleFunc("/stream", s.handleStream)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe err: ", err)
		return err
	}
	return nil
}
