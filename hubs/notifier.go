package hubs

import (
	"context"
	"fmt"
	"fr_book_api/actors"
	"fr_book_api/models"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	// -- imports --
	// -- end --
)

// NotifierSetup sets up things for the hub.
func NotifierSetup(sugar string, mongoDb *mongo.Database, logger *zap.Logger) error {
	// -- init --
	actors.NewHub("notifier", &NotifierController{
		db: mongoDb,
	}, 100, time.Second,
		logger).Start()
	return nil
	// -- end --
}

// NotifierController is controller for the hub.
type NotifierController struct {
	h   *actors.Hub
	log *zap.Logger
	// -- declarations --
	db         *mongo.Database
	chId       *models.IDNode
	callStates []*callCurr
	// -- end --
}

func (nc *NotifierController) Connect(h *actors.Hub, ct time.Time) {
	nc.h = h
	// -- connect --
	nc.log = h.Log()
	nc.chId, _ = models.NewIDNode(6)
	// -- end --
}

func (nc *NotifierController) OnShutdown() {
	// -- shutdown --
	nc.log.Info("Shutting down the co-hub")
	// -- end --
}

func (nc *NotifierController) ParseUser(b []byte) (actors.Serializable, error) {
	// -- parse-user --
	var ev models.CallEvent

	err := ev.Parse(b)
	if err != nil {
		return nil, err
	}

	return &ev, nil
	// -- end --
}

func (nc *NotifierController) ParseHub(b []byte, client actors.HubClient) (actors.Serializable, error) {
	// -- parse-hub --
	var ev models.CallEvent

	err := ev.Parse(b)
	if err != nil {
		return nil, err
	}

	return &ev, nil
	// -- end --
}

func (nc *NotifierController) ProcessUser(d actors.Serializable, userId int64, ct time.Time) {
	// -- process-user --
	if d == nil {
		nc.log.Error("No data")
		return
	}
	ev := d.(*models.CallEvent)
	if ev.Kind == models.CallEventTypeStartCall {
		call := nc.GetCall(int(userId), ev.ToId)
		if call != nil {
			nc.log.Error("Call already exists", zap.Int64("from", userId), zap.Int("to", ev.ToId))
			nc.Init(userId)
			return
		}

		nc.log.Info("Call Requested Started", zap.Int64("from", userId), zap.Int("to", ev.ToId))

		chId := nc.GetChannelId()

		var userFrom models.User
		if err := nc.db.Collection("users").FindOne(context.Background(), bson.M{"_id": userId}).Decode(&userFrom); err != nil {
			nc.log.Error("User not found", zap.Int64("user_id", userId))
			return
		}

		var userTo models.User
		if err := nc.db.Collection("users").FindOne(context.Background(), bson.M{"_id": ev.ToId}).Decode(&userTo); err != nil {
			return
		}

		nc.h.UserCustom(userId, &models.CallEvent{
			Kind:    models.CallEventTypeStartCall,
			Channel: chId,
			ToId:    ev.ToId,
			ToPic:   userTo.ProfilePic,
			FromPic: userFrom.ProfilePic,
		})

		if nc.h.HasUser(int64(ev.ToId)) {
			nc.h.UserCustom(int64(ev.ToId), &models.CallEvent{
				Kind:    models.CallEventTypeIncoming,
				Channel: chId,
				ToId:    int(userId),
				ToPic:   userFrom.ProfilePic,
				FromPic: userTo.ProfilePic,
			})
		}

		nc.callStates = append(nc.callStates, &callCurr{
			FromId:    userId,
			fromPic:   userFrom.ProfilePic,
			toPic:     userTo.ProfilePic,
			ToId:      int64(ev.ToId),
			State:     callStateWaiting,
			channelId: chId,
		})
	}

	if ev.Kind == models.CallEventTypeAcceptCall {
		call := nc.GetCall(int(userId), ev.ToId)
		if call == nil {
			nc.log.Error("Call not found", zap.Int64("from", userId), zap.Int("to", ev.ToId))
			return
		}
		nc.log.Info("Call Accepted", zap.Int64("from", userId), zap.Int("to", ev.ToId))

		if call.State == callStateWaiting {
			if call.ToId == int64(userId) {
				nc.h.UserCustom(call.ToId, &models.CallEvent{
					Kind:    models.CallEventTypeStartCall,
					ToId:    call.getToId(int(userId)),
					Channel: call.channelId,
				})
				call.State = callStateRunning
			}
		} else {
			nc.log.Error("Call not in waiting state", zap.Int64("from", userId), zap.Int("to", ev.ToId))
		}
	}

	if ev.Kind == models.CallEventTypeEndCall {
		call := nc.GetCall(int(userId), ev.ToId)
		if call == nil {
			nc.log.Error("Call not found to end", zap.Int64("from", userId), zap.Int("to", ev.ToId))
			return
		}
		nc.h.UserCustom(int64(call.FromId), &models.CallEvent{
			Kind:    models.CallEventTypeEndCall,
			Channel: call.channelId,
			ToId:    call.getToId(int(userId)),
		})

		nc.h.UserCustom(int64(call.ToId), &models.CallEvent{
			Kind:    models.CallEventTypeEndCall,
			Channel: call.channelId,
			ToId:    call.getToId(int(userId)),
		})
		// delete it

		beforeLen := len(nc.callStates)
		newStates := make([]*callCurr, 0)
		for _, v := range nc.callStates {
			if (v.FromId != call.FromId && v.ToId != call.ToId) && (v.FromId != call.ToId && v.ToId != call.FromId) {
				newStates = append(newStates, v)
			}
		}
		nc.callStates = newStates

		nc.log.Info("Call Ended", zap.Int("before", beforeLen), zap.Int("after", len(nc.callStates)))
	}

	if ev.Kind == models.CallEventTypeInit {
		nc.Init(userId)
	}

	// -- end --
}

func (nc *NotifierController) State(userId int64) []actors.Serializable {
	// -- state --
	return nil
	// -- end --
}

func (nc *NotifierController) ProcessDefault(o actors.Serializable, c *actors.OneTimeClient, ct time.Time) {
	// -- process-default --

	fmt.Println("Processing Default")
	// -- end --
}

func (nc *NotifierController) ProcessHub(d actors.Serializable, c actors.HubClient, ct time.Time) {
	// -- process-hub --

	fmt.Println("Processing Hub")
	// -- end --
}

func (nc *NotifierController) Tick(ct time.Time) {
	// -- tick --
	// -- end --
}

func (nc *NotifierController) OnDisconnect(userId int64) {
	// -- disconnect --

	newStates := make([]*callCurr, 0)
	for _, v := range nc.callStates {
		if v.FromId == userId || v.ToId == userId {
			if nc.h.HasUser(int64(v.getToId(int(userId)))) {
				nc.h.UserCustom(int64(v.getToId(int(userId))), &models.CallEvent{
					Kind:    models.CallEventTypeEndCall,
					Channel: v.channelId,
					ToId:    v.getToId(int(userId)),
				})
			}
		} else {
			newStates = append(newStates, v)
		}
	}

	nc.callStates = newStates
	// -- end --
}

func (nc *NotifierController) OnPanic() {
	// -- panic --
	// -- end --
}

func (nc *NotifierController) Healthz() string {
	// -- health --
	return ""
	// -- end --
}

// -- code --

// create a unique id using 2 integers
func (nc *NotifierController) GetChannelId() string {
	// a and b can be swapped and still create the same id
	return nc.chId.Generate().String()
}

type callState int

const (
	callStateNone callState = iota
	callStateWaiting
	callStateRunning
	callStateEnded
)

type callCurr struct {
	FromId    int64
	ToId      int64
	State     callState
	channelId string
	fromPic   string
	toPic     string
}

func (nc *NotifierController) GetCall(a, b int) *callCurr {
	for _, v := range nc.callStates {
		if v.FromId == int64(a) && v.ToId == int64(b) {
			return v
		}
		if v.FromId == int64(b) && v.ToId == int64(a) {
			return v
		}
	}
	return nil
}

func (c *callCurr) getToId(myId int) int {
	if myId == int(c.ToId) {
		return int(c.FromId)
	}
	return int(c.ToId)
}

func (nc *NotifierController) Init(userId int64) {
	nc.h.UserCustom(userId, &models.CallEvent{
		Kind: models.CallEventTypeInit,
	})

	for _, v := range nc.callStates {
		if v.FromId == userId || v.ToId == userId {
			if v.State == callStateWaiting {
				if v.ToId == int64(userId) {
					nc.h.UserCustom(v.ToId, &models.CallEvent{
						Kind:    models.CallEventTypeIncoming,
						Channel: v.channelId,
						ToId:    v.getToId(int(userId)),
					})
				} else {
					nc.h.UserCustom(v.FromId, &models.CallEvent{
						Kind:    models.CallEventTypeStartCall,
						Channel: v.channelId,
						ToId:    v.getToId(int(userId)),
					})
				}
			}
			if v.State == callStateRunning {
				nc.h.UserCustom(userId, &models.CallEvent{
					Kind:    models.CallEventTypeEndCall,
					Channel: v.channelId,
					ToId:    v.getToId(int(userId)),
				})
			}
		}
	}

}

// -- end --
