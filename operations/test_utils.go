package operations

import (
	"context"
	"net/http"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func addQueryParams(r *http.Request, params map[string]string) *http.Request {
	q := r.URL.Query()
	for k, v := range params {
		q.Add(k, v)
	}
	r.URL.RawQuery = q.Encode()
	return r
}

func addTokenParams(r *http.Request, tokens jwt.MapClaims, sugar string) (*http.Request, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, tokens)

	signedToken, err := token.SignedString([]byte(sugar))
	if err != nil {
		return nil, err
	}
	r.Header.Set("jwt", signedToken)
	return r, nil
}

func mongoDB(dbName string) (*mongo.Database, error) {
	ctx, c1 := context.WithTimeout(context.Background(), 10*time.Second)
	defer c1()
	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		return nil, err
	}

	ctx, c2 := context.WithTimeout(context.Background(), 10*time.Second)
	defer c2()
	err = mongoClient.Ping(ctx, readpref.Primary())
	if err != nil {
		return nil, err
	}

	db := mongoClient.Database(dbName)
	err = db.Drop(context.Background())
	return db, err
}
