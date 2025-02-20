package operations

import (
	"net/http"
	"strings"
	"time"

	"fr_book_api/models"

	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	// -- imports --
	// -- end --
)

// CreateArticle
func CreateArticle(sugar string, mongoDb *mongo.Database, logger *zap.Logger) http.Handler {
	oLog := logger.With(zap.String("op", "createArticle"))
	// -- init --
	articlesId, _ := models.NewIDNode(4)
	// -- end --
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		v := models.NewValidator(r).Secret(sugar)

		userId := v.Token("user_id").Int()

		title := v.Form("title").Optional().String()

		authorName := v.Form("author_name").Optional().String()

		description := v.Form("description").Optional().String()

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
		log.Debug("Start Operation", zap.Any("user_id", userId), zap.Any("title", title), zap.Any("content", content), zap.Any("photo", photo))
		ct := time.Now()
		article := &models.Article{
			Content:     content,
			Photo:       photo,
			Tags:        strings.Split(tags, ","),
			AuthorName:  authorName,
			Description: description,
			CreatedAt:   &ct,
			Title:       title,
			UserId:      userId,
			Pdf:         pdf,
			Id:          int(articlesId.Generate().Int64()),
		}

		_, err := mongoDb.Collection("articles").InsertOne(r.Context(), article)
		if err != nil {
			log.Error("Unable to insert article", zap.Error(err))
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
