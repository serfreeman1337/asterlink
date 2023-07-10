package suitecrm

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

type streamData struct {
	Show bool    `json:"show"`
	Data *entity `json:"data"`
}

type streamsMap map[string][]chan *streamData

func (s *suitecrm) notify(ext string, show bool, e *entity) {
	s.m.Lock()

	chans, ok := s.streams[ext]

	if !ok {
		s.m.Unlock()
		return
	}

	data := streamData{show, e}

	for _, ch := range chans {
		ch <- &data
	}

	s.m.Unlock()
}

func (s *suitecrm) streamHandler(w http.ResponseWriter, r *http.Request) {
	cLog := s.log.WithField("api", "stream")

	ctl := http.NewResponseController(w)

	ext := r.Context().Value(ExtKey{}).(string)
	cLog = cLog.WithFields(log.Fields{"remote_addr": r.RemoteAddr, "ext": ext})

	cLog.Trace("connected")

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	s.m.Lock()

	ch := make(chan *streamData)
	s.streams[ext] = append(s.streams[ext], ch)

	s.m.Unlock()

	defer func() {
		cLog.Trace("disconnected")

		s.m.Lock()

		close(ch)
		chans := s.streams[ext]

		if l := len(s.streams[ext]); l == 1 {
			delete(s.streams, ext)
		} else {
			for i, other := range chans {
				if other == ch {
					chans[i] = chans[l-1]
					break
				}
			}

			s.streams[ext] = chans[:l-1]
		}

		s.m.Unlock()
	}()

	enc := json.NewEncoder(w)

	// Send active calls.
	for _, e := range s.ent {
		e.exts.Range(func(key interface{}, value interface{}) bool {
			if key != ext { // For streamed extension only.
				return true
			}

			ctl.SetWriteDeadline(time.Now().Add(10 * time.Second))

			fmt.Fprint(w, "data: ")
			enc.Encode(streamData{true, e})
			fmt.Fprint(w, "\n\n")

			return true
		})
	}

	ctl.Flush()

	for {
		select {
		case data, ok := <-ch:
			if !ok {
				return
			}

			ctl.SetWriteDeadline(time.Now().Add(10 * time.Second))

			fmt.Fprint(w, "data: ")
			enc.Encode(data)
			fmt.Fprint(w, "\n\n")

			ctl.Flush()
		case <-r.Context().Done(): // Stream has disconnected.
			return
		}
	}
}
