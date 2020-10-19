package connect

import (
	"time"

	log "github.com/sirupsen/logrus"
)

// Direction int
type Direction int

// Direction out or in
const (
	In Direction = iota
	Out
)

// Call struct
type Call struct {
	LID        string
	Dir        Direction
	CID        string
	DID        string
	Ext        string
	TimeCall   time.Time
	TimeDial   time.Time
	TimeAnswer time.Time
	Ch         string
	ChDest     string
	Rec        string
	Vote       string
	O          bool
	Log        *log.Entry
}

// OrigFunc ?
type OrigFunc func(ext string, dest string, oID string)

// Connecter interface
type Connecter interface {
	Init()
	Start(call *Call)
	OrigStart(call *Call, oID string)
	Dial(call *Call, ext string)
	StopDial(call *Call, ext string)
	Answer(call *Call, ext string)
	End(call *Call, cause string)
}
