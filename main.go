package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	StatusActive   = "active"
	StatusInactive = "inactive"
)

type Country struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name           string             `bson:"name" json:"name"`
	IsoCode        string             `bson:"iso_code" json:"iso_code"`
	DialCode       string             `bson:"dial_code" json:"dial_code"`
	FlagURL        string             `bson:"flag_url" json:"flag_url"`
	CurrencyCode   string             `bson:"currency_code" json:"currency_code"`
	CurrencySymbol string             `bson:"currency_symbol" json:"currency_symbol"`
	CurrencyName   string             `bson:"currency_name" json:"currency_name"`
	Status         string             `bson:"status" json:"status"`
}

type State struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name      string             `bson:"name" json:"name"`
	CountryID primitive.ObjectID `bson:"country_id" json:"country_id"`
}

type LGA struct {
	ID      primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name    string             `bson:"name" json:"name"`
	StateID primitive.ObjectID `bson:"state_id" json:"state_id"`
}

type City struct {
	ID      primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name    string             `bson:"name" json:"name"`
	StateID primitive.ObjectID `bson:"state_id" json:"state_id"`
}

type ExternalCountry struct {
	CountryID      int    `json:"countryId"`
	Name           string `json:"name"`
	IsoCode        string `json:"isoCode"`
	DialCode       string `json:"dialCode"`
	FlagURL        string `json:"flagUrl"`
	CurrencyCode   string `json:"currencyCode"`
	CurrencySymbol string `json:"currencySymbol"`
	CurrencyName   string `json:"currencyName"`
	StatusEnum     int    `json:"statusEnum"`
}

type ExternalState struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type ExternalLGA struct {
	Name string `json:"name"`
}

func connectMongo() *mongo.Database {
	client, err := mongo.NewClient(options.Client().ApplyURI(os.Getenv("MONGO_URI")))
	if err != nil {
		log.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := client.Connect(ctx); err != nil {
		log.Fatal(err)
	}
	return client.Database("geo_data")
}

func seedData(db *mongo.Database, baseURL string) error {
	ctx := context.Background()

	// clear old data to avoid duplicates
	db.Collection("countries").DeleteMany(ctx, bson.M{})
	db.Collection("states").DeleteMany(ctx, bson.M{})
	db.Collection("lgas").DeleteMany(ctx, bson.M{})
	db.Collection("cities").DeleteMany(ctx, bson.M{})

	countriesCol := db.Collection("countries")
	statesCol := db.Collection("states")
	lgasCol := db.Collection("lgas")

	resp, err := http.Get(fmt.Sprintf("%s/api/Countries?status=active&page=1", baseURL))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var countryResp struct {
		Items []ExternalCountry `json:"items"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&countryResp); err != nil {
		return err
	}

	for _, c := range countryResp.Items {
		status := StatusInactive
		if c.StatusEnum == 1 {
			status = StatusActive
		}

		countryOID := primitive.NewObjectID()
		_, err := countriesCol.InsertOne(ctx, Country{
			ID:             countryOID,
			Name:           c.Name,
			IsoCode:        c.IsoCode,
			DialCode:       c.DialCode,
			FlagURL:        c.FlagURL,
			CurrencyCode:   c.CurrencyCode,
			CurrencySymbol: c.CurrencySymbol,
			CurrencyName:   c.CurrencyName,
			Status:         status,
		})
		if err != nil {
			return err
		}

		// Get states
		stateResp, err := http.Get(fmt.Sprintf("%s/api/Countries/get-states-by-countryId/%d", baseURL, c.CountryID))
		if err != nil {
			return err
		}
		defer stateResp.Body.Close()

		var states []ExternalState
		if err := json.NewDecoder(stateResp.Body).Decode(&states); err != nil {
			return err
		}

		for _, s := range states {
			stateOID := primitive.NewObjectID()
			_, err := statesCol.InsertOne(ctx, State{
				ID:        stateOID,
				Name:      s.Name,
				CountryID: countryOID,
			})
			if err != nil {
				return err
			}

			// Get LGAs
			lgaResp, err := http.Get(fmt.Sprintf("%s/api/Countries/lga-by-stateId/%d", baseURL, s.ID))
			if err != nil {
				return err
			}
			defer lgaResp.Body.Close()

			var lgas []ExternalLGA
			if err := json.NewDecoder(lgaResp.Body).Decode(&lgas); err != nil {
				return err
			}

			for _, l := range lgas {
				_, err := lgasCol.InsertOne(ctx, LGA{
					ID:      primitive.NewObjectID(),
					Name:    l.Name,
					StateID: stateOID,
				})
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func main() {
	godotenv.Load()
	db := connectMongo()

	if len(os.Args) > 1 && os.Args[1] == "--seed" {
		fmt.Println("Seeding data...")
		if err := seedData(db, os.Getenv("BACEND_BASEURL")); err != nil {
			log.Fatal(err)
		}
		fmt.Println("Seeding completed.")
		return
	}

	r := gin.Default()

	// Get all countries
	r.GET("/countries", func(c *gin.Context) {
		cursor, err := db.Collection("countries").Find(context.Background(), bson.M{})
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		var countries []Country
		if err := cursor.All(context.Background(), &countries); err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, countries)
	})

	// Get active countries
	r.GET("/countries/active", func(c *gin.Context) {
		cursor, err := db.Collection("countries").Find(context.Background(), bson.M{"status": StatusActive})
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		var countries []Country
		if err := cursor.All(context.Background(), &countries); err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, countries)
	})

	// Get states by country ID
	r.GET("/states/:countryId", func(c *gin.Context) {
		oid, err := primitive.ObjectIDFromHex(c.Param("countryId"))
		if err != nil {
			c.JSON(400, gin.H{"error": "invalid country id"})
			return
		}
		cursor, err := db.Collection("states").Find(context.Background(), bson.M{"country_id": oid})
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		var states []State
		if err := cursor.All(context.Background(), &states); err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, states)
	})

	// Get LGAs by state ID
	r.GET("/lgas/:stateId", func(c *gin.Context) {
		oid, err := primitive.ObjectIDFromHex(c.Param("stateId"))
		if err != nil {
			c.JSON(400, gin.H{"error": "invalid state id"})
			return
		}
		cursor, err := db.Collection("lgas").Find(context.Background(), bson.M{"state_id": oid})
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		var lgas []LGA
		if err := cursor.All(context.Background(), &lgas); err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, lgas)
	})

	// Add cities for a state
	r.POST("/cities/:stateId", func(c *gin.Context) {
		oid, err := primitive.ObjectIDFromHex(c.Param("stateId"))
		if err != nil {
			c.JSON(400, gin.H{"error": "invalid state id"})
			return
		}
		var cityNames []string
		if err := c.BindJSON(&cityNames); err != nil {
			c.JSON(400, gin.H{"error": "invalid input"})
			return
		}

		citiesCol := db.Collection("cities")
		for _, name := range cityNames {
			trimmed := strings.TrimSpace(name)
			if trimmed == "" {
				continue
			}
			// case-insensitive check
			filter := bson.M{
				"state_id": oid,
				"$expr": bson.M{
					"$eq": []interface{}{
						bson.M{"$toLower": "$name"},
						strings.ToLower(trimmed),
					},
				},
			}
			count, err := citiesCol.CountDocuments(context.Background(), filter)
			if err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}
			if count == 0 {
				_, err := citiesCol.InsertOne(context.Background(), City{
					ID:      primitive.NewObjectID(),
					Name:    trimmed,
					StateID: oid,
				})
				if err != nil {
					c.JSON(500, gin.H{"error": err.Error()})
					return
				}
			}
		}
		c.JSON(200, gin.H{"message": "cities added successfully"})
	})

	fmt.Println("Server running on :8080")
	r.Run(":8080")
}
