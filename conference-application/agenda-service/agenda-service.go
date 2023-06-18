package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"

	httpSwagger "github.com/swaggo/http-swagger/v2"

	"math/rand"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	_ "github.com/salaboy/platforms-on-k8s/conference-application/agenda-service/docs"
	"golang.org/x/exp/slices"
)

type Proposal struct {
	Id string
}

type AgendaItem struct {
	Id       string
	Proposal Proposal
	Title    string
	Author   string
	Day      string
	Time     string
}

type ServiceInfo struct {
	Name         string
	Version      string
	Source       string
	PodId        string
	PodNamespace string
	PodNodeName  string
}

var rdb *redis.Client
var KEY = "AGENDAITEMS"
var VERSION = getEnv("VERSION", "1.0.0")
var SOURCE = getEnv("SOURCE", "https://github.com/salaboy/platforms-on-k8s/tree/main/conference-application/agenda-service")
var POD_ID = getEnv("POD_ID", "N/A")
var POD_NAMESPACE = getEnv("POD_NAMESPACE", "N/A")
var POD_NODENAME = getEnv("POD_NODENAME", "N/A")
var REDIS_HOST = getEnv("REDIS_HOST", "localhost")
var REDIS_PORT = getEnv("REDIS_PORT", "6379")
var REDIS_PASSOWRD = getEnv("REDIS_PASSWORD", "")

func getAgendaByDayHandler(w http.ResponseWriter, r *http.Request) {

}

// getHighlightsHandler gets highlights from Redis.
// @Summary Show highlights
// @Description Get all highlights
// @Tags Highlight
// @Accept json
// @Product json
// @Router /highlights [get]
// @Success 200 {array} AgendaItem
func getHighlightsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	agendaItemsHashes, err := rdb.HGetAll(ctx, KEY).Result()
	if err != nil {
		panic(err)
	}
	higlights := 4
	min := 0
	max := len(agendaItemsHashes)

	var chosenOnes []int
	counter := 0
	for {
		if len(chosenOnes) == higlights {
			break
		}
		random := rand.Intn(max-min) + min
		if !slices.Contains(chosenOnes, random) {
			chosenOnes = append(chosenOnes, random)
		}

	}
	log.Printf("Chosen ones: %d", chosenOnes)

	counter = 0
	var agendaItems []AgendaItem
	for _, ai := range agendaItemsHashes {
		if slices.Contains(chosenOnes, counter) {
			var agendaItem AgendaItem
			err = json.Unmarshal([]byte(ai), &agendaItem)
			if err != nil {
				log.Printf("There was an error decoding the AgendaItem into the struct: %v", err)
			}
			agendaItems = append(agendaItems, agendaItem)
		}
		counter++
	}

	respondWithJSON(w, http.StatusOK, agendaItems)

}

// getAllAgendaItemsHandler gets all agenda item from database.
// @Summary Show highlights
// @Description Get all highlights
// @Tags Highlight
// @Accept json
// @Product json
// @Router /highlights [get]
// @Success 200 {array} AgendaItem
func getAllAgendaItemsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	agendaItemsHashs, err := rdb.HGetAll(ctx, KEY).Result()
	if err != nil {
		panic(err)
	}
	var agendaItems []AgendaItem

	for _, ai := range agendaItemsHashs {
		var agendaItem AgendaItem
		err = json.Unmarshal([]byte(ai), &agendaItem)
		if err != nil {
			log.Printf("There was an error decoding the AgendaItem into the struct: %v", err)
		}
		agendaItems = append(agendaItems, agendaItem)
	}
	log.Printf("Agenda Items retrieved from Database: %d", len(agendaItems))
	respondWithJSON(w, http.StatusOK, agendaItems)

}

func NewGetAllAgendaItemsHandler(redisClient *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()
		agendaItemsHashs, err := redisClient.HGetAll(ctx, KEY).Result()
		if err != nil {
			panic(err)
		}
		var agendaItems []AgendaItem

		for _, ai := range agendaItemsHashs {
			var agendaItem AgendaItem
			err = json.Unmarshal([]byte(ai), &agendaItem)
			if err != nil {
				log.Printf("There was an error decoding the AgendaItem into the struct: %v", err)
			}
			agendaItems = append(agendaItems, agendaItem)
		}
		log.Printf("Agenda Items retrieved from Database: %d", len(agendaItems))
		respondWithJSON(w, http.StatusOK, agendaItems)
	}
}

// getAgendaItemByIdHandler gets an agenda item by ID.
// @Summary Show Agenda Item
// @Description Gets an AgendaItem by ID
// @Tags Agenda
// @Accept json
// @Produce json
// @Param ID path string true "AgendaItem ID"
// @Router /{ID} [get]
// @Success 200 {object} AgendaItem
func getAgendaItemByIdHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	agendaItemId := chi.URLParam(r, "id")
	log.Printf("Fetching Agenda Item By Id: %s", agendaItemId)
	agendaItemById, err := rdb.HGet(ctx, KEY, agendaItemId).Result()
	if err != nil {
		panic(err)
	}
	var agendaItem AgendaItem
	err = json.Unmarshal([]byte(agendaItemById), &agendaItem)
	if err != nil {
		log.Printf("There was an error decoding the request body into the struct: %v", err)
	}
	log.Printf("Agenda Item retrieved from Database: %s", agendaItem)
	respondWithJSON(w, http.StatusOK, agendaItem)
}

// newAgendaItemHandler creates a new agenda item.
// @Summary Creates a new agenda item
// @Description Creates a new agenda item
// @Tags Agenda
// @Accept json
// @Produce json
// @Router / [post]
// @Success 200 {object} AgendaItem
func newAgendaItemHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Creating new AgendaItem")
	ctx := context.Background()

	var agendaItem AgendaItem
	err := json.NewDecoder(r.Body).Decode(&agendaItem)
	if err != nil {
		log.Printf("There was an error decoding the request body into the struct: %v", err)
	}

	// @TODO: write fail scenario (check for fail string in title return 500)

	agendaItem.Id = uuid.New().String()

	err = rdb.HSetNX(ctx, KEY, agendaItem.Id, agendaItem).Err()
	if err != nil {
		panic(err)
	}

	log.Printf("Agenda Item Stored in Database: %s", agendaItem)

	// @TODO avoid doing two marshals to json
	respondWithJSON(w, http.StatusOK, agendaItem)
}

func NewCreateAgendaItemHandler(redisClient *redis.Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Creating new AgendaItem")
		ctx := context.Background()

		var agendaItem AgendaItem
		err := json.NewDecoder(r.Body).Decode(&agendaItem)
		if err != nil {
			log.Printf("There was an error decoding the request body into the struct: %v", err)
		}

		// @TODO: write fail scenario (check for fail string in title return 500)

		agendaItem.Id = uuid.New().String()

		err = redisClient.HSetNX(ctx, KEY, agendaItem.Id, agendaItem).Err()
		if err != nil {
			panic(err)
		}

		log.Printf("Agenda Item Stored in Database: %s", agendaItem)

		// @TODO avoid doing two marshals to json
		respondWithJSON(w, http.StatusOK, agendaItem)
	}
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

// @title Agenda Service API
// @version 0.0.1
// @description REST API to provide Agenda features.
// @termsOfService	http://swagger.io/terms/
// @contact.name	Mauricio Salatino @salaboy
// @contact.url	https://www.salaboy.com
// @contact.email	salaboy@salaboy.com
// @license.name	Apache 2.0
// @license.url	http://www.apache.org/licenses/LICENSE-2.0.html
// @host		localhost:8080
// @BasePath	/
func main() {
	appPort := os.Getenv("APP_PORT")
	if appPort == "" {
		appPort = "8080"
	}

	log.Printf("Starting Agenda Service in Port: %s", appPort)

	rdb = redis.NewClient(&redis.Options{
		Addr:     REDIS_HOST + ":" + REDIS_PORT,
		Password: REDIS_PASSOWRD, // no password set
		DB:       0,              // use default DB
	})

	log.Printf("Connected to Redis.")

	server := NewAgendaServer(rdb)

	// Add Swagger 2.0
	server.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("http://localhost:"+appPort+"/swagger/doc.json"),
	))

	err := http.ListenAndServe(":"+appPort, server)
	if err != nil {
		log.Panic(err)
	}
}

func (p Proposal) MarshalBinary() ([]byte, error) {
	return json.Marshal(p)
}

func (p Proposal) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, &p); err != nil {
		return err
	}

	return nil
}

func (ai AgendaItem) MarshalBinary() ([]byte, error) {
	return json.Marshal(ai)
}

func (ai AgendaItem) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, &ai); err != nil {
		return err
	}

	return nil
}

func getEnv(key, fallback string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		value = fallback
	}
	return value
}

// NewAgendaServer creates a new chi.Mux with all necessaries http.HandlerFunc handlers, to make the agenda-service
// running.
func NewAgendaServer(redisClient *redis.Client) *chi.Mux {

	r := chi.NewRouter()

	// Dapr subscription routes orders topic to this route
	r.Post("/", NewCreateAgendaItemHandler(redisClient))
	r.Get("/", NewGetAllAgendaItemsHandler(redisClient))
	r.Get("/highlights", getHighlightsHandler)
	r.Get("/{id}", getAgendaItemByIdHandler)
	r.Get("/day/{day}", getAgendaByDayHandler)

	// r.Delete("/{id}", deleteAgendaItemHandler)
	// r.Delete("/", deleteAllHandler)

	// Add health check
	r.HandleFunc("/health/{endpoint:readiness|liveness}", healthCheck)

	// Service info
	r.HandleFunc("/service/info", serviceInfo)

	return r
}

func serviceInfo(w http.ResponseWriter, r *http.Request) {
	var info = ServiceInfo{
		Name:         "AGENDA",
		Version:      VERSION,
		Source:       SOURCE,
		PodId:        POD_ID,
		PodNamespace: POD_NODENAME,
	}
	json.NewEncoder(w).Encode(info)
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}
