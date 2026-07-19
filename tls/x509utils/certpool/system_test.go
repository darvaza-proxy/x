package certpool_test

import (
	"bytes"
	"context"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"fmt"
	"testing"

	"darvaza.org/core"

	"darvaza.org/x/tls/x509utils/certpool"
)

func TestSystemCertPool(t *testing.T) {
	pool, err := certpool.SystemCertPool()
	if errors.Is(err, core.ErrTODO) {
		t.Skipf("system cert pool loader not implemented: %v", err)
	}
	core.AssertMustNoError(t, err, "system pool")
	core.AssertMustNotNil(t, pool, "pool")

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
	core.AssertEqual(t, count, i-1, "visited all")
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
