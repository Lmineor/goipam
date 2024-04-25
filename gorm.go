package ipam

import (
	"context"
	"fmt"
	"gorm.io/gorm"
)

type gormImpl struct {
	db          *gorm.DB
	maxIdLength int
}

type Namespace struct {
	Namespace string `gorm:"namespace;uniqueIndex:namespace_idx"`
}

func (g *gormImpl) prefixExists(ctx context.Context, prefix Prefix) (*Prefix, bool) {
	p, err := g.ReadPrefix(ctx, prefix.Cidr)
	if err != nil {
		return nil, false
	}
	return &p, true
}

func (g *gormImpl) checkNamespaceExists(namespace string) bool {
	if err := g.db.Where("namespace = ?", namespace).First(&Namespace{}).Error; err != nil {
		return false
	}
	return true
}

func (g *gormImpl) CreatePrefix(ctx context.Context, prefix Prefix) (Prefix, error) {
	namespace := namespaceFromContext(ctx)
	if !g.checkNamespaceExists(namespace) {
		return Prefix{}, ErrNamespaceDoesNotExist
	}
	existingPrefix, exists := g.prefixExists(ctx, prefix)
	if exists {
		return *existingPrefix.deepCopy(), nil
	}
	prefix.Version = int64(0)

	if prefix.Namespace != "" && prefix.Namespace != namespace {
		return Prefix{}, fmt.Errorf("prefix namespace inconsistent with context")
	}
	prefix.Namespace = namespace
	err := g.db.Create(&prefix).Error
	return *prefix.deepCopy(), err
}
func (g *gormImpl) ReadPrefixByID(_ context.Context, id uint) (Prefix, error) {
	var p Prefix
	err := g.db.Where("id = ?", id).First(&p).Error
	if err != nil {
		return p, fmt.Errorf("unable to read prefix:%w", err)
	}
	return *p.deepCopy(), nil
}
func (g *gormImpl) ReadAllChildPrefixByParentID(_ context.Context, id uint) (Prefixes, error) {
	var pixes Prefixes
	err := g.db.Find(&pixes, "parent_id = ?", id).Error
	if err != nil {
		return pixes, fmt.Errorf("unable to read prefix:%w", err)
	}
	ps := make([]Prefix, 0, len(pixes))
	for _, v := range pixes {
		ps = append(ps, *v.deepCopy())
	}
	return ps, nil
}

func (g *gormImpl) ReadPrefix(ctx context.Context, cidr string) (Prefix, error) {
	var p Prefix
	namespace := namespaceFromContext(ctx)
	if !g.checkNamespaceExists(namespace) {
		return p, ErrNamespaceDoesNotExist
	}
	err := g.db.Where("namespace = ? AND cidr = ?", namespace, cidr).First(&p).Error
	if err != nil {
		return p, fmt.Errorf("unable to read prefix:%w", err)
	}
	return *p.deepCopy(), nil
}

func (g *gormImpl) DeleteAllPrefixes(ctx context.Context) error {
	namespace := namespaceFromContext(ctx)
	if !g.checkNamespaceExists(namespace) {
		return ErrNamespaceDoesNotExist
	}
	return g.db.Delete(&Prefix{}, "namespace = ?", namespace).Error
}

// ReadAllPrefixes returns all known prefixes.
func (g *gormImpl) ReadAllPrefixes(ctx context.Context) (Prefixes, error) {
	namespace := namespaceFromContext(ctx)
	if !g.checkNamespaceExists(namespace) {
		return nil, ErrNamespaceDoesNotExist
	}
	var prefixes Prefixes

	err := g.db.Find(&prefixes, "namespace = ?", namespace).Error
	if err != nil {
		return nil, fmt.Errorf("unable to read prefixes:%w", err)
	}
	ps := make([]Prefix, 0, len(prefixes))
	for _, v := range prefixes {
		ps = append(ps, *v.deepCopy())
	}
	return ps, nil
}

// ReadAllPrefixCidrs is cheaper that ReadAllPrefixes because it only returns the Cidrs.
func (g *gormImpl) ReadAllPrefixCidrs(ctx context.Context) ([]string, error) {
	namespace := namespaceFromContext(ctx)
	if !g.checkNamespaceExists(namespace) {
		return nil, ErrNamespaceDoesNotExist
	}
	var cidrs []string
	err := g.db.Model(Prefix{}).Where("namespace = ?", namespace).Pluck("cidr", &cidrs).Error
	if err != nil {
		return nil, fmt.Errorf("unable to read prefixes in namespace:%s: %w", namespace, err)
	}
	return cidrs, nil
}

// UpdatePrefix tries to update the prefix.
// Returns OptimisticLockError if it does not succeed due to a concurrent update.
func (g *gormImpl) UpdatePrefix(ctx context.Context, prefix Prefix) (Prefix, error) {
	namespace := namespaceFromContext(ctx)
	if !g.checkNamespaceExists(namespace) {
		return Prefix{}, ErrNamespaceDoesNotExist
	}
	oldVersion := prefix.Version
	prefix.Version = oldVersion + 1
	err := g.db.Save(&prefix).Error
	if err != nil {
		return Prefix{}, err
	}

	return *prefix.deepCopy(), nil
}

func (g *gormImpl) DeletePrefix(ctx context.Context, prefix Prefix) (Prefix, error) {
	namespace := namespaceFromContext(ctx)
	if !g.checkNamespaceExists(namespace) {
		return Prefix{}, ErrNamespaceDoesNotExist
	}
	err := g.db.Delete(&Prefix{}, "cidr = ? AND namespace = ?", prefix.Cidr, namespace).Error
	if err != nil {
		return Prefix{}, fmt.Errorf("unable delete prefix: %w", err)
	}
	return *prefix.deepCopy(), nil
}
func (g *gormImpl) Name() string {
	return "gorm"
}

func (g *gormImpl) CreateNamespace(_ context.Context, namespace string) error {
	if len(namespace) > g.maxIdLength {
		return ErrNameTooLong
	}
	newNamespace := Namespace{
		Namespace: namespace,
	}
	return g.db.Create(&newNamespace).Error
}

func (g *gormImpl) ListNamespaces(_ context.Context) ([]string, error) {
	var result []string
	err := g.db.Model(&Namespace{}).Pluck("namespace", &result).Error
	return result, err
}

func (g *gormImpl) DeleteNamespace(_ context.Context, namespace string) error {
	if !g.checkNamespaceExists(namespace) {
		return ErrNamespaceDoesNotExist
	}
	return g.db.Transaction(func(tx *gorm.DB) error {
		// delete prefix first
		if err := tx.Delete(&Prefix{}, "namespace = ?", namespace).Error; err != nil {
			return err
		}
		if err := tx.Delete(&Namespace{}, "namespace = ?", namespace).Error; err != nil {
			return err
		}
		return nil
	})
}

func (g *gormImpl) DeleteIPAddress(_ context.Context, prefix Prefix, ip string) error {
	return g.db.Delete(IPStorage{}, "ip = ? and parent_prefix = ? AND namespace = ?", ip, prefix.Cidr, prefix.Namespace).Error
}

func (g *gormImpl) PutIPAddress(ctx context.Context, prefix Prefix, ip string) error {
	if g.IPAllocated(ctx, prefix, ip) {
		// already exist
		return fmt.Errorf("ip %s aleady exist", ip)
	}
	ipD := IPStorage{
		IP:           ip,
		ParentPrefix: prefix.Cidr,
		Namespace:    prefix.Namespace,
	}
	return g.db.Create(&ipD).Error
}

func (g *gormImpl) IPAllocated(ctx context.Context, prefix Prefix, ip string) bool {
	namespace := namespaceFromContext(ctx)
	if err := g.db.First(&IPStorage{}, "ip = ? AND parent_prefix = ? AND namespace = ?", ip, prefix.Cidr, namespace).Error; err == nil {
		// already exist
		return true
	}
	return false
}

func (g *gormImpl) AllocatedIPS(ctx context.Context, prefix Prefix) ([]IPStorage, error) {
	namespace := namespaceFromContext(ctx)
	var ips []IPStorage
	if err := g.db.Find(&ips, "parent_prefix = ? AND namespace = ?", prefix.Cidr, namespace).Error; err != nil {
		return ips, err
	}
	return ips, nil
}
