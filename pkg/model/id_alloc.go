package model

// IdAlloc DB号段表模型
type IdAlloc struct {
	ID      int64  `gorm:"column:id;primary_key"`
	BizTag  string `gorm:"column:biz_tag;uniqueIndex"` // 业务标识
	MaxID   int64  `gorm:"column:max_id"`              // 已分配到的最大ID
	Step    int32  `gorm:"column:step"`                // 每次申领步长
	StartID int64  `gorm:"column:start_id"`            // 起始ID
	EndID   int64  `gorm:"column:end_id"`              // 上限ID
}

func (IdAlloc) TableName() string {
	return "id_alloc"
}
