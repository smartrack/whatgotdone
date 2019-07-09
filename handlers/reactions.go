package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/mtlynch/whatgotdone/types"
)

func (s *defaultServer) reactionsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
		} else if r.Method == "GET" {
			s.handleReactionsGet(w, r)
		} else if r.Method == "POST" {
			s.handleReactionsPost(w, r)
		} else {
			log.Printf("Invalid method for drafts handler: %s", r.Method)
			http.Error(w, "Invalid operation", http.StatusBadRequest)
		}
	}
}

func (s defaultServer) handleReactionsGet(w http.ResponseWriter, r *http.Request) {
	date, err := dateFromRequestPath(r)
	if err != nil {
		log.Printf("Invalid date: %s", date)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	reactions, err := s.datastore.GetReactions(usernameFromRequestPath(r), date)
	if err != nil {
		log.Printf("Failed to retrieve reactions: %s", err)
		http.Error(w, "Failed to retrieve reactions", http.StatusInternalServerError)
		return
	}

	reactionsFiltered := []types.Reaction{}
	for _, reaction := range reactions {
		if reaction.Symbol != "" {
			reactionsFiltered = append(reactionsFiltered, reaction)
		}
	}

	if err := json.NewEncoder(w).Encode(reactionsFiltered); err != nil {
		panic(err)
	}
}

func (s defaultServer) handleReactionsPost(w http.ResponseWriter, r *http.Request) {
	username, err := s.loggedInUser(r)
	if err != nil {
		http.Error(w, "You must log in to provide a reaction", http.StatusForbidden)
		return
	}

	reactionSymbol, err := reactionSymbolFromRequest(r)
	if err != nil {
		log.Printf("Invalid reactions request: %v", err)
		http.Error(w, "Invalid reactions request", http.StatusBadRequest)
	}

	entryAuthor := usernameFromRequestPath(r)

	entryDate, err := dateFromRequestPath(r)
	if err != nil {
		log.Printf("Invalid date: %s", entryDate)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	reaction := types.Reaction{
		Username:  username,
		Timestamp: time.Now().Format(time.RFC3339),
		Symbol:    reactionSymbol,
	}
	err = s.datastore.AddReaction(entryAuthor, entryDate, reaction)
	if err != nil {
		log.Printf("Failed to add reaction: %s", err)
		http.Error(w, "Failed to add reaction", http.StatusInternalServerError)
		return
	}

	type reactionResponse struct {
		Ok bool `json:"ok"`
	}
	resp := reactionResponse{
		Ok: true,
	}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		panic(err)
	}
}

func reactionSymbolFromRequest(r *http.Request) (string, error) {
	type reactionRequest struct {
		ReactionSymbol string `json:"reactionSymbol"`
	}
	var rr reactionRequest
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&rr)
	if err != nil {
		return "", err
	}

	if !isValidReaction(rr.ReactionSymbol) {
		return "", fmt.Errorf("Invalid reaction choice: %s", rr.ReactionSymbol)
	}

	return rr.ReactionSymbol, nil
}

func isValidReaction(reaction string) bool {
	validReactionSymbols := [...]string{"", "👍", "🙁", "🎉"}
	for _, v := range validReactionSymbols {
		if reaction == v {
			return true
		}
	}
	return false
}
