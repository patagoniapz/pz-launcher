// Package download descarga ficheros a una ruta temporal verificando su
// SHA-256 ANTES de que el llamante los mueva a su destino final.
//
// Para tolerar conexiones inestables:
//   - Un "vigilante de cuelgue": si no llegan datos durante stallTimeout, se
//     cancela la petición (en vez de quedarse esperando indefinidamente).
//   - Reintentos automáticos con espera creciente ante errores de red.
//
// Un hash que no coincide NO se reintenta: significa que el fichero del servidor
// es otro, no un corte de red.
package download

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	maxAttempts  = 4
	stallTimeout = 20 * time.Second
)

// TempPrefix es el prefijo de los ficheros temporales de descarga. Se escriben
// en la carpeta destino y se renombran al nombre final tras verificar el hash.
// Expuesto para que los llamantes puedan barrer temporales huérfanos de una
// ejecución interrumpida.
const TempPrefix = ".pzl-download-"

var errHashMismatch = errors.New("el hash no coincide")

// Verified descarga url en un fichero temporal dentro de dir, comprueba que su
// SHA-256 coincide con wantHash y devuelve la ruta del temporal. El llamante es
// responsable de moverlo a su destino (os.Rename) y de borrarlo si algo falla.
//
// onProgress (si no es nil) recibe los bytes acumulados de ESTE fichero; se
// reinicia a 0 si una descarga se reintenta desde cero.
// onEvent (si no es nil) recibe mensajes de estado (p. ej. reintentos).
func Verified(url, dir, wantHash string, onProgress func(fileDone int64), onEvent func(msg string)) (string, error) {
	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if attempt > 1 && onEvent != nil {
			onEvent(fmt.Sprintf("Conexión interrumpida; reintentando (%d/%d)…", attempt, maxAttempts))
		}
		path, err := downloadOnce(url, dir, wantHash, onProgress)
		if err == nil {
			return path, nil
		}
		lastErr = err
		if errors.Is(err, errHashMismatch) {
			return "", err // fichero equivocado: no tiene sentido reintentar
		}
		if attempt < maxAttempts {
			time.Sleep(time.Duration(attempt) * time.Second) // espera creciente
		}
	}
	return "", fmt.Errorf("tras %d intentos: %w", maxAttempts, lastErr)
}

func downloadOnce(url, dir, wantHash string, onProgress func(fileDone int64)) (string, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("descargando %s: %w", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("descargando %s: HTTP %d", url, resp.StatusCode)
	}

	tmp, err := os.CreateTemp(dir, TempPrefix+"*")
	if err != nil {
		return "", fmt.Errorf("creando temporal: %w", err)
	}
	tmpPath := tmp.Name()

	// Vigilante: si no llegan datos durante stallTimeout, cancela el contexto
	// para que el io.Copy falle y se dispare un reintento.
	stall := time.AfterFunc(stallTimeout, cancel)

	h := sha256.New()
	pw := &progressWriter{
		onProgress: onProgress,
		onData:     func() { stall.Reset(stallTimeout) },
	}
	_, copyErr := io.Copy(io.MultiWriter(tmp, h, pw), resp.Body)
	stall.Stop()
	closeErr := tmp.Close()

	if copyErr != nil {
		os.Remove(tmpPath)
		if ctx.Err() != nil {
			return "", fmt.Errorf("descarga de %s detenida: sin datos durante %s", filepath.Base(url), stallTimeout)
		}
		return "", fmt.Errorf("guardando descarga: %w", copyErr)
	}
	if closeErr != nil {
		os.Remove(tmpPath)
		return "", fmt.Errorf("cerrando temporal: %w", closeErr)
	}

	got := hex.EncodeToString(h.Sum(nil))
	if !strings.EqualFold(got, wantHash) {
		os.Remove(tmpPath)
		return "", fmt.Errorf("%w en %s: esperado %s, obtenido %s",
			errHashMismatch, filepath.Base(url), wantHash, got)
	}
	return tmpPath, nil
}

// progressWriter cuenta los bytes acumulados del fichero y refresca el vigilante
// de cuelgue en cada bloque recibido.
type progressWriter struct {
	done       int64
	onProgress func(fileDone int64)
	onData     func()
}

func (p *progressWriter) Write(b []byte) (int, error) {
	n := len(b)
	p.done += int64(n)
	if p.onData != nil {
		p.onData()
	}
	if p.onProgress != nil {
		p.onProgress(p.done)
	}
	return n, nil
}
