package certpool

import (
	"bytes"
	"context"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"testing"
)

func TestSystemCertPool(t *testing.T) {
	pool, err := SystemCertPool()
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	if deadLine, ok := t.Deadline(); ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithDeadline(ctx, deadLine)
		defer cancel()
	}

	i, count := 1, pool.Count()
	pool.ForEach(ctx, func(_ context.Context, cert *x509.Certificate) bool {
		printSystemCertTest(t, i, count, cert)
		i++
		return true
	})
}

func printSystemCertTest(t *testing.T, i, count int, cert *x509.Certificate) {
	var buf bytes.Buffer
	_, _ = fmt.Fprintf(&buf, "[%v/%v] ", i, count)
	if cert.IsCA {
		_, _ = buf.WriteString("CA ")
	}
	if len(cert.SubjectKeyId) > 0 {
		_, _ = buf.WriteString(base64.StdEncoding.EncodeToString(cert.SubjectKeyId))
		_, _ = buf.WriteRune(' ')
	}

	_, _ = fmt.Fprintf(&buf, "%q", cert.Subject)

	t.Log(buf.String())
}
