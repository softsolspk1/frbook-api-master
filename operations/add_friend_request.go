package operations

import (
	"net/http"

	"fr_book_api/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	// -- imports --
	// -- end --
)

// AddFriendRequest
func AddFriendRequest(sugar string, mongoDb *mongo.Database, logger *zap.Logger) http.Handler {
	oLog := logger.With(zap.String("op", "addFriendRequest"))
	// -- init --
	frId, _ := models.NewIDNode(10)
	// -- end --
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		v := models.NewValidator(r).Secret(sugar)

		userId := v.Token("user_id").Int()

		toId := v.Form("to_id").Int()

		log := oLog.With(zap.String("ip", r.Header.Get("X-Real-IP")))
		// -- code --
		if !v.Valid() {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		log.Debug("Start Operation", zap.Any("user_id", userId), zap.Any("to_id", toId))

		// count existing friend requests

		count, err := mongoDb.Collection("friend_requests").CountDocuments(r.Context(), bson.M{
			"$or": []bson.M{
				{"from_id": userId, "to_id": toId},
				{"from_id": toId, "to_id": userId},
			},
		})

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if count > 0 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		_, err = mongoDb.Collection("friend_requests").InsertOne(r.Context(), &models.FriendRequest{
			FromId: userId,
			ToId:   toId,
			Id:     int(frId.Generate().Int64()),
		})

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
