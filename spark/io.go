package spark

import (
	"io"
	"os"
	"time"
)

// TODO(aybabtme): implement io.Closer interface to call sprk.Stop()
// properly.

type reader func(b []byte) (int, error)

func (r reader) Read(b []byte) (int, error) { return r(b) }

type writer func(b []byte) (int, error)

func (w writer) Write(b []byte) (int, error) { return w(b) }

// Reader wraps the reads of r with a SparkStream. The stream will
// have Bytes units and refresh every 33ms.
//
// It will stop printing when the reader returns an error.
func Reader(r io.Reader) io.Reader {
	sprk := Spark(time.Millisecond * 33)
	sprk.Units = Bytes
	started := false
	return reader(func(b []byte) (int, error) {
		if !started {
			sprk.Start()
			started = true
		}
		n, err := r.Read(b)
		if err == nil {
			sprk.Add(float64(n))
		} else {
			sprk.Stop()
		}
		return n, err
	})
}

// ReaderOut wraps the reads of r with a SparkStream. The stream will
// have Bytes units and refresh every 33ms.
//
// It will stop printing when the reader returns an error.
func ReaderOut(r io.Reader, out *os.File) io.Reader {
	sprk := Spark(time.Millisecond * 33)
	sprk.Out = out
	sprk.Units = Bytes
	started := false
	return reader(func(b []byte) (int, error) {
		if !started {
			sprk.Start()
			started = true
		}
		n, err := r.Read(b)
		if err == nil {
			sprk.Add(float64(n))
		} else {
			sprk.Stop()
		}
		return n, err
	})
}

// Writer wraps the writes to w with a SparkStream. The stream will
// have Bytes units and refresh every 33ms.
//
// It will stop printing when the writer returns an error.
func Writer(w io.Writer) (io.Writer, func()) {
	var out = os.Stderr
	if f, ok := w.(*os.File); ok && f == os.Stderr {
		out = os.Stdout
	}
	sprk := Spark(time.Millisecond * 33)
	sprk.Units = Bytes
	sprk.Out = out
	started := false
	return writer(func(b []byte) (int, error) {
		if !started {
			sprk.Start()
			started = true
		}
		n, err := w.Write(b)
		if err == nil {
			sprk.Add(float64(n))
		} else {
			sprk.Stop()
		}
		return n, err
	}), sprk.Stop
}

// WriteSeeker wraps the writes to w with a SparkStream. The stream
// will have Bytes units and refresh every 33ms.
//
// It will stop printing when the writer returns an error.
func WriteSeeker(ws io.WriteSeeker) (io.WriteSeeker, func()) {
	var out = os.Stderr
	if f, ok := ws.(*os.File); ok && f == os.Stderr {
		out = os.Stdout
	}
	sprk := Spark(time.Millisecond * 33)
	sprk.Units = Bytes
	sprk.Out = out
	return writeSeeker{
		WriteSeeker: ws,
		sprk:        sprk,
		started:     false,
	}, sprk.Stop
}

type writeSeeker struct {
	io.WriteSeeker
	sprk    *SparkStream
	started bool
}

func (w writeSeeker) Write(b []byte) (int, error) {
	if !w.started {
		w.sprk.Start()
		w.started = true
	}
	n, err := w.WriteSeeker.Write(b)
	if err == nil {
		w.sprk.Add(float64(n))
	} else {
		w.sprk.Stop()
	}
	return n, err
}
