package server_test

import (
	"testing"

	"github.com/fjnkt98/todo-go/server"
)

func TestNewServer(t *testing.T) {
	_, err := server.NewServer(8000)
	if err != nil {
		t.Errorf("expected nil, but got %v", err)
	}
}

func TestNewServeMux(t *testing.T) {
	_, err := server.NewServeMux()
	if err != nil {
		t.Errorf("expected nil, but got %v", err)
	}
}
