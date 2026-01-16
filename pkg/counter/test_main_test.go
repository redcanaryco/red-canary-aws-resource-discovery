package counter

import (
	"io"
	"log"
	"os"
	"testing"
)

// TestMain silences package-level log output during tests so mocked error
// responses don't clutter test output. Tests can still use package loggers
// (e.g. interfaces.Logger) which are mocked explicitly in each test.
func TestMain(m *testing.M) {
	log.SetOutput(io.Discard)
	os.Exit(m.Run())
}
