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

// GetFriends
func GetFriends(sugar string, mongoDb *mongo.Database, logger *zap.Logger) http.Handler {
	oLog := logger.With(zap.String("op", "getFriends"))
	// -- init --
	// -- end --
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		v := models.NewValidator(r).Secret(sugar)

		userId := v.Token("user_id").Int()

		log := oLog.With(zap.String("ip", r.Header.Get("X-Real-IP")))
		// -- code --
		if !v.Valid() {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		log.Debug("Start Operation", zap.Any("user_id", userId))

		var friends []*models.User

		c, err := mongoDb.Collection("friends").Find(r.Context(), bson.M{"$or": []bson.M{
			{"from_id": userId},
			{"to_id": userId},
		}})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		for c.Next(r.Context()) {
			var f models.FriendEntry
			if err := c.Decode(&f); err != nil {
				continue
			}
			reqUsr := f.FromId
			if f.FromId == userId {
				reqUsr = f.ToId
			}
			var user models.User
			err := mongoDb.Collection("users").FindOne(r.Context(), bson.M{"_id": reqUsr}).Decode(&user)
			if err != nil {
				continue
			}
			friends = append(friends, &user)
		}

		JSON(&models.UserListResponse{
			Code:   200,
			Result: friends,
		}, w)
		// -- end --
	})
}

// -- extra --
// -- end --
