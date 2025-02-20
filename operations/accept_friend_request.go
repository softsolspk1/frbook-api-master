package operations

import (
	"net/http"
	"time"

	"fr_book_api/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	// -- imports --
	// -- end --
)

// AcceptFriendRequest
func AcceptFriendRequest(sugar string, mongoDb *mongo.Database, logger *zap.Logger) http.Handler {
	oLog := logger.With(zap.String("op", "acceptFriendRequest"))
	// -- init --
	frId, _ := models.NewIDNode(9)
	// -- end --
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		v := models.NewValidator(r).Secret(sugar)

		userId := v.Token("user_id").Int()

		id := v.Path("id").Int()

		log := oLog.With(zap.String("ip", r.Header.Get("X-Real-IP")))
		// -- code --
		if !v.Valid() {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		log.Debug("Start Operation", zap.Any("user_id", userId), zap.Any("id", id))

		// get the friend request

		var fr models.FriendRequest
		err := mongoDb.Collection("friend_requests").FindOne(r.Context(), bson.M{"_id": id}).Decode(&fr)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// add the friend
		newFr := models.FriendEntry{
			CreatedAt: time.Now(),
			FromId:    fr.FromId,
			ToId:      fr.ToId,
			Id:        int(frId.Generate().Int64()),
		}

		_, err = mongoDb.Collection("friends").InsertOne(r.Context(), newFr)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		JSON(&models.StatusResponse{
			Code: 200,
		}, w)
		// -- end --
	})
}

// -- extra --
// -- end --
