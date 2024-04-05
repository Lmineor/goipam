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
	NS string `gorm:"ns;uniqueIndex:user_title_idx"`
}

func (s *gormImpl) prefixExists(ctx context.Context, prefix Prefix) (*Prefix, bool) {
	p, err := s.ReadPrefix(ctx, prefix.Cidr)
	if err != nil {
		return nil, false
	}
	return &p, true
}

func (s *gormImpl) checkNamespaceExists(namespace string) error {
	return s.db.First(&Namespace{}).Where("namespace = ?", namespace).Error
}

func (s *gormImpl) CreatePrefix(ctx context.Context, prefix Prefix) (Prefix, error) {
	namespace := namespaceFromContext(ctx)
	if err := s.checkNamespaceExists(namespace); err != nil {
		return Prefix{}, err
	}
	existingPrefix, exists := s.prefixExists(ctx, prefix)
	if exists {
		return *existingPrefix, nil
	}
	prefix.version = int64(0)

	err := s.db.Create(prefix).Error
	return prefix, err
}

func (s *gormImpl) ReadPrefix(ctx context.Context, prefix string) (Prefix, error) {
	var p Prefix
	namespace := namespaceFromContext(ctx)
	if err := s.checkNamespaceExists(namespace); err != nil {
		return p, err
	}
	err := s.db.First(&prefix).Where("ns = ? AND prefix = ?", namespace, prefix).Error
	if err != nil {
		return p, fmt.Errorf("unable to read prefix:%w", err)
	}
	return p, nil
}

func (s *gormImpl) DeleteAllPrefixes(ctx context.Context) error {
	namespace := namespaceFromContext(ctx)
	if err := s.checkNamespaceExists(namespace); err != nil {
		return err
	}
	return s.db.Delete(&Prefix{}, "namespace = ?", namespace).Error
}

// ReadAllPrefixes returns all known prefixes.
func (s *gormImpl) ReadAllPrefixes(ctx context.Context) (Prefixes, error) {
	namespace := namespaceFromContext(ctx)
	if err := s.checkNamespaceExists(namespace); err != nil {
		return nil, err
	}
	var prefixes Prefixes
	err := s.db.Find(&prefixes, "namespace = ?", namespace).Error
	if err != nil {
		return nil, fmt.Errorf("unable to read prefixes:%w", err)
	}
	return prefixes, nil
}

// ReadAllPrefixCidrs is cheaper that ReadAllPrefixes because it only returns the Cidrs.
func (s *gormImpl) ReadAllPrefixCidrs(ctx context.Context) ([]string, error) {
	namespace := namespaceFromContext(ctx)
	if err := s.checkNamespaceExists(namespace); err != nil {
		return nil, err
	}
	var cidrs []string
	err := s.db.Model(Prefix{}).Where("namespace = ?", namespace).Pluck("cidr", &cidrs).Error
	if err != nil {
		return nil, fmt.Errorf("unable to read prefixes in namespace:%s :%w", namespace, err)
	}
	return cidrs, nil
}

// UpdatePrefix tries to update the prefix.
// Returns OptimisticLockError if it does not succeed due to a concurrent update.
func (s *gormImpl) UpdatePrefix(ctx context.Context, prefix Prefix) (Prefix, error) {
	namespace := namespaceFromContext(ctx)
	if err := s.checkNamespaceExists(namespace); err != nil {
		return Prefix{}, err
	}
	return Prefix{}, nil
	//oldVersion := prefix.version
	//prefix.version = oldVersion + 1
	//prefix
	//result, err = tx.ExecContext(ctx, "UPDATE "+getTableName(namespace)+" SET prefix=$1 WHERE cidr=$2 AND prefix->>'Version'=$3", pn, prefix.Cidr, oldVersion)
	//if err != nil {
	//	return Prefix{}, fmt.Errorf("%w: unable to update prefix:%s", ErrOptimisticLockError, prefix.Cidr)
	//}
	//rows, err = result.RowsAffected()
	//if err != nil {
	//	return Prefix{}, err
	//}
	//if rows == 0 {
	//	// Rollback, but ignore error, if rollback is omitted, the row lock created by SELECT FOR UPDATE will not get released.
	//	_ = tx.Rollback()
	//	return Prefix{}, fmt.Errorf("%w: updatePrefix did not effect any row", ErrOptimisticLockError)
	//}
	//return prefix, tx.Commit()
}

func (s *gormImpl) DeletePrefix(ctx context.Context, prefix Prefix) (Prefix, error) {
	namespace := namespaceFromContext(ctx)
	if err := s.checkNamespaceExists(namespace); err != nil {
		return Prefix{}, err
	}
	err := s.db.Delete(Prefix{}, "cidr = ? AND namespace = ?", prefix.Cidr, namespace).Error
	if err != nil {
		return Prefix{}, fmt.Errorf("unable delete prefix: %w", err)
	}
	return prefix, nil
}
func (s *gormImpl) Name() string {
	return "gorm"
}

func (s *gormImpl) CreateNamespace(ctx context.Context, namespace string) error {
	if len(namespace) > s.maxIdLength {
		return ErrNameTooLong
	}
	newNamespace := Namespace{
		NS: namespace,
	}
	return s.db.Create(&newNamespace).Error
}

func (s *gormImpl) ListNamespaces(ctx context.Context) ([]string, error) {
	var result []string
	err := s.db.Model(&Namespace{}).Pluck("ns", &result).Error
	return result, err
}

func (s *gormImpl) DeleteNamespace(ctx context.Context, namespace string) error {
	if err := s.checkNamespaceExists(namespace); err != nil {
		return err
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
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

func (s *gormImpl) DeleteIPAddress(ctx context.Context, prefix Prefix, ip string) error {
	return s.db.Delete(IPStorage{}, "ip = ? and parent_prefix = ? AND namespace = ?", ip, prefix.Cidr, prefix.Namespace).Error
}

func (s *gormImpl) PutIPAddress(ctx context.Context, prefix Prefix, ip string) error {
	if s.IPAllocated(ctx, prefix, ip) {
		// already exist
		return fmt.Errorf("ip %s aleady exist", ip)
	}
	ipD := IPStorage{
		IP:           ip,
		ParentPrefix: prefix.Cidr,
		Namespace:    prefix.Namespace,
	}
	return s.db.Create(&ipD).Error
}

func (s *gormImpl) IPAllocated(ctx context.Context, prefix Prefix, ip string) bool {
	namespace := namespaceFromContext(ctx)
	if err := s.db.First(&IPStorage{}, "ip = ? AND parent_prefix = ? AND namespace = ?", ip, prefix.Cidr, namespace).Error; err == nil {
		// already exist
		return true
	}
	return false
}

func (s *gormImpl) AllocatedIPS(ctx context.Context, prefix Prefix) ([]IPStorage, error) {
	namespace := namespaceFromContext(ctx)
	var ips []IPStorage
	if err := s.db.Find(&ips, "cidr = ? AND namespace = ?", prefix.Cidr, namespace).Error; err != nil {
		return nil, err
	}
	return ips, nil
}
