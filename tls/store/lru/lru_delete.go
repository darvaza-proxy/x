package lru

import (
	"context"

	"darvaza.org/core"
	"darvaza.org/x/tls"
)

func (s *LRU) Delete(ctx context.Context, _ *tls.Certificate) error {
	if err := s.checkInit(); err != nil {
		return err
	} else if err := ctx.Err(); err != nil {
		return err
	}

	return core.ErrTODO
}
