package hubs

import (
	"crypto/tls"
	"fr_book_api/actors"
	"fr_book_api/models"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"

	// -- imports --
	"gopkg.in/mail.v2"
	// -- end --
)

// SmtpServerSetup sets up things for the hub.
func SmtpServerSetup(sugar string, mongoDb *mongo.Database, logger *zap.Logger) error {
	// -- init --
	h := actors.NewHub("smtp", &SmtpServerController{}, 100, time.Minute, logger)
	h.Start()
	return nil
	// -- end --
}

// SmtpServerController is controller for the hub.
type SmtpServerController struct {
	h   *actors.Hub
	log *zap.Logger
	// -- declarations --
	d *mail.Dialer
	// -- end --
}

func (sc *SmtpServerController) Connect(h *actors.Hub, ct time.Time) {
	sc.h = h
	// -- connect --
	sc.log = h.Log()
	sender := "otp@demowebsite.world"
	password := "23Otpismaster!"

	d := mail.NewDialer("mail.demowebsite.world", 587, sender, password)

	d.TLSConfig = &tls.Config{InsecureSkipVerify: true} // Skip certificate verification for simplicity

	sc.d = d
	sc.log.Info("Connected")
	// -- end --
}

func (sc *SmtpServerController) OnShutdown() {
	// -- shutdown --
	// -- end --
}

func (sc *SmtpServerController) ParseUser(b []byte) (actors.Serializable, error) {
	// -- parse-user --
	return nil, nil
	// -- end --
}

func (sc *SmtpServerController) ParseHub(b []byte, client actors.HubClient) (actors.Serializable, error) {
	// -- parse-hub --
	return nil, nil
	// -- end --
}

func (sc *SmtpServerController) ProcessUser(d actors.Serializable, userId int64, ct time.Time) {
	// -- process-user --
	// -- end --
}

func (sc *SmtpServerController) State(userId int64) []actors.Serializable {
	// -- state --
	return nil
	// -- end --
}

func (sc *SmtpServerController) ProcessDefault(o actors.Serializable, c *actors.OneTimeClient, ct time.Time) {
	// -- process-default --

	event := o.(*models.SmsEvent)

	sender := "otp@demowebsite.world"
	recipient := event.Email
	m := mail.NewMessage()
	m.SetHeader("From", sender)
	m.SetHeader("To", recipient)
	m.SetHeader("Subject", event.Subject)
	m.SetBody("text/plain", event.Content)

	if err := sc.d.DialAndSend(m); err != nil {
		sc.log.Error("Error sending email", zap.Error(err), zap.String("email", event.Email))
		c.Msg(&models.SmsEvent{
			Success: false,
		})
	} else {
		// sc.log.Info("Email sent", zap.String("email", in.User.Email))
		c.Msg(&models.SmsEvent{
			Success: true,
		})
	}

	// -- end --
}

func (sc *SmtpServerController) ProcessHub(d actors.Serializable, c actors.HubClient, ct time.Time) {
	// -- process-hub --
	// -- end --
}

func (sc *SmtpServerController) Tick(ct time.Time) {
	// -- tick --
	// -- end --
}

func (sc *SmtpServerController) OnDisconnect(userId int64) {
	// -- disconnect --
	// -- end --
}

func (sc *SmtpServerController) OnPanic() {
	// -- panic --
	// -- end --
}

func (sc *SmtpServerController) Healthz() string {
	// -- health --
	return ""
	// -- end --
}

// -- code --
// -- end --
