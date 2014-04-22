package ioutil

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"testing"
)

func TestSelect(t *testing.T) {
	rdrs := []io.Reader{bytes.NewBuffer([]byte("hello there\n")), bytes.NewBuffer([]byte("goodbye now\n"))}

	// Try both rdrs[0] and rdrs[1]
	rdrs[0] = blockingReader{}

	s, _ := NewSelect(rdrs...)

	n, r := s.Select()
	fmt.Printf("rdrs: %v\n: %v\r: %v\n\n", rdrs, n, r)

	o, err := io.Copy(os.Stdout, r)
	fmt.Printf("\no: %v\nerr: %v\n", o, err)
}

type blockingReader struct{}

func (b blockingReader) Read(p []byte) (int, error) {
	select {}
}
