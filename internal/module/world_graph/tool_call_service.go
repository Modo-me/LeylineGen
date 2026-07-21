package world_graph

import "context"

// ObservePlayerNearbyWorld returns the villages connected to the player node.
func (s *service) ObservePlayerNearbyWorld(ctx context.Context) (*NearbyWorldInfo, error) {
	villages, err := s.Repository.GetPlayerConnectedVillages(ctx)
	if err != nil {
		return nil, err
	}

	infos := make([]villageInfo, 0, len(villages))
	for _, v := range villages {
		infos = append(infos, villageInfo{
			ID:   v.ID,
			Name: v.Name,
		})
	}
	return &NearbyWorldInfo{VillagesInfo: infos}, nil
}

// ObserveVillageNearbyWorld returns the villages connected to the given
// village, each with its compass direction from the center.
func (s *service) ObserveVillageNearbyWorld(ctx context.Context, villageID int64) (*NearbyWorldInfo, error) {
	center := VillageNode{ID: villageID}

	connected, err := s.Repository.GetConnectedVillages(ctx, center)
	if err != nil {
		return nil, err
	}

	infos := make([]villageInfo, 0, len(connected))
	for _, v := range connected {
		rel, err := s.Repository.GetConnectedRel(ctx, center, v)
		if err != nil {
			return nil, err
		}

		dir := ""
		if rel != nil {
			dir = string(rel.Direction)
		}

		infos = append(infos, villageInfo{
			ID:        v.ID,
			Name:      v.Name,
			Direction: dir,
		})
	}
	return &NearbyWorldInfo{VillagesInfo: infos}, nil
}

func (s *service) CreateNewNpc(ctx context.Context, npcName string, VillageID int64) error {
	npc := &Npc{}
	ID, err := s.Repository.CreateNpc(ctx, npc)
	npcNode := &NpcNode{
		ID:   ID,
		Name: npcName,
	}
	err = s.Repository.CreateNpcNodeByVillage(ctx, npcNode, VillageID)
	return err
}
