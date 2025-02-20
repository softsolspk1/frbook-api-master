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

// AddComment
func AddComment(sugar string, mongoDb *mongo.Database, logger *zap.Logger) http.Handler {
	oLog := logger.With(zap.String("op", "addComment"))
	// -- init --
	commentId, _ := models.NewIDNode(5)
	// -- end --
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		v := models.NewValidator(r).Secret(sugar)

		userId := v.Token("user_id").Int()

		id := v.Path("id").Int()

		content := v.Form("content").String()

		log := oLog.With(zap.String("ip", r.Header.Get("X-Real-IP")))
		// -- code --
		if !v.Valid() {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		log.Debug("Start Operation", zap.Any("user_id", userId), zap.Any("id", id), zap.Any("content", content))

		if len(content) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		var u models.User

		if err := mongoDb.Collection("users").FindOne(r.Context(), bson.M{"_id": userId}).Decode(&u); err != nil {
			if err == mongo.ErrNoDocuments {
				w.WriteHeader(http.StatusNotFound)
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}
			return
		}

		comment := models.Comment{
			PostId:    id,
			UserId:    userId,
			Content:   content,
			Name:      u.Name,
			CreatedAt: time.Now(),
			Id:        int(commentId.Generate().Int64()),
		}

		mongoDb.Collection("comments").InsertOne(r.Context(), comment)

		mongoDb.Collection("posts").UpdateOne(r.Context(), bson.M{"_id": id}, bson.M{"$inc": bson.M{"comments_count": 1}})

		JSON(&models.StatusResponse{
			Code: 200,
		}, w)
		// -- end --
	})
}

// -- extra --
// -- end --
