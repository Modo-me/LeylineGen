package world_graph

import "context"

// ObservePlayerNearbyWorld returns the villages connected to the player node.
func (s *Service) ObservePlayerNearbyWorld(ctx context.Context) (*NearbyWorldInfo, error) {
	villages, err := s.Repository.GetPlayerConnectedVillages(ctx)
	if err != nil {
		return nil, err
	}

	infos := make([]VillageInfo, 0, len(villages))
	for _, v := range villages {
		infos = append(infos, VillageInfo{
			ID:   v.ID,
			Name: v.Name,
		})
	}
	return &NearbyWorldInfo{VillagesInfo: infos}, nil
}

// ObserveVillageNearbyWorld returns the villages connected to the given
// village, each with its compass direction from the center.
func (s *Service) ObserveVillageNearbyWorld(ctx context.Context, villageID int64) (*NearbyWorldInfo, error) {
	center := VillageNode{ID: villageID}

	connected, err := s.Repository.GetConnectedVillages(ctx, center)
	if err != nil {
		return nil, err
	}

	infos := make([]VillageInfo, 0, len(connected))
	for _, v := range connected {
		rel, err := s.Repository.GetConnectedRel(ctx, center, v)
		if err != nil {
			return nil, err
		}

		dir := ""
		if rel != nil {
			dir = string(rel.Direction)
		}

		infos = append(infos, VillageInfo{
			ID:        v.ID,
			Name:      v.Name,
			Direction: dir,
		})
	}
	return &NearbyWorldInfo{VillagesInfo: infos}, nil
}

func (s *Service) CreateNewNpc(ctx context.Context, npcName string, villageID int64) (int64, error) {
	npc := &Npc{}
	ID, err := s.Repository.CreateNpc(ctx, npc)
	npcNode := &NpcNode{
		ID:   ID,
		Name: npcName,
	}
	err = s.Repository.CreateNpcNodeByVillage(ctx, npcNode, villageID)
	if err != nil {
		return 0, err
	}
	return ID, nil
}

func (s *Service) CreateStepWithNpc(ctx context.Context, npcID int64, dialogueLines []string) error {
	step := &Step{
		NpcID:         npcID,
		DialogueLines: dialogueLines,
		Npc:           Npc{ID: npcID},
	}
	err := s.Repository.CreateStep(ctx, step)
	if err != nil {
		return err
	}
	steps = append(steps, *step)
	return nil
}
