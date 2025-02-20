package operations

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/http"

	"fr_book_api/models"

	jwt "github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	// -- imports --
	// -- end --
)

// Login
func Login(sugar string, mongoDb *mongo.Database, logger *zap.Logger) http.Handler {
	oLog := logger.With(zap.String("op", "login"))
	// -- init --
	// -- end --
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		v := models.NewValidator(r).Secret(sugar)

		email := v.Form("email").String()

		password := v.Form("password").String()

		log := oLog.With(zap.String("ip", r.Header.Get("X-Real-IP")))
		// -- code --
		if !v.Valid() {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		log.Debug("Start Operation", zap.Any("username", email), zap.Any("password", password))

		var u models.User
		if err := mongoDb.Collection("users").FindOne(r.Context(), bson.M{"email": email}).Decode(&u); err != nil {
			if err == mongo.ErrNoDocuments {
				w.WriteHeader(http.StatusNotFound)
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}
			return
		}

		hash := md5.Sum([]byte(password))

		if u.Password != hex.EncodeToString(hash[:]) {
			JSON(&models.StatusResponse{
				Code:  400,
				Error: "Invalid Password",
			}, w)
			return
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"user_id": fmt.Sprintf("%d", u.Id),
		})
		signedToken, err := token.SignedString([]byte(sugar))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("jwt", signedToken)

		JSON(&models.StatusResponse{
			Code: 200,
		}, w)
		// -- end --
	})
}

// -- extra --
// -- end --
