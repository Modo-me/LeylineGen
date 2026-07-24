package world_graph

import "context"

func (r *Repository) CreateNpc(ctx context.Context, npc *Npc) (int64, error) {
	result := r.db.WithContext(ctx).Create(npc)
	if result.Error != nil {
		return 0, result.Error
	}
	return npc.ID, nil
}

func (r *Repository) QueryNpcByID(ctx context.Context, id int64) (*Npc, error) {
	var npc Npc
	result := r.db.WithContext(ctx).First(&npc, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &npc, nil
}

func (r *Repository) UpdateNpc(ctx context.Context, npc *Npc) error {
	return r.db.WithContext(ctx).Save(npc).Error
}

func (r *Repository) DeleteNpc(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&Npc{}, id).Error
}

func (r *Repository) CreateStep(ctx context.Context, step *Step) error {
	return r.db.WithContext(ctx).Create(step).Error
}

func (r *Repository) QueryStepByID(ctx context.Context, id int64) (*Step, error) {
	var step Step
	result := r.db.WithContext(ctx).Preload("Npc").First(&step, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &step, nil
}

func (r *Repository) UpdateStep(ctx context.Context, step *Step) error {
	return r.db.WithContext(ctx).Save(step).Error
}

func (r *Repository) DeleteStep(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&Step{}, id).Error
}

func (r *Repository) CreateQuest(ctx context.Context, quest *Quest) error {
	return r.db.WithContext(ctx).Create(quest).Error
}

func (r *Repository) QueryQuest(ctx context.Context, id int64) (*Quest, error) {
	var quest Quest
	result := r.db.WithContext(ctx).Preload("Steps.Npc").First(&quest, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &quest, nil
}

func (r *Repository) UpdateQuest(ctx context.Context, quest *Quest) error {
	return r.db.WithContext(ctx).Save(quest).Error
}

func (r *Repository) DeleteQuest(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&Quest{}, id).Error
}

func (r *Repository) CreateVillage(ctx context.Context, village *Village) error {
	return r.db.WithContext(ctx).Create(village).Error
}
