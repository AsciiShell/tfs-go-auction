package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/gorilla/websocket"

	"github.com/go-chi/chi"
	"github.com/pkg/errors"

	"gitlab.com/asciishell/tfs-go-auction/internal/auth"
	"gitlab.com/asciishell/tfs-go-auction/internal/errs"
	"gitlab.com/asciishell/tfs-go-auction/internal/lot"
	"gitlab.com/asciishell/tfs-go-auction/internal/services"
	"gitlab.com/asciishell/tfs-go-auction/internal/storage"
	"gitlab.com/asciishell/tfs-go-auction/internal/template"
	"gitlab.com/asciishell/tfs-go-auction/internal/user"
	"gitlab.com/asciishell/tfs-go-auction/pkg/log"
)

type AuctionHandler struct {
	storage     *storage.Storage
	logger      log.Logger
	temps       template.Templates
	wsClients   WSClients
	priceTickCh chan lot.Lot
	upgrader    websocket.Upgrader
}
type WSClients struct {
	wsConn []*websocket.Conn
	*sync.Mutex
}

func (clients *WSClients) AddClient(h *AuctionHandler, conn *websocket.Conn) {
	clients.Mutex.Lock()
	clients.wsConn = append(clients.wsConn, conn)
	clients.Mutex.Unlock()
	h.logger.Infof("added client, total clients: %d\n", len(h.wsClients.wsConn))
}

func (clients *WSClients) removeClientByID(h *AuctionHandler, id int) {
	clients.Mutex.Lock()
	clients.wsConn = append(clients.wsConn[:id], clients.wsConn[id+1:]...)
	clients.Mutex.Unlock()
	h.logger.Infof("removed client #%d, total clients %d\n", id, len(clients.wsConn))
}

func (clients *WSClients) BroadcastMessage(h *AuctionHandler, message []byte) {
	for i, c := range clients.wsConn {
		err := c.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			h.logger.Infof("can't broadcast message: %+v\n", err)
			clients.removeClientByID(h, i)
		}
	}
}

type key int

const userKey key = 0

func NewAuctionHandler(storage storage.Storage, logger *log.Logger, temps template.Templates) *AuctionHandler {
	h := AuctionHandler{storage: &storage, logger: *logger, temps: temps}
	h.priceTickCh = make(chan lot.Lot)
	h.upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	return &h
}

func (h AuctionHandler) logError(r *http.Request, err error) {
	h.logger.Errorf("%s %s - %s:%+v", r.Method, r.RequestURI, r.Context().Value(userKey), err)
}

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
	http.SetCookie(w, &http.Cookie{Name: "BearerToken", Value: sess.SessionID, Path: "/", Expires: sess.ValidUntil})
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
	lotType := r.URL.Query().Get("status")
	data, err := services.GetLots(lotType, *h.storage)
	if err != nil {
		http.Error(w, errs.NewError(err).StringJSON(), http.StatusInternalServerError)
		h.logError(r, err)
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
	h.priceTickCh <- newLot
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
	lotType := strings.ToLower(r.URL.Query().Get("type"))
	lots, err := services.GetUserLots(id, lotType, *h.storage)
	if err != nil || len(lots) == 0 {
		http.Error(w, errs.NewErrorStr("no data").StringJSON(), http.StatusNotFound)
		return
	}
	err = json.NewEncoder(w).Encode(lots)
	if err != nil {
		h.logError(r, errors.Wrap(err, "can't write lot"))
		return
	}
}
func (h *AuctionHandler) NotImplemented(w http.ResponseWriter, r *http.Request) {
	h.logInfo(r, "Request not implemented")
	_, _ = w.Write([]byte("not implemented"))
}
func (h *AuctionHandler) HTMLGetLots(w http.ResponseWriter, r *http.Request) {
	lotType := r.URL.Query().Get("status")
	data, err := services.GetLots(lotType, *h.storage)
	if err != nil {
		http.Error(w, errs.NewError(err).StringJSON(), http.StatusInternalServerError)
		h.logError(r, err)
		return
	}
	h.temps.Render(w, "all_lots", struct {
		LotType string
		Data    []lot.Lot
	}{LotType: lotType, Data: data})
}
func (h *AuctionHandler) HTMLGetUserLots(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, errs.NewError(err).StringJSON(), http.StatusBadRequest)
		return
	}
	if id == 0 {
		id = r.Context().Value(userKey).(int)
	}
	lotType := strings.ToLower(r.URL.Query().Get("type"))
	lots, err := services.GetUserLots(id, lotType, *h.storage)
	if err != nil {
		http.Error(w, errs.NewErrorStr("no data").StringJSON(), http.StatusNotFound)
		return
	}
	h.temps.Render(w, "user_lots", struct {
		LotType string
		Data    []lot.Lot
	}{LotType: lotType, Data: lots})
}
func (h *AuctionHandler) HTMLGetLot(w http.ResponseWriter, r *http.Request) {
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
	h.temps.Render(w, "lot_details", lotData)
}
func (h *AuctionHandler) WSLotUpdate(w http.ResponseWriter, r *http.Request) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logError(r, err)
		return
	}
	defer func() {
		_ = conn.Close()
	}()
	h.wsClients.AddClient(h, conn)

	for {
		priceTick := <-h.priceTickCh
		res, err := json.Marshal(priceTick)
		if err != nil {
			h.logger.Infof("can't marshal message: %+v\n", err)
			continue
		}
		h.wsClients.BroadcastMessage(h, res)
	}

}
