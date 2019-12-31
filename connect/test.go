package connect

type du struct {
}

func (con *du) Init(originate OrigFunc) {
}

func (con *du) Start(c *Call) {
}

func (con *du) OrigStart(c *Call, oID string) {
}

func (con *du) Dial(c *Call, ext string) {
}

func (con *du) StopDial(c *Call, ext string) {
}

func (con *du) Answer(c *Call, ext string) {
}

func (con *du) End(c *Call) {
}

// NewDummyConnector func
func NewDummyConnector() Connecter {
	c := &du{}
	return c
}
