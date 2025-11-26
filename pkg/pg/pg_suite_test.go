package pg

import (
	"bytes"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/rs/zerolog"
)

var testLogBuffer *bytes.Buffer

func TestInternal(t *testing.T) {
	testLogBuffer = new(bytes.Buffer)
	logger = zerolog.New(testLogBuffer).
		With().Timestamp().Logger().Level(zerolog.PanicLevel)
	RegisterFailHandler(Fail)
	RunSpecs(t, "Pg Suite")
}
