package operations

import (
	"net/http"

	"fr_book_api/models"

	"github.com/thoas/go-funk"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	// -- imports --
	// -- end --
)

// GetPosts
func GetPosts(sugar string, mongoDb *mongo.Database, logger *zap.Logger) http.Handler {
	oLog := logger.With(zap.String("op", "getPosts"))
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
		log.Debug("Start Operation")

		c, err := mongoDb.Collection("posts").Find(r.Context(), bson.M{}, options.Find().SetSort(bson.M{"created_at": -1}))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		var posts []*models.Post
		users := make(map[int]models.User)

		defer c.Close(r.Context())

		for c.Next(r.Context()) {
			var p models.Post
			if err := c.Decode(&p); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			p.LikesCount = len(p.Likes)

			if funk.Contains(p.Likes, userId) {
				p.Liked = true
			}

			if _, ok := users[p.UserId]; !ok {
				var u models.User
				if err := mongoDb.Collection("users").FindOne(r.Context(), bson.M{"_id": p.UserId}).Decode(&u); err != nil {
					continue
				}
				users[p.UserId] = u
			}

			p.Name = users[p.UserId].Name
			p.ProfilePic = users[p.UserId].ProfilePic
			p.Likes = nil

			posts = append(posts, &p)
		}

		JSON(&models.PostListResponse{
			Code:   200,
			Result: posts,
		}, w)
		// -- end --
	})
}

// -- extra --
// -- end --
