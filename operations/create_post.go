package operations

import (
	"net/http"
	"time"

	"fr_book_api/models"

	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	// -- imports --
	// -- end --
)

// CreatePost
func CreatePost(sugar string, mongoDb *mongo.Database, logger *zap.Logger) http.Handler {
	oLog := logger.With(zap.String("op", "createPost"))
	// -- init --
	postsId, _ := models.NewIDNode(3)
	// -- end --
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		v := models.NewValidator(r).Secret(sugar)

		userId := v.Token("user_id").Int()

		title := v.Form("title").Optional().String()

		content := v.Form("content").Optional().String()

		video := v.Form("video").Optional().String()

		image := v.Form("image").Optional().String()

		log := oLog.With(zap.String("ip", r.Header.Get("X-Real-IP")))
		// -- code --
		if !v.Valid() {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		log.Debug("Start Operation", zap.Any("user_id", userId), zap.Any("content", content), zap.Any("image", image))

		post := &models.Post{
			Content:   content,
			Image:     image,
			CreatedAt: time.Now(),
			Title:     title,
			UserId:    userId,
			Video:     video,
			Id:        int(postsId.Generate().Int64()),
		}

		mongoDb.Collection("posts").InsertOne(r.Context(), post)

		JSON(&models.StatusResponse{
			Code: 200,
		}, w)
		// -- end --
	})
}

// -- extra --
// -- end --
