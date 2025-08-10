package main

import (
	"context"
	"embed"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

//go:embed static/*
var staticFS embed.FS

type Question struct {
	ID     string `json:"id" bson:"id"`
	Game   string `json:"game" bson:"game"`
	Prompt string `json:"prompt" bson:"prompt"`
	Hint   string `json:"hint,omitempty" bson:"hint,omitempty"`
}

var seedQuestions = []interface{}{
	Question{ID: "wwyd-1", Game: "wwyd", Prompt: "You find ₦200,000 in a taxi you just left. Do you: A) Try to find the owner, B) Hand it into the nearest police station, C) Keep it?", Hint: "Discuss honesty vs risk"},
	Question{ID: "wwyd-2", Game: "wwyd", Prompt: "Your friend asks you to cover for them at work for a day while they attend an important interview. Their boss is strict. Do you help?", Hint: "Boundaries and loyalty"},
	Question{ID: "wwyd-3", Game: "wwyd", Prompt: "You have the chance to move abroad for 2 years with a big pay bump but your partner doesn't want to go. What do you do?", Hint: "Career vs relationships"},
	Question{ID: "hot-1", Game: "hotseat", Prompt: "What’s the biggest mistake you made that taught you the most?"},
	Question{ID: "hot-2", Game: "hotseat", Prompt: "What’s a secret hobby you’ve never told anyone about?"},
	Question{ID: "hot-3", Game: "hotseat", Prompt: "If you could go back and give your 18-year-old self one piece of advice, what would it be?"},
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
	}

	clientOpts := options.Client().ApplyURI(mongoURI)
	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		log.Fatalf("mongo connect: %v", err)
	}
	defer client.Disconnect(ctx)

	db := client.Database("wwyd_game")
	col := db.Collection("questions")

	count, err := col.CountDocuments(ctx, bson.M{})
	if err != nil {
		log.Fatalf("count: %v", err)
	}
	if count == 0 {
		log.Println("seeding questions into MongoDB...")
		if _, err := col.InsertMany(ctx, seedQuestions); err != nil {
			log.Fatalf("insert many: %v", err)
		}
	} else {
		log.Printf("questions already present: %d\n", count)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			data, err := staticFS.ReadFile("static/default.html")
			if err != nil {
				http.Error(w, "file not found", http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Type", "text/html")
			w.Write(data)
			return
		}
		http.FileServer(http.FS(staticFS)).ServeHTTP(w, r)
	})

	http.HandleFunc("/api/questions", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		cur, err := col.Find(ctx, bson.M{})
		if err != nil {
			http.Error(w, "db find error", http.StatusInternalServerError)
			return
		}
		defer cur.Close(ctx)

		var out []Question
		for cur.Next(ctx) {
			var q Question
			if err := cur.Decode(&q); err != nil {
				log.Println("decode err:", err)
				continue
			}
			out = append(out, q)
		}
		json.NewEncoder(w).Encode(out)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("listening on :%s (serving static files)", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
