package actors

import (
	"bytes"
	"errors"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// -- settings --
const (
	writeWait          = 5 * time.Second
	pongWait           = 10 * time.Second
	pingPeriod         = (pongWait * 9) / 10
	maxMessageSize     = 8192
	maxMessagesInQueue = 256
)

// -- end --

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Client points to a connected client.
type HubClient interface {
	//Custom(o models.Serializable, c Client) bool
	//MsgFromUser(b []byte, c UserClient) bool
	//MsgFromHub(b []byte, o Serializable, c HubClient) bool
	//MsgFromHubAsync(o Serializable, c HubClient)
	Custom(o Serializable, c HubClient) bool
	Default(o Serializable, c *OneTimeClient)

	Close()
	GetHubID() string
	Connected() bool
	LogFields() []zapcore.Field
}

type UserClient interface {
	Close()
	GetUserID() int64
	//Custom(o Serializable) bool
	Msg(b []byte) bool
}

// WSClient points to a connected client.
type WSClient struct {
	hub *Hub

	UserID   int64
	UserName string

	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte

	log *zap.Logger
}

// Hub maintains the set of active clients and broadcasts messages to the
// clients.

// IncomingMessage Broadcast message
type IncomingMessage struct {
	B []byte
	J Serializable
	U UserClient
	H HubClient
	D *OneTimeClient
}

// NewWSClient starts a new client connected to the given hub.
func NewWSClient(hub *Hub, userID int64, userName string, w http.ResponseWriter, r *http.Request, logger *zap.Logger) (*WSClient, error) {
	//if userID == 0 {
	//	return nil, errors.New("UserID can not be Nil")
	//}
	conn, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		logger.Error("Upgrade Failed", zap.Error(err))
		return nil, err
	}

	client := &WSClient{
		hub:      hub,
		UserID:   userID,
		UserName: userName,
		conn:     conn,
		send:     make(chan []byte, maxMessagesInQueue),
		log:      logger,
	}
	select {
	case hub.register <- client:
	default:
		client.Close()
		logger.Error("Hub closed while registering", zap.String("hub_id", hub.ID))
		return nil, errors.New("Hub Closed")
	}

	go client.writePump()
	go client.readPump()
	return client, nil
}

func (c *WSClient) Msg(b []byte) bool {
	select {
	case c.send <- b:
		return true
	default:
		go func() {
			c.hub.unregister <- c
		}()
		return false
	}
}

// func (c *WSClient) Custom(o Serializable) bool {
// 	//err := c.hub.UserCustom(c.UserID, o)
// 	//return err == nil
// 	return false
// }

func (c *WSClient) Close() {
	close(c.send)
}

func (c *WSClient) GetUserID() int64 {
	return c.UserID
}

func (c *WSClient) GetUserName() string {
	return c.UserName
}

func (c *WSClient) LogFields() []zapcore.Field {
	return []zapcore.Field{
		zap.String("ctype", "websocket"),
		zap.Int64("user_id", c.GetUserID()),
	}
}

// readPump pumps messages from the websocket connection to the hub.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *WSClient) readPump() {
	defer func() {
		c.log.Debug("WS Close")
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.log.Warn("Unexpected Close", zap.Error(err))
			}
			break
		}
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		c.hub.msgFromUser(message, c)
	}
}

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *WSClient) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
		c.log.Debug("DISCONNECTED")
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				c.log.Debug("Hub Channel Closed")
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				c.log.Warn("Next Writer Return failed")
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				c.log.Warn("Write Close Failed")
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				c.log.Warn("Ping Failed")
				return
			}
		}
	}
}

type OneTimeClient struct {
	Response chan Serializable
	UserID   int
}

func NewOneTimeClient(buffer int) *OneTimeClient {
	c := &OneTimeClient{
		Response: make(chan Serializable, buffer),
	}
	return c
}

func (otc *OneTimeClient) Close() {
	close(otc.Response)
}
func (otc *OneTimeClient) GetUserID() int      { return otc.UserID }
func (otc *OneTimeClient) GetUserName() string { return "" }
func (otc *OneTimeClient) Msg(o Serializable) bool {
	sent := false
	select {
	case otc.Response <- o:
		sent = true
	default:
	}
	return sent
}

func (otc *OneTimeClient) LogFields() []zapcore.Field {
	return []zapcore.Field{}
}

func (otc *OneTimeClient) Read(timeout int) (Serializable, error) {
	t := time.After(time.Duration(timeout) * time.Second)
	select {
	case ret := <-otc.Response:
		return ret, nil
	case <-t:
		return nil, errors.New("Timeout")
	}
}
