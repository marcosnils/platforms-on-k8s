package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redismock/v9"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Redis for testing
var db, mock = redismock.NewClientMock()

func testServer() *httptest.Server {
	server := NewAgendaServer(db)
	return httptest.NewServer(server)
}

func Test_healthCheck(t *testing.T) {

	ts := testServer()
	defer ts.Close()
	t.Run("it should return 200 when GET '/health/readiness'", func(t *testing.T) {
		// arrange, act
		res, _ := http.Get(fmt.Sprintf("%s/health/readiness", ts.URL))

		// assert
		assert.Equal(t, http.StatusOK, res.StatusCode)
	})

	t.Run("it should return 200 when GET '/health/liveness'", func(t *testing.T) {
		// arrange, act
		res, _ := http.Get(fmt.Sprintf("%s/health/liveness", ts.URL))

		// assert
		assert.Equal(t, http.StatusOK, res.StatusCode)
	})
}

func Test_newAgendaItemHandler(t *testing.T) {

	ts := testServer()
	defer ts.Close()

	t.Run("it should return 201 when POST '/' execute successfully", func(t *testing.T) {

		// arrange
		mock.Regexp().ExpectHSetNX(KEY, "", ".*").SetVal(true)
		agendaItem := AgendaItem{
			Proposal: Proposal{
				Id: uuid.NewString(),
			},
			Title:  "Platform Engineering on K8S",
			Author: "Mauricio Salatino",
			Day:    "2023-12-18",
			Time:   "20:00:00Z",
		}
		agendaItemAsBytes, _ := json.Marshal(agendaItem)
		body := bytes.NewBuffer(agendaItemAsBytes)

		// act
		res, _ := http.Post(ts.URL, "application/json", body)

		// assert
		assert.Equal(t, http.StatusOK, res.StatusCode)

		// check
		var newAgendaItem AgendaItem
		defer res.Body.Close()
		json.NewDecoder(res.Body).Decode(&newAgendaItem)

		assert.Equal(t, newAgendaItem.Time, agendaItem.Time)

		// clear mock
		mock.ClearExpect()
	})
}

func Test_newGetAllAgendaItemsHandler(t *testing.T) {
	ts := testServer()
	defer ts.Close()

	t.Run("it should return 200 when GET '/' execute successfully", func(t *testing.T) {
		// arrange
		mock.ExpectHGetAll(KEY).SetVal(map[string]string{})

		// act
		res, _ := http.Get(fmt.Sprintf("%s/", ts.URL))

		// assert
		assert.Equal(t, res.StatusCode, http.StatusOK)

		// clear mock
		mock.ClearExpect()
	})

	t.Run("it should return 200 with one AgendaItem", func(t *testing.T) {
		// arrange
		agendaItem := []AgendaItem{
			{
				Proposal: Proposal{
					Id: uuid.NewString(),
				},
				Title:  "Platform Engineering on K8S",
				Author: "Mauricio Salatino",
				Day:    "2023-12-18",
				Time:   "20:00:00Z",
			},
		}
		agendaItemAsBytes, _ := json.Marshal(agendaItem)
		mock.ExpectHGetAll(KEY).SetVal(map[string]string{
			KEY: string(agendaItemAsBytes),
		})

		// act
		res, _ := http.Get(fmt.Sprintf("%s/", ts.URL))
		var agendaItems []AgendaItem
		defer res.Body.Close()
		json.NewDecoder(res.Body).Decode(&agendaItems)

		assert.Equal(t, 1, len(agendaItems))
	})
}
