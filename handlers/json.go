package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/olimeme/helpers"
)

var profanity = map[string]string{
	"kerfuffle": "****", 
	"sharbert": "****", 
	"fornax": "****", 
}

func replaceProfanity(msg string) string {
	words := strings.Split(msg, " ")

	for i, word := range words {
		if replacement, exists := profanity[strings.ToLower(word)]; exists {
			words[i] = replacement
		}
	}
	return strings.Join(words, " ")
}


func ValidateChirtp(res http.ResponseWriter, req *http.Request) {
	jsonBody := helpers.BodyJson{}
	decoder := json.NewDecoder(req.Body)
    err := decoder.Decode(&jsonBody)
    if err != nil {
		helpers.RespondWithError(res, 400, "Something went wrong")
		return
    }

	if len(jsonBody.Body) > 140 {
		helpers.RespondWithError(res, 400, "Chirp is too long")
		return 
	}

	processedMsg := replaceProfanity(jsonBody.Body)

	helpers.RespondWithJSON(res, 200, map[string]any{
		"cleaned_body": processedMsg,
	})
	return 
}

