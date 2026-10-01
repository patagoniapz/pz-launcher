// Servidor estático de PRUEBAS para probar el launcher en local.
// Sirve devtest/www en http://127.0.0.1:8099 y limita la velocidad del .jar
// para que la barra de progreso y el ETA se vean con claridad.
//
// Uso (desde la raíz del repo):
//
//	go run ./devtest/server
package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

const rate = 1 << 20 // 1 MiB/s

func main() {
	www, err := filepath.Abs("devtest/www")
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		p := filepath.Join(www, filepath.Clean(r.URL.Path))
		f, err := os.Open(p)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		defer f.Close()

		fi, _ := f.Stat()
		w.Header().Set("Content-Length", strconv.FormatInt(fi.Size(), 10))

		// Solo limitamos la velocidad del .jar (para ver el ETA); el resto va directo.
		if filepath.Ext(p) != ".jar" {
			io.Copy(w, f)
			return
		}

		buf := make([]byte, 64<<10)
		for {
			n, readErr := f.Read(buf)
			if n > 0 {
				if _, err := w.Write(buf[:n]); err != nil {
					return
				}
				if fl, ok := w.(http.Flusher); ok {
					fl.Flush()
				}
				time.Sleep(time.Duration(float64(n) / float64(rate) * float64(time.Second)))
			}
			if readErr != nil {
				return
			}
		}
	})

	log.Println("Servidor de pruebas en http://127.0.0.1:8099  (Ctrl+C para parar)")
	log.Fatal(http.ListenAndServe("127.0.0.1:8099", nil))
}
