package e2e_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/gonzaloccnc/marketplace-go/internal/product"
	"github.com/gonzaloccnc/marketplace-go/pkg/database"
)

// shared across the e2e tests, initialised once in TestMain.
var (
	testPool   *pgxpool.Pool
	testRouter *gin.Engine
)

// TestMain spins up a real Postgres in a throwaway container, runs the project
// migrations against it and wires the real product module onto a gin router.
func TestMain(m *testing.M) {
	ctx := context.Background()

	container, err := tcpostgres.Run(ctx,
		"postgres:17-alpine",
		tcpostgres.WithDatabase("marketplace"),
		tcpostgres.WithUsername("gonza"),
		tcpostgres.WithPassword("strong_password"),
		testcontainers.WithWaitStrategy(
			wait.ForListeningPort("5432/tcp").WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to start postgres container: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		if err := testcontainers.TerminateContainer(container); err != nil {
			fmt.Fprintf(os.Stderr, "failed to terminate container: %v\n", err)
		}
	}()

	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get connection string: %v\n", err)
		os.Exit(1)
	}

	testPool, err = pgxpool.New(ctx, dsn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create pool: %v\n", err)
		os.Exit(1)
	}
	defer testPool.Close()

	if err := database.RunMigrations(ctx, testPool); err != nil {
		fmt.Fprintf(os.Stderr, "failed to run migrations: %v\n", err)
		os.Exit(1)
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	slog.SetDefault(logger)

	gin.SetMode(gin.TestMode)
	testRouter = gin.New()
	product.Register(testRouter, testPool)

	os.Exit(m.Run())
}

// decodeData unmarshals the "data" field of the standard httpx.ApiResponse
// envelope ({"status":..., "data":...}) into v.
func decodeData(t *testing.T, body []byte, v any) {
	t.Helper()
	var env struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(body, &env); err != nil {
		t.Fatalf("failed to unmarshal envelope: %v", err)
	}
	if err := json.Unmarshal(env.Data, v); err != nil {
		t.Fatalf("failed to unmarshal data: %v", err)
	}
}

// resetProducts empties the products table so each test starts from a known state.
func resetProducts(t *testing.T) {
	t.Helper()
	if _, err := testPool.Exec(context.Background(), "TRUNCATE products"); err != nil {
		t.Fatalf("failed to truncate products: %v", err)
	}
}

// seedProduct inserts a product directly and returns its generated id.
func seedProduct(t *testing.T, name string, price float64, stock int, attrs map[string]string) string {
	t.Helper()
	var id string
	err := testPool.QueryRow(
		context.Background(),
		"INSERT INTO products (name, price, stock, attributes) VALUES ($1, $2, $3, $4) RETURNING id",
		name, price, stock, attrs,
	).Scan(&id)
	if err != nil {
		t.Fatalf("failed to seed product: %v", err)
	}
	return id
}

func TestGetProducts(t *testing.T) {
	resetProducts(t)
	id1 := seedProduct(t, "Keyboard", 49.99, 10, map[string]string{"color": "black"})
	id2 := seedProduct(t, "Mouse", 19.99, 5, map[string]string{"color": "white"})

	req := httptest.NewRequest(http.MethodGet, "/products", nil)
	rec := httptest.NewRecorder()
	testRouter.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d (body: %s)", rec.Code, http.StatusOK, rec.Body.String())
	}

	var got []product.ProductDTO
	decodeData(t, rec.Body.Bytes(), &got)

	if len(got) != 2 {
		t.Fatalf("got %d products, want 2", len(got))
	}

	byID := make(map[string]product.ProductDTO, len(got))
	for _, p := range got {
		byID[p.ID] = p
	}

	kb, ok := byID[id1]
	if !ok {
		t.Fatalf("keyboard (id %s) missing from response", id1)
	}
	if kb.Name != "Keyboard" || kb.Price != 49.99 || kb.Stock != 10 || kb.Attributes["color"] != "black" {
		t.Errorf("keyboard = %+v, unexpected fields", kb)
	}

	mouse, ok := byID[id2]
	if !ok {
		t.Fatalf("mouse (id %s) missing from response", id2)
	}
	if mouse.Name != "Mouse" || mouse.Price != 19.99 || mouse.Stock != 5 || mouse.Attributes["color"] != "white" {
		t.Errorf("mouse = %+v, unexpected fields", mouse)
	}
}

func TestGetProductsEmpty(t *testing.T) {
	resetProducts(t)

	req := httptest.NewRequest(http.MethodGet, "/products", nil)
	rec := httptest.NewRecorder()
	testRouter.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d (body: %s)", rec.Code, http.StatusOK, rec.Body.String())
	}

	var got []product.ProductDTO
	decodeData(t, rec.Body.Bytes(), &got)
	if len(got) != 0 {
		t.Errorf("got %d products, want 0", len(got))
	}
}

func TestGetProductByID(t *testing.T) {
	resetProducts(t)

	id1 := seedProduct(t, "Keyboard", 49.99, 10, map[string]string{"color": "black"})

	req := httptest.NewRequest(http.MethodGet, "/products/"+id1, nil)
	rec := httptest.NewRecorder()
	testRouter.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d (body: %s)", rec.Code, http.StatusOK, rec.Body.String())
	}

	var got product.ProductDTO
	decodeData(t, rec.Body.Bytes(), &got)
	if got.ID != id1 {
		t.Errorf("got product ID %s, want %s", got.ID, id1)
	}

}

func TestGetProductByIDNotFound(t *testing.T) {
	resetProducts(t)

	id1 := uuid.New().String()
	req := httptest.NewRequest(http.MethodGet, "/products/"+id1, nil)
	rec := httptest.NewRecorder()
	testRouter.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status code = %d, want %d (body: %s)", rec.Code, http.StatusNotFound, rec.Body.String())
	}
}

// updateProduct sends a PUT /products/:id with the given JSON body.
func updateProduct(t *testing.T, id, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPut, "/products/"+id, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	testRouter.ServeHTTP(rec, req)
	return rec
}

func TestUpdateProduct(t *testing.T) {
	resetProducts(t)
	id := seedProduct(t, "Keyboard", 49.99, 10, map[string]string{"color": "black"})

	rec := updateProduct(t, id, `{"name":"Keyboard Pro","price":79.5,"stock":3,"attributes":{"color":"red"}}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d (body: %s)", rec.Code, http.StatusOK, rec.Body.String())
	}

	var got product.ProductDTO
	decodeData(t, rec.Body.Bytes(), &got)
	if got.ID != id || got.Name != "Keyboard Pro" || got.Price != 79.5 || got.Stock != 3 || got.Attributes["color"] != "red" {
		t.Errorf("updated product = %+v, unexpected fields", got)
	}
	if got.State != product.ProductStateActive {
		t.Errorf("state = %q, want %q (state omitted must be kept)", got.State, product.ProductStateActive)
	}
}

func TestUpdateProductChangesState(t *testing.T) {
	resetProducts(t)
	id := seedProduct(t, "Keyboard", 49.99, 10, map[string]string{"color": "black"})

	rec := updateProduct(t, id, `{"name":"Keyboard","price":49.99,"stock":10,"attributes":{"color":"black"},"state":"I"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d (body: %s)", rec.Code, http.StatusOK, rec.Body.String())
	}

	var got product.ProductDTO
	decodeData(t, rec.Body.Bytes(), &got)
	if got.State != product.ProductStateInactive {
		t.Errorf("state = %q, want %q", got.State, product.ProductStateInactive)
	}
}

func TestUpdateProductErrors(t *testing.T) {
	resetProducts(t)
	id := seedProduct(t, "Keyboard", 49.99, 10, map[string]string{"color": "black"})

	tests := []struct {
		name string
		id   string
		body string
		want int
	}{
		{"not found", uuid.New().String(), `{"name":"x","price":1,"stock":1,"attributes":{}}`, http.StatusNotFound},
		{"invalid id", "not-a-uuid", `{"name":"x","price":1,"stock":1,"attributes":{}}`, http.StatusBadRequest},
		{"zero stock", id, `{"name":"x","price":1,"stock":0,"attributes":{}}`, http.StatusBadRequest},
		{"negative stock", id, `{"name":"x","price":1,"stock":-5,"attributes":{}}`, http.StatusBadRequest},
		{"invalid state", id, `{"name":"x","price":1,"stock":1,"attributes":{},"state":"X"}`, http.StatusBadRequest},
		{"zero price", id, `{"name":"x","price":0,"stock":1,"attributes":{}}`, http.StatusBadRequest},
		{"missing name", id, `{"price":1,"stock":1,"attributes":{}}`, http.StatusBadRequest},
		{"malformed json", id, `{"name":`, http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := updateProduct(t, tt.id, tt.body)
			if rec.Code != tt.want {
				t.Errorf("status code = %d, want %d (body: %s)", rec.Code, tt.want, rec.Body.String())
			}
		})
	}
}

// inactiveProduct sends a PATCH /products/:id/inactive.
func inactiveProduct(t *testing.T, id string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPatch, "/products/"+id+"/inactive", nil)
	rec := httptest.NewRecorder()
	testRouter.ServeHTTP(rec, req)
	return rec
}

func TestInactiveProduct(t *testing.T) {
	resetProducts(t)
	id := seedProduct(t, "Keyboard", 49.99, 10, map[string]string{"color": "black"})

	rec := inactiveProduct(t, id)
	if rec.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d (body: %s)", rec.Code, http.StatusOK, rec.Body.String())
	}

	var got product.ProductDTO
	decodeData(t, rec.Body.Bytes(), &got)
	if got.ID != id || got.State != product.ProductStateInactive {
		t.Errorf("inactivated product = %+v, want id %s with state %q", got, id, product.ProductStateInactive)
	}

	// a second call must report that the product is already inactive
	rec = inactiveProduct(t, id)
	if rec.Code != http.StatusConflict {
		t.Errorf("second call status code = %d, want %d (body: %s)", rec.Code, http.StatusConflict, rec.Body.String())
	}
}

func TestInactiveProductErrors(t *testing.T) {
	resetProducts(t)

	tests := []struct {
		name string
		id   string
		want int
	}{
		{"not found", uuid.New().String(), http.StatusNotFound},
		{"invalid id", "not-a-uuid", http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := inactiveProduct(t, tt.id)
			if rec.Code != tt.want {
				t.Errorf("status code = %d, want %d (body: %s)", rec.Code, tt.want, rec.Body.String())
			}
		})
	}
}

func TestGetProductsExcludesInactive(t *testing.T) {
	resetProducts(t)
	activeID := seedProduct(t, "Keyboard", 49.99, 10, map[string]string{"color": "black"})
	inactiveID := seedProduct(t, "Mouse", 19.99, 5, map[string]string{"color": "white"})

	if rec := inactiveProduct(t, inactiveID); rec.Code != http.StatusOK {
		t.Fatalf("inactive status code = %d, want %d (body: %s)", rec.Code, http.StatusOK, rec.Body.String())
	}

	req := httptest.NewRequest(http.MethodGet, "/products", nil)
	rec := httptest.NewRecorder()
	testRouter.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d (body: %s)", rec.Code, http.StatusOK, rec.Body.String())
	}

	var got []product.ProductDTO
	decodeData(t, rec.Body.Bytes(), &got)
	if len(got) != 1 || got[0].ID != activeID {
		t.Errorf("got %+v, want only the active product %s", got, activeID)
	}

	// the inactive product is still reachable by id so it can be reactivated
	req = httptest.NewRequest(http.MethodGet, "/products/"+inactiveID, nil)
	rec = httptest.NewRecorder()
	testRouter.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("get inactive by id status code = %d, want %d", rec.Code, http.StatusOK)
	}
}
