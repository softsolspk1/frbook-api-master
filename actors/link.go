package actors

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func HubLinks(logger *zap.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		linksLock.RLock()
		defer linksLock.RUnlock()
		var links []*HubLink
		for _, link := range links {
			links = append(links, link)
		}

		b, err := json.Marshal(links)
		if err != nil {
			logger.Error("Error Marshalling Links", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Write(b)
	})
}

func NewCustomOutgoingLink(src string, dest string, url string, subprotocols []string, logger *zap.Logger) (*HubLink, error) {
	srcHub := HubById(src)
	if srcHub == nil {
		return nil, fmt.Errorf("Source Hub %s not found", src)
	}
	hl := HubLink{
		SrcID:     src,
		DestID:    dest,
		Type:      "outgoing",
		Url:       url,
		Protocols: subprotocols,
		C:         make(chan []byte, 100),
		H:         srcHub,
		log:       logger.With(zap.String("type", "link-out"), zap.String("src", src), zap.String("dest", dest)),
	}

	installLink(&hl)

	go hl.Connect()

	return &hl, nil
}

func NewOutgoingHubLink(src string, dest string, host string, port int, sugar string, logger *zap.Logger) (*HubLink, error) {
	srcHub := HubById(src)
	if srcHub == nil {
		return nil, fmt.Errorf("Source Hub %s not found", src)
	}

	var url string
	if port == 443 {
		url = fmt.Sprintf("wss://%s/ws/%s", host, dest)
	} else {
		url = fmt.Sprintf("ws://%s:%d/ws/%s", host, port, dest)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"hub_id": src,
	})

	signedToken, err := token.SignedString([]byte(sugar))
	if err != nil {
		return nil, err
	}

	url = fmt.Sprintf("%s?__jwt=%s", url, signedToken)

	hl := HubLink{
		SrcID:  src,
		DestID: dest,
		Type:   "outgoing",
		Url:    url,
		C:      make(chan []byte, 100),
		H:      srcHub,
		log:    logger.With(zap.String("type", "link-out"), zap.String("src", src), zap.String("dest", dest)),
	}

	installLink(&hl)

	go hl.Connect()
	return &hl, nil
}

func NewIncomingHubLink(src string, dest string, logger *zap.Logger) (*HubLink, error) {
	destHub := HubById(dest)
	hl := HubLink{
		SrcID:  dest,
		DestID: src,
		Type:   "incoming",
		C:      make(chan []byte, 100),
		H:      destHub,
		log:    logger.With(zap.String("type", "link-in"), zap.String("src", src), zap.String("dest", dest)),
	}

	installLink(&hl)

	return &hl, nil
}

func (hl *HubLink) Connected() bool {
	return hl.conn != nil
}

func (hl *HubLink) Accept(conn *websocket.Conn) {
	// If we already have a connection.
	if hl.conn != nil {
		hl.conn.Close()
	}
	hl.conn = conn
	hl.LastConnected = time.Now()
	go hl.readPump(conn)
	go hl.writePump(conn)
}

func (hl *HubLink) Connect() error {
	if hl.closed {
		return nil
	}
	// what if hl.conn is not nil ?
	hl.conn = nil

	// TODO: Improve retry logic.
	if time.Now().Sub(hl.LastConnected) < 5*time.Second {
		time.Sleep(5 * time.Second)
	}
	hl.LastConnected = time.Now()
	dialer := &websocket.Dialer{
		ReadBufferSize:  4096,
		WriteBufferSize: 4096,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
		Subprotocols: hl.Protocols,
	}

	// TODO(harsh) :- Handle Security
	conn, _, err := dialer.Dial(hl.Url, nil)

	if err != nil {
		hl.log.Warn("Reconnecting", zap.Error(err))
		return hl.Connect()
	}

	go hl.readPump(conn)
	go hl.writePump(conn)

	hl.log.Info("Connected")

	hl.conn = conn
	return nil
}

func (hl *HubLink) readPump(conn *websocket.Conn) {
	defer func() {
		hl.log.Info("Exiting connected while reading")
		hl.conn = nil
		if hl.Url != "" {
			hl.Connect()
		}
		conn.Close()
	}()
	conn.SetReadLimit(4096)
	conn.SetReadDeadline(time.Now().Add(30 * time.Second))
	conn.SetPongHandler(func(string) error {
		hl.LastPong = time.Now()
		conn.SetReadDeadline(time.Now().Add(30 * time.Second))
		return nil
	})
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				hl.log.Error("Unexpected Close", zap.Error(err))
			}
			break
		}

		hl.H.auditIncomingLink(hl, message)

		//message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))

		if !hl.H.bytes(message, hl) {
			hl.log.Warn("Dropping message")
		}

	}
}

func (hl *HubLink) writePump(conn *websocket.Conn) {
	ticker := time.NewTicker(10 * time.Second)
	defer func() {
		ticker.Stop()
		conn.Close()
	}()
	for {
		select {
		case message, ok := <-hl.C:
			{
				conn.SetWriteDeadline(time.Now().Add(60 * time.Second))
				if !ok {
					// The hub closed the channel.
					conn.WriteMessage(websocket.CloseMessage, []byte{})
					hl.log.Error("Hub Channel Closed")
					return
				}

				w, err := conn.NextWriter(websocket.BinaryMessage)
				if err != nil {
					hl.log.Warn("Next Writer Return failed", zap.Error(err))
					return
				}
				w.Write(message)
				hl.H.auditOutgoingLink(hl, message)

				if err := w.Close(); err != nil {
					hl.log.Warn("Write Close Failed")
					return
				}
			}
		case <-ticker.C:
			{
				conn.SetWriteDeadline(time.Now().Add(60 * time.Second))
				if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					hl.log.Warn("Ping Failed", zap.Error(err))
					return
				}
				hl.LastPing = time.Now()
			}
		}
	}
}

func (hl *HubLink) Custom(o Serializable, c HubClient) bool {
	return hl.MsgFromHub(nil, o, c)
}

func (hl *HubLink) Default(o Serializable, c *OneTimeClient) {

}

func (hl *HubLink) MsgFromHub(b []byte, o Serializable, c HubClient) bool {
	if b == nil && o == nil {
		return false
	}
	if b == nil {
		var err error
		b, err = o.Serialize()
		if err != nil {
			return false
		}
	}
	select {
	case hl.C <- b:
		return true
	default:
		return false
	}
}

func (hl *HubLink) MsgFromHubAsync(o Serializable, c HubClient) {
	go hl.MsgFromHub(nil, o, c)
}

func (hl *HubLink) MsgFromUser(b []byte, c UserClient) bool {
	return false
}

func (hl *HubLink) Close() {
	hl.lock.Lock()
	defer hl.lock.Unlock()
	hl.closed = true
	if hl.conn != nil {
		hl.conn.Close()
	}
	RemoveLink(hl)
}

func (hl *HubLink) GetHubID() string {
	return hl.DestID
}

func (hl *HubLink) LogFields() []zapcore.Field {
	return []zapcore.Field{}
}
