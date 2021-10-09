package main

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type User struct {
	ID       primitive.ObjectID `json: "_id,omitempty"`
	Name     string             `json: "name"`
	Email    string             `json: "email"`
	Password string             `json: "password"`
}

type Post struct {
	ID        primitive.ObjectID `json: "_id,omitempty"`
	Caption   string             `json: "caption"`
	ImageURL  string             `json: "imageURL"`
	Timestamp time.Time          `json: "timestamp"`
}

var client *mongo.Client

func hashPassword(pswd string) string {
	data := []byte(pswd)
	hash := sha256.Sum256(data)
	return base64.StdEncoding.EncodeToString(hash[:])
}

func createUserHandler(res http.ResponseWriter, req *http.Request) {
	var user User
	json.NewDecoder(req.Body).Decode(&user)
	usersCol := client.Database("Anuvab_Insta").Collection("users")
	ctx, _ := context.WithTimeout(context.Background(), 15*time.Second)
	cursor, err := usersCol.Find(ctx, bson.M{})
	for cursor.Next(ctx) {
		var backlogUser User
		cursor.Decode(&backlogUser)
		if backlogUser.Email == user.Email {
			res.WriteHeader(http.StatusInternalServerError)
			res.Write([]byte(`{"This e-mail is already registered!":"` + err.Error() + `"}`))
			return
		}
	}
	hashedPswd := hashPassword(user.Password)
	user.Password = hashedPswd

	userResult, insertErrorUser := usersCol.InsertOne(ctx, user)
	if insertErrorUser != nil {
		fmt.Println("Error while creating user: ", insertErrorUser)
	} else {
		json.NewEncoder(res).Encode(userResult)
		userID := userResult.InsertedID
		fmt.Println("New user id: ", userID)
	}

	res.Header().Add("content-type", "application/json")
	res.WriteHeader(http.StatusOK)
}

func createPostHandler(res http.ResponseWriter, req *http.Request) {
	var post Post
	post.Timestamp = time.Now()
	json.NewDecoder(req.Body).Decode(&post)

	postsCol := client.Database("Anuvab_Insta").Collection("posts")
	ctx, _ := context.WithTimeout(context.Background(), 15*time.Second)
	postResult, insertErrorPost := postsCol.InsertOne(ctx, post)

	if insertErrorPost != nil {
		fmt.Println("Error while creating post: ", insertErrorPost)
	} else {
		newPostID := postResult.InsertedID
		fmt.Println("New post ID: ", newPostID)
	}

	res.Header().Add("content-type", "application/json")
	res.WriteHeader(http.StatusOK)
}

func main() {
	fmt.Println("Main.go up and running!")

	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	ctx, _ := context.WithTimeout(context.Background(), 15*time.Second)
	client, _ = mongo.Connect(ctx, clientOptions)

	http.HandleFunc("/users", createUserHandler)
	http.HandleFunc("/posts", createPostHandler)

	httpErr := http.ListenAndServe(":5000", nil)

	if httpErr != nil {
		panic(httpErr)
	}
}
