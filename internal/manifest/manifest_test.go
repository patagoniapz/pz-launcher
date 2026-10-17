package manifest

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestFetch_BustsCaches comprueba que la petición del manifiesto pide contenido
// fresco (cabeceras no-cache y parámetro único anti-caché) y que el JSON se
// decodifica igual con el parámetro añadido.
func TestFetch_BustsCaches(t *testing.T) {
	var gotReq *http.Request
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotReq = r.Clone(r.Context())
		w.Write([]byte(`{"schema":1,"launcher_version":"1.2.3"}`))
	}))
	defer srv.Close()

	m, err := Fetch(srv.URL + "/manifest.json")
	if err != nil {
		t.Fatalf("Fetch devolvió error: %v", err)
	}
	if m.LauncherVersion != "1.2.3" {
		t.Fatalf("LauncherVersion = %q, se esperaba 1.2.3", m.LauncherVersion)
	}
	if gotReq.Header.Get("Cache-Control") != "no-cache" {
		t.Fatal("falta la cabecera Cache-Control: no-cache")
	}
	if gotReq.URL.Query().Get("ts") == "" {
		t.Fatal("falta el parámetro anti-caché ts en la URL")
	}
}

// TestFetch_AppendsToExistingQuery comprueba que el anti-caché no rompe URLs
// que ya llevan parámetros (se añade con & en vez de un segundo ?).
func TestFetch_AppendsToExistingQuery(t *testing.T) {
	var gotReq *http.Request
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotReq = r.Clone(r.Context())
		w.Write([]byte(`{"schema":1}`))
	}))
	defer srv.Close()

	if _, err := Fetch(srv.URL + "/manifest.json?canal=beta"); err != nil {
		t.Fatalf("Fetch devolvió error: %v", err)
	}
	q := gotReq.URL.Query()
	if q.Get("canal") != "beta" {
		t.Fatal("se perdió el parámetro original de la URL")
	}
	if q.Get("ts") == "" {
		t.Fatal("falta el parámetro anti-caché ts en la URL")
	}
}
