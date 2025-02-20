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

// UpdateArticle
func UpdateArticle(sugar string, mongoDb *mongo.Database, logger *zap.Logger) http.Handler {
	oLog := logger.With(zap.String("op", "updateArticle"))
	// -- init --
	// -- end --
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		v := models.NewValidator(r).Secret(sugar)

		userId := v.Token("user_id").Int()

		id := v.Path("id").Int()

		authorName := v.Form("author_name").Optional().String()

		description := v.Form("description").Optional().String()

		title := v.Form("title").Optional().String()

		tags := v.Form("tags").Optional().String()

		content := v.Form("content").Optional().String()

		photo := v.Form("photo").Optional().String()

		pdf := v.Form("pdf").Optional().String()

		log := oLog.With(zap.String("ip", r.Header.Get("X-Real-IP")))
		// -- code --
		if !v.Valid() {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		log.Debug("Start Operation", zap.Any("user_id", userId), zap.Any("id", id), zap.Any("title", title), zap.Any("tags", tags), zap.Any("content", content), zap.Any("photo", photo))

		_, err := mongoDb.Collection("articles").UpdateOne(r.Context(), bson.M{"_id": id}, bson.M{"$set": bson.M{
			"title":       title,
			"tags":        tags,
			"content":     content,
			"photo":       photo,
			"description": description,
			"pdf":         pdf,
			"author_name": authorName,
		}})
		if err != nil {
			log.Error("Unable to update article", zap.Error(err))
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
