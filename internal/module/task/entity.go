package task

type Task struct {
	ID        uint   `gorm:"primaryKey" json:"id"`
	WorldName string `gorm:"type:varchar(200)" json:"world_name"`
	WorldDesc string `gorm:"type:text" json:"world_desc"`
	Emotion   string `gorm:"type:varchar(100)" json:"emotion"`
	State     string `gorm:"type:varchar(20)" json:"state"`
}
