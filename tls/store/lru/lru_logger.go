package lru

import "darvaza.org/slog"

func (s *LRU) logError(err error, note string) {
	if l, ok := s.withError(slog.Error, err); ok {
		l.Println(note)
	}
}

func (s *LRU) getLogger(level slog.LogLevel) (slog.Logger, bool) {
	if s != nil && s.logger != nil {
		return s.logger.WithLevel(level).WithEnabled()
	}
	return nil, false
}

func (s *LRU) withError(level slog.LogLevel, err error) (slog.Logger, bool) {
	if l, ok := s.getLogger(level); ok {
		if err != nil {
			l = l.WithField(slog.ErrorFieldName, err)
		}
		return l, true
	}

	return nil, false
}
