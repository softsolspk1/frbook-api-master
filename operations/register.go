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

// Register
func Register(sugar string, mongoDb *mongo.Database, logger *zap.Logger) http.Handler {
	oLog := logger.With(zap.String("op", "register"))
	// -- init --
	usersID, _ := models.NewIDNode(2)
	// -- end --
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		v := models.NewValidator(r).Secret(sugar)

		name := v.Form("name").String()

		email := v.Form("email").String()

		password := v.Form("password").String()

		log := oLog.With(zap.String("ip", r.Header.Get("X-Real-IP")))
		// -- code --
		if !v.Valid() {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		log.Debug("Start Operation", zap.Any("username", name), zap.Any("password", password), zap.Any("name", name), zap.Any("email", email))

		count, err := mongoDb.Collection("users").CountDocuments(r.Context(), bson.M{"email": email})
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if count > 0 {
			JSON(&models.StatusResponse{
				Code:  400,
				Error: "Email already exists",
			}, w)
			return
		}

		hash := md5.Sum([]byte(password))
		password = hex.EncodeToString(hash[:])

		user := &models.User{
			Password: password,
			Name:     name,
			Email:    email,
			Verified: false,
		}

		user.Id = int(usersID.Generate().Int64())

		if _, err := mongoDb.Collection("users").InsertOne(r.Context(), user); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"user_id": fmt.Sprintf("%d", user.Id),
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
