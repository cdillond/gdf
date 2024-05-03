package gdf

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"io"
	"slices"
)

func writeHeader(p *PDF, w io.Writer) error {
	if p.info == nil && len(p.catalog.Acroform.acrofields) == 0 {
		n, err := w.Write([]byte("%PDF-2.0\n\x81\x81\x81\x81\n"))
		p.n += n
		return err
	}
	n, err := w.Write([]byte("%PDF-1.7\n\x81\x81\x81\x81\n"))
	p.n += n
	return err
}

func writeObjects(p *PDF, w io.Writer) error {
	for _, obj := range p.objects {
		p.xref = append(p.xref, p.n)
		t, err := w.Write([]byte(itoa(obj.id()) + "\x200\x20obj\n"))
		p.n += t
		if err != nil {
			return err
		}
		t, err = obj.encode(w)
		p.n += t
		if err != nil {
			return err
		}
		t, err = w.Write([]byte("endobj\n"))
		p.n += t
		if err != nil {
			return err
		}
	}
	return nil
}

func writeXref(p *PDF, w io.Writer) error {
	p.xref = append(p.xref, p.n) // adding the offset even though it won't be included yet
	t, err := w.Write([]byte("xref\n" +
		"0\x20" + itoa(len(p.xref)) + "\n" +
		"0000000000\x2065536\x20f\n\r"))
	p.n += t
	if err != nil {
		return err
	}
	b := []byte("0000000000\x2000000\x20n\n\r")
	for i := 0; i < len(p.xref)-1; i++ {
		pad10(p.xref[i], b)
		t, err := w.Write(b)
		p.n += t
		if err != nil {
			return err
		}
	}
	return nil
}

func writeTrailer(p *PDF, w io.Writer) error {
	t, err := w.Write([]byte("trailer\n"))
	if err != nil {
		return err
	}
	p.n += t

	// 14.5: The ID string should be based on the current time, the PDF's file location, or the size of the file in bytes.
	// Since the first two options are not WASM compatible, we'll go with the third, but it will be annoying.

	// The md5 hash produces a 16 byte output, which is 32 bytes when hex-encoded. idx will serve as a placeholder.
	idx := make([]byte, 32)

	fields := []field{
		{"/Size", len(p.xref)},
		{"/ID", slices.Concat([]byte("[<"), idx, []byte("> <"), idx, []byte(">]"))},
		{"/Root", "1\x200\x20R"},
	}
	if p.info != nil {
		fields = append(fields, field{"/Info", iref(p.info)})
	}
	buf := dict(512, fields)
	buf = append(buf, "startxref\n"...)
	buf = itobuf(p.xref[len(p.xref)-1], buf)
	buf = append(buf, "\n%%EOF\n"...)

	// now that everything has been appended to buf, we know the final file length.
	h := md5.New()
	h.Write(itob(p.n + len(buf)))
	idx = hex.AppendEncode(nil, h.Sum(nil))

	// Overwrite the 0 bytes.
	i := bytes.IndexByte(buf, 0)
	var n int
	if i > -1 {
		n = copy(buf[i:], idx)
	}
	i = bytes.IndexByte(buf[i+n:], 0)
	if i > -1 {
		copy(buf[i:], idx)
	}

	t, err = w.Write(buf)

	p.n += t
	return err
}
