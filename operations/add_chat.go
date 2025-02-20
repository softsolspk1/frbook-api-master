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

// AddChat
func AddChat(sugar string, mongoDb *mongo.Database, logger *zap.Logger) http.Handler {
	oLog := logger.With(zap.String("op", "addChat"))
	// -- init --

	chatId, _ := models.NewIDNode(8)

	// -- end --
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		v := models.NewValidator(r).Secret(sugar)

		userId := v.Token("user_id").Int()

		toId := v.Form("to_id").Int()

		content := v.Form("content").String()

		log := oLog.With(zap.String("ip", r.Header.Get("X-Real-IP")))
		// -- code --
		if !v.Valid() {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		log.Debug("Start Operation", zap.Any("user_id", userId), zap.Any("to_id", toId), zap.Any("content", content))

		if content == "" {
			return
		}

		mongoDb.Collection("chats").InsertOne(r.Context(), &models.Chat{
			Id:        int(chatId.Generate().Int64()),
			FromId:    userId,
			ToId:      toId,
			Content:   content,
			CreatedAt: time.Now(),
		})

		JSON(&models.StatusResponse{
			Code: 200,
		}, w)
		// -- end --
	})
}

// -- extra --
// -- end --
