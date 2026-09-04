package template

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/kouleen/portable-chat/pkg/model"
	"gorm.io/gorm"
)

// SegmentAlloc 分布式号段分配器
type SegmentAlloc struct {
	db     *gorm.DB
	bizTag string

	mu         sync.Mutex
	current    int64 // 当前待发放
	segmentEnd int64 // 当前号段结束(不含)

	startID int64
	endID   int64
}

// NewSegmentAlloc 创建分配器并预拉取第一段号段
func NewSegmentAlloc(ctx context.Context, db *gorm.DB, bizTag string) (*SegmentAlloc, error) {
	var rec model.IdAlloc
	if err := db.WithContext(ctx).Where("biz_tag = ?", bizTag).Take(&rec).Error; err != nil {
		return nil, fmt.Errorf("load id_alloc record fail: %w", err)
	}
	s := &SegmentAlloc{
		db:      db,
		bizTag:  bizTag,
		startID: rec.StartID,
		endID:   rec.EndID,
	}
	if err := s.fetchSegment(ctx); err != nil {
		return nil, fmt.Errorf("first fetch segment fail: %w", err)
	}
	return s, nil
}

// fetchSegment 事务+行锁原子抢占一段号段
func (s *SegmentAlloc) fetchSegment(ctx context.Context) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var rec model.IdAlloc
		if err := tx.Set("gorm:query_option", "FOR UPDATE").
			Where("biz_tag = ?", s.bizTag).Take(&rec).Error; err != nil {
			return err
		}

		oldMax := rec.MaxID
		if oldMax >= s.endID {
			return errors.New("id pool exhausted")
		}
		newMax := oldMax + int64(rec.Step)
		if newMax > s.endID {
			newMax = s.endID
		}
		if err := tx.Model(&model.IdAlloc{}).
			Where("biz_tag = ?", s.bizTag).
			Update("max_id", newMax).Error; err != nil {
			return err
		}
		s.mu.Lock()
		s.current = oldMax
		s.segmentEnd = newMax
		s.mu.Unlock()
		return nil
	})
}

// Next 获取下一个全局唯一内部有序ID
func (s *SegmentAlloc) Next(ctx context.Context) (int64, error) {
	s.mu.Lock()
	if s.current >= s.segmentEnd {
		s.mu.Unlock()
		if err := s.fetchSegment(ctx); err != nil {
			return 0, err
		}
		s.mu.Lock()
	}
	id := s.current
	s.current++
	s.mu.Unlock()

	if id < s.startID || id > s.endID {
		return 0, errors.New("id out of range")
	}
	return id, nil
}

// NextUIN 直接获取对外乱序9位UIN(内部seq + Obfuscate)
func (s *SegmentAlloc) NextUIN(ctx context.Context) (int64, error) {
	seq, err := s.Next(ctx)
	if err != nil {
		return 0, err
	}
	return Obfuscate(seq)
}
