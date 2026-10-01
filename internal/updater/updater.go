// Package updater sincroniza los ficheros del juego gestionados contra el
// manifiesto: compara hashes y reemplaza solo lo que haya cambiado, mostrando
// progreso agregado y una estimación del tiempo restante.
package updater

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"pzlauncher/internal/download"
	"pzlauncher/internal/hashutil"
	"pzlauncher/internal/manifest"
	"pzlauncher/internal/ui"
)

// SyncFiles recorre los ficheros del manifiesto, comparando el hash local con
// el esperado. Descarga y reemplaza únicamente los que falten o difieran.
// baseDir es la carpeta del launcher; los paths del manifiesto son relativos a ella.
func SyncFiles(baseDir string, files []manifest.File, rep ui.Reporter) error {
	// 1. Calcular el conjunto de trabajo (ficheros que cambian) y el total de
	//    bytes, para poder mostrar un porcentaje y ETA global.
	type job struct {
		f    manifest.File
		dest string
	}
	var jobs []job
	var totalBytes int64
	for _, f := range files {
		dest := filepath.Join(baseDir, filepath.FromSlash(f.Path))
		local, err := hashutil.SHA256File(dest)
		if err != nil {
			return fmt.Errorf("hash local de %s: %w", f.Path, err)
		}
		if local == f.SHA256 {
			continue
		}
		jobs = append(jobs, job{f, dest})
		totalBytes += f.Size
	}
	if len(jobs) == 0 {
		return nil // todo al día
	}

	// 2. Descargar y reemplazar, actualizando el progreso agregado.
	tr := &tracker{rep: rep, total: totalBytes, start: time.Now()}
	tr.render(time.Now())
	for _, j := range jobs {
		if err := replaceFile(j.dest, j.f, tr.add); err != nil {
			return err
		}
	}
	rep.Progress(totalBytes, totalBytes)
	return nil
}

// replaceFile descarga el fichero a un temporal (verificando su hash), hace una
// copia de seguridad .bak del actual si existía, y mueve el nuevo a su sitio.
func replaceFile(dest string, f manifest.File, onChunk func(n int64)) error {
	// El temporal se crea en la carpeta destino para que os.Rename sea atómico
	// (mismo volumen) y no un copy-across-devices.
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return fmt.Errorf("creando carpeta de %s: %w", f.Path, err)
	}

	tmpPath, err := download.Verified(f.URL, filepath.Dir(dest), f.SHA256, onChunk)
	if err != nil {
		return err
	}

	bak := dest + ".bak"
	if _, err := os.Stat(dest); err == nil {
		_ = os.Remove(bak)
		if err := os.Rename(dest, bak); err != nil {
			os.Remove(tmpPath)
			return fmt.Errorf("respaldando %s: %w", f.Path, err)
		}
	}

	if err := os.Rename(tmpPath, dest); err != nil {
		// Intentar restaurar el respaldo si el swap falla.
		_ = os.Rename(bak, dest)
		os.Remove(tmpPath)
		return fmt.Errorf("reemplazando %s: %w", f.Path, err)
	}

	// Éxito: el .bak queda como red de seguridad hasta la próxima actualización.
	return nil
}

// tracker acumula los bytes descargados y refresca la UI (porcentaje + ETA),
// limitando la frecuencia de refresco para no saturar el diálogo.
type tracker struct {
	rep      ui.Reporter
	total    int64
	done     int64
	start    time.Time
	lastUIAt time.Time
}

func (t *tracker) add(n int64) {
	t.done += n
	now := time.Now()
	if now.Sub(t.lastUIAt) < 250*time.Millisecond && t.done < t.total {
		return
	}
	t.lastUIAt = now
	t.render(now)
}

func (t *tracker) render(now time.Time) {
	if t.total <= 0 {
		t.rep.Stage("Actualizando ficheros…")
		return
	}
	pct := int(float64(t.done) / float64(t.total) * 100)
	if pct > 100 {
		pct = 100
	}

	eta := ""
	elapsed := now.Sub(t.start).Seconds()
	if elapsed > 0.5 && t.done > 0 && t.done < t.total {
		speed := float64(t.done) / elapsed // bytes/seg
		if speed > 0 {
			eta = " · faltan ~" + formatDuration(float64(t.total-t.done)/speed)
		}
	}

	t.rep.Stage(fmt.Sprintf("Actualizando ficheros… %d%%%s", pct, eta))
	t.rep.Progress(t.done, t.total)
}

// formatDuration da una duración legible y breve a partir de segundos.
func formatDuration(seconds float64) string {
	s := int(seconds + 0.5)
	if s < 60 {
		return fmt.Sprintf("%d s", s)
	}
	return fmt.Sprintf("%d min %d s", s/60, s%60)
}
