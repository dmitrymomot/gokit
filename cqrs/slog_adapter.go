package cqrs

import (
	"log/slog"

	"github.com/ThreeDotsLabs/watermill"
)

func NewSlogAdapter(log *slog.Logger) watermill.LoggerAdapter {
	return &slogAdapter{log: log}
}

type slogAdapter struct {
	log    *slog.Logger
	fields watermill.LogFields
}

func (s *slogAdapter) Error(msg string, err error, fields watermill.LogFields) {
	fields["error"] = err
	attr := s.allFields(fields)
	s.log.Error(msg, attr...)
}

func (s *slogAdapter) Info(msg string, fields watermill.LogFields) {
	attr := s.allFields(fields)
	s.log.Info(msg, attr...)
}

func (s *slogAdapter) Debug(msg string, fields watermill.LogFields) {
	attr := s.allFields(fields)
	s.log.Debug(msg, attr...)
}

func (s *slogAdapter) Trace(msg string, fields watermill.LogFields) {
	attr := s.allFields(fields)
	s.log.Debug(msg, attr...)
}

func (s *slogAdapter) With(fields watermill.LogFields) watermill.LoggerAdapter {
	return &slogAdapter{
		log:    s.log,
		fields: s.mergeFields(fields),
	}
}

func (s *slogAdapter) mergeFields(fields watermill.LogFields) watermill.LogFields {
	merged := make(watermill.LogFields, len(s.fields)+len(fields))
	for k, v := range s.fields {
		merged[k] = v
	}
	for k, v := range fields {
		merged[k] = v
	}
	return merged
}

func (s *slogAdapter) allFields(fields watermill.LogFields) []any {
	allFields := s.mergeFields(fields)
	attr := make([]any, 0, len(allFields))
	for k, v := range fields {
		attr = append(attr, slog.Any(k, v))
	}
	return attr
}
