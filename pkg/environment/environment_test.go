package environment

import (
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestGetStr(t *testing.T) {
	r := require.New(t)
	type testCase struct {
		Name     string
		Key      string
		ValueW   string
		ValueD   string
		Expected string
	}
	const KeyWrite = "TEST_STR_ENV"
	const KeyEmpty = "TEST_STR_ENV_NO"
	const ValueTrue = "True"
	const ValueFalse = "False"
	testCases := []testCase{
		{Name: "Normal", Key: KeyWrite, ValueW: ValueTrue, ValueD: ValueFalse, Expected: ValueTrue},
		{Name: "Read default", Key: KeyEmpty, ValueW: ValueTrue, ValueD: ValueFalse, Expected: ValueFalse},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			err := os.Setenv(tc.Key, tc.ValueW)
			r.NoError(err)

			actual := GetStr(KeyWrite, tc.ValueD)
			if actual != tc.Expected {
				t.Errorf("Written: %s=%s. Get(%s, %v) = %v, want %v", tc.Key, tc.ValueW, KeyWrite, tc.ValueD, actual, tc.Expected)
			}
			os.Clearenv()
		})
	}
}

func TestGetInt(t *testing.T) {
	r := require.New(t)
	type testCase struct {
		Name     string
		Key      string
		ValueW   string
		ValueD   int
		Expected int
	}
	const KeyWrite = "TEST_STR_ENV"
	const KeyEmpty = "TEST_STR_ENV_NO"
	const ValueTrue = 100
	const ValueFalse = 200
	testCases := []testCase{
		{Name: "Normal", Key: KeyWrite, ValueW: strconv.Itoa(ValueTrue), ValueD: ValueFalse, Expected: ValueTrue},
		{Name: "Read default", Key: KeyEmpty, ValueW: strconv.Itoa(ValueTrue), ValueD: ValueFalse, Expected: ValueFalse},
		{Name: "Bad write, use default", Key: KeyWrite, ValueW: "Not int", ValueD: ValueFalse, Expected: ValueFalse},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			err := os.Setenv(tc.Key, tc.ValueW)
			r.NoError(err)

			actual := GetInt(KeyWrite, tc.ValueD)
			if actual != tc.Expected {
				t.Errorf("Written: %s=%s. Get(%s, %v) = %v, want %v", tc.Key, tc.ValueW, KeyWrite, tc.ValueD, actual, tc.Expected)
			}
			os.Clearenv()
		})
	}
}

func TestGetBool(t *testing.T) {
	r := require.New(t)
	type testCase struct {
		Name     string
		Key      string
		ValueW   string
		ValueD   bool
		Expected bool
	}
	const KeyWrite = "TEST_STR_ENV"
	const KeyEmpty = "TEST_STR_ENV_NO"
	testCases := []testCase{
		{Name: "Normal", Key: KeyWrite, ValueW: "1", ValueD: false, Expected: true},
		{Name: "Read default", Key: KeyEmpty, ValueW: "1", ValueD: false, Expected: false},
		{Name: "Bad write, use default", Key: KeyWrite, ValueW: " bool", ValueD: true, Expected: true},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			err := os.Setenv(tc.Key, tc.ValueW)
			r.NoError(err)

			actual := GetBool(KeyWrite, tc.ValueD)
			if actual != tc.Expected {
				t.Errorf("Written: %s=%s. Get(%s, %v) = %v, want %v", tc.Key, tc.ValueW, KeyWrite, tc.ValueD, actual, tc.Expected)
			}
			os.Clearenv()
		})
	}
}

func TestGetDuration(t *testing.T) {
	r := require.New(t)
	type testCase struct {
		Name     string
		Key      string
		ValueW   string
		ValueD   time.Duration
		Expected time.Duration
	}
	const KeyWrite = "TEST_STR_ENV"
	const KeyEmpty = "TEST_STR_ENV_NO"
	const ValueTrue = 100
	const ValueFalse = 200
	testCases := []testCase{
		{Name: "Normal", Key: KeyWrite, ValueW: time.Duration(ValueTrue).String(), ValueD: ValueFalse, Expected: ValueTrue},
		{Name: "Read default", Key: KeyEmpty, ValueW: time.Duration(ValueTrue).String(), ValueD: ValueFalse, Expected: ValueFalse},
		{Name: "Bad write, use default", Key: KeyWrite, ValueW: "Not duration", ValueD: ValueFalse, Expected: ValueFalse},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			err := os.Setenv(tc.Key, tc.ValueW)
			r.NoError(err)

			actual := GetDuration(KeyWrite, tc.ValueD)
			if actual != tc.Expected {
				t.Errorf("Written: %s=%s. Get(%s, %v) = %v, want %v", tc.Key, tc.ValueW, KeyWrite, tc.ValueD, actual, tc.Expected)
			}
			os.Clearenv()
		})
	}
}
