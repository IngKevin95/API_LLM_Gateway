// Package wal implementa un Write-Ahead Log append-only en disco con rotación
// por tamaño y lectura de recuperación, para durabilidad de la auditoría.
package wal

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"api-llm-gateway/internal/audit"
)

// WAL es un log append-only con rotación. Seguro concurrente.
type WAL struct {
	mu       sync.Mutex
	dir      string
	maxBytes int64
	file     *os.File
	written  int64
	seq      int
}

// New abre (o crea) un WAL en dir con rotación al superar maxBytes.
func New(dir string, maxBytes int64) (*WAL, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	w := &WAL{dir: dir, maxBytes: maxBytes}
	if err := w.openNew(); err != nil {
		return nil, err
	}
	return w, nil
}

// Dir devuelve el directorio del WAL.
func (w *WAL) Dir() string { return w.dir }

func (w *WAL) openNew() error {
	w.seq++
	name := fmt.Sprintf("wal-%s-%03d.log", time.Now().UTC().Format("20060102T150405.000000"), w.seq)
	f, err := os.OpenFile(filepath.Join(w.dir, name), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return err
	}
	w.file = f
	w.written = 0
	return nil
}

// Append escribe un evento (longitud-prefijada) y rota si supera el umbral.
// No fuerza fsync en el camino crítico (overhead mínimo; durabilidad vía journal FS).
func (w *WAL) Append(e audit.Event) error {
	data, err := json.Marshal(e)
	if err != nil {
		return err
	}
	w.mu.Lock()
	defer w.mu.Unlock()

	var lenbuf [4]byte
	binary.BigEndian.PutUint32(lenbuf[:], uint32(len(data)))
	if _, err := w.file.Write(lenbuf[:]); err != nil {
		return err
	}
	if _, err := w.file.Write(data); err != nil {
		return err
	}
	w.written += int64(4 + len(data))
	if w.written >= w.maxBytes {
		if err := w.file.Close(); err != nil {
			return err
		}
		return w.openNew() // archiva el actual (ya nombrado) y abre uno nuevo
	}
	return nil
}

// Close cierra el archivo activo.
func (w *WAL) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.file != nil {
		return w.file.Close()
	}
	return nil
}

// Recover lee todos los registros de todos los archivos WAL del directorio
// (activo + archivados), en orden, para el replay de recuperación.
func (w *WAL) Recover() ([]audit.Event, error) {
	files, err := filepath.Glob(filepath.Join(w.dir, "wal-*.log"))
	if err != nil {
		return nil, err
	}
	sort.Strings(files) // el timestamp+seq en el nombre da orden cronológico

	var out []audit.Event
	for _, f := range files {
		events, err := readFile(f)
		if err != nil {
			return nil, err
		}
		out = append(out, events...)
	}
	return out, nil
}

func readFile(path string) ([]audit.Event, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r := bufio.NewReader(f)
	var out []audit.Event
	for {
		var lenbuf [4]byte
		if _, err := io.ReadFull(r, lenbuf[:]); err != nil {
			if err == io.EOF {
				break
			}
			if err == io.ErrUnexpectedEOF {
				break // registro parcial por crash mid-write: se ignora la cola
			}
			return nil, err
		}
		n := binary.BigEndian.Uint32(lenbuf[:])
		data := make([]byte, n)
		if _, err := io.ReadFull(r, data); err != nil {
			break // registro truncado: recuperamos hasta el último completo
		}
		var e audit.Event
		if err := json.Unmarshal(data, &e); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, nil
}
