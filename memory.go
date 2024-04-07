package ipam

//
//import (
//	"context"
//	"fmt"
//	"sync"
//)
//
//type memory struct {
//	prefixes map[string]map[string]Prefix
//	lock     sync.RWMutex
//	IPs      map[string][]string
//}
//
//// NewMemory create a memory storage for ipam
//func NewMemory(ctx context.Context) Storage {
//	m := &memory{
//		prefixes: make(map[string]map[string]Prefix),
//		lock:     sync.RWMutex{},
//		IPs:      make(map[string][]string, 0),
//	}
//	namespace := namespaceFromContext(ctx)
//	_ = m.CreateNamespace(ctx, namespace)
//	return m
//}
//func (m *memory) Name() string {
//	return "memory"
//}
//func (m *memory) CreatePrefix(ctx context.Context, prefix Prefix) (Prefix, error) {
//	m.lock.Lock()
//	defer m.lock.Unlock()
//	namespace := namespaceFromContext(ctx)
//	if _, ok := m.prefixes[namespace]; !ok {
//		return Prefix{}, ErrNamespaceDoesNotExist
//	}
//	_, ok := m.prefixes[namespace][prefix.Cidr]
//	if ok {
//		return Prefix{}, fmt.Errorf("prefix already created:%v", prefix)
//	}
//	m.prefixes[namespace][prefix.Cidr] = *prefix.deepCopy()
//	return prefix, nil
//}
//func (m *memory) ReadPrefix(ctx context.Context, prefix string) (Prefix, error) {
//	m.lock.RLock()
//	defer m.lock.RUnlock()
//	namespace := namespaceFromContext(ctx)
//	if _, ok := m.prefixes[namespace]; !ok {
//		return Prefix{}, ErrNamespaceDoesNotExist
//	}
//	result, ok := m.prefixes[namespace][prefix]
//	if !ok {
//		return Prefix{}, fmt.Errorf("prefix %s not found", prefix)
//	}
//	return *result.deepCopy(), nil
//}
//
//func (m *memory) DeleteAllPrefixes(ctx context.Context) error {
//	m.lock.RLock()
//	defer m.lock.RUnlock()
//	namespace := namespaceFromContext(ctx)
//	m.prefixes[namespace] = make(map[string]Prefix)
//	return nil
//}
//
//func (m *memory) ReadAllPrefixes(ctx context.Context) (Prefixes, error) {
//	m.lock.RLock()
//	defer m.lock.RUnlock()
//	namespace := namespaceFromContext(ctx)
//	if _, ok := m.prefixes[namespace]; !ok {
//		return nil, ErrNamespaceDoesNotExist
//	}
//	ps := make([]Prefix, 0, len(m.prefixes[namespace]))
//	for _, v := range m.prefixes[namespace] {
//		ps = append(ps, *v.deepCopy())
//	}
//	return ps, nil
//}
//
//func (m *memory) ReadAllPrefixCidrs(ctx context.Context) ([]string, error) {
//	m.lock.RLock()
//	defer m.lock.RUnlock()
//	namespace := namespaceFromContext(ctx)
//	if _, ok := m.prefixes[namespace]; !ok {
//		return nil, ErrNamespaceDoesNotExist
//	}
//
//	ps := make([]string, 0, len(m.prefixes[namespace]))
//	for _, v := range m.prefixes[namespace] {
//		ps = append(ps, v.Cidr)
//	}
//	return ps, nil
//}
//
//func (m *memory) UpdatePrefix(ctx context.Context, prefix Prefix) (Prefix, error) {
//	m.lock.Lock()
//	defer m.lock.Unlock()
//	namespace := namespaceFromContext(ctx)
//	oldVersion := prefix.Version
//	prefix.Version = oldVersion + 1
//
//	if prefix.Cidr == "" {
//		return Prefix{}, fmt.Errorf("prefix not present:%v", prefix)
//	}
//
//	if _, ok := m.prefixes[namespace]; !ok {
//		return Prefix{}, ErrNamespaceDoesNotExist
//	}
//	oldPrefix, ok := m.prefixes[namespace][prefix.Cidr]
//	if !ok {
//		return Prefix{}, fmt.Errorf("prefix not found:%s", prefix.Cidr)
//	}
//
//	if oldPrefix.Version != oldVersion {
//		return Prefix{}, fmt.Errorf("%w: unable to update prefix:%s", ErrOptimisticLockError, prefix.Cidr)
//	}
//	m.prefixes[namespace][prefix.Cidr] = *prefix.deepCopy()
//	return prefix, nil
//}
//func (m *memory) DeletePrefix(ctx context.Context, prefix Prefix) (Prefix, error) {
//	m.lock.Lock()
//	defer m.lock.Unlock()
//	namespace := namespaceFromContext(ctx)
//	if _, ok := m.prefixes[namespace]; !ok {
//		return Prefix{}, ErrNamespaceDoesNotExist
//	}
//	delete(m.prefixes[namespace], prefix.Cidr)
//	return *prefix.deepCopy(), nil
//}
//
//func (m *memory) CreateNamespace(_ context.Context, namespace string) error {
//	m.lock.Lock()
//	defer m.lock.Unlock()
//	if _, ok := m.prefixes[namespace]; !ok {
//		m.prefixes[namespace] = make(map[string]Prefix)
//	}
//	return nil
//}
//
//func (m *memory) ListNamespaces(_ context.Context) ([]string, error) {
//	m.lock.Lock()
//	defer m.lock.Unlock()
//	var result []string
//	for n := range m.prefixes {
//		result = append(result, n)
//	}
//	return result, nil
//}
//
//func (m *memory) DeleteNamespace(_ context.Context, namespace string) error {
//	if _, ok := m.prefixes[namespace]; !ok {
//		return ErrNamespaceDoesNotExist
//	}
//	delete(m.prefixes, namespace)
//	return nil
//}
//
//func (m *memory) PutIPAddress(ctx context.Context, prefix Prefix, ip string) error {
//	namespace := namespaceFromContext(ctx)
//	m.lock.Lock()
//	defer m.lock.Unlock()
//
//	key := fmt.Sprintf("/%s/%s/", namespace, prefix.Cidr)
//	m.IPs[key] = append(m.IPs[key], ip)
//	return nil
//}
//func (m *memory) DeleteIPAddress(ctx context.Context, prefix Prefix, ip string) error {
//	namespace := namespaceFromContext(ctx)
//	m.lock.Lock()
//	defer m.lock.Unlock()
//
//	key := fmt.Sprintf("/%s/%s/", namespace, prefix.Cidr)
//	originIPs := m.IPs[key]
//	var newIPs []string
//	for _, v := range originIPs {
//		if v == ip {
//			continue
//		}
//		newIPs = append(newIPs, v)
//	}
//	m.IPs[key] = newIPs
//	return nil
//
//}
//func (m *memory) IPAllocated(ctx context.Context, prefix Prefix, ip string) bool {
//	namespace := namespaceFromContext(ctx)
//	key := fmt.Sprintf("/%s/%s/", namespace, prefix.Cidr)
//	allocated := m.IPs[key]
//	for _, v := range allocated {
//		if v == ip {
//			return true
//		}
//	}
//	return false
//
//}
//func (m *memory) AllocatedIPS(ctx context.Context, prefix Prefix) ([]IPStorage, error) {
//	namespace := namespaceFromContext(ctx)
//	key := fmt.Sprintf("/%s/%s/", namespace, prefix.Cidr)
//	var ips []IPStorage
//	for _, v := range m.IPs[key] {
//		ipStorage := IPStorage{
//			IP:           v,
//			ParentPrefix: prefix.Cidr,
//			Namespace:    prefix.Namespace,
//		}
//		ips = append(ips, ipStorage)
//	}
//	return ips, nil
//}
