package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/pkg/errors"

	"gitlab.com/asciishell/tfs-go-auction/internal/lot"

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
	logger  log.Logger
}

type key int

const userKey key = 0

func NewAuctionHandler(storage storage.Storage, logger *log.Logger) *AuctionHandler {
	return &AuctionHandler{storage: &storage, logger: *logger}
}

func (h AuctionHandler) logError(r *http.Request, err error) {
	h.logger.Errorf("%s %s - %s:%+v", r.Method, r.RequestURI, r.Context().Value(userKey), err)
}

// nolint:unparam,unused
func (h AuctionHandler) logInfo(r *http.Request, msg string, args ...interface{}) {
	h.logger.Infof("%s %s - %+v:%s", r.Method, r.RequestURI, r.Context().Value(userKey), fmt.Sprintf(msg, args...))
}

func (h *AuctionHandler) Authenticator(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sess, err := auth.HandleToken(r, h.storage)
		if err != nil {
			switch err {
			case errs.ErrUnauthorized:
				http.Error(w, errs.NewErrorStr("Не авторизован").StringJSON(), http.StatusUnauthorized)
			case errs.ErrNotFound:
				http.Error(w, errs.NewError(err).StringJSON(), http.StatusUnauthorized)
			default:
				http.Error(w, errs.NewErrorStr("Неизвестная ошибка авторизации").StringJSON(), http.StatusUnauthorized)
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
		http.Error(w, errs.NewError(errors.Wrapf(err, "Неверные входные данные")).StringJSON(), http.StatusBadRequest)
		return
	}
	err = services.Registry(&userData, h.storage)
	switch err {
	case nil:
		http.Error(w, "", http.StatusCreated)
	case errs.ErrEmptyCredits:
		http.Error(w, errs.NewError(err).StringJSON(), http.StatusBadRequest)
	default:
		http.Error(w, errs.NewError(errors.Wrapf(err, "Невозможно зарегистрировать пользователя, конфликт. Например, email уже существует в системе")).StringJSON(), http.StatusConflict)
	}
}

func (h *AuctionHandler) PostSignin(w http.ResponseWriter, r *http.Request) {
	var userData user.User
	if err := json.NewDecoder(r.Body).Decode(&userData); err != nil {
		http.Error(w, errs.NewError(err).StringJSON(), http.StatusBadRequest)
		return
	}
	sess, err := auth.Signin(userData.Email, userData.Password, h.storage)
	if err != nil {
		http.Error(w, errs.NewError(errors.Wrapf(err, "Пользователь не авторизован")).StringJSON(), http.StatusUnauthorized)
		return
	}
	err = json.NewEncoder(w).Encode(sess)
	if err != nil {
		h.logError(r, errors.Wrap(err, "can't write session"))
		return
	}

}

func (h *AuctionHandler) PutUser(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, errs.NewError(err).StringJSON(), http.StatusBadRequest)
		return
	}
	if id != 0 && id != r.Context().Value(userKey).(int) {
		http.Error(w, errs.NewErrorStr("Запрещено").StringJSON(), http.StatusForbidden)
		return
	}
	userData := user.User{ID: r.Context().Value(userKey).(int)}
	err = (*h.storage).GetUser(&userData)
	if err != nil {
		http.Error(w, errs.NewError(err).StringJSON(), http.StatusNotFound)
		return
	}

	var newUser user.User
	err = json.NewDecoder(r.Body).Decode(&newUser)
	if err != nil {
		http.Error(w, errs.NewError(err).StringJSON(), http.StatusBadRequest)
		return
	}
	userData.Update(newUser)
	err = json.NewEncoder(w).Encode(userData)
	if err != nil {
		h.logError(r, errors.Wrap(err, "can't write user"))
		return
	}
}

func (h *AuctionHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, errs.NewError(err).StringJSON(), http.StatusBadRequest)
		return
	}
	if id == 0 {
		id = r.Context().Value(userKey).(int)
	}
	usr, err := services.FindUserByID(id, h.storage)
	if err != nil {
		http.Error(w, errs.NewError(err).StringJSON(), http.StatusNotFound)
		return
	}
	err = json.NewEncoder(w).Encode(usr)
	if err != nil {
		h.logError(r, errors.Wrap(err, "can't write user"))
		return
	}
}
func (h *AuctionHandler) GetLots(w http.ResponseWriter, r *http.Request) {
	t, err := lot.NewStatus(r.URL.Query().Get("status"))
	selector := lot.Lot{}
	if err == nil {
		selector.Status = t.String()
	}
	data, err := (*h.storage).GetLots(selector)
	if err != nil {
		http.Error(w, errs.NewError(err).StringJSON(), http.StatusInternalServerError)
		h.logError(r, errors.Wrap(err, "can't select lots"))
		return
	}
	err = json.NewEncoder(w).Encode(data)
	if err != nil {
		h.logError(r, errors.Wrap(err, "can't write lots"))
		return
	}
}
func (h *AuctionHandler) PostLots(w http.ResponseWriter, r *http.Request) {
	var lotData lot.Lot
	err := json.NewDecoder(r.Body).Decode(&lotData)
	if err != nil {
		http.Error(w, errs.NewError(err).StringJSON(), http.StatusBadRequest)
		return
	}
	lotData.CreatorID = r.Context().Value(userKey).(int)
	err = (*h.storage).AddLot(&lotData)
	if err != nil {
		http.Error(w, errs.NewError(err).StringJSON(), http.StatusBadRequest)
		return
	}
	err = json.NewEncoder(w).Encode(lotData)
	if err != nil {
		h.logError(r, errors.Wrap(err, "can't write lot"))
		return
	}

}
func (h *AuctionHandler) GetLot(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, errs.NewError(err).StringJSON(), http.StatusBadRequest)
		return
	}
	lotData := lot.Lot{ID: id}
	err = (*h.storage).GetLot(&lotData)
	if err != nil {
		http.Error(w, errs.NewError(err).StringJSON(), http.StatusNotFound)
		return
	}
	err = json.NewEncoder(w).Encode(lotData)
	if err != nil {
		h.logError(r, errors.Wrap(err, "can't write lot"))
		return
	}
}
func (h *AuctionHandler) PutLot(w http.ResponseWriter, r *http.Request) {
	// Лот в БД
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, errs.NewError(err).StringJSON(), http.StatusBadRequest)
		return
	}
	lotData := lot.Lot{ID: id, Status: lot.Created.String()}
	err = (*h.storage).GetLot(&lotData)
	if err != nil {
		http.Error(w, errs.NewError(err).StringJSON(), http.StatusNotFound)
		return
	}
	if lotData.CreatorID != r.Context().Value(userKey) {
		http.Error(w, errs.NewErrorStr("пользователь не соответствует создателю").StringJSON(), http.StatusNotFound)
		return
	}
	// Новый лот
	var newLot lot.Lot
	err = json.NewDecoder(r.Body).Decode(&newLot)
	if err != nil {
		http.Error(w, errs.NewError(err).StringJSON(), http.StatusBadRequest)
		return
	}
	newLot.ID = id
	err = (*h.storage).UpdateLot(&newLot)
	if err != nil {
		http.Error(w, errs.NewError(err).StringJSON(), http.StatusNotFound)
		return
	}
	err = json.NewEncoder(w).Encode(newLot)
	if err != nil {
		h.logError(r, errors.Wrap(err, "can't write lot"))
		return
	}
}
func (h *AuctionHandler) DeleteLot(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, errs.NewError(err).StringJSON(), http.StatusBadRequest)
		return
	}
	lotData := lot.Lot{ID: id, Status: lot.Created.String(), CreatorID: r.Context().Value(userKey).(int)}
	err = (*h.storage).DeleteLot(&lotData)
	if err != nil {
		http.Error(w, errs.NewError(err).StringJSON(), http.StatusNotFound)
		return
	}
	http.Error(w, "", http.StatusNoContent)
}

func (h *AuctionHandler) BuyLot(w http.ResponseWriter, r *http.Request) {
	type BuyLot struct {
		Price int
	}
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, errs.NewError(err).StringJSON(), http.StatusBadRequest)
		return
	}
	var price BuyLot
	err = json.NewDecoder(r.Body).Decode(&price)
	if err != nil {
		http.Error(w, errs.NewError(err).StringJSON(), http.StatusBadRequest)
		return
	}
	newLot, err := (*h.storage).BuyLot(id, r.Context().Value(userKey).(int), price.Price)
	if err != nil {
		http.Error(w, errs.NewError(err).StringJSON(), http.StatusConflict)
		return
	}
	err = json.NewEncoder(w).Encode(newLot)
	if err != nil {
		h.logError(r, errors.Wrap(err, "can't write lots"))
		return
	}
}
func (h *AuctionHandler) GetUserLots(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, errs.NewError(err).StringJSON(), http.StatusBadRequest)
		return
	}
	if id == 0 {
		id = r.Context().Value(userKey).(int)
	}
	var lots []lot.Lot
	switch strings.ToLower(r.URL.Query().Get("type")) {
	case "own":
		lots, err = (*h.storage).GetOwnLots(&lot.Lot{CreatorID: id}, &lot.Lot{})
	case "buyed":
		lots, err = (*h.storage).GetOwnLots(&lot.Lot{BuyerID: &id}, &lot.Lot{})
	default:
		lots, err = (*h.storage).GetOwnLots(&lot.Lot{CreatorID: id}, &lot.Lot{BuyerID: &id})

	}
	if err != nil || len(lots) == 0 {
		http.Error(w, "no data", http.StatusNotFound)
		return
	}
	err = json.NewEncoder(w).Encode(lots)
	if err != nil {
		h.logError(r, errors.Wrap(err, "can't write lot"))
		return
	}
}
