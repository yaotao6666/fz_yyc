package category

import (
	"fz_yyc_api/internal/models"
	"time"

	"gorm.io/gorm"
)

// Node 分类树节点（JSON 就绪，用于管理端树形展示与级联选择）
type Node struct {
	ID           uint64    `json:"id"`
	Name         string    `json:"name"`
	ParentID     *uint64   `json:"parent_id"`
	Level        uint8     `json:"level"`
	Sort         uint      `json:"sort"`
	Status       uint8     `json:"status"`
	ProductCount int64     `json:"product_count"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Children     []*Node   `json:"children,omitempty"`
}

// CollectSubtreeIDs 返回 rootID 及其所有直接/间接子孙分类的 id 集合（平铺去重）。
// 模型最多三级，这里做两层下钻再加一层兜底循环保护，避免异常数据死循环。
func CollectSubtreeIDs(db *gorm.DB, rootID uint64) ([]uint64, error) {
	ids := []uint64{rootID}
	frontier := []uint64{rootID}
	for i := 0; i < 4 && len(frontier) > 0; i++ {
		var children []uint64
		if err := db.Model(&models.Category{}).
			Where("parent_id IN ?", frontier).
			Pluck("id", &children).Error; err != nil {
			return nil, err
		}
		if len(children) == 0 {
			break
		}
		ids = append(ids, children...)
		frontier = children
	}
	return ids, nil
}

// SubtreeProductCount 统计 rootID 及其全部子孙分类下直接挂载的商品数量（上架未删除）。
func SubtreeProductCount(db *gorm.DB, rootID uint64) (int64, error) {
	ids, err := CollectSubtreeIDs(db, rootID)
	if err != nil {
		return 0, err
	}
	var count int64
	err = db.Model(&models.Product{}).
		Where("category_id IN ? AND deleted_at IS NULL", ids).
		Count(&count).Error
	return count, err
}

// BuildTree 从全表构建以 simpled 根节点（parent_id IS NULL）为入口的分类树。
// 每个节点的 ProductCount 为该分类自身+全部子孙分类直接挂载的商品数。
func BuildTree(db *gorm.DB) ([]*Node, error) {
	var all []models.Category
	if err := db.Order("sort ASC, id ASC").Find(&all).Error; err != nil {
		return nil, err
	}

	children := map[uint64][]*models.Category{}
	var roots []*models.Category
	for i := range all {
		n := &all[i]
		if n.ParentID == nil {
			roots = append(roots, n)
		} else {
			pid := *n.ParentID
			children[pid] = append(children[pid], n)
		}
	}

	var build func(*models.Category) (*Node, error)
	build = func(n *models.Category) (*Node, error) {
		node := &Node{
			ID:        n.ID,
			Name:      n.Name,
			ParentID:  n.ParentID,
			Level:     n.Level,
			Sort:      n.Sort,
			Status:    n.Status,
			CreatedAt: n.CreatedAt,
			UpdatedAt: n.UpdatedAt,
		}
		count, err := SubtreeProductCount(db, n.ID)
		if err != nil {
			return nil, err
		}
		node.ProductCount = count
		for _, ch := range children[n.ID] {
			childNode, err := build(ch)
			if err != nil {
				return nil, err
			}
			node.Children = append(node.Children, childNode)
		}
		return node, nil
	}

	var result []*Node
	for _, r := range roots {
		node, err := build(r)
		if err != nil {
			return nil, err
		}
		result = append(result, node)
	}
	return result, nil
}