package todo

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fjnkt98/todo-go/api"
	"github.com/fjnkt98/todo-go/repository"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var pool *pgxpool.Pool

func TestMain(m *testing.M) {
	ctx := context.Background()

	schemaFile, err := filepath.Abs(filepath.Join("..", "..", "schema.sql"))
	if err != nil {
		log.Fatalf("failed to specify schema.sql file: %s", err)
	}
	r, err := os.Open(schemaFile)
	if err != nil {
		log.Fatalf("failed to open schema.sql file: %s", err)
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image: "postgres:15-bullseye",
			Env: map[string]string{
				"POSTGRES_PASSWORD":         "todo",
				"POSTGRES_USER":             "todo",
				"POSTGRES_DB":               "todo",
				"POSTGRES_HOST_AUTH_METHOD": "password",
				"TZ":                        "Asia/Tokyo",
			},
			Files: []testcontainers.ContainerFile{
				{
					Reader:            r,
					ContainerFilePath: "/docker-entrypoint-initdb.d/schema.sql",
					FileMode:          0o666,
				},
			},
			ExposedPorts: []string{"5432/tcp"},
			WaitingFor:   wait.ForListeningPort("5432/tcp"),
		},
		Started: true,
	})
	if err != nil {
		log.Fatalf("failed to create database container: %s", err)
	}

	defer func() {
		if err := container.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate database container: %s", err)
		}
	}()

	host, err := container.Host(ctx)
	if err != nil {
		log.Fatalf("failed to get host: %s", err)
	}
	port, err := container.MappedPort(ctx, "5432/tcp")
	if err != nil {
		log.Fatalf("failed to get port: %s", err)
	}

	dsn := fmt.Sprintf(
		"postgres://todo:todo@%s:%d/todo?sslmode=disable",
		host,
		port.Int(),
	)
	pool, err = repository.NewPool(ctx, dsn)
	if err != nil {
		log.Fatalf("failed to create connection pool: %s", err)
	}
	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("failed to verify a connection: %s", err)
	}

	os.Exit(m.Run())
}

func TestScenario(t *testing.T) {
	e := echo.New()
	e.Validator = new(api.Validator)

	h := NewHandler(pool)

	// Add first post
	{
		body := `{"title":"first todo"}`
		req := httptest.NewRequest(http.MethodPost, "/api/todo", strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// Add first item
		if err := h.POST(c); err != nil {
			t.Errorf("failed to request: %s", err)
		}
		if rec.Code != http.StatusOK {
			t.Fatalf("expected %d but got %d", http.StatusOK, rec.Code)
		}

		var res PostResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
			t.Errorf("failed to unmarshal the response: %s", err)
		}

		if res.Item.ID != 1 {
			t.Errorf(`expected 1, but got %d`, res.Item.ID)
		}
		if res.Item.Title != "first todo" {
			t.Errorf(`expected "first todo", but got %s`, res.Item.Title)
		}
	}

	// Add second item
	{
		body := `{"title":"second todo"}`
		req := httptest.NewRequest(http.MethodPost, "/api/todo", strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// Add first post
		if err := h.POST(c); err != nil {
			t.Errorf("failed to request: %s", err)
		}
		if rec.Code != http.StatusOK {
			t.Fatalf("expected %d but got %d", http.StatusOK, rec.Code)
		}

		var res PostResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
			t.Errorf("failed to unmarshal the response: %s", err)
		}

		if res.Item.ID != 2 {
			t.Errorf(`expected 2, but got %d`, res.Item.ID)
		}
		if res.Item.Title != "second todo" {
			t.Errorf(`expected "second todo", but got %s`, res.Item.Title)
		}
	}

	// Get all items
	{
		req := httptest.NewRequest(http.MethodGet, "/api/todo", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		if err := h.GET(c); err != nil {
			t.Errorf("failed to request: %s", err)
		}
		if rec.Code != http.StatusOK {
			t.Fatalf("expected %d but got %d", http.StatusOK, rec.Code)
		}

		var res GetResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
			t.Errorf("failed to unmarshal the response: %s", err)
		}

		if len(res.Items) != 2 {
			t.Fatalf("expected length 2 but got %d", len(res.Items))
		}

		if res.Items[0].ID != 1 {
			t.Errorf("expected id 1, but got %d", res.Items[0].ID)
		}
		if res.Items[0].Title != "first todo" {
			t.Errorf(`expected id "first todo", but got %s`, res.Items[0].Title)
		}

		if res.Items[1].ID != 2 {
			t.Errorf("expected id 2, but got %d", res.Items[1].ID)
		}
		if res.Items[1].Title != "second todo" {
			t.Errorf(`expected id "second todo", but got %s`, res.Items[1].Title)
		}
	}

	// Update first item
	{
		body := `{"title":"updated first todo"}`
		req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		c.SetPath("/api/todo/:id")
		c.SetParamNames("id")
		c.SetParamValues("1")

		if err := h.PUT(c); err != nil {
			t.Errorf("failed to request: %s", err)
		}
		if rec.Code != http.StatusOK {
			t.Fatalf("expected %d, but got %d", http.StatusOK, rec.Code)
		}

		var res PutResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
			t.Errorf("failed to unmarshal the response: %s", err)
		}

		if res.Item.ID != 1 {
			t.Errorf("expected id 1, but got %d", res.Item.ID)
		}
		if res.Item.Title != "updated first todo" {
			t.Errorf(`expected id "updated first todo", but got %s`, res.Item.Title)
		}
	}

	// Get all items again
	{
		req := httptest.NewRequest(http.MethodGet, "/api/todo", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		if err := h.GET(c); err != nil {
			t.Errorf("failed to request: %s", err)
		}
		if rec.Code != http.StatusOK {
			t.Fatalf("expected %d but got %d", http.StatusOK, rec.Code)
		}

		var res GetResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
			t.Errorf("failed to unmarshal the response: %s", err)
		}

		if len(res.Items) != 2 {
			t.Fatalf("expected length 2 but got %d", len(res.Items))
		}

		if res.Items[0].ID != 1 {
			t.Errorf("expected id 1, but got %d", res.Items[0].ID)
		}
		if res.Items[0].Title != "updated first todo" {
			t.Errorf(`expected id "updated first todo", but got %s`, res.Items[0].Title)
		}

		if res.Items[1].ID != 2 {
			t.Errorf("expected id 2, but got %d", res.Items[1].ID)
		}
		if res.Items[1].Title != "second todo" {
			t.Errorf(`expected id "second todo", but got %s`, res.Items[1].Title)
		}
	}
}
