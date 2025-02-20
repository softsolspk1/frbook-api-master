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

// CallNotifierSetup sets up things for the hub.
func CallNotifierSetup(sugar string, mongoDb *mongo.Database, logger *zap.Logger) error {
	// -- init --
	actors.NewHub("callnotifier", &CallNotifierController{
		db: mongoDb,
	}, 100, time.Second, logger).Start()
	return nil
	// -- end --
}

// CallNotifierController is controller for the hub.
type CallNotifierController struct {
	h   *actors.Hub
	log *zap.Logger
	// -- declarations --
	db         *mongo.Database
	chId       *models.IDNode
	callStates []*callCurr
	// -- end --
}

func (nc *CallNotifierController) Connect(h *actors.Hub, ct time.Time) {
	nc.h = h
	// -- connect --
	nc.log = h.Log()
	nc.chId, _ = models.NewIDNode(6)
	// -- end --
}

func (nc *CallNotifierController) OnShutdown() {
	// -- shutdown --
	// -- end --
}

func (nc *CallNotifierController) ParseUser(b []byte) (actors.Serializable, error) {
	// -- parse-user --
	var ev models.CallEvent

	err := ev.Parse(b)
	if err != nil {
		return nil, err
	}

	return &ev, nil
	// -- end --
}

func (nc *CallNotifierController) ParseHub(b []byte, client actors.HubClient) (actors.Serializable, error) {
	// -- parse-hub --
	var ev models.CallEvent

	err := ev.Parse(b)
	if err != nil {
		return nil, err
	}

	return &ev, nil
	// -- end --
}

func (nc *CallNotifierController) ProcessUser(d actors.Serializable, userId int64, ct time.Time) {
	// -- process-user --
	if d == nil {
		nc.log.Error("No data")
		return
	}
	fmt.Println("Processing User")

	ev := d.(*models.CallEvent)
	if ev.Kind == 1 {

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
			Channel: nc.GetChannelId(int(userId), ev.ToId),
			ToId:    ev.ToId,
			ToPic:   userTo.ProfilePic,
			FromPic: userFrom.ProfilePic,
		})

		if nc.h.HasUser(int64(ev.ToId)) {
			nc.h.UserCustom(int64(ev.ToId), &models.CallEvent{
				Kind:    models.CallEventTypeIncoming,
				Channel: nc.GetChannelId(int(userId), ev.ToId),
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
			channelId: nc.GetChannelId(int(userId), ev.ToId),
		})
		nc.log.Info("Call Started", zap.Int64("from", userId), zap.Int("to", ev.ToId))
	}

	if ev.Kind == models.CallEventTypeAcceptCall {
		call := nc.GetCall(int(userId), ev.ToId)
		if call == nil {
			return
		}
		if call.State == callStateWaiting {
			if call.ToId == int64(userId) {
				nc.h.UserCustom(call.ToId, &models.CallEvent{
					Kind:    models.CallEventTypeStartCall,
					ToId:    call.getToId(int(userId)),
					Channel: call.channelId,
					FromPic: call.fromPic,
					ToPic:   call.toPic,
				})
				call.State = callStateRunning
			}
		}
	}

	if ev.Kind == models.CallEventTypeEndCall {
		call := nc.GetCall(int(userId), ev.ToId)
		if call == nil {
			return
		}
		nc.h.UserCustom(int64(call.FromId), &models.CallEvent{
			Kind:    models.CallEventTypeEndCall,
			Channel: call.channelId,
			FromPic: call.fromPic,
			ToPic:   call.toPic,
			ToId:    call.getToId(int(userId)),
		})

		nc.h.UserCustom(int64(call.ToId), &models.CallEvent{
			Kind:    models.CallEventTypeEndCall,
			Channel: call.channelId,
			ToId:    call.getToId(int(userId)),
			FromPic: call.fromPic,
			ToPic:   call.toPic,
		})
		// delete it
		newStates := make([]*callCurr, 0)
		for _, v := range nc.callStates {
			if v.FromId != int64(userId) && v.ToId != int64(userId) {
				newStates = append(newStates, v)
			}
		}
		nc.callStates = newStates
	}

	if ev.Kind == models.CallEventTypeInit {
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
							FromPic: v.fromPic,
							ToPic:   v.toPic,
						})
					} else {
						nc.h.UserCustom(v.FromId, &models.CallEvent{
							Kind:    models.CallEventTypeStartCall,
							Channel: v.channelId,
							ToId:    v.getToId(int(userId)),
							FromPic: v.fromPic,
							ToPic:   v.toPic,
						})
					}
				}
				if v.State == callStateRunning {
					nc.h.UserCustom(userId, &models.CallEvent{
						Kind:    models.CallEventTypeEndCall,
						Channel: v.channelId,
						ToId:    v.getToId(int(userId)),
						FromPic: v.fromPic,
						ToPic:   v.toPic,
					})
				}
			}
		}
	}

	// -- end --
}

func (nc *CallNotifierController) State(userId int64) []actors.Serializable {
	// -- state --
	return nil
	// -- end --
}

func (nc *CallNotifierController) ProcessDefault(o actors.Serializable, c *actors.OneTimeClient, ct time.Time) {
	// -- process-default --

	fmt.Println("Processing Default")
	// -- end --
}

func (nc *CallNotifierController) ProcessHub(d actors.Serializable, c actors.HubClient, ct time.Time) {
	// -- process-hub --

	fmt.Println("Processing Hub")
	// -- end --
}

func (nc *CallNotifierController) Tick(ct time.Time) {
	// -- tick --
	// -- end --
}

func (nc *CallNotifierController) OnDisconnect(userId int64) {
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

func (nc *CallNotifierController) OnPanic() {
	// -- panic --
	// -- end --
}

func (nc *CallNotifierController) Healthz() string {
	// -- health --
	return ""
	// -- end --
}

// -- code --

// create a unique id using 2 integers
func (nc *CallNotifierController) GetChannelId(a, b int) string {
	// a and b can be swapped and still create the same id
	if a > b {
		a, b = b, a
	}

	return fmt.Sprintf("%d:%d", a, b)
}

func (nc *CallNotifierController) GetCall(a, b int) *callCurr {
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

// -- end --
