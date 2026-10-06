// Package updater sincroniza los ficheros del juego gestionados contra el
// manifiesto: compara hashes y reemplaza solo lo que haya cambiado, mostrando
// progreso agregado y una estimación del tiempo restante.
package updater

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
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
	// 0. Barrer temporales de descarga huérfanos de una ejecución anterior que
	//    se interrumpiera a mitad (p. ej. el launcher cerrado durante una bajada).
	//    En funcionamiento normal no queda ninguno; esto evita que se acumulen.
	sweepLeftoverTemps(baseDir)

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
		if err := replaceFile(j.dest, j.f, tr); err != nil {
			return err
		}
		tr.finishFile(j.f.Size)
	}
	rep.Progress(totalBytes, totalBytes)
	return nil
}

// replaceFile descarga el fichero a un temporal (verificando su hash), hace una
// copia de seguridad .bak del actual si existía, y mueve el nuevo a su sitio.
func replaceFile(dest string, f manifest.File, tr *tracker) error {
	// El temporal se crea en la carpeta destino para que os.Rename sea atómico
	// (mismo volumen) y no un copy-across-devices.
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return fmt.Errorf("creando carpeta de %s: %w", f.Path, err)
	}

	tmpPath, err := download.Verified(f.URL, filepath.Dir(dest), f.SHA256,
		tr.onFileProgress,
		func(msg string) { tr.rep.Stage(msg) },
	)
	if err != nil {
		return err
	}

	// Respaldo TEMPORAL solo para poder revertir si el intercambio falla.
	// Se elimina al terminar: no dejamos ningún .bak.
	bak := dest + ".bak"
	hadDest := false
	if _, err := os.Stat(dest); err == nil {
		hadDest = true
		_ = os.Remove(bak)
		if err := os.Rename(dest, bak); err != nil {
			os.Remove(tmpPath)
			return fmt.Errorf("respaldando %s: %w", f.Path, err)
		}
	}

	if err := os.Rename(tmpPath, dest); err != nil {
		if hadDest {
			_ = os.Rename(bak, dest) // restaurar
		}
		os.Remove(tmpPath)
		return fmt.Errorf("reemplazando %s: %w", f.Path, err)
	}

	// Éxito: borramos el respaldo temporal.
	if hadDest {
		_ = os.Remove(bak)
	}
	return nil
}

// sweepLeftoverTemps borra, de forma silenciosa y best-effort, los temporales de
// descarga (download.TempPrefix) que hubieran quedado colgados bajo root por una
// ejecución anterior interrumpida a mitad de una bajada. No es un error que no
// haya ninguno ni que alguno no se pueda borrar.
func sweepLeftoverTemps(root string) {
	_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // carpeta ilegible: la saltamos sin abortar el barrido
		}
		if !d.IsDir() && strings.HasPrefix(d.Name(), download.TempPrefix) {
			_ = os.Remove(path)
		}
		return nil
	})
}

// tracker muestra el progreso agregado. Lleva por separado los bytes de los
// ficheros ya completados (completed) y los del fichero en curso (current), de
// modo que si una descarga se reintenta desde cero el porcentaje no se descuadra.
type tracker struct {
	rep       ui.Reporter
	total     int64
	completed int64
	current   int64
	start     time.Time
	lastUIAt  time.Time
}

// onFileProgress recibe los bytes acumulados del fichero en curso.
func (t *tracker) onFileProgress(fileDone int64) {
	t.current = fileDone
	now := time.Now()
	if now.Sub(t.lastUIAt) < 250*time.Millisecond && (t.completed+t.current) < t.total {
		return
	}
	t.lastUIAt = now
	t.render(now)
}

// finishFile contabiliza un fichero terminado.
func (t *tracker) finishFile(size int64) {
	t.completed += size
	t.current = 0
}

func (t *tracker) render(now time.Time) {
	if t.total <= 0 {
		t.rep.Stage("Actualizando ficheros…")
		return
	}
	done := t.completed + t.current
	if done > t.total {
		done = t.total
	}
	pct := int(float64(done) / float64(t.total) * 100)

	eta := ""
	elapsed := now.Sub(t.start).Seconds()
	if elapsed > 0.5 && done > 0 && done < t.total {
		speed := float64(done) / elapsed // bytes/seg
		if speed > 0 {
			eta = " · faltan ~" + formatDuration(float64(t.total-done)/speed)
		}
	}

	t.rep.Stage(fmt.Sprintf("Actualizando ficheros… %d%%%s", pct, eta))
	t.rep.Progress(done, t.total)
}

// formatDuration da una duración legible y breve a partir de segundos.
func formatDuration(seconds float64) string {
	s := int(seconds + 0.5)
	if s < 60 {
		return fmt.Sprintf("%d s", s)
	}
	return fmt.Sprintf("%d min %d s", s/60, s%60)
}
