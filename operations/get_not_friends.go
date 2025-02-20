package operations

import (
	"net/http"

	"fr_book_api/models"

	"github.com/thoas/go-funk"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	// -- imports --
	// -- end --
)

// GetNotFriends
func GetNotFriends(sugar string, mongoDb *mongo.Database, logger *zap.Logger) http.Handler {
	oLog := logger.With(zap.String("op", "getNotFriends"))
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

		pendingResp := make(map[int]int)
		takeAction := make(map[int]int)

		frc, err := mongoDb.Collection("friend_requests").Find(r.Context(), bson.M{})

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		for frc.Next(r.Context()) {
			var fr models.FriendRequest
			if err := frc.Decode(&fr); err != nil {
				continue
			}
			if fr.FromId == userId {
				pendingResp[fr.ToId] = fr.Id
			}
			if fr.ToId == userId {
				takeAction[fr.FromId] = fr.Id
			}
		}

		idsToIgnore := []int{userId}

		fc, err := mongoDb.Collection("friends").Find(r.Context(), bson.M{})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		for fc.Next(r.Context()) {
			var fr models.FriendEntry
			if err := fc.Decode(&fr); err != nil {
				continue
			}
			if fr.FromId == userId {
				idsToIgnore = append(idsToIgnore, fr.ToId)
			}
			if fr.ToId == userId {
				idsToIgnore = append(idsToIgnore, fr.FromId)
			}
		}

		var friends []*models.User

		c, err := mongoDb.Collection("users").Find(r.Context(), bson.M{})

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		defer c.Close(r.Context())

		for c.Next(r.Context()) {
			var user models.User
			if err := c.Decode(&user); err != nil {
				continue
			}
			if funk.Contains(idsToIgnore, user.Id) {
				continue
			}

			if _, ok := pendingResp[user.Id]; ok {
				user.Status = models.ReqStatusPending
				user.ReqId = pendingResp[user.Id]
			}

			if _, ok := takeAction[user.Id]; ok {
				user.Status = models.ReqStatusTakeAction
				user.ReqId = takeAction[user.Id]
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
