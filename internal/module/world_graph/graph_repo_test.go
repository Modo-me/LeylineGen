package world_graph

import (
	"context"
	"errors"
	"testing"

	"github.com/neo4j/neo4j-go-driver/v6/neo4j"
	"github.com/neo4j/neo4j-go-driver/v6/neo4j/db"
)

// ---------------------------------------------------------------------------
// mocks for grapher / graphSession / graphResult
// ---------------------------------------------------------------------------

type mockCounters struct {
	nodesCreated         int
	nodesDeleted         int
	relationshipsCreated int
	relationshipsDeleted int
	propertiesSet        int
	labelsAdded          int
	labelsRemoved        int
	indexesAdded         int
	indexesRemoved       int
	constraintsAdded     int
	constraintsRemoved   int
	systemUpdates        int
}

func (m *mockCounters) ContainsUpdates() bool                      { return m.nodesCreated > 0 || m.relationshipsCreated > 0 }
func (m *mockCounters) NodesCreated() int                          { return m.nodesCreated }
func (m *mockCounters) NodesDeleted() int                          { return m.nodesDeleted }
func (m *mockCounters) RelationshipsCreated() int                  { return m.relationshipsCreated }
func (m *mockCounters) RelationshipsDeleted() int                  { return m.relationshipsDeleted }
func (m *mockCounters) PropertiesSet() int                         { return m.propertiesSet }
func (m *mockCounters) LabelsAdded() int                           { return m.labelsAdded }
func (m *mockCounters) LabelsRemoved() int                         { return m.labelsRemoved }
func (m *mockCounters) IndexesAdded() int                          { return m.indexesAdded }
func (m *mockCounters) IndexesRemoved() int                        { return m.indexesRemoved }
func (m *mockCounters) ConstraintsAdded() int                      { return m.constraintsAdded }
func (m *mockCounters) ConstraintsRemoved() int                    { return m.constraintsRemoved }
func (m *mockCounters) SystemUpdates() int                         { return m.systemUpdates }
func (m *mockCounters) ContainsSystemUpdates() bool                { return m.systemUpdates > 0 }

type mockResultSummary struct {
	neo4j.ResultSummary
	counters neo4j.Counters
}

func (m *mockResultSummary) Counters() neo4j.Counters { return m.counters }

type mockGraphResult struct {
	records []*db.Record
	index   int
	err     error
	summary neo4j.ResultSummary
}

func (m *mockGraphResult) Next(_ context.Context) bool {
	m.index++
	return m.index <= len(m.records)
}

func (m *mockGraphResult) Record() *db.Record {
	if m.index > 0 && m.index <= len(m.records) {
		return m.records[m.index-1]
	}
	return nil
}

func (m *mockGraphResult) Err() error { return m.err }

func (m *mockGraphResult) Consume(_ context.Context) (neo4j.ResultSummary, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.summary != nil {
		return m.summary, nil
	}
	return &mockResultSummary{counters: &mockCounters{nodesCreated: 1, relationshipsCreated: 1}}, nil
}

type mockGraphSession struct {
	runFunc func(ctx context.Context, cypher string, params map[string]any) (graphResult, error)
}

func (m *mockGraphSession) Run(ctx context.Context, cypher string, params map[string]any, _ ...func(*neo4j.TransactionConfig)) (graphResult, error) {
	return m.runFunc(ctx, cypher, params)
}

func (m *mockGraphSession) Close(_ context.Context) error { return nil }

type mockGrapher struct {
	sessionFunc func(ctx context.Context, config neo4j.SessionConfig) graphSession
}

func (m *mockGrapher) NewSession(ctx context.Context, config neo4j.SessionConfig) graphSession {
	return m.sessionFunc(ctx, config)
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func newRepoWithMocks(g grapher) *Repository {
	return &Repository{driver: g}
}

func recordVillage(id int64, name string, x, z int64) *db.Record {
	return &db.Record{
		Keys:   []string{"Id", "Name", "X", "Z"},
		Values: []any{id, name, x, z},
	}
}

func recordNpc(id int64, name string) *db.Record {
	return &db.Record{
		Keys:   []string{"Id", "Name"},
		Values: []any{id, name},
	}
}

func recordDirection(dir string) *db.Record {
	return &db.Record{
		Keys:   []string{"Direction"},
		Values: []any{dir},
	}
}

func recordPlayer(x, z int64) *db.Record {
	return &db.Record{
		Keys:   []string{"X", "Z"},
		Values: []any{x, z},
	}
}

// ---------------------------------------------------------------------------
// Tests — Read operations
// ---------------------------------------------------------------------------

func TestGetConnectedVillages(t *testing.T) {
	t.Run("returns villages", func(t *testing.T) {
		repo := newRepoWithMocks(&mockGrapher{
			sessionFunc: func(_ context.Context, _ neo4j.SessionConfig) graphSession {
				return &mockGraphSession{
					runFunc: func(_ context.Context, _ string, params map[string]any) (graphResult, error) {
						if params["id"] != int64(1) {
							t.Errorf("expected id 1, got %v", params["id"])
						}
						return &mockGraphResult{
							records: []*db.Record{
								recordVillage(2, "v2", 10, 20),
								recordVillage(3, "v3", 30, 40),
							},
						}, nil
					},
				}
			},
		})

		nodes, err := repo.GetConnectedVillages(context.Background(), VillageNode{ID: 1})
		if err != nil {
			t.Fatal(err)
		}
		if len(nodes) != 2 {
			t.Fatalf("expected 2 villages, got %d", len(nodes))
		}
		if nodes[0].ID != 2 || nodes[0].Name != "v2" || nodes[0].X != 10 || nodes[0].Z != 20 {
			t.Errorf("unexpected first village: %+v", nodes[0])
		}
		if nodes[1].ID != 3 || nodes[1].Name != "v3" || nodes[1].X != 30 || nodes[1].Z != 40 {
			t.Errorf("unexpected second village: %+v", nodes[1])
		}
	})

	t.Run("run error", func(t *testing.T) {
		repo := newRepoWithMocks(&mockGrapher{
			sessionFunc: func(_ context.Context, _ neo4j.SessionConfig) graphSession {
				return &mockGraphSession{
					runFunc: func(_ context.Context, _ string, _ map[string]any) (graphResult, error) {
						return nil, errors.New("neo4j down")
					},
				}
			},
		})
		_, err := repo.GetConnectedVillages(context.Background(), VillageNode{ID: 1})
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("iteration error", func(t *testing.T) {
		repo := newRepoWithMocks(&mockGrapher{
			sessionFunc: func(_ context.Context, _ neo4j.SessionConfig) graphSession {
				return &mockGraphSession{
					runFunc: func(_ context.Context, _ string, _ map[string]any) (graphResult, error) {
						return &mockGraphResult{
							records: []*db.Record{recordVillage(2, "v2", 10, 20)},
							err:     errors.New("iteration failed"),
						}, nil
					},
				}
			},
		})
		_, err := repo.GetConnectedVillages(context.Background(), VillageNode{ID: 1})
		if err == nil {
			t.Fatal("expected iteration error")
		}
	})

	t.Run("empty results", func(t *testing.T) {
		repo := newRepoWithMocks(&mockGrapher{
			sessionFunc: func(_ context.Context, _ neo4j.SessionConfig) graphSession {
				return &mockGraphSession{
					runFunc: func(_ context.Context, _ string, _ map[string]any) (graphResult, error) {
						return &mockGraphResult{records: nil}, nil
					},
				}
			},
		})
		nodes, err := repo.GetConnectedVillages(context.Background(), VillageNode{ID: 1})
		if err != nil {
			t.Fatal(err)
		}
		if len(nodes) != 0 {
			t.Fatalf("expected 0 villages, got %d", len(nodes))
		}
	})
}

func TestGetConnectedVillageByDirection(t *testing.T) {
	t.Run("returns village", func(t *testing.T) {
		repo := newRepoWithMocks(&mockGrapher{
			sessionFunc: func(_ context.Context, _ neo4j.SessionConfig) graphSession {
				return &mockGraphSession{
					runFunc: func(_ context.Context, _ string, params map[string]any) (graphResult, error) {
						if params["dir"] != "east" {
							t.Errorf("expected dir east, got %v", params["dir"])
						}
						return &mockGraphResult{
							records: []*db.Record{recordVillage(2, "v2", 10, 20)},
						}, nil
					},
				}
			},
		})

		v, err := repo.GetConnectedVillageByDirection(context.Background(), VillageNode{ID: 1}, ConnectedRel{Direction: East})
		if err != nil {
			t.Fatal(err)
		}
		if v == nil {
			t.Fatal("expected non-nil village")
		}
		if v.ID != 2 || v.Name != "v2" || v.X != 10 || v.Z != 20 {
			t.Errorf("unexpected village: %+v", v)
		}
	})

	t.Run("not found returns nil", func(t *testing.T) {
		repo := newRepoWithMocks(&mockGrapher{
			sessionFunc: func(_ context.Context, _ neo4j.SessionConfig) graphSession {
				return &mockGraphSession{
					runFunc: func(_ context.Context, _ string, _ map[string]any) (graphResult, error) {
						return &mockGraphResult{records: nil}, nil
					},
				}
			},
		})
		v, err := repo.GetConnectedVillageByDirection(context.Background(), VillageNode{ID: 1}, ConnectedRel{Direction: North})
		if err != nil {
			t.Fatal(err)
		}
		if v != nil {
			t.Fatal("expected nil village")
		}
	})

	t.Run("run error", func(t *testing.T) {
		repo := newRepoWithMocks(&mockGrapher{
			sessionFunc: func(_ context.Context, _ neo4j.SessionConfig) graphSession {
				return &mockGraphSession{
					runFunc: func(_ context.Context, _ string, _ map[string]any) (graphResult, error) {
						return nil, errors.New("neo4j down")
					},
				}
			},
		})
		_, err := repo.GetConnectedVillageByDirection(context.Background(), VillageNode{ID: 1}, ConnectedRel{Direction: South})
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestGetPlayerConnectedVillages(t *testing.T) {
	t.Run("returns villages", func(t *testing.T) {
		repo := newRepoWithMocks(&mockGrapher{
			sessionFunc: func(_ context.Context, _ neo4j.SessionConfig) graphSession {
				return &mockGraphSession{
					runFunc: func(_ context.Context, _ string, _ map[string]any) (graphResult, error) {
						return &mockGraphResult{
							records: []*db.Record{
								recordVillage(1, "v1", 0, 0),
								recordVillage(2, "v2", 100, 200),
							},
						}, nil
					},
				}
			},
		})
		vs, err := repo.GetPlayerConnectedVillages(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if len(vs) != 2 {
			t.Fatalf("expected 2 villages, got %d", len(vs))
		}
	})

	t.Run("run error", func(t *testing.T) {
		repo := newRepoWithMocks(&mockGrapher{
			sessionFunc: func(_ context.Context, _ neo4j.SessionConfig) graphSession {
				return &mockGraphSession{
					runFunc: func(_ context.Context, _ string, _ map[string]any) (graphResult, error) {
						return nil, errors.New("error")
					},
				}
			},
		})
		_, err := repo.GetPlayerConnectedVillages(context.Background())
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("iteration error", func(t *testing.T) {
		repo := newRepoWithMocks(&mockGrapher{
			sessionFunc: func(_ context.Context, _ neo4j.SessionConfig) graphSession {
				return &mockGraphSession{
					runFunc: func(_ context.Context, _ string, _ map[string]any) (graphResult, error) {
						return &mockGraphResult{
							records: []*db.Record{recordVillage(1, "v1", 0, 0)},
							err:     errors.New("iter err"),
						}, nil
					},
				}
			},
		})
		_, err := repo.GetPlayerConnectedVillages(context.Background())
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestGetAllNpcsInVillage(t *testing.T) {
	t.Run("returns npcs", func(t *testing.T) {
		repo := newRepoWithMocks(&mockGrapher{
			sessionFunc: func(_ context.Context, _ neo4j.SessionConfig) graphSession {
				return &mockGraphSession{
					runFunc: func(_ context.Context, _ string, params map[string]any) (graphResult, error) {
						if params["id"] != int64(10) {
							t.Errorf("expected id 10, got %v", params["id"])
						}
						return &mockGraphResult{
							records: []*db.Record{
								recordNpc(1, "npc1"),
								recordNpc(2, "npc2"),
							},
						}, nil
					},
				}
			},
		})
		npcs, err := repo.GetAllNpcsInVillage(context.Background(), VillageNode{ID: 10})
		if err != nil {
			t.Fatal(err)
		}
		if len(npcs) != 2 {
			t.Fatalf("expected 2 npcs, got %d", len(npcs))
		}
		if npcs[0].ID != 1 || npcs[0].Name != "npc1" {
			t.Errorf("unexpected first npc: %+v", npcs[0])
		}
	})

	t.Run("empty", func(t *testing.T) {
		repo := newRepoWithMocks(&mockGrapher{
			sessionFunc: func(_ context.Context, _ neo4j.SessionConfig) graphSession {
				return &mockGraphSession{
					runFunc: func(_ context.Context, _ string, _ map[string]any) (graphResult, error) {
						return &mockGraphResult{records: nil}, nil
					},
				}
			},
		})
		npcs, err := repo.GetAllNpcsInVillage(context.Background(), VillageNode{ID: 5})
		if err != nil {
			t.Fatal(err)
		}
		if len(npcs) != 0 {
			t.Fatalf("expected 0 npcs, got %d", len(npcs))
		}
	})

	t.Run("run error", func(t *testing.T) {
		repo := newRepoWithMocks(&mockGrapher{
			sessionFunc: func(_ context.Context, _ neo4j.SessionConfig) graphSession {
				return &mockGraphSession{
					runFunc: func(_ context.Context, _ string, _ map[string]any) (graphResult, error) {
						return nil, errors.New("err")
					},
				}
			},
		})
		_, err := repo.GetAllNpcsInVillage(context.Background(), VillageNode{ID: 5})
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestGetNpcNodeByID(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		repo := newRepoWithMocks(&mockGrapher{
			sessionFunc: func(_ context.Context, _ neo4j.SessionConfig) graphSession {
				return &mockGraphSession{
					runFunc: func(_ context.Context, _ string, params map[string]any) (graphResult, error) {
						if params["id"] != int64(42) {
							t.Errorf("expected id 42, got %v", params["id"])
						}
						return &mockGraphResult{
							records: []*db.Record{recordNpc(42, "alice")},
						}, nil
					},
				}
			},
		})
		npc, err := repo.GetNpcNodeByID(context.Background(), 42)
		if err != nil {
			t.Fatal(err)
		}
		if npc == nil {
			t.Fatal("expected non-nil npc")
		}
		if npc.ID != 42 || npc.Name != "alice" {
			t.Errorf("unexpected npc: %+v", npc)
		}
	})

	t.Run("not found returns nil", func(t *testing.T) {
		repo := newRepoWithMocks(&mockGrapher{
			sessionFunc: func(_ context.Context, _ neo4j.SessionConfig) graphSession {
				return &mockGraphSession{
					runFunc: func(_ context.Context, _ string, _ map[string]any) (graphResult, error) {
						return &mockGraphResult{records: nil}, nil
					},
				}
			},
		})
		npc, err := repo.GetNpcNodeByID(context.Background(), 999)
		if err != nil {
			t.Fatal(err)
		}
		if npc != nil {
			t.Fatal("expected nil npc")
		}
	})

	t.Run("run error", func(t *testing.T) {
		repo := newRepoWithMocks(&mockGrapher{
			sessionFunc: func(_ context.Context, _ neo4j.SessionConfig) graphSession {
				return &mockGraphSession{
					runFunc: func(_ context.Context, _ string, _ map[string]any) (graphResult, error) {
						return nil, errors.New("err")
					},
				}
			},
		})
		_, err := repo.GetNpcNodeByID(context.Background(), 1)
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestGetVillageByNpc(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		repo := newRepoWithMocks(&mockGrapher{
			sessionFunc: func(_ context.Context, _ neo4j.SessionConfig) graphSession {
				return &mockGraphSession{
					runFunc: func(_ context.Context, _ string, params map[string]any) (graphResult, error) {
						if params["id"] != int64(7) {
							t.Errorf("expected id 7, got %v", params["id"])
						}
						return &mockGraphResult{
							records: []*db.Record{recordVillage(3, "home", 50, 60)},
						}, nil
					},
				}
			},
		})
		v, err := repo.GetVillageByNpc(context.Background(), NpcNode{ID: 7})
		if err != nil {
			t.Fatal(err)
		}
		if v == nil {
			t.Fatal("expected non-nil village")
		}
		if v.ID != 3 || v.Name != "home" || v.X != 50 || v.Z != 60 {
			t.Errorf("unexpected village: %+v", v)
		}
	})

	t.Run("not found", func(t *testing.T) {
		repo := newRepoWithMocks(&mockGrapher{
			sessionFunc: func(_ context.Context, _ neo4j.SessionConfig) graphSession {
				return &mockGraphSession{
					runFunc: func(_ context.Context, _ string, _ map[string]any) (graphResult, error) {
						return &mockGraphResult{records: nil}, nil
					},
				}
			},
		})
		v, err := repo.GetVillageByNpc(context.Background(), NpcNode{ID: 7})
		if err != nil {
			t.Fatal(err)
		}
		if v != nil {
			t.Fatal("expected nil")
		}
	})

	t.Run("run error", func(t *testing.T) {
		repo := newRepoWithMocks(&mockGrapher{
			sessionFunc: func(_ context.Context, _ neo4j.SessionConfig) graphSession {
				return &mockGraphSession{
					runFunc: func(_ context.Context, _ string, _ map[string]any) (graphResult, error) {
						return nil, errors.New("err")
					},
				}
			},
		})
		_, err := repo.GetVillageByNpc(context.Background(), NpcNode{ID: 7})
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestGetConnectedRel(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		repo := newRepoWithMocks(&mockGrapher{
			sessionFunc: func(_ context.Context, _ neo4j.SessionConfig) graphSession {
				return &mockGraphSession{
					runFunc: func(_ context.Context, _ string, params map[string]any) (graphResult, error) {
						if params["sourceId"] != int64(1) || params["targetId"] != int64(2) {
							t.Errorf("unexpected params: %v", params)
						}
						return &mockGraphResult{
							records: []*db.Record{recordDirection("north")},
						}, nil
					},
				}
			},
		})
		rel, err := repo.GetConnectedRel(context.Background(), VillageNode{ID: 1}, VillageNode{ID: 2})
		if err != nil {
			t.Fatal(err)
		}
		if rel == nil {
			t.Fatal("expected non-nil rel")
		}
		if rel.Direction != North {
			t.Errorf("expected north, got %s", rel.Direction)
		}
	})

	t.Run("not found", func(t *testing.T) {
		repo := newRepoWithMocks(&mockGrapher{
			sessionFunc: func(_ context.Context, _ neo4j.SessionConfig) graphSession {
				return &mockGraphSession{
					runFunc: func(_ context.Context, _ string, _ map[string]any) (graphResult, error) {
						return &mockGraphResult{records: nil}, nil
					},
				}
			},
		})
		rel, err := repo.GetConnectedRel(context.Background(), VillageNode{ID: 1}, VillageNode{ID: 2})
		if err != nil {
			t.Fatal(err)
		}
		if rel != nil {
			t.Fatal("expected nil rel")
		}
	})
}

func TestQueryAllVillages(t *testing.T) {
	t.Run("returns all", func(t *testing.T) {
		repo := newRepoWithMocks(&mockGrapher{
			sessionFunc: func(_ context.Context, _ neo4j.SessionConfig) graphSession {
				return &mockGraphSession{
					runFunc: func(_ context.Context, _ string, _ map[string]any) (graphResult, error) {
						return &mockGraphResult{
							records: []*db.Record{
								recordVillage(1, "a", 0, 0),
								recordVillage(2, "b", 10, 20),
							},
						}, nil
					},
				}
			},
		})
		vs, err := repo.QueryAllVillages(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if len(vs) != 2 {
			t.Fatalf("expected 2, got %d", len(vs))
		}
	})

	t.Run("empty", func(t *testing.T) {
		repo := newRepoWithMocks(&mockGrapher{
			sessionFunc: func(_ context.Context, _ neo4j.SessionConfig) graphSession {
				return &mockGraphSession{
					runFunc: func(_ context.Context, _ string, _ map[string]any) (graphResult, error) {
						return &mockGraphResult{records: nil}, nil
					},
				}
			},
		})
		vs, err := repo.QueryAllVillages(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if len(vs) != 0 {
			t.Fatalf("expected 0, got %d", len(vs))
		}
	})

	t.Run("run error", func(t *testing.T) {
		repo := newRepoWithMocks(&mockGrapher{
			sessionFunc: func(_ context.Context, _ neo4j.SessionConfig) graphSession {
				return &mockGraphSession{
					runFunc: func(_ context.Context, _ string, _ map[string]any) (graphResult, error) {
						return nil, errors.New("err")
					},
				}
			},
		})
		_, err := repo.QueryAllVillages(context.Background())
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestQueryPlayerNode(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		repo := newRepoWithMocks(&mockGrapher{
			sessionFunc: func(_ context.Context, _ neo4j.SessionConfig) graphSession {
				return &mockGraphSession{
					runFunc: func(_ context.Context, _ string, _ map[string]any) (graphResult, error) {
						return &mockGraphResult{
							records: []*db.Record{recordPlayer(100, 200)},
						}, nil
					},
				}
			},
		})
		p, err := repo.QueryPlayerNode(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if p == nil {
			t.Fatal("expected non-nil player")
		}
		if p.X != 100 || p.Z != 200 {
			t.Errorf("expected (100,200), got (%d,%d)", p.X, p.Z)
		}
	})

	t.Run("not found returns nil", func(t *testing.T) {
		repo := newRepoWithMocks(&mockGrapher{
			sessionFunc: func(_ context.Context, _ neo4j.SessionConfig) graphSession {
				return &mockGraphSession{
					runFunc: func(_ context.Context, _ string, _ map[string]any) (graphResult, error) {
						return &mockGraphResult{records: nil}, nil
					},
				}
			},
		})
		p, err := repo.QueryPlayerNode(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if p != nil {
			t.Fatal("expected nil player")
		}
	})

	t.Run("run error", func(t *testing.T) {
		repo := newRepoWithMocks(&mockGrapher{
			sessionFunc: func(_ context.Context, _ neo4j.SessionConfig) graphSession {
				return &mockGraphSession{
					runFunc: func(_ context.Context, _ string, _ map[string]any) (graphResult, error) {
						return nil, errors.New("err")
					},
				}
			},
		})
		_, err := repo.QueryPlayerNode(context.Background())
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

// ---------------------------------------------------------------------------
// Tests — Write operations
// ---------------------------------------------------------------------------

func TestCreateNpcNodeByVillage(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo := newRepoWithMocks(&mockGrapher{
			sessionFunc: func(_ context.Context, _ neo4j.SessionConfig) graphSession {
				return &mockGraphSession{
					runFunc: func(_ context.Context, _ string, params map[string]any) (graphResult, error) {
						if params["villageId"] != int64(5) {
							t.Errorf("expected villageId 5, got %v", params["villageId"])
						}
						if params["npcId"] != int64(100) {
							t.Errorf("expected npcId 100, got %v", params["npcId"])
						}
						return &mockGraphResult{
							summary: &mockResultSummary{counters: &mockCounters{nodesCreated: 1}},
						}, nil
					},
				}
			},
		})
		err := repo.CreateNpcNodeByVillage(context.Background(), &NpcNode{ID: 100, Name: "bob"}, 5)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("nil node", func(t *testing.T) {
		repo := newRepoWithMocks(&mockGrapher{
			sessionFunc: func(_ context.Context, _ neo4j.SessionConfig) graphSession {
				return &mockGraphSession{}
			},
		})
		err := repo.CreatePlayerNode(context.Background(), nil)
		if err == nil {
			t.Fatal("expected error for nil node")
		}
	})

	t.Run("run error", func(t *testing.T) {
		repo := newRepoWithMocks(&mockGrapher{
			sessionFunc: func(_ context.Context, _ neo4j.SessionConfig) graphSession {
				return &mockGraphSession{
					runFunc: func(_ context.Context, _ string, _ map[string]any) (graphResult, error) {
						return nil, errors.New("err")
					},
				}
			},
		})
		err := repo.CreateNpcNodeByVillage(context.Background(), &NpcNode{ID: 1, Name: "x"}, 5)
		if err == nil {
			t.Fatal("expected error")
		}
	})

}

func TestCreateVillageNode(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo := newRepoWithMocks(&mockGrapher{
			sessionFunc: func(_ context.Context, _ neo4j.SessionConfig) graphSession {
				return &mockGraphSession{
					runFunc: func(_ context.Context, _ string, params map[string]any) (graphResult, error) {
						if params["id"] != int64(1) || params["name"] != "test" {
							t.Errorf("unexpected params: %v", params)
						}
						return &mockGraphResult{
							summary: &mockResultSummary{counters: &mockCounters{nodesCreated: 1}},
						}, nil
					},
				}
			},
		})
		err := repo.CreateVillageNode(context.Background(), &VillageNode{ID: 1, Name: "test", X: 10, Z: 20})
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("nil node", func(t *testing.T) {
		repo := newRepoWithMocks(&mockGrapher{})
		err := repo.CreateVillageNode(context.Background(), nil)
		if err == nil {
			t.Fatal("expected error for nil node")
		}
	})

	t.Run("run error", func(t *testing.T) {
		repo := newRepoWithMocks(&mockGrapher{
			sessionFunc: func(_ context.Context, _ neo4j.SessionConfig) graphSession {
				return &mockGraphSession{
					runFunc: func(_ context.Context, _ string, _ map[string]any) (graphResult, error) {
						return nil, errors.New("err")
					},
				}
			},
		})
		err := repo.CreateVillageNode(context.Background(), &VillageNode{ID: 1, Name: "t"})
		if err == nil {
			t.Fatal("expected error")
		}
	})

}

func TestCreatePlayerNode(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo := newRepoWithMocks(&mockGrapher{
			sessionFunc: func(_ context.Context, _ neo4j.SessionConfig) graphSession {
				return &mockGraphSession{
					runFunc: func(_ context.Context, _ string, params map[string]any) (graphResult, error) {
						if params["x"] != int64(50) || params["z"] != int64(100) {
							t.Errorf("unexpected params: %v", params)
						}
						return &mockGraphResult{}, nil
					},
				}
			},
		})
		err := repo.CreatePlayerNode(context.Background(), &PlayerNode{X: 50, Z: 100})
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("nil node", func(t *testing.T) {
		repo := newRepoWithMocks(&mockGrapher{
			sessionFunc: func(_ context.Context, _ neo4j.SessionConfig) graphSession {
				return &mockGraphSession{}
			},
		})
		err := repo.CreatePlayerNode(context.Background(), nil)
		if err == nil {
			t.Fatal("expected error for nil node")
		}
	})

	t.Run("run error", func(t *testing.T) {
		repo := newRepoWithMocks(&mockGrapher{
			sessionFunc: func(_ context.Context, _ neo4j.SessionConfig) graphSession {
				return &mockGraphSession{
					runFunc: func(_ context.Context, _ string, _ map[string]any) (graphResult, error) {
						return nil, errors.New("err")
					},
				}
			},
		})
		err := repo.CreatePlayerNode(context.Background(), &PlayerNode{X: 0, Z: 0})
		if err == nil {
			t.Fatal("expected error")
		}
	})

}

func TestCreateVillageToVillageRel(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo := newRepoWithMocks(&mockGrapher{
			sessionFunc: func(_ context.Context, _ neo4j.SessionConfig) graphSession {
				return &mockGraphSession{
					runFunc: func(_ context.Context, _ string, params map[string]any) (graphResult, error) {
						if params["sid"] != int64(1) || params["tid"] != int64(2) {
							t.Errorf("unexpected params: %v", params)
						}
						return &mockGraphResult{
							summary: &mockResultSummary{counters: &mockCounters{relationshipsCreated: 1}},
						}, nil
					},
				}
			},
		})
		err := repo.CreateVillageToVillageRel(context.Background(), 1, 2, North)
		if err != nil {
			t.Fatal(err)
		}
	})


	t.Run("run error", func(t *testing.T) {
		repo := newRepoWithMocks(&mockGrapher{
			sessionFunc: func(_ context.Context, _ neo4j.SessionConfig) graphSession {
				return &mockGraphSession{
					runFunc: func(_ context.Context, _ string, _ map[string]any) (graphResult, error) {
						return nil, errors.New("err")
					},
				}
			},
		})
		err := repo.CreateVillageToVillageRel(context.Background(), 1, 2, South)
		if err == nil {
			t.Fatal("expected error")
		}
	})

}

func TestCreatePlayerToVillageRel(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo := newRepoWithMocks(&mockGrapher{
			sessionFunc: func(_ context.Context, _ neo4j.SessionConfig) graphSession {
				return &mockGraphSession{
					runFunc: func(_ context.Context, _ string, params map[string]any) (graphResult, error) {
						if params["vid"] != int64(10) {
							t.Errorf("expected vid 10, got %v", params["vid"])
						}
						return &mockGraphResult{
							summary: &mockResultSummary{counters: &mockCounters{relationshipsCreated: 1}},
						}, nil
					},
				}
			},
		})
		err := repo.CreatePlayerToVillageRel(context.Background(), 10, Northeast)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("not found", func(t *testing.T) {
		repo := newRepoWithMocks(&mockGrapher{
			sessionFunc: func(_ context.Context, _ neo4j.SessionConfig) graphSession {
				return &mockGraphSession{
					runFunc: func(_ context.Context, _ string, _ map[string]any) (graphResult, error) {
						return &mockGraphResult{
							summary: &mockResultSummary{counters: &mockCounters{relationshipsCreated: 0}},
						}, nil
					},
				}
			},
		})
		err := repo.CreatePlayerToVillageRel(context.Background(), 999, North)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("run error", func(t *testing.T) {
		repo := newRepoWithMocks(&mockGrapher{
			sessionFunc: func(_ context.Context, _ neo4j.SessionConfig) graphSession {
				return &mockGraphSession{
					runFunc: func(_ context.Context, _ string, _ map[string]any) (graphResult, error) {
						return nil, errors.New("err")
					},
				}
			},
		})
		err := repo.CreatePlayerToVillageRel(context.Background(), 1, North)
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestCreateVillageToPlayerRel(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo := newRepoWithMocks(&mockGrapher{
			sessionFunc: func(_ context.Context, _ neo4j.SessionConfig) graphSession {
				return &mockGraphSession{
					runFunc: func(_ context.Context, _ string, params map[string]any) (graphResult, error) {
						if params["vid"] != int64(5) {
							t.Errorf("expected vid 5, got %v", params["vid"])
						}
						return &mockGraphResult{
							summary: &mockResultSummary{counters: &mockCounters{relationshipsCreated: 1}},
						}, nil
					},
				}
			},
		})
		err := repo.CreateVillageToPlayerRel(context.Background(), 5, Southwest)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("not found", func(t *testing.T) {
		repo := newRepoWithMocks(&mockGrapher{
			sessionFunc: func(_ context.Context, _ neo4j.SessionConfig) graphSession {
				return &mockGraphSession{
					runFunc: func(_ context.Context, _ string, _ map[string]any) (graphResult, error) {
						return &mockGraphResult{
							summary: &mockResultSummary{counters: &mockCounters{relationshipsCreated: 0}},
						}, nil
					},
				}
			},
		})
		err := repo.CreateVillageToPlayerRel(context.Background(), 999, North)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("run error", func(t *testing.T) {
		repo := newRepoWithMocks(&mockGrapher{
			sessionFunc: func(_ context.Context, _ neo4j.SessionConfig) graphSession {
				return &mockGraphSession{
					runFunc: func(_ context.Context, _ string, _ map[string]any) (graphResult, error) {
						return nil, errors.New("err")
					},
				}
			},
		})
		err := repo.CreateVillageToPlayerRel(context.Background(), 1, North)
		if err == nil {
			t.Fatal("expected error")
		}
	})
}
