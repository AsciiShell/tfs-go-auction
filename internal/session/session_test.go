package session

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenerateToken(t *testing.T) {
	r := require.New(t)
	str, err := GenerateToken()
	r.NoError(err)
	r.Equal(len(str), tokenLen)
}
