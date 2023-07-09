package suitecrm

import (
	"net/http"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/gorilla/websocket"
)

type wsClient struct {
	conn *websocket.Conn
	mux  sync.Mutex
}

const (
	writeWait  = 10 * time.Second
	pongWait   = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true }, // use default options
}

func (s *suitecrm) wsHandler(w http.ResponseWriter, r *http.Request) {
	cLog := s.log.WithField("api", "ws")

	ext := r.Context().Value(ExtKey{}).(string)
	cLog = cLog.WithFields(log.Fields{"remote_addr": r.RemoteAddr, "ext": ext})

	var c *wsClient
	if conn, err := upgrader.Upgrade(w, r, nil); err == nil { // new websocket connection
		c = &wsClient{conn: conn}

		if s.wsRoom[ext] == nil {
			s.wsRoom[ext] = make(map[*wsClient]bool)
		}

		s.wsRoom[ext][c] = true
	} else {
		cLog.Warn(err)
		return
	}

	cLog.Trace("connected")

	d := make(chan bool)
	defer func() {
		cLog.Trace("disconnected")

		delete(s.wsRoom[ext], c)
		c.conn.Close()

		// stop pinger routine
		d <- true
	}()

	// send active calls
	for _, e := range s.ent {
		e.exts.Range(func(key interface{}, value interface{}) bool {
			s.wsBroadcast(key.(string), true, e)

			return true
		})
	}

	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	// pinger routine
	go func() {
		ticker := time.NewTicker(pingPeriod)

		defer func() {
			ticker.Stop()
			c.conn.Close()
		}()

		for {
			select {
			case <-ticker.C:
				c.mux.Lock()

				c.conn.SetWriteDeadline(time.Now().Add(writeWait))
				err := c.conn.WriteMessage(websocket.PingMessage, nil)

				c.mux.Unlock()

				if err != nil {
					return
				}
			case <-d: // websocket disconnected
				return
			}
		}
	}()

	// read loop is required in order to ping function to work
	for {
		if _, _, err := c.conn.ReadMessage(); err != nil {
			return
		}
	}
}

func (s *suitecrm) wsBroadcast(ext string, show bool, e *entity) {
	if s.wsRoom[ext] == nil || len(s.wsRoom[ext]) == 0 { // no connected sockets for that extension
		return
	}

	var d struct {
		Show bool    `json:"show"`
		Data *entity `json:"data"`
	}

	d.Show = show
	d.Data = e

	for c := range s.wsRoom[ext] {
		c.mux.Lock()

		c.conn.SetWriteDeadline(time.Now().Add(writeWait))
		c.conn.WriteJSON(d)

		c.mux.Unlock()
	}
}
