package operations

import (
	"net/http"

	"fr_book_api/models"

	"github.com/thoas/go-funk"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	// -- imports --
	// -- end --
)

// GetChats
func GetChats(sugar string, mongoDb *mongo.Database, logger *zap.Logger) http.Handler {
	oLog := logger.With(zap.String("op", "getChats"))
	// -- init --
	// -- end --
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		v := models.NewValidator(r).Secret(sugar)

		userId := v.Token("user_id").Int()

		toId := v.Query("to_id").Int()

		log := oLog.With(zap.String("ip", r.Header.Get("X-Real-IP")))
		// -- code --
		if !v.Valid() {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		log.Debug("Start Operation", zap.Any("user_id", userId), zap.Any("to_id", toId))

		var chats []*models.Chat

		c, err := mongoDb.Collection("chats").Find(r.Context(), bson.M{"$or": []bson.M{{"from_id": userId, "to_id": toId}, {"from_id": toId, "to_id": userId}}}, options.Find().SetSort(bson.M{"created_at": -1}).SetLimit(10))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		for c.Next(r.Context()) {
			var chat models.Chat
			if err := c.Decode(&chat); err != nil {
				continue
			}
			chats = append(chats, &chat)
		}

		chats = funk.Reverse(chats).([]*models.Chat)

		JSON(&models.ChatListResponse{
			Code:   200,
			Result: chats,
		}, w)
		// -- end --
	})
}

// -- extra --
// -- end --
