package tests

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"testing"
	"time"

	. "github.com/Eun/go-hit"
)

const (
	host = "localhost:8088"

	basePath = "http://" + host + "/api/v1"

	attempts = 20
)

var (
	token     string
	capsuleID string
	imageID   string

	headers IStep
)

func TestMain(m *testing.M) {
	if err := healthCheck(attempts); err != nil {
		log.Fatalf("Integration tests: host %s is not available: %v", host, err)
	}

	code := m.Run()
	os.Exit(code)
}

func healthCheck(attempts int) error {
	for i := 0; i < attempts; i++ {
		fmt.Println()
	}

	var err error

	for i := 0; i < attempts; i++ {
		err = Do(Get(basePath), Expect().Status().Equal(http.StatusNotFound))
		if err == nil {
			return nil
		}

		log.Printf("Integration tests: url %s is not available, attempts left: %d\n", basePath, i+1)

		time.Sleep(500 * time.Millisecond)
	}

	return err
}
