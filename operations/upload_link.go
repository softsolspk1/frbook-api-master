package operations

import (
	"net/http"
	"path/filepath"

	"fr_book_api/models"

	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	// -- imports --
	// -- end --
)

// UploadLink
func UploadLink(mongoDb *mongo.Database, logger *zap.Logger) http.Handler {
	oLog := logger.With(zap.String("op", "uploadLink"))
	// -- init --
	// -- end --
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		v := models.NewValidator(r)

		name := v.Form("name").String()

		log := oLog.With(zap.String("ip", r.Header.Get("X-Real-IP")))
		// -- code --
		if !v.Valid() {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		log.Debug("Start Operation", zap.Any("name", name))

		response, err := http.Get(name)
		if err != nil {
			log.Error("Uplaod faieled", zap.Error(err))
			return
		}
		if response.StatusCode != 200 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		nm := filepath.Ext(name)

		url, err := UploadIO(r.Context(), "z"+nm, response.Body)

		JSON(&models.StringResponse{
			Code:   200,
			Result: url,
		}, w)
		// -- end --
	})
}

// -- extra --
// -- end --
