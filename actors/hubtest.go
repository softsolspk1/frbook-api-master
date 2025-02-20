package actors

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type DummyUser struct {
	ID       int64
	Outgoing chan []byte
	// -- user-fields --
	// -- end --

}

func NewDummyUser(id int64) *DummyUser {
	return &DummyUser{
		ID:       id,
		Outgoing: make(chan []byte, 100),
	}
}

func (du *DummyUser) Msg(b []byte) bool {
	sent := false
	select {
	case du.Outgoing <- b:
		sent = true
	default:
	}
	return sent
}

func (du *DummyUser) Close() {
	close(du.Outgoing)
}

func (du *DummyUser) Custom(o Serializable) bool {
	b, _ := o.Serialize()
	return du.Msg(b)
}

func (du *DummyUser) GetUserID() int64    { return du.ID }
func (du *DummyUser) GetUserName() string { return "" }

func (du *DummyUser) Drain() [][]byte {
	var ret [][]byte
	for {
		select {
		case m := <-du.Outgoing:
			{
				ret = append(ret, m)
			}
		default:
			return ret
		}
	}
}
func (du *DummyUser) LogFields() []zapcore.Field {
	return []zapcore.Field{}
}

func NewDummyHub(id string, buffer int, logger *zap.Logger) *Hub {
	h := &Hub{
		ID:       id,
		incoming: make(chan *IncomingMessage, buffer),
		register: make(chan UserClient, buffer/2),
		clients:  make(map[int64][]*UserClientWrapper),
		log:      logger,
	}
	installHub(id, h)

	return h
}

func (h *Hub) RegisterDummy(u *DummyUser) {
	h.clients[u.ID] = append(h.clients[u.ID], &UserClientWrapper{
		C:      u,
		userID: u.ID,
	})
}

// -- code --
// -- end --
