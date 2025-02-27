package main

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"fr_book_api/operations"
	"net/http"

	"fr_book_api/actors"
	"fr_book_api/hubs"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/jessevdk/go-flags"
	"github.com/ztrue/shutdown"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	// -- imports --
	// -- end --
)

type MongoOptions struct {
	Uri      string `long:"uri" description:"Enter Mongo URI to connect to" default:"mongodb+srv://softsolspak:SFWZ9evKS69CdQSx@respire.9xsja.mongodb.net/"`
	Database string `long:"database" description:"Which MongoDatabse to connect to" default:"frbook"`
}

type Options struct {
	Host  string `short:"h" long:"host" description:"What host" default:""`
	Port  int    `short:"p" long:"port" description:"Enter port to run the server on" default:"8000"`
	Sugar string `long:"sugar" description:"Secret Sugar for Signing Tokens"`
	Debug bool   `long:"debug" short:"d"`

	Mongo *MongoOptions `group:"mongo" namespace:"mongo"`

	ScreenshotScript string `long:"screenshot_script" `
	UploadBucket     string `long:"upload_bucket" `

	// -- options --
	// -- end --
}

func main() {
	var opts Options
	_, err := flags.ParseArgs(&opts, os.Args)
	if err != nil {
		panic("Could not parse command line args")
	}
	var logger *zap.Logger

	var config zap.Config
	if opts.Debug {
		config = zap.NewDevelopmentConfig()
	} else {
		config = zap.NewProductionConfig()
	}
	config.EncoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format("02-01-2006 15:04:05 -0700 MST"))
	}
	config.OutputPaths = []string{"stdout"}

	logger, _ = config.Build()

	defer logger.Sync()

	// -- before-setup --
	// -- end --

	mongoDb, err := mongoDB(opts.Mongo)
	if err != nil {
		// panic("Could not initialize mongo:" + err.Error())
	}

	// -- cache-init --
	// -- end --

	if err := hubs.CallNotifierSetup(opts.Sugar, mongoDb, logger); err != nil {
		panic(err.Error())
	}
	if err := hubs.NotifierSetup(opts.Sugar, mongoDb, logger); err != nil {
		panic(err.Error())
	}
	if err := hubs.SmtpServerSetup(opts.Sugar, mongoDb, logger); err != nil {
		panic(err.Error())
	}
	if err := hubs.BackgroundSetup(opts.Sugar, mongoDb, logger); err != nil {
		panic(err.Error())
	}

	r := mux.NewRouter()
	r.Handle("/add-chat", operations.AddChat(opts.Sugar, mongoDb, logger)).Methods("POST")
	r.Handle("/articles", operations.GetArticles(opts.Sugar, mongoDb, logger)).Methods("GET")
	r.Handle("/articles", operations.CreateArticle(opts.Sugar, mongoDb, logger)).Methods("POST")
	r.Handle("/articles/{id}", operations.GetArticle(opts.Sugar, mongoDb, logger)).Methods("GET")
	r.Handle("/articles/{id}", operations.UpdateArticle(opts.Sugar, mongoDb, logger)).Methods("POST")
	r.Handle("/assets/{name}", operations.GetAsset(mongoDb, logger)).Methods("GET")
	r.Handle("/chats", operations.GetChats(opts.Sugar, mongoDb, logger)).Methods("GET")
	r.Handle("/complete-registration", operations.CompleteRegistration(opts.Sugar, mongoDb, logger)).Methods("POST")
	r.Handle("/friend-requests", operations.GetFriendRequests(opts.Sugar, mongoDb, logger)).Methods("GET")
	r.Handle("/friend-requests", operations.AddFriendRequest(opts.Sugar, mongoDb, logger)).Methods("POST")
	r.Handle("/friend-requests/{id}/accept", operations.AcceptFriendRequest(opts.Sugar, mongoDb, logger)).Methods("POST")
	r.Handle("/friend-requests/{id}/reject", operations.RejectFriendRequest(opts.Sugar, mongoDb, logger)).Methods("POST")
	r.Handle("/friends", operations.GetFriends(opts.Sugar, mongoDb, logger)).Methods("GET")
	r.Handle("/login", operations.Login(opts.Sugar, mongoDb, logger)).Methods("POST")
	r.Handle("/me", operations.Me(opts.Sugar, mongoDb, logger)).Methods("GET")
	r.Handle("/me", operations.UpdateMe(opts.Sugar, mongoDb, logger)).Methods("POST")
	r.Handle("/notfriends", operations.GetNotFriends(opts.Sugar, mongoDb, logger)).Methods("GET")
	r.Handle("/posts", operations.GetPosts(opts.Sugar, mongoDb, logger)).Methods("GET")
	r.Handle("/posts", operations.CreatePost(opts.Sugar, mongoDb, logger)).Methods("POST")
	r.Handle("/posts/{id}/comment", operations.GetComments(opts.Sugar, mongoDb, logger)).Methods("GET")
	r.Handle("/posts/{id}/comment", operations.AddComment(opts.Sugar, mongoDb, logger)).Methods("POST")
	r.Handle("/posts/{id}/like", operations.LikePost(opts.Sugar, mongoDb, logger)).Methods("POST")
	r.Handle("/posts/{id}/unlike", operations.UnlikePost(opts.Sugar, mongoDb, logger)).Methods("POST")
	r.Handle("/start-verification", operations.StartVerification(opts.Sugar, mongoDb, logger)).Methods("POST")
	r.Handle("/uploadlink", operations.UploadLink(mongoDb, logger)).Methods("POST")
	r.Handle("/users", operations.GetUsers(opts.Sugar, mongoDb, logger)).Methods("GET")
	r.Handle("/users", operations.Register(opts.Sugar, mongoDb, logger)).Methods("POST")

	uploadHandler, err := operations.Upload(opts.UploadBucket, logger)
	if err != nil {
		panic("Upload bucket is not correctly configured")
	}
	r.Handle("/upload", uploadHandler).Methods("POST")

	r.Handle("/ws/{id}", actors.HubHandler(opts.Sugar, logger)).Methods("GET")
	r.Handle("/_hub/links", actors.HubLinks(logger)).Methods("GET")
	r.Handle("/_hub/healthz", actors.HubHealthzHandler()).Methods("GET")
	r.Handle("/_hub/kick", actors.HubKickHandler()).Methods("GET")

	// -- routes --
	// -- end --
	allowedMethods := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS", "DELETE"})
	allowedHeaders := handlers.AllowedHeaders([]string{"jwt", "build", "Content-Type", "content-type"})
	exposedHeaders := handlers.ExposedHeaders([]string{"jwt", "build", "Content-Type", "content-type"})

	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", opts.Host, opts.Port),
		Handler: handlers.CORS(exposedHeaders, allowedHeaders, allowedMethods)(r),
	}
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		server.ListenAndServe()
		wg.Done()
	}()

	shutdown.AddWithParam(func(sig os.Signal) {
		logger.Info("Received Signal", zap.Any("signal", sig))
		server.Shutdown(context.Background())
		wg.Wait()
		logger.Info("Server ShutDown complete")
		actors.ShutdownAllHubs()
		logger.Info("Hubs ShutDown complete")
	})

	logger.Info("Starting Server")
	shutdown.Listen(os.Interrupt, os.Kill)
}

func ContentHandler(f string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, f)
	})
}

func mongoDB(opts *MongoOptions) (*mongo.Database, error) {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(opts.Uri))
	if err != nil {
		return nil, err
	}

	ctx, _ = context.WithTimeout(context.Background(), 10*time.Second)
	err = mongoClient.Ping(ctx, readpref.Primary())
	if err != nil {
		return nil, err
	}

	return mongoClient.Database(opts.Database), nil
}

// -- extra --
// -- end --
