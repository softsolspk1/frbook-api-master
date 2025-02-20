package actors

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type pendingMsg struct {
	id int
	b  []byte
	t  time.Time
}

type UserClientWrapper struct {
	userID         int64
	sessionID      int
	C              UserClient
	nextMsgID      int
	disconnectedTs time.Time
	reconnects     int

	pending []pendingMsg
}

func (uc *UserClientWrapper) initialMessages(after int) ([]pendingMsg, error) {
	firstMsgId := 0
	lastMsgId := 0
	if len(uc.pending) > 0 {
		firstMsgId = uc.pending[0].id
		lastMsgId = uc.pending[len(uc.pending)-1].id
	}
	if after < firstMsgId-1 {
		return nil, fmt.Errorf("Invalid after %d", after)
	}
	if after > lastMsgId {
		return nil, fmt.Errorf("Invalid after %d", after)
	}
	return uc.pending[after-firstMsgId+1:], nil
}

type Hub struct {
	ID string

	// Registered clients
	clients map[int64][]*UserClientWrapper

	// Inbound messages from the clients.
	incoming chan *IncomingMessage

	// Register requests from the clients.
	register chan UserClient

	// Unregister requests from clients.
	done         chan bool
	unregister   chan UserClient
	controller   Controller
	ticker       *time.Ticker
	tickChan     chan time.Time
	lastTick     time.Time
	log          *zap.Logger
	tickDuration time.Duration
	shutdownWg   *sync.WaitGroup

	kickChan chan int64

	statusChan chan *StatusQuery

	nextSessionID int

	auditChan chan string
	audit     bool
}

type StatusQuery struct {
	R chan string
}

// func (h *Hub) TestAccessClients() map[Client]bool {
// 	var ret = make(map[Client]bool)
// 	for _, v := range h.clients {
// 		//for _, c := range v {
// 		ret[v.C] = true
// 		//}
// 	}
// 	return ret
// }

func (h *Hub) TickDuration() time.Duration {
	return h.tickDuration
}

func (h *Hub) Start() {
	go h.run()
}

func (h *Hub) Tick(dt time.Duration) {
	if h.lastTick.IsZero() {
		h.lastTick = time.Now()
		h.tickChan <- h.lastTick
	} else {
		h.lastTick = h.lastTick.Add(dt)
		h.tickChan <- h.lastTick
	}
}

func (h *Hub) Connected() bool {
	return true
}

func (h *Hub) getWrapper(userId int64, C UserClient) *UserClientWrapper {
	wrapper, ok := h.clients[userId]
	if !ok {
		return nil
	}
	for _, w := range wrapper {
		if w.C == C {
			return w
		}
	}
	return nil
}

func (h *Hub) StartAudit(folder string) {
	os.MkdirAll(folder, os.ModePerm)
	h.audit = true
	h.auditChan = make(chan string, 1000)
	h.log.Info("Starting Audit", zap.String("folder", folder))
	go h.auditThread(folder, h.ID)
}

func (h *Hub) auditThread(folder string, hubID string) {
	f, _ := os.OpenFile(filepath.Join(folder, fmt.Sprintf("%s.log", hubID)), os.O_APPEND|os.O_CREATE|os.O_WRONLY, os.ModePerm)
	lastSync := time.Now()
	for b := range h.auditChan {
		f.WriteString(b)
		if time.Since(lastSync).Seconds() > 10 {
			f.Sync()
			lastSync = time.Now()
		}
	}
}

func (h *Hub) auditBytes(b string) {
	select {
	case h.auditChan <- b:
	default:
	}
}

func (h *Hub) auditOutgoingLink(link *HubLink, msg []byte) {
	if !h.audit {
		return
	}
	auditText := fmt.Sprintf("%s - OUTGOING-LINK\n", time.Now().Format("15:04:05"))
	auditText += fmt.Sprintf("DEST=%s\n", link.DestID)
	auditText += fmt.Sprintf("RAW=%s\n", msg)
	auditText += "=======\n"
	h.auditBytes(auditText)
}

func (h *Hub) auditIncomingLink(link *HubLink, msg []byte) {
	if !h.audit {
		return
	}
	auditText := fmt.Sprintf("%s - INCOMING-LINK\n", time.Now().Format("15:04:05"))
	auditText += fmt.Sprintf("SRC=%s\n", link.SrcID)
	auditText += fmt.Sprintf("RAW=%s\n", msg)
	auditText += "=======\n"
	h.auditBytes(auditText)
}

func (h *Hub) auditIncomingUser(m *IncomingMessage) {
	if !h.audit {
		return
	}
	if m.U == nil {
		return
	}

	auditText := fmt.Sprintf("%s - INCOMING-USER\n", time.Now().Format("15:04:05"))
	auditText += fmt.Sprintf("USERID=%d\n", m.U.GetUserID())
	s := strings.TrimSpace(string(m.B))
	auditText += fmt.Sprintf("RAW=%s\n", s)
	if strings.HasPrefix(s, "#") {
		auditText += fmt.Sprintf("SYSTEM=YES\n")
	} else {
		o, err := h.controller.ParseUser(m.B)
		if err != nil {
			auditText += fmt.Sprintf("PARSE=FAILED\n")
		} else {
			b, _ := json.MarshalIndent(o, "", "  ")
			auditText += "JSON=\n"
			auditText += string(b)
			auditText += "\n"
		}
	}
	auditText += "=======\n"
	h.auditBytes(auditText)
}

func (h *Hub) auditUser(id int64, o Serializable) {
	if !h.audit {
		return
	}

	auditText := fmt.Sprintf("%s - USER\n", time.Now().Format("15:04:05"))
	auditText += fmt.Sprintf("USERID=%d\n", id)
	b, _ := json.MarshalIndent(o, "", "  ")
	auditText += "JSON=\n"
	auditText += string(b)
	auditText += "\n"
	auditText += "=======\n"
	h.auditBytes(auditText)
}

func (h *Hub) auditBroadcast(o Serializable, ignoreUserId int64) {
	if !h.audit {
		return
	}

	auditText := fmt.Sprintf("%s - BROADCAST\n", time.Now().Format("15:04:05"))
	if ignoreUserId != 0 {
		auditText += fmt.Sprintf("IGNORE=%d\n", ignoreUserId)
	}
	b, _ := json.MarshalIndent(o, "", "  ")
	auditText += "JSON=\n"
	auditText += string(b)
	auditText += "\n"
	auditText += "=======\n"
	h.auditBytes(auditText)
}

func (h *Hub) auditWrapperCleanup(w *UserClientWrapper) {
	if !h.audit {
		return
	}

	auditText := fmt.Sprintf("%s - WRAPPER-CLEANUP\n", time.Now().Format("15:04:05"))
	auditText += fmt.Sprintf("USERID=%d\n", w.userID)
	auditText += fmt.Sprintf("DISCONNECTED=%s\n", w.disconnectedTs.Format("15:04:05"))
	auditText += fmt.Sprintf("SESSIONID=%d\n", w.sessionID)
	auditText += fmt.Sprintf("PENDING=%d\n", len(w.pending))
	auditText += fmt.Sprintf("RECONNECTS=%d\n", w.reconnects)
	auditText += "======\n"

	h.auditBytes(auditText)
}

func (h *Hub) auditIncomingDefault(m *IncomingMessage) {
	if !h.audit {
		return
	}
	if m.D == nil {
		return
	}

	auditText := fmt.Sprintf("%s - INCOMING-DEFAULT\n", time.Now().Format("15:04:05"))
	b, _ := json.MarshalIndent(m.J, "", "  ")
	auditText += "JSON=\n"
	auditText += string(b)
	auditText += "\n"
	auditText += "=======\n"
	h.auditBytes(auditText)
}

func (h *Hub) auditIncomingHub(m *IncomingMessage) {
	if !h.audit {
		return
	}
	if m.H == nil {
		return
	}

	auditText := fmt.Sprintf("%s - INCOMING-HUB\n", time.Now().Format("15:04:05"))
	auditText += fmt.Sprintf("HUBID=%s\n", m.H.GetHubID())
	o := m.J
	if o == nil {
		s := strings.TrimSpace(string(m.B))
		auditText += fmt.Sprintf("RAW=%s\n", s)
		var err error
		o, err = h.controller.ParseHub(m.B, m.H)
		if err != nil {
			auditText += fmt.Sprintf("PARSE=FAILED\n")
		}
	}
	if o != nil {
		b, _ := json.MarshalIndent(o, "", "  ")
		auditText += "JSON=\n"
		auditText += string(b)
		auditText += "\n"
	}
	auditText += "=======\n"
	h.auditBytes(auditText)
}

func (h *Hub) auditRegister(c UserClient) {
	if !h.audit {
		return
	}

	auditText := fmt.Sprintf("%s - REGISTRATION\n", time.Now().Format("15:04:05"))
	auditText += fmt.Sprintf("USERID=%d\n", c.GetUserID())
	auditText += fmt.Sprintf("ACTIVE=%d\n", len(h.clients[c.GetUserID()]))
	auditText += "======\n"

	h.auditBytes(auditText)
}

func (h *Hub) auditUnregister(c UserClient) {
	if !h.audit {
		return
	}

	auditText := fmt.Sprintf("%s - LEAVING\n", time.Now().Format("15:04:05"))
	auditText += fmt.Sprintf("USERID=%d\n", c.GetUserID())
	wrapper := h.getWrapper(c.GetUserID(), c)
	if wrapper == nil {
		auditText += "WRAPPER=MISSING\n"
	} else {
		auditText += fmt.Sprintf("PENDING=%d\n", len(wrapper.pending))
		auditText += fmt.Sprintf("RECONNECTS=%d\n", wrapper.reconnects)
	}
	auditText += fmt.Sprintf("ACTIVE=%d\n", len(h.clients[c.GetUserID()]))
	auditText += "======\n"

	h.auditBytes(auditText)
}

func (h *Hub) loop(closed, crashed bool) (rClosed bool, rCrashed bool) {
	defer func() {
		if r := recover(); r != nil {
			rCrashed = true
			h.log.Error("Hub Crashed", zap.Stack("crash"))

		}
	}()
	select {
	case <-h.done:
		{
			h.log.Info("Hub Closing")
			rClosed = true
		}
	case client := <-h.register:
		{
			h.auditRegister(client)
			if closed {
				client.Close()
			} else {
				h.nextSessionID++
				wrapper := &UserClientWrapper{
					userID:    client.GetUserID(),
					sessionID: h.nextSessionID,
					C:         client,
					nextMsgID: 1,
					pending:   []pendingMsg{},
				}
				h.clients[client.GetUserID()] = append(h.clients[client.GetUserID()], wrapper)
			}
		}
	case client := <-h.unregister:
		{
			h.auditUnregister(client)
			wrapper := h.getWrapper(client.GetUserID(), client)
			if wrapper != nil {
				if wrapper.C != nil {
					wrapper.C.Close()
					wrapper.C = nil
				}
				wrapper.disconnectedTs = time.Now()
				// we don't remove wrapper yet cause who knows we get reconnected.
				h.log.Info("Client Unregistered", zap.Int64("user_id", wrapper.userID), zap.Int("session_id", wrapper.sessionID))
				//
			} else {
				h.log.Error("Client Not Found", zap.Int64("user_id", client.GetUserID()))
			}
		}
	case message := <-h.incoming:
		{
			if !crashed {
				if message.U != nil {
					h.auditIncomingUser(message)

					s := strings.TrimSpace(string(message.B))
					if strings.HasPrefix(s, "#") {
						wrapper := h.getWrapper(message.U.GetUserID(), message.U)
						if wrapper == nil {
							return
						}
						comps := strings.Split(s, "#")
						if len(comps) == 3 {
							sessionID, _ := strconv.Atoi(comps[1])
							initCode, _ := strconv.Atoi(comps[2])
							if sessionID == 0 || initCode == 0 {
								h.Reset(wrapper)
								return
							}
							var oldWrapper *UserClientWrapper
							for _, w := range h.clients[message.U.GetUserID()] {
								if w.sessionID == sessionID {
									oldWrapper = w
									break
								}
							}
							if oldWrapper == nil {
								h.Reset(wrapper)
								return
							}
							h.log.Info("Reusing old session", zap.Int("msgid", oldWrapper.nextMsgID), zap.Int("initcode", initCode))
							if oldWrapper.C != nil {
								oldWrapper.C.Close()
							}
							oldWrapper.C = wrapper.C
							wrapper.C = nil
							wrapper.disconnectedTs = time.Now()

							wrapper = oldWrapper
							initial, err := wrapper.initialMessages(initCode)
							if err == nil {
								if len(initial) == 0 {
									wrapper.C.Msg([]byte("@"))
								}
								for _, m := range initial {
									wrapper.C.Msg(m.b)
								}
								return
							} else {
								h.Reset(wrapper)
								return
							}
						} else {
							h.Reset(wrapper)
							return
						}
					}

					var err error
					message.J, err = h.controller.ParseUser(message.B)
					if err != nil {
						h.log.Error("Parse Failed", zap.Error(err))
						return
					}
					h.controller.ProcessUser(message.J, message.U.GetUserID(), time.Now())
				} else if message.H != nil {
					h.auditIncomingHub(message)

					if message.J == nil {
						var err error
						message.J, err = h.controller.ParseHub(message.B, message.H)
						if err != nil {
							h.log.Error("Parse Failed", zap.Error(err))
							return
						}
					}
					h.controller.ProcessHub(message.J, message.H, time.Now())
				} else if message.D != nil {
					h.controller.ProcessDefault(message.J, message.D, time.Now())
				}
			} else {

			}
		}
	case userId := <-h.kickChan:
		{
			for _, wrapper := range h.clients[userId] {
				if wrapper.C != nil {
					conn := wrapper.C.(*WSClient).conn
					if conn != nil {
						conn.Close()
					}
				}
			}
		}
	case ctime := <-h.ticker.C:
		{
			if !closed && !crashed {
				h.controller.Tick(ctime)
			}
			for cid, wrappers := range h.clients {
				dirty := false
				for _, wrapper := range wrappers {
					if wrapper.C == nil && time.Since(wrapper.disconnectedTs) > 30*time.Second {
						dirty = true
					}
				}
				if dirty {
					var leftOvers []*UserClientWrapper
					for _, wrapper := range wrappers {
						if wrapper.C == nil && time.Since(wrapper.disconnectedTs) > 30*time.Second {
							h.auditWrapperCleanup(wrapper)
							continue
						}
						leftOvers = append(leftOvers, wrapper)
					}
					if len(leftOvers) == 0 {
						delete(h.clients, cid)
						h.controller.OnDisconnect(cid)
					} else {
						h.clients[cid] = leftOvers
					}
				}
			}
		}
	case ctime := <-h.tickChan:
		h.controller.Tick(ctime)
	case query := <-h.statusChan:
		{
			resp := fmt.Sprintf("%s - %d clients - %s", h.ID, len(h.clients), h.controller.Healthz())
			select {
			case query.R <- resp:
			default:
			}
		}
	}
	return
}

func (h *Hub) run() {
	h.controller.Connect(h, time.Now())
	closed := false
	crashed := false

	for {
		closed, crashed = h.loop(closed, crashed)
		if closed && len(h.clients) != 0 {
			h.log.Warn("Closed with actives")
			break
		}
		if closed && len(h.clients) == 0 {
			break
		}
	}

	h.controller.OnShutdown()
	h.log.Info("Hub Terminated")
	RemoveHub(h)
	if h.shutdownWg != nil {
		h.shutdownWg.Done()
		h.shutdownWg = nil
	}
	h.ticker.Stop()
	close(h.done)

	var myLinks []*HubLink
	linksLock.RLock()
	for _, link := range links {
		if link.SrcID == h.ID {
			myLinks = append(myLinks, link)
		}
	}
	linksLock.RUnlock()

	for _, link := range myLinks {
		link.Close()
	}
}

func (h *Hub) Log() *zap.Logger {
	return h.log
}

func (h *Hub) Custom(o Serializable, from HubClient) bool {
	if h.ID == from.GetHubID() {
		go func() {
			h.incoming <- &IncomingMessage{
				J: o,
				B: nil,
				H: from,
			}
		}()
	} else {
		h.incoming <- &IncomingMessage{
			J: o,
			B: nil,
			H: from,
		}
	}
	return true
}

// Msg a message to the given client. If we can't send a message we close
// the client's channel, which would result in client cleaning up the websocket
func (h *Hub) msgFromUser(msg []byte, from UserClient) bool {
	h.incoming <- &IncomingMessage{
		B: msg,
		U: from,
	}
	return true
}

func (h *Hub) MsgFromDefault(o Serializable, from *OneTimeClient) bool {
	if from == nil {
		from = &OneTimeClient{}
	}
	h.incoming <- &IncomingMessage{
		J: o,
		D: from,
	}
	return true
}

func (h *Hub) bytes(b []byte, from HubClient) bool {
	msg := &IncomingMessage{
		J: nil,
		B: b,
		H: from,
	}
	select {
	case h.incoming <- msg:
	default:
		return false
	}
	return true
}

func (h *Hub) Default(o Serializable, from *OneTimeClient) {
	if from == nil {
		from = &OneTimeClient{}
	}
	h.incoming <- &IncomingMessage{
		J: o,
		D: from,
	}
}

func (h *Hub) Status(r chan string) {
	h.statusChan <- &StatusQuery{
		R: r,
	}
}

// Close does nothing for a hub. Incase a hub gets registered as a client, we don't want it to
// shutdown if a connected hub merely wants to close the connection. For shutting down a hub
// please refer to Shutdown function.
func (h *Hub) Close() {
}

// Shutdown shuts the hub down and closes all clients connected to it.
func (h *Hub) Shutdown(wg *sync.WaitGroup) {
	h.shutdownWg = wg
	h.done <- true
}

func (h *Hub) DrainIncoming() []*IncomingMessage {
	var ret []*IncomingMessage
	for {
		select {
		case m := <-h.incoming:
			{
				ret = append(ret, m)
			}
		default:
			return ret
		}
	}
}

func (h *Hub) IsUser() bool {
	return false
}

func (h *Hub) IsHub() bool {
	return true
}

func (h *Hub) GetUserID() int {
	return 0
}

func (h *Hub) GetUserName() string {
	return ""
}

func (h *Hub) GetHubID() string {
	return h.ID
}

func (h *Hub) Reset(wrapper *UserClientWrapper) {
	wrapper.nextMsgID = 1
	wrapper.pending = []pendingMsg{}
	wrapper.C.Msg([]byte(fmt.Sprintf("#%d", wrapper.sessionID)))
	dt := time.Now()
	for _, msg := range h.controller.State(wrapper.userID) {
		b, _ := msg.Serialize()
		wrapper.pending = append(wrapper.pending, pendingMsg{
			id: wrapper.nextMsgID,
			b:  b,
			t:  dt,
		})
		wrapper.C.Msg(b)
		wrapper.nextMsgID++
	}
}

//	func (h *Hub) TestJson(o models.Serializable, from Client) bool {
//		h.controller.Process(nil, o, from)
//		return true
//	}
func (h *Hub) HasUser(id int64) bool {
	if _, ok := h.clients[id]; !ok {
		return false
	}
	return true
}

func (h *Hub) UserCustom(id int64, o Serializable) error {
	h.auditUser(id, o)
	return h.internalUserCustom(id, o)
}

// User should only be called from Hub Processing Loop.
func (h *Hub) internalUserCustom(id int64, o Serializable) error {
	b, err := o.Serialize()
	if err != nil {
		return err
	}
	for _, wrapper := range h.clients[id] {
		wrapper.pending = append(wrapper.pending, pendingMsg{
			id: wrapper.nextMsgID,
			b:  b,
			t:  time.Now(),
		})
		wrapper.nextMsgID++

		if wrapper.C != nil {
			wrapper.C.Msg(b)
		}
	}
	return nil
}

// func (h *Hub) Custom(o models.Serializable, from Client) bool {
// 	h.incoming <- &IncomingMessage{
// 		J:    o,
// 		From: from,
// 	}
// 	return true
// }

func (h *Hub) Simulated() bool {
	return h.tickDuration > 10*time.Minute
}

func (h *Hub) BroadcastExcept(o Serializable, userId int64) error {
	h.auditBroadcast(o, userId)
	// loop over all clients in hub
	for uid := range h.clients {
		if uid == userId {
			continue
		}
		h.internalUserCustom(uid, o)
	}
	return nil
}

func (h *Hub) BroadcastCustom(o Serializable) error {
	h.auditBroadcast(o, 0)
	// loop over all clients in hub
	for uid := range h.clients {
		h.internalUserCustom(uid, o)
	}
	return nil
}

func (h *Hub) IsConnected(userID int64) bool {
	for _, wrapper := range h.clients[userID] {
		if wrapper.C != nil {
			return true
		}
	}
	return false
}

func (h *Hub) LogFields() []zapcore.Field {
	return []zapcore.Field{
		zap.String("ctype", "hub"),
		zap.String("hub_id", h.GetHubID()),
	}
}

func (h *Hub) Kick(userID int64) {
	h.kickChan <- userID
}

var hubs map[string]*Hub
var links map[string]*HubLink
var hubLock sync.RWMutex
var linksLock sync.RWMutex

func HubClientById(id string, src string) HubClient {
	hub := HubById(id)
	if hub != nil {
		return hub
	}
	if src == "" {
		return nil
	}
	link := LinkById(src, id)
	return link
}

func HubById(id string) *Hub {
	hubLock.RLock()
	defer hubLock.RUnlock()

	return hubs[id]
}

func LinkById(srcHub string, destHub string) *HubLink {
	linksLock.RLock()
	defer linksLock.RUnlock()

	return links[fmt.Sprintf("%s-%s", srcHub, destHub)]
}

func HubByIdWait(id string, d time.Duration) *Hub {
	for {
		h := HubById(id)
		if h != nil {
			return h
		}
		time.Sleep(100 * time.Millisecond)
		d = d - (100 * time.Millisecond)
		if d <= 0 {
			return nil
		}
	}
}

func RemoveHub(c HubClient) {
	hubLock.Lock()
	defer hubLock.Unlock()
	delete(hubs, c.GetHubID())
}

func RemoveLink(c *HubLink) {
	linksLock.Lock()
	defer linksLock.Unlock()
	delete(links, fmt.Sprintf("%s-%s", c.SrcID, c.DestID))
}

func Hubs() []*Hub {
	hubLock.RLock()
	defer hubLock.RUnlock()

	var ret []*Hub
	for _, v := range hubs {
		ret = append(ret, v)
	}

	return ret
}

func ShutdownAllHubs() {
	var shutdownWg sync.WaitGroup
	for _, h := range Hubs() {
		shutdownWg.Add(1)
		// start all shutdowns in parallel.
		go func(h *Hub) {
			h.Shutdown(&shutdownWg)
		}(h)
	}
	shutdownWg.Wait()
}

func installLink(link *HubLink) {
	linksLock.Lock()
	defer linksLock.Unlock()
	links[link.SrcID+"-"+link.DestID] = link
}

func installHub(id string, c *Hub) {
	hubLock.Lock()
	defer hubLock.Unlock()
	hubs[id] = c
}

// NewHub starts a new Hub with given ID
func NewHub(id string, ctrl Controller, buffer int, tickDuration time.Duration, logger *zap.Logger) *Hub {
	h := &Hub{
		ID:           id,
		incoming:     make(chan *IncomingMessage, buffer),
		register:     make(chan UserClient, buffer/2),
		unregister:   make(chan UserClient, buffer/2),
		clients:      make(map[int64][]*UserClientWrapper),
		done:         make(chan bool, 2),
		controller:   ctrl,
		ticker:       time.NewTicker(tickDuration),
		tickChan:     make(chan time.Time, 10),
		tickDuration: tickDuration,
		statusChan:   make(chan *StatusQuery, 10),
		kickChan:     make(chan int64, 10),
		log:          logger.With(zap.String("hub_id", id)),
	}

	installHub(id, h)
	return h
}

func (h *Hub) SetupCustomLogging(folder string) {
	if len(folder) == 0 {
		h.log.Error("Attempting to setup with empty folder")
		return
	}
	h.log, _ = zap.Config{
		Level:            zap.NewAtomicLevelAt(zap.InfoLevel),
		Development:      false,
		Encoding:         "json",
		EncoderConfig:    zap.NewProductionEncoderConfig(),
		OutputPaths:      []string{filepath.Join(folder, h.ID+".log")},
		ErrorOutputPaths: []string{filepath.Join(folder, h.ID+".error.log")},
	}.Build()
}

func init() {
	hubs = make(map[string]*Hub)
	links = make(map[string]*HubLink)
}

func HubKickHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		hubId := r.Form.Get("id")
		userId, _ := strconv.ParseInt(r.Form.Get("user_id"), 10, 64)

		hub := HubById(hubId)
		if hub == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		hub.Kick(userId)
	})
}

func HubHealthzHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)

		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("Hubs:\n"))
		statuses := make(chan string, 10)
		for _, hub := range Hubs() {
			hub.Status(statuses)
			status := <-statuses
			w.Write([]byte(status))
			w.Write([]byte("\n"))
			w.Write([]byte("========================\n"))
		}
	})
}

func HubHandler(sugar string, logger *zap.Logger) http.Handler {
	oLog := logger.With(zap.String("op", "hub_handler"))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]
		tokenStr := r.Header.Get("jwt")
		if tokenStr == "" {
			q := r.URL.Query()
			tokenStr = q.Get("__jwt")
		}
		if tokenStr == "" || tokenStr == "null" || tokenStr == "nil" {
			oLog.Error("Missing Token")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New(fmt.Sprintf("Unexpected signing method: %v", t.Header["alg"]))
			}
			return []byte(sugar), nil
		})
		if err != nil {
			oLog.Error("Invalid Token", zap.Error(err))
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || !token.Valid {
			oLog.Error("Invalid Token", zap.String("token", tokenStr))
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		expiry, hasExpiry := claims["expiry"]
		if hasExpiry {
			unixExpiry, err := strconv.Atoi(expiry.(string))
			if err == nil {
				t := time.Unix(int64(unixExpiry), 0)
				if t.Before(time.Now()) {
					oLog.Error("Token Expired", zap.String("token", tokenStr))
					w.WriteHeader(http.StatusUnauthorized)
					return
				}
			}
		}

		hub := HubById(id)
		if hub == nil {
			oLog.Error("Hub Not Found", zap.String("hub_id", id))
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if hubID, found := claims["hub_id"]; found {
			oLog.Info("Incoming Connection")

			link := LinkById(id, hubID.(string))
			if link == nil {
				oLog.Error("Hub Not Found")
				w.WriteHeader(http.StatusNotFound)
				return
			}
			srcHub := link
			conn, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				oLog.Error("Can't upgrade", zap.Error(err))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			srcHub.Accept(conn)
			return
		}
		if userID, found := claims["user_id"]; found {
			uID, _ := strconv.ParseInt(userID.(string), 10, 64)
			oLog.Info("Incoming Connection", zap.Int64("user_id", uID))

			uName := ""
			if _, exists := claims["user_name"]; exists {
				uName = claims["user_name"].(string)
			}
			_, err := NewWSClient(hub, uID, uName, w, r, oLog)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			}
			return
		}
		oLog.Error("Invalid Token", zap.String("token", tokenStr))
		w.WriteHeader(http.StatusUnauthorized)
	})
}

// HubLink
type HubLink struct {
	SrcID     string   `json:"src"`
	DestID    string   `json:"dest"`
	Type      string   `json:"type"`
	Protocols []string `json:"protocols"`

	Url string `json:"url,omitempty"`

	LastConnected time.Time `json:"last_connected"`
	LastPong      time.Time `json:"last_pong"`
	LastPing      time.Time `json:"last_ping"`

	H *Hub        `json:"-"` // The hub we are connected to on the SRC side.
	C chan []byte `json:"-"`

	conn   *websocket.Conn `json:"-"`
	closed bool            `json:"-"`

	log *zap.Logger `json:"-"`

	lock sync.Mutex `json:"-"`
}

// Controller controls a hub and has complete logic for handling the messages
// coming to a hub.
type Controller interface {
	Connect(h *Hub, t time.Time)
	Tick(t time.Time)
	OnDisconnect(userID int64)
	ParseUser(b []byte) (Serializable, error)
	ParseHub(b []byte, client HubClient) (Serializable, error)

	State(userId int64) []Serializable
	ProcessUser(d Serializable, userId int64, ct time.Time)
	ProcessHub(d Serializable, client HubClient, ct time.Time)
	ProcessDefault(d Serializable, client *OneTimeClient, ct time.Time)

	Healthz() string

	OnPanic()
	OnShutdown()
}
