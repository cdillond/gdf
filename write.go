package gdf

import (
	"crypto/md5"
	"fmt"
	"io"
	"time"
)

func writeHeader(p *PDF, w io.Writer) error {
	n, err := w.Write([]byte("%PDF-2.0\n\x81\x81\x81\x81\n"))
	p.n += n
	return err
}

func writeObjects(p *PDF, w io.Writer) error {
	for _, obj := range p.objects {
		p.xref = append(p.xref, p.n)
		t, err := fmt.Fprintf(w, "%d 0 obj\n", obj.refNum())
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
	t, err := fmt.Fprintf(w, "xref\n0 %d\n", len(p.xref))
	p.n += t
	if err != nil {
		return err
	}

	t, err = w.Write([]byte("0000000000 65536 f\n\r"))
	p.n += t
	if err != nil {
		return err
	}
	for i := 0; i < len(p.xref)-1; i++ {
		t, err = fmt.Fprintf(w, "%010d 00000 n\n\r", p.xref[i])
		p.n += t
		if err != nil {
			return err
		}
	}
	return nil
}

func writeTrailer(p *PDF, w io.Writer) error {
	_, err := w.Write([]byte("trailer\n"))
	if err != nil {
		return err
	}
	h := md5.New()
	id := h.Sum([]byte(fmt.Sprintf("%d", time.Now().Nanosecond())))
	_, err = fmt.Fprintf(w, "<<\n/Size %d\n/ID [<%X> <%X>]\n/Root 1 0 R\n>>\nstartxref\n%d\n%%%%EOF\n",
		len(p.xref), id, id, p.xref[len(p.xref)-1])
	return err
}
