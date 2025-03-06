package connect

type du struct {
}

func (con *du) Init() {
}

func (con *du) Reload() {
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

func (con *du) End(c *Call, cause string) {
}

func (con *du) SetOriginate(orig OrigFunc) {
}

// NewDummyConnector func
func NewDummyConnector() Connecter {
	c := &du{}
	return c
}
