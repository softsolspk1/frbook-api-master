package operations

import (
	"fmt"
	"net/http"
	"time"

	"fr_book_api/actors"
	"fr_book_api/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	"golang.org/x/exp/rand"
	// -- imports --
	// -- end --
)

// StartVerification
func StartVerification(sugar string, mongoDb *mongo.Database, logger *zap.Logger) http.Handler {
	oLog := logger.With(zap.String("op", "startVerification"))
	// -- init --
	rand.Seed(uint64(time.Now().UnixNano()))
	verficationOtps = make(map[int]int)
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

		min := 100000
		max := 999999
		randomNumber := rand.Intn(max-min+1) + min

		verficationOtps[userId] = randomNumber

		c := actors.NewOneTimeClient(10)

		hb := actors.HubById("smtp")

		hb.Default(&models.SmsEvent{
			Email:   u.Email,
			Subject: "Verification OTP",
			Content: fmt.Sprintf("Your OTP is %d", randomNumber),
		}, c)

		resp := (<-c.Response).(*models.SmsEvent)
		log.Info("OTP Sent", zap.Any("otp", randomNumber))

		if !resp.Success {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		JSON(&models.IntResponse{
			Code: 200,
		}, w)
		// -- end --
	})
}

// -- extra --

var verficationOtps = make(map[int]int)

// -- end --
