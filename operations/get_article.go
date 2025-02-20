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

// GetArticle
func GetArticle(sugar string, mongoDb *mongo.Database, logger *zap.Logger) http.Handler {
	oLog := logger.With(zap.String("op", "getArticle"))
	// -- init --
	// -- end --
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		v := models.NewValidator(r).Secret(sugar)

		userId := v.Token("user_id").Int()

		id := v.Path("id").Int()

		log := oLog.With(zap.String("ip", r.Header.Get("X-Real-IP")))
		// -- code --
		if !v.Valid() {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		log.Debug("Start Operation", zap.Any("user_id", userId), zap.Any("id", id))
		var article models.Article

		err := mongoDb.Collection("articles").FindOne(r.Context(), bson.M{"_id": id}).Decode(&article)

		if err != nil {
			log.Error("Unable to get article", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		JSON(&models.ArticleResponse{
			Code:   200,
			Result: &article,
		}, w)
		// -- end --
	})
}

// -- extra --
// -- end --
