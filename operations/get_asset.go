package operations

import (
	"net/http"
	"os"
	"path/filepath"

	"fr_book_api/models"

	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	// -- imports --
	// -- end --
)

// GetAsset
func GetAsset(mongoDb *mongo.Database, logger *zap.Logger) http.Handler {
	oLog := logger.With(zap.String("op", "getAsset"))
	// -- init --
	// -- end --
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		v := models.NewValidator(r)

		name := v.Path("name").String()

		log := oLog.With(zap.String("ip", r.Header.Get("X-Real-IP")))
		// -- code --
		if !v.Valid() {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		log.Debug("Start Operation")

		b, _ := os.ReadFile(filepath.Join("assets", name))
		w.Write(b)
		// -- end --
	})
}

// -- extra --
// -- end --
