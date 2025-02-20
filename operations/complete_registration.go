package operations

import (
	"net/http"
	"strconv"

	"fr_book_api/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	// -- imports --
	// -- end --
)

// CompleteRegistration
func CompleteRegistration(sugar string, mongoDb *mongo.Database, logger *zap.Logger) http.Handler {
	oLog := logger.With(zap.String("op", "completeRegistration"))
	// -- init --
	// -- end --
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		v := models.NewValidator(r).Secret(sugar)

		userId := v.Token("user_id").Int()

		otp := v.Form("otp").String()

		log := oLog.With(zap.String("ip", r.Header.Get("X-Real-IP")))
		// -- code --
		if !v.Valid() {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		log.Debug("Start Operation", zap.Any("user_id", userId), zap.Any("otp", otp))

		// convert otp to int

		otpInt, _ := strconv.Atoi(otp)

		if verficationOtps[userId] != otpInt {
			JSON(&models.StatusResponse{
				Code:  400,
				Error: "Invalid OTP",
			}, w)
			return
		}

		delete(verficationOtps, userId)

		mongoDb.Collection("users").UpdateOne(r.Context(), bson.M{"_id": userId}, bson.M{"$set": bson.M{"verified": true}})

		JSON(&models.StatusResponse{
			Code: 200,
		}, w)
		// -- end --
	})
}

// -- extra --
// -- end --
