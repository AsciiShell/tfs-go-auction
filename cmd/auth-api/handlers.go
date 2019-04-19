package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"gitlab.com/asciishell/tfs-go-auction/internal/errs"

	"github.com/go-chi/chi"
	"gitlab.com/asciishell/tfs-go-auction/internal/auth"
	"gitlab.com/asciishell/tfs-go-auction/internal/services"
	"gitlab.com/asciishell/tfs-go-auction/internal/storage"
	"gitlab.com/asciishell/tfs-go-auction/internal/user"
	"gitlab.com/asciishell/tfs-go-auction/pkg/log"
)

type AuctionHandler struct {
	storage *storage.Storage
}

type key int

const userKey key = 0

func NewAuctionHandler(storage storage.Storage) *AuctionHandler {
	return &AuctionHandler{storage: &storage}
}

func (h *AuctionHandler) Authenticator(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sess, err := auth.HandleToken(r, h.storage)
		if err != nil {
			switch err {
			case errs.ErrUnauthorized:
				http.Error(w, "", http.StatusUnauthorized)
			case errs.ErrNotFound:
				http.Error(w, err.Error(), http.StatusUnauthorized)
			default:
				http.Error(w, "Неизвестная ошибка авторизации", http.StatusUnauthorized)
			}
			return
		}
		ctx := context.WithValue(r.Context(), userKey, sess.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (h *AuctionHandler) PostSignup(w http.ResponseWriter, r *http.Request) {
	var userData user.User
	err := json.NewDecoder(r.Body).Decode(&userData)
	if err != nil {
		http.Error(w, fmt.Sprintf("Неверные входные данные %s", err), http.StatusBadRequest)
		return
	}
	err = services.Registry(&userData, h.storage)
	switch err {
	case nil:
		http.Error(w, "Пользователь зарегистрирован", http.StatusCreated)
	case errs.ErrEmptyCredits:
		http.Error(w, err.Error(), http.StatusBadRequest)
	default:
		http.Error(w, "Невозможно зарегистрировать пользователя, конфликт. Например, email уже существует в системе", http.StatusConflict)
	}
}

func (h *AuctionHandler) PostSignin(w http.ResponseWriter, r *http.Request) {
	var userData user.User
	if err := json.NewDecoder(r.Body).Decode(&userData); err != nil {
		http.Error(w, "Неверные входные данные", http.StatusBadRequest)
		return
	}
	sess, err := auth.Signin(userData.Email, userData.Password, h.storage)
	if err != nil {
		http.Error(w, "Пользователь не авторизован", http.StatusUnauthorized)
		return
	}
	bytes, err := sess.MarshalJSON()
	if err != nil {
		http.Error(w, "БРРР", http.StatusInternalServerError)
		return
	}
	if _, err = w.Write(bytes); err != nil {
		log.New().Warnf("Error writing bytes: %v", err)
	}
}

func (h *AuctionHandler) PutUser(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Неверные входные данные", http.StatusBadRequest)
		return
	}
	if id != 0 && id != r.Context().Value(userKey).(int) {
		http.Error(w, "Запрещено", http.StatusForbidden)
		return
	}
	userData := user.User{ID: r.Context().Value(userKey).(int)}
	err = (*h.storage).GetUser(&userData)
	if err != nil {
		http.Error(w, "Не найден", http.StatusNotFound)
		return
	}

	var newUser user.User
	err = json.NewDecoder(r.Body).Decode(&newUser)
	if err != nil {
		http.Error(w, "Неверные входные данные", http.StatusBadRequest)
		return
	}
	userData.Update(newUser)
	bytes, err := json.Marshal(userData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if _, err := w.Write(bytes); err != nil {
		log.New().Errorf("can't write bytes")
		return
	}
}

func (h *AuctionHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if id == 0 {
		id = r.Context().Value(userKey).(int)
	}
	usr, err := services.FindUserByID(id, h.storage)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	bytes, err := json.Marshal(usr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if _, err := w.Write(bytes); err != nil {
		log.New().Errorf("can't write bytes")
		return
	}
}
func (h *AuctionHandler) NotImplemented(w http.ResponseWriter, r *http.Request) {
	log.New().Infof("Request not implemented %s %s: %s", r.Method, r.RequestURI, r.Context().Value("User"))
}
