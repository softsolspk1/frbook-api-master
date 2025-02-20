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

// Me
func Me(sugar string, mongoDb *mongo.Database, logger *zap.Logger) http.Handler {
	oLog := logger.With(zap.String("op", "me"))
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

		var u models.User

		if err := mongoDb.Collection("users").FindOne(r.Context(), bson.M{"_id": userId}).Decode(&u); err != nil {
			if err == mongo.ErrNoDocuments {
				w.WriteHeader(http.StatusNotFound)
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}
			return
		}

		JSON(&models.UserResponse{
			Code:   200,
			Result: &u,
		}, w)
		// -- end --
	})
}

// -- extra --
// -- end --
