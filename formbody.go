package multipart

import (
	"bytes"
	"fmt"
	"io"
	"net/textproto"
	"os"
	"sort"
)

type FormBody struct {
	boundary string
	b        *bytes.Buffer
	mr       multiReader
	closed   bool
}

func NewFormBody() *FormBody {
	b := &FormBody{
		b:        bytes.NewBuffer(nil),
		boundary: randomBoundary(),
	}
	return b
}

func (w *FormBody) Read(p []byte) (n int, err error) {
	if len(w.mr.readers) == 0 {
		return 0, io.EOF
	}
	if !w.closed {
		w.Close()
	}
	return w.mr.Read(p)
}

// Boundary returns the Writer's boundary.
func (w *FormBody) Boundary() string {
	return w.boundary
}

func (w *FormBody) FormDataContentType() string {
	return "multipart/form-data; boundary=" + w.boundary
}

func (w *FormBody) Close() error {
	if !w.closed {
		w.closed = true
		fmt.Fprintf(w.b, "\r\n--%s--\r\n", w.boundary)
		w.mr.readers = append(w.mr.readers, w.b)
	}
	return nil
}

func (f *FormBody) getPart() bool {
	if f.b.Len() != 0 {
		raw := make([]byte, f.b.Len())
		copy(raw, f.b.Bytes())
		f.mr.readers = append(f.mr.readers, bytes.NewReader(raw))
		f.b.Reset()
		return true
	}
	return false
}

func (f *FormBody) CreatePart(header textproto.MIMEHeader) (io.Writer, error) {
	if f.closed {
		return nil, fmt.Errorf("Closed")
	}
	if f.getPart() {
		fmt.Fprintf(f.b, "\r\n--%s\r\n", f.boundary)
	} else {
		fmt.Fprintf(f.b, "--%s\r\n", f.boundary)
	}
	keys := make([]string, 0, len(header))
	for k := range header {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		for _, v := range header[k] {
			fmt.Fprintf(f.b, "%s: %s\r\n", k, v)
		}
	}
	fmt.Fprintf(f.b, "\r\n")
	f.getPart()
	return f.b, nil

}

func (w *FormBody) CreateFromByPath(fieldname, filename, filePath string) error {
	fd, err := os.Open(filePath)
	if err != nil {
		return err
	}
	return w.CreateFromReader(fieldname, filename, fd)
}

// CreateFromReader use ioutil.NopCloser(rd) for Reader or do not want to close .
func (w *FormBody) CreateFromReader(fieldname, filename string, rd io.ReadCloser) error {
	if _, err := w.CreateFormFile(fieldname, filename); err != nil {
		return err
	}
	w.mr.readers = append(w.mr.readers, rd)
	return nil
}

func (w *FormBody) CreateFormFile(fieldname, filename string) (io.Writer, error) {
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition",
		fmt.Sprintf(`form-data; name="%s"; filename="%s"`,
			escapeQuotes(fieldname), escapeQuotes(filename)))
	h.Set("Content-Type", "application/octet-stream")
	return w.CreatePart(h)
}

// CreateFormField calls CreatePart with a header using the
// given field name.
func (w *FormBody) CreateFormField(fieldname string) (io.Writer, error) {
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition",
		fmt.Sprintf(`form-data; name="%s"`, escapeQuotes(fieldname)))
	return w.CreatePart(h)
}

// WriteField calls CreateFormField and then writes the given value.
func (w *FormBody) WriteField(fieldname, value string) error {
	p, err := w.CreateFormField(fieldname)
	if err != nil {
		return err
	}
	_, err = p.Write([]byte(value))
	return err
}

type eofReader struct{}

func (eofReader) Read([]byte) (int, error) {
	return 0, io.EOF
}

type multiReader struct {
	readers []io.Reader
}

func (mr *multiReader) Read(p []byte) (n int, err error) {
	for len(mr.readers) > 0 {
		// Optimization to flatten nested multiReaders (Issue 13558).
		if len(mr.readers) == 1 {
			if r, ok := mr.readers[0].(*multiReader); ok {
				mr.readers = r.readers
				continue
			}
		}
		n, err = mr.readers[0].Read(p)
		if err == io.EOF {
			if closer, ok := mr.readers[0].(io.Closer); ok {
				closer.Close()
			}
			// Use eofReader instead of nil to avoid nil panic
			// after performing flatten (Issue 18232).
			mr.readers[0] = eofReader{} // permit earlier GC
			mr.readers = mr.readers[1:]
		}
		if n > 0 || err != io.EOF {
			if err == io.EOF && len(mr.readers) > 0 {
				// Don't return EOF yet. More readers remain.
				err = nil
			}
			return
		}
	}
	return 0, io.EOF
}
