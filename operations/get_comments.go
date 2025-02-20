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

// GetComments
func GetComments(sugar string, mongoDb *mongo.Database, logger *zap.Logger) http.Handler {
	oLog := logger.With(zap.String("op", "getComments"))
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

		var comments []*models.Comment

		c, err := mongoDb.Collection("comments").Find(r.Context(), bson.M{"post_id": id}, options.Find().SetSort(bson.M{"created_at": -1}).SetLimit(10))

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		users := make(map[int]models.User)

		defer c.Close(r.Context())

		for c.Next(r.Context()) {
			var comment models.Comment
			if err := c.Decode(&comment); err != nil {
				continue
			}

			if _, ok := users[comment.UserId]; !ok {
				var user models.User
				if err := mongoDb.Collection("users").FindOne(r.Context(), bson.M{"_id": comment.UserId}).Decode(&user); err != nil {
					continue
				}
				users[comment.UserId] = user
			}

			comment.ProfilePic = users[comment.UserId].ProfilePic

			comments = append(comments, &comment)
		}

		comments = funk.Reverse(comments).([]*models.Comment)

		JSON(&models.CommentListResponse{
			Code:   200,
			Result: comments,
		}, w)
		// -- end --
	})
}

// -- extra --
// -- end --
