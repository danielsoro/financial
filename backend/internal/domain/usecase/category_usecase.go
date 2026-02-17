package usecase

import (
	"context"

	"github.com/dcunha/finance/backend/internal/domain"
	"github.com/dcunha/finance/backend/internal/domain/entity"
	"github.com/dcunha/finance/backend/internal/domain/repository"
	"github.com/google/uuid"
)

type CategoryUsecase struct {
	categoryRepo repository.CategoryRepository
}

func NewCategoryUsecase(repo repository.CategoryRepository) *CategoryUsecase {
	return &CategoryUsecase{categoryRepo: repo}
}

func (uc *CategoryUsecase) List(ctx context.Context, userID uuid.UUID, catType string) ([]entity.Category, error) {
	return uc.categoryRepo.FindAllForUser(ctx, userID, catType)
}

func (uc *CategoryUsecase) ListTree(ctx context.Context, userID uuid.UUID, catType string) ([]entity.Category, error) {
	cats, err := uc.categoryRepo.FindAllForUser(ctx, userID, catType)
	if err != nil {
		return nil, err
	}
	return buildTree(cats), nil
}

func (uc *CategoryUsecase) Create(ctx context.Context, userID uuid.UUID, name, catType string, parentID *uuid.UUID) (*entity.Category, error) {
	if parentID != nil {
		parent, err := uc.categoryRepo.FindByID(ctx, *parentID)
		if err != nil {
			return nil, err
		}
		// Parent must be accessible by the user (default or owned)
		if parent.UserID != nil && *parent.UserID != userID {
			return nil, domain.ErrForbidden
		}
		// Subcategory inherits type from parent
		catType = parent.Type
	}

	cat := &entity.Category{
		UserID:    &userID,
		ParentID:  parentID,
		Name:      name,
		Type:      catType,
		IsDefault: false,
	}
	if err := uc.categoryRepo.Create(ctx, cat); err != nil {
		return nil, err
	}
	return cat, nil
}

func (uc *CategoryUsecase) Update(ctx context.Context, userID uuid.UUID, id uuid.UUID, name, catType string, parentID *uuid.UUID) (*entity.Category, error) {
	cat, err := uc.categoryRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if cat.IsDefault || cat.UserID == nil || *cat.UserID != userID {
		return nil, domain.ErrForbidden
	}

	// If parent is being changed, validate
	if parentID != nil {
		// Cannot set parent to self
		if *parentID == id {
			return nil, domain.ErrCyclicCategory
		}
		// Validate parent exists and is accessible
		parent, err := uc.categoryRepo.FindByID(ctx, *parentID)
		if err != nil {
			return nil, err
		}
		if parent.UserID != nil && *parent.UserID != userID {
			return nil, domain.ErrForbidden
		}
		// Check for cycles: walk up from the proposed parent to make sure we don't reach `id`
		if err := uc.checkCycle(ctx, *parentID, id); err != nil {
			return nil, err
		}
	}

	cat.Name = name
	cat.Type = catType
	cat.ParentID = parentID
	if err := uc.categoryRepo.Update(ctx, cat); err != nil {
		return nil, err
	}
	return cat, nil
}

// checkCycle walks up the ancestor chain from startID and returns ErrCyclicCategory if it reaches targetID.
func (uc *CategoryUsecase) checkCycle(ctx context.Context, startID, targetID uuid.UUID) error {
	currentID := startID
	for {
		cat, err := uc.categoryRepo.FindByID(ctx, currentID)
		if err != nil {
			return err
		}
		if cat.ParentID == nil {
			return nil // reached root, no cycle
		}
		if *cat.ParentID == targetID {
			return domain.ErrCyclicCategory
		}
		currentID = *cat.ParentID
	}
}

func (uc *CategoryUsecase) Delete(ctx context.Context, userID uuid.UUID, id uuid.UUID) error {
	cat, err := uc.categoryRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if cat.IsDefault || cat.UserID == nil || *cat.UserID != userID {
		return domain.ErrForbidden
	}
	inUse, err := uc.categoryRepo.IsSubtreeInUse(ctx, id)
	if err != nil {
		return err
	}
	if inUse {
		return domain.ErrCategoryInUse
	}
	return uc.categoryRepo.Delete(ctx, id)
}

// buildTree converts a flat sorted list into a tree structure.
func buildTree(cats []entity.Category) []entity.Category {
	catMap := make(map[uuid.UUID]*entity.Category, len(cats))
	var roots []entity.Category

	// First pass: put all in map
	for i := range cats {
		cats[i].Children = []entity.Category{}
		catMap[cats[i].ID] = &cats[i]
	}

	// Second pass: build parent-child relationships
	for i := range cats {
		if cats[i].ParentID != nil {
			if parent, ok := catMap[*cats[i].ParentID]; ok {
				parent.Children = append(parent.Children, cats[i])
			}
		} else {
			roots = append(roots, cats[i])
		}
	}

	// Recursively set children from map (since we modified pointers)
	var setChildren func(cats []entity.Category) []entity.Category
	setChildren = func(cats []entity.Category) []entity.Category {
		for i := range cats {
			if mapped, ok := catMap[cats[i].ID]; ok {
				cats[i].Children = setChildren(mapped.Children)
			}
		}
		return cats
	}

	return setChildren(roots)
}
