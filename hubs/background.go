package hubs

import (
	"fr_book_api/actors"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	// -- imports --
	// -- end --
)

// BackgroundSetup sets up things for the hub.
func BackgroundSetup(sugar string, mongoDb *mongo.Database, logger *zap.Logger) error {
	// -- init --
	return nil
	// -- end --
}

// BackgroundController is controller for the hub.
type BackgroundController struct {
	h   *actors.Hub
	log *zap.Logger
	// -- declarations --
	// -- end --
}

func (bc *BackgroundController) Connect(h *actors.Hub, ct time.Time) {
	bc.h = h
	// -- connect --
	bc.log = h.Log()
	// -- end --
}

func (bc *BackgroundController) OnShutdown() {
	// -- shutdown --
	// -- end --
}

func (bc *BackgroundController) ParseUser(b []byte) (actors.Serializable, error) {
	// -- parse-user --
	return nil, nil
	// -- end --
}

func (bc *BackgroundController) ParseHub(b []byte, client actors.HubClient) (actors.Serializable, error) {
	// -- parse-hub --
	return nil, nil
	// -- end --
}

func (bc *BackgroundController) ProcessUser(d actors.Serializable, userId int64, ct time.Time) {
	// -- process-user --
	// -- end --
}

func (bc *BackgroundController) State(userId int64) []actors.Serializable {
	// -- state --
	return nil
	// -- end --
}

func (bc *BackgroundController) ProcessDefault(o actors.Serializable, c *actors.OneTimeClient, ct time.Time) {
	// -- process-default --
	// -- end --
}

func (bc *BackgroundController) ProcessHub(d actors.Serializable, c actors.HubClient, ct time.Time) {
	// -- process-hub --
	// -- end --
}

func (bc *BackgroundController) Tick(ct time.Time) {
	// -- tick --
	// -- end --
}

func (bc *BackgroundController) OnDisconnect(userId int64) {
	// -- disconnect --
	// -- end --
}

func (bc *BackgroundController) OnPanic() {
	// -- panic --
	// -- end --
}

func (bc *BackgroundController) Healthz() string {
	// -- health --
	return ""
	// -- end --
}

// -- code --
// -- end --
