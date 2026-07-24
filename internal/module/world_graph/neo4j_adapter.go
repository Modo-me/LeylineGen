package world_graph

import (
	"context"

	"github.com/neo4j/neo4j-go-driver/v6/neo4j"
	"github.com/neo4j/neo4j-go-driver/v6/neo4j/db"
)

// graphResult is a local abstraction over neo4j.Result that only exposes
// the methods the Repository actually needs.  neo4j.Result contains
// unexported methods (buffer, errorHandler) that prevent external
// implementations, so we define our own.
type graphResult interface {
	Next(ctx context.Context) bool
	Record() *db.Record
	Err() error
	Consume(ctx context.Context) (neo4j.ResultSummary, error)
}

// graphSession is a local abstraction over neo4j.Session.
type graphSession interface {
	Run(ctx context.Context, cypher string, params map[string]any, configurers ...func(*neo4j.TransactionConfig)) (graphResult, error)
	Close(ctx context.Context) error
}

// grapher is a local abstraction over neo4j.Driver.
type grapher interface {
	NewSession(ctx context.Context, config neo4j.SessionConfig) graphSession
}

// ---- Production adapter: wraps neo4j.Driver into grapher ---------------

type neo4jGrapher struct {
	inner neo4j.Driver
}

func (g *neo4jGrapher) NewSession(ctx context.Context, config neo4j.SessionConfig) graphSession {
	s := g.inner.NewSession(ctx, config)
	return &neo4jGraphSession{inner: s}
}

type neo4jGraphSession struct {
	inner neo4j.Session
}

func (s *neo4jGraphSession) Run(ctx context.Context, cypher string, params map[string]any, configurers ...func(*neo4j.TransactionConfig)) (graphResult, error) {
	r, err := s.inner.Run(ctx, cypher, params, configurers...)
	if err != nil {
		return nil, err
	}
	return &neo4jGraphResult{inner: r}, nil
}

func (s *neo4jGraphSession) Close(ctx context.Context) error {
	return s.inner.Close(ctx)
}

type neo4jGraphResult struct {
	inner neo4j.Result
}

func (r *neo4jGraphResult) Next(ctx context.Context) bool {
	return r.inner.Next(ctx)
}

func (r *neo4jGraphResult) Record() *db.Record {
	return r.inner.Record()
}

func (r *neo4jGraphResult) Err() error {
	return r.inner.Err()
}

func (r *neo4jGraphResult) Consume(ctx context.Context) (neo4j.ResultSummary, error) {
	return r.inner.Consume(ctx)
}
