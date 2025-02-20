package operations

import (
	"net/http"
	"sort"

	"fr_book_api/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	// -- imports --
	// -- end --
)

// GetArticles
func GetArticles(sugar string, mongoDb *mongo.Database, logger *zap.Logger) http.Handler {
	oLog := logger.With(zap.String("op", "getArticles"))
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

		var articles []*models.Article
		users := make(map[int]models.User)

		c, err := mongoDb.Collection("articles").Find(r.Context(), bson.M{})
		if err != nil {
			log.Error("Unable to get articles", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		for c.Next(r.Context()) {
			var article models.Article
			err := c.Decode(&article)
			if err != nil {
				log.Error("Unable to decode article", zap.Error(err))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			if _, ok := users[article.UserId]; !ok {
				var u models.User
				if err := mongoDb.Collection("users").FindOne(r.Context(), bson.M{"_id": article.UserId}).Decode(&u); err != nil {
					continue
				}
				users[article.UserId] = u
			}
			article.ProfilePic = users[article.UserId].ProfilePic

			articles = append(articles, &article)
		}

		sort.Slice(articles, func(i, j int) bool {
			return articles[i].CreatedAt.After(*articles[j].CreatedAt)
		})

		JSON(&models.ArticleListResponse{
			Code:   200,
			Result: articles,
		}, w)
		// -- end --
	})
}

// -- extra --
// -- end --
