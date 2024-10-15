package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/mistermoe/httpr"
)

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterResponse struct {
	ID    int    `json:"id"`
	Token string `json:"token"`
}

type RegisterErrorResponse struct {
	Error string `json:"error"`
}

func main() {
	client := httpr.NewClient(
		httpr.BaseURL("https://reqres.in"),
		httpr.Inspect(),
	)

	reqBody := RegisterRequest{
		Email: "eve.holt@reqres.in",
		// Email:    "moegrammer@hehe.gov", // uncomment for 400 response
		Password: "wowsosecret",
	}

	var respBody RegisterResponse
	var errBody RegisterErrorResponse

	resp, err := client.Post(
		context.Background(),
		"/api/register",
		httpr.RequestBodyJSON(reqBody),
		httpr.ResponseBodyJSON(&respBody, &errBody),
	)

	if err != nil {
		log.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode == http.StatusBadRequest {
		fmt.Printf("(%v): %v\n", resp.StatusCode, errBody.Error)
	} else {
		fmt.Printf("registration successful: %v\n", respBody.ID)
	}
}
