package download

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"sync/atomic"
	"testing"
)

func hashOf(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

// El servidor falla la primera petición y responde bien la segunda: Verified
// debe reintentar y acabar con éxito.
func TestVerified_RetriesOnServerError(t *testing.T) {
	content := []byte("payload que llega al segundo intento")
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&calls, 1) == 1 {
			http.Error(w, "boom", http.StatusInternalServerError)
			return
		}
		w.Write(content)
	}))
	defer srv.Close()

	path, err := Verified(srv.URL+"/file", t.TempDir(), hashOf(content), nil, nil)
	if err != nil {
		t.Fatalf("esperaba éxito tras reintentar, error: %v", err)
	}
	got, _ := os.ReadFile(path)
	if string(got) != string(content) {
		t.Fatalf("contenido incorrecto: %q", got)
	}
	if calls < 2 {
		t.Fatalf("esperaba al menos 2 intentos, hubo %d", calls)
	}
}

// Un hash que no coincide NO debe reintentarse: una sola petición y error.
func TestVerified_NoRetryOnHashMismatch(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.Write([]byte("contenido real"))
	}))
	defer srv.Close()

	_, err := Verified(srv.URL+"/file", t.TempDir(), hashOf([]byte("otra cosa")), nil, nil)
	if err == nil {
		t.Fatal("esperaba error por hash que no coincide")
	}
	if calls != 1 {
		t.Fatalf("un hash incorrecto no debe reintentarse; hubo %d peticiones", calls)
	}
}
