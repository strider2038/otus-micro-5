package racetest

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
)

const (
	domain       = "http://arch.homework"
	requestCount = 10
)

func TestCreateOrder_RaceTest(t *testing.T) {
	email := gofakeit.Email()
	registerUserByEmail(t, email)
	accessToken := loginUserByEmail(t, email)
	eTag := getOrdersETag(t, email, accessToken)

	responses := make(chan int, requestCount)
	for i := 0; i < requestCount; i++ {
		go createUserOrder(t, email, eTag, accessToken, responses)
	}
	var i, numOf202, numOf400 int
	for code := range responses {
		if code == http.StatusAccepted {
			numOf202++
		} else if code == http.StatusPreconditionFailed || code == http.StatusConflict {
			numOf400++
		} else {
			t.Errorf("unexpected response code: %d", code)
		}
		i++
		if i == requestCount {
			break
		}
	}

	if numOf202 != 1 {
		t.Errorf("expected only one 202 code, actual count is %d", numOf202)
	}
	if numOf400 != requestCount-1 {
		t.Errorf("expected count of 409/412 code is %d, actual count is %d", requestCount-1, numOf400)
	}
}

func registerUserByEmail(t *testing.T, email string) {
	response, err := http.DefaultClient.Post(
		domain+"/api/v1/identity/register",
		"application/json",
		strings.NewReader(
			fmt.Sprintf(
				`{
					"email": "%s",
					"password": "test_pa$$word",
					"firstName": "John",
					"lastName": "Doe"
				}`,
				email,
			),
		),
	)
	if err != nil {
		t.Fatalf("failed to register user %s: %s", email, err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusCreated {
		s, _ := io.ReadAll(response.Body)
		t.Fatalf("failed to register user %s: http code %d, body: %s", email, response.StatusCode, string(s))
	}
}

func loginUserByEmail(t *testing.T, email string) string {
	response, err := http.DefaultClient.Post(
		domain+"/api/v1/identity/login",
		"application/json",
		strings.NewReader(
			fmt.Sprintf(
				`{
					"email": "%s",
					"password": "test_pa$$word"
				}`,
				email,
			),
		),
	)
	if err != nil {
		t.Fatalf("failed to login user %s: %s", email, err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		s, _ := io.ReadAll(response.Body)
		t.Fatalf("failed to login user %s: http code %d, body: %s", email, response.StatusCode, string(s))
	}

	var loginResponse struct {
		AccessToken string `json:"accessToken"`
	}
	err = json.NewDecoder(response.Body).Decode(&loginResponse)
	if err != nil {
		t.Fatal("failed to decode login response:", err)
	}
	if loginResponse.AccessToken == "" {
		t.Fatal("access token is empty")
	}

	return loginResponse.AccessToken
}

func getOrdersETag(t *testing.T, email, token string) string {
	request, err := http.NewRequest(http.MethodGet, domain+"/api/v1/ordering/orders", nil)
	if err != nil {
		t.Fatalf("failed to make request: %s", err)
	}
	request.Header.Set("Authorization", "Bearer "+token)
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatalf("failed to get user %s orders: %s", email, err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		s, _ := io.ReadAll(response.Body)
		t.Fatalf("failed to get user %s orders: http code %d, body: %s", email, response.StatusCode, string(s))
	}

	etag := response.Header.Get("etag")
	if etag == "" {
		t.Fatalf("ETag header is empty")
	}

	return etag
}

func createUserOrder(t *testing.T, email, etag, token string, statuses chan int) {
	request, err := http.NewRequest(http.MethodPost, domain+"/api/v1/ordering/orders", strings.NewReader(`{"price": 10}`))
	if err != nil {
		t.Fatalf("failed to make request: %s", err)
	}
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("If-Match", etag)
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatalf("failed to create user %s order: %s", email, err)
	}
	defer response.Body.Close()

	statuses <- response.StatusCode
}
