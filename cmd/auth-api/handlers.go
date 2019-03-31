package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"gitlab.com/asciishell/tfs-go-auktion/internal/auth"
	"gitlab.com/asciishell/tfs-go-auktion/internal/user"
)

func PostSignup(w http.ResponseWriter, r *http.Request) {
	var userData user.User
	err := json.NewDecoder(r.Body).Decode(&userData)
	if err != nil {
		w.WriteHeader(400)
		return
	}
	err = userData.Registry()
	if err != nil {
		w.WriteHeader(400)
		return
	}
	w.WriteHeader(201)
}

func PostSignin(w http.ResponseWriter, r *http.Request) {
	var userData user.User
	err := json.NewDecoder(r.Body).Decode(&userData)
	if err != nil {
		w.WriteHeader(400)
		return
	}
	sess, err := auth.Signin(userData.Email, userData.Password)
	if err != nil {
		w.WriteHeader(404)
		return
	}
	bytes, err := sess.Marshal()
	if err != nil {
		w.WriteHeader(500)
		return
	}
	_, err = w.Write(bytes)
	if err != nil {
		log.Printf("Error writing bytes: %v", err)
	}
}

func PutUser(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		w.WriteHeader(400)
		return
	}
	if !auth.ValidateTokenUser(r, userID) {
		w.WriteHeader(401)
		return
	}

	userData, err := user.FindUserByID(userID)
	if err != nil {
		w.WriteHeader(404)
		return
	}
	if userData.ID != userID {
		w.WriteHeader(403)
		return
	}
	var newUser user.User
	err = json.NewDecoder(r.Body).Decode(&newUser)
	if err != nil {
		w.WriteHeader(400)
		return
	}
	userData.Update(newUser)
}
