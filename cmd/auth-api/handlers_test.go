package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"gitlab.com/asciishell/tfs-go-auction/internal/mock_storage"
	"gitlab.com/asciishell/tfs-go-auction/internal/user"
	"gitlab.com/asciishell/tfs-go-auction/pkg/environment"
)

const TimeOut = 2 * time.Second

func RaceTimeout() time.Duration {
	m := environment.GetInt("TIMEOUT_MULTIPLY", 1)
	return time.Duration(m) * TimeOut
}
func TestAuctionHandler_PostSignup(t *testing.T) {
	type testCase struct {
		Name        string
		ContentType string
		Request     string
		Ok          bool
		Code        int
	}
	testCases := []testCase{
		{Name: "Normal",
			Request:     `{"first_name": "Павел","last_name": "Дуров","birthday": "1984-10-10","email": "durov@telegram.org","password": "qwerty"}`,
			ContentType: "application/json",
			Ok:          true,
			Code:        http.StatusCreated},
		{Name: "Bad date",
			Request:     `{"first_name": "Павел","last_name": "Дуров","birthday": "2019-04-21T13:28:39.414Z","email": "durov@telegram.org","password": "qwerty"}`,
			ContentType: "application/json",
			Ok:          false,
			Code:        http.StatusBadRequest},
		{Name: "Empty email",
			Request:     `{"first_name": "Павел","last_name": "Дуров","birthday": "1984-10-10","email": "","password": "qwerty"} `,
			ContentType: "application/json",
			Ok:          false,
			Code:        http.StatusBadRequest},
		{Name: "Empty password",
			Request:     `{"first_name": "Павел","last_name": "Дуров","birthday": "1984-10-10","email": "durov@telegram.org","password": ""} `,
			ContentType: "application/json",
			Ok:          false,
			Code:        http.StatusBadRequest},
		{Name: "Bad request",
			Request:     `I am not a good JSON struct`,
			ContentType: "plain/text",
			Ok:          false,
			Code:        http.StatusBadRequest},
	}
	r := require.New(t)
	timeout := RaceTimeout()
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			m := mock_storage.NewMockStorage(ctrl)
			if tc.Ok {
				m.EXPECT().AddUser(gomock.Any()).Return(nil).Times(1)
			}
			handler := NewAuctionHandler(m)
			mux := http.NewServeMux()
			mux.HandleFunc("/", handler.PostSignup)
			ts := httptest.NewServer(mux)
			client := http.Client{Timeout: timeout}

			reader := bytes.NewReader([]byte(tc.Request))

			resp, err := client.Post(ts.URL, tc.ContentType, reader)
			r.NoError(err)
			r.Equal(resp.StatusCode, tc.Code)
		})
	}

}

func TestAuctionHandler_PostSignin(t *testing.T) {
	type testCase struct {
		Name        string
		ContentType string
		Request     string
		Exists      bool
		Ok          bool
		Code        int
	}
	testCases := []testCase{
		{Name: "Normal",
			Request:     `{"email": "durov@telegram.org","password": "correct"}`,
			ContentType: "application/json",
			Exists:      true,
			Ok:          true,
			Code:        http.StatusOK},
		{Name: "Wrong email/password",
			Request:     `{"email": "durov@telegram.org","password": "incorrect"}`,
			ContentType: "application/json",
			Exists:      false,
			Ok:          true,
			Code:        http.StatusUnauthorized},
		{Name: "Bad request",
			Request:     `I am not a good JSON struct`,
			ContentType: "plaint/text",
			Ok:          false,
			Exists:      false,
			Code:        http.StatusBadRequest},
	}
	r := require.New(t)
	timeout := RaceTimeout()
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			m := mock_storage.NewMockStorage(ctrl)
			if tc.Ok {
				if tc.Exists {
					m.EXPECT().GetUser(gomock.Any()).DoAndReturn(func(u *user.User) error {
						hash, _ := user.HashPassword("correct")
						*u = user.User{Email: "durov@telegram.org", Password: hash}
						return nil
					}).Times(1)
					m.EXPECT().AddSession(gomock.Any()).Return(nil).Times(1)

				} else {
					m.EXPECT().GetUser(gomock.Any()).DoAndReturn(func(u *user.User) error {
						hash, _ := user.HashPassword("correct")
						*u = user.User{Email: "durov@telegram.org", Password: hash}
						return nil
					}).Times(1)
				}
			}
			handler := NewAuctionHandler(m)
			mux := http.NewServeMux()
			mux.HandleFunc("/", handler.PostSignin)
			ts := httptest.NewServer(mux)
			client := http.Client{Timeout: timeout}

			reader := bytes.NewReader([]byte(tc.Request))

			resp, err := client.Post(ts.URL, tc.ContentType, reader)
			r.NoError(err)

			r.Equal(resp.StatusCode, tc.Code)
		})
	}

}
