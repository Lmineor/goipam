package ipam

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"math"
	"net/netip"
	"strings"

	"go4.org/netipx"
)

// Prefix is a expression of a ip with length and forms a classless network.
// nolint:musttag
type Prefix struct {
	ID        uint   `gorm:"primarykey"`
	Cidr      string `gorm:"primaryKey;uniqueIndex:cidr_parent_cidr_namespace_idx"` // The Cidr of this prefix
	ParentID  uint   `gorm:"column:parent_id;foreignKey:ID"`
	IsParent  bool   // if this Prefix has child prefixes, this is set to true
	Version   int64  // Version is used for optimistic locking
	Namespace string `gorm:"uniqueIndex:cidr_parent_cidr_namespace_idx"` // the namespace of this prefix
}

type Prefixes []Prefix

// deepCopy to a new Prefix
func (p Prefix) deepCopy() *Prefix {
	return &Prefix{
		ID:        p.ID,
		Cidr:      p.Cidr,
		ParentID:  p.ParentID,
		IsParent:  p.IsParent,
		Version:   p.Version,
		Namespace: p.Namespace,
	}
}

// GobEncode implements GobEncode for Prefix
func (p *Prefix) GobEncode() ([]byte, error) {
	w := new(bytes.Buffer)
	encoder := gob.NewEncoder(w)
	if err := encoder.Encode(p.IsParent); err != nil {
		return nil, err
	}
	if err := encoder.Encode(p.Version); err != nil {
		return nil, err
	}
	if err := encoder.Encode(p.Cidr); err != nil {
		return nil, err
	}
	if err := encoder.Encode(p.ParentID); err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

// GobDecode implements GobDecode for Prefix
func (p *Prefix) GobDecode(buf []byte) error {
	r := bytes.NewBuffer(buf)
	decoder := gob.NewDecoder(r)
	if err := decoder.Decode(&p.IsParent); err != nil {
		return err
	}
	if err := decoder.Decode(&p.Version); err != nil {
		return err
	}
	if err := decoder.Decode(&p.Cidr); err != nil {
		return err
	}
	return decoder.Decode(&p.ParentID)
}

func copyMap(m map[string]bool) map[string]bool {
	cm := make(map[string]bool, len(m))
	for k, v := range m {
		cm[k] = v
	}
	return cm
}

// Usage of ips and child Prefixes of a Prefix
type Usage struct {
	// AvailableIPs the number of available IPs if this is not a parent prefix
	// No more than 2^31 available IPs are reported
	AvailableIPs uint64
	// AcquiredIPs the number of acquired IPs if this is not a parent prefix
	AcquiredIPs uint64
	// AvailableSmallestPrefixes is the count of available Prefixes with 2 countable Bits
	// No more than 2^31 available Prefixes are reported
	AvailableSmallestPrefixes uint64
	// AvailablePrefixes is a list of prefixes which are available
	AvailablePrefixes []string
	// AcquiredPrefixes the number of acquired prefixes if this is a parent prefix
	AcquiredPrefixes uint64
}

func (i *ipamer) NewPrefix(ctx context.Context, cidr string) (*Prefix, error) {
	namespace := namespaceFromContext(ctx)
	existingPrefixes, err := i.storage.ReadAllPrefixCidrs(ctx)
	if err != nil {
		return nil, err
	}
	p, err := i.newPrefix(cidr, 0, namespace)
	if err != nil {
		return nil, err
	}
	err = PrefixesOverlapping(existingPrefixes, []string{p.Cidr})
	if err != nil {
		return nil, err
	}
	newPrefix, err := i.storage.CreatePrefix(ctx, *p)
	if err != nil {
		return nil, err
	}

	return &newPrefix, nil
}

func (i *ipamer) DeletePrefix(ctx context.Context, cidr string) (*Prefix, error) {
	p := i.GetPrefixByCidr(ctx, cidr)
	if p == nil {
		return nil, fmt.Errorf("%w: delete prefix:%s", ErrNotFound, cidr)
	}

	ips, _ := i.storage.AllocatedIPS(ctx, *p)
	if len(ips) > 0 {
		return nil, fmt.Errorf("prefix %s has ips, delete prefix not possible", p.Cidr)
	}
	prefix, err := i.storage.DeletePrefix(ctx, *p)
	if err != nil {
		return nil, fmt.Errorf("delete prefix:%s %w", p.Cidr, err)
	}

	return &prefix, nil
}

func (i *ipamer) AcquireChildPrefix(ctx context.Context, parentID uint, length uint8) (*Prefix, error) {
	namespace := namespaceFromContext(ctx)
	return i.acquireChildPrefixInternalByParentID(ctx, namespace, parentID, "", int(length))
}

func (i *ipamer) AcquireSpecificChildPrefix(ctx context.Context, parentID uint, childCidr string) (*Prefix, error) {
	namespace := namespaceFromContext(ctx)
	return i.acquireChildPrefixInternalByParentID(ctx, namespace, parentID, childCidr, 0)
}

func (i *ipamer) acquireAllocatedChildPrefixes(ctx context.Context, parentID uint) (Prefixes, error) {
	if parentID == 0 {
		return nil, fmt.Errorf("parent prefix cannot be zero")
	}
	return i.storage.ReadAllChildPrefixByParentID(ctx, parentID)
}

// acquireChildPrefixInternal will return a Prefix with a smaller length from the given Prefix.
func (i *ipamer) acquireChildPrefixInternalByParentID(ctx context.Context, namespace string, parentID uint, childCidr string, length int) (*Prefix, error) {
	specificChildRequest := childCidr != ""
	var childPrefix netip.Prefix
	parent := i.GetPrefixByID(ctx, parentID)
	if parent == nil {
		return nil, fmt.Errorf("unable to find prefix for cidr:%d", parentID)
	}
	ipPrefix, err := netip.ParsePrefix(parent.Cidr)
	if err != nil {
		return nil, err
	}
	if specificChildRequest {
		childPrefix, err = netip.ParsePrefix(childCidr)
		if err != nil {
			return nil, err
		}
		length = childPrefix.Bits()
	}
	if ipPrefix.Bits() >= length {
		return nil, fmt.Errorf("given length:%d must be greater than prefix length:%d", length, ipPrefix.Bits())
	}
	allocatedIPS, _ := i.storage.AllocatedIPS(ctx, *parent)
	if len(allocatedIPS) > 0 {
		return nil, fmt.Errorf("prefix %s has ips, acquire child prefix not possible", parent.Cidr)
	}

	var ipsetBuilder netipx.IPSetBuilder
	ipsetBuilder.AddPrefix(ipPrefix)

	storedAllChildPrefixes, _ := i.acquireAllocatedChildPrefixes(ctx, parentID)
	for _, stored := range storedAllChildPrefixes {
		cpipprefix, err := netip.ParsePrefix(stored.Cidr)
		if err != nil {
			return nil, err
		}
		ipsetBuilder.RemovePrefix(cpipprefix)
	}

	ipset, err := ipsetBuilder.IPSet()
	if err != nil {
		return nil, fmt.Errorf("error constructing ipset:%w", err)
	}

	var cp netip.Prefix
	if !specificChildRequest {
		var ok bool
		cp, _, ok = ipset.RemoveFreePrefix(uint8(length))
		if !ok {
			pfxs := ipset.Prefixes()
			if len(pfxs) == 0 {
				return nil, fmt.Errorf("no prefix found in %s with length:%d", parent.Cidr, length)
			}

			var availablePrefixes []string
			for _, p := range pfxs {
				availablePrefixes = append(availablePrefixes, p.String())
			}
			adj := "are"
			if len(availablePrefixes) == 1 {
				adj = "is"
			}

			return nil, fmt.Errorf("no prefix found in %s with length:%d, but %s %s available", parent.Cidr, length, strings.Join(availablePrefixes, ","), adj)
		}
	} else {
		if ok := ipset.ContainsPrefix(childPrefix); !ok {
			// Parent prefix does not contain specific child prefix
			return nil, fmt.Errorf("specific prefix %s is not available in prefix %s", childCidr, parent.Cidr)
		}
		cp = childPrefix
	}

	child := &Prefix{
		Cidr:     cp.String(),
		ParentID: parentID,
	}

	parent.IsParent = true

	_, err = i.storage.UpdatePrefix(ctx, *parent)
	if err != nil {
		return nil, fmt.Errorf("unable to update parent prefix:%v error:%w", parent, err)
	}
	child, err = i.newPrefix(child.Cidr, parentID, namespace)
	if err != nil {
		return nil, fmt.Errorf("unable to persist created child:%w", err)
	}
	created, err := i.storage.CreatePrefix(ctx, *child)
	if err != nil {
		return nil, fmt.Errorf("unable to update parent prefix:%v error:%w", child, err)
	}

	return &created, nil
}

func (i *ipamer) ReleaseChildPrefix(ctx context.Context, child *Prefix) error {
	namespace := namespaceFromContext(ctx)
	return i.releaseChildPrefixInternal(ctx, namespace, child)
}

// releaseChildPrefixInternal will mark this child Prefix as available again.
func (i *ipamer) releaseChildPrefixInternal(ctx context.Context, _ string, child *Prefix) error {
	parent := i.GetPrefixByID(ctx, child.ParentID)

	if parent == nil {
		return fmt.Errorf("prefix %s is not child prefix", child.Cidr)
	}
	ips, _ := i.storage.AllocatedIPS(ctx, *child)
	if len(ips) > 0 {
		return fmt.Errorf("prefix %s has ips, deletion not possible", child.Cidr)
	}
	_, err := i.DeletePrefix(ctx, child.Cidr)
	if err != nil {
		return fmt.Errorf("unable to release prefix %v:%w", child, err)
	}
	allChiledPrefixes, err := i.storage.ReadAllChildPrefixByParentID(ctx, child.ParentID)
	if err != nil {
		return fmt.Errorf("unable to read all child prefix %d:%w", child.ParentID, err)
	}
	if len(allChiledPrefixes) == 0 {
		parent.IsParent = false
		_, err = i.storage.UpdatePrefix(ctx, *parent)
		if err != nil {
			return fmt.Errorf("unable to update parent prefix %s:%w", parent.Cidr, err)
		}
	}

	return nil
}

func (i *ipamer) GetPrefixByID(ctx context.Context, id uint) *Prefix {
	prefix, err := i.storage.ReadPrefixByID(ctx, id)
	if err != nil {
		return nil
	}
	return &prefix
}

func (i *ipamer) GetPrefixByCidr(ctx context.Context, cidr string) *Prefix {
	ipprefix, err := netip.ParsePrefix(cidr)
	if err != nil {
		return nil
	}
	prefix, err := i.storage.ReadPrefix(ctx, ipprefix.Masked().String())
	if err != nil {
		return nil
	}
	return &prefix
}

func (i *ipamer) AcquireSpecificIP(ctx context.Context, cidr string, specificIP string) (*IPStorage, error) {
	return i.acquireSpecificIPInternal(ctx, cidr, specificIP)
}

// acquireSpecificIPInternal will acquire given IP and store it to DB, if already in use, return nil.
// If specificIP is empty, the next free IP is returned.
// If there is no free IP an NoIPAvailableError is returned.
// If the Prefix is not found an NotFoundError is returned.
func (i *ipamer) acquireSpecificIPInternal(ctx context.Context, cidr string, specificIP string) (*IPStorage, error) {
	parent := i.GetPrefixByCidr(ctx, cidr)
	if parent == nil {
		return nil, fmt.Errorf("%w: unable to find prefix for cidr :%s", ErrNotFound, cidr)
	}
	if parent.IsParent {
		return nil, fmt.Errorf("prefix %s has childprefixes, acquire ip not possible", parent.Cidr)
	}
	ipnet, err := netip.ParsePrefix(parent.Cidr)
	if err != nil {
		return nil, err
	}

	var specificIPnet netip.Addr
	if specificIP != "" {
		specificIPnet, err = netip.ParseAddr(specificIP)
		if err != nil {
			return nil, fmt.Errorf("given ip:%s in not valid", specificIP)
		}
		if !ipnet.Contains(specificIPnet) {
			return nil, fmt.Errorf("given ip:%s is not in %s", specificIP, parent.Cidr)
		}
		allocated := i.storage.IPAllocated(ctx, *parent, specificIPnet.String())
		if allocated {
			return nil, fmt.Errorf("%w: given ip:%s is already allocated", ErrAlreadyAllocated, specificIPnet)
		}
	}

	iprange := netipx.RangeOfPrefix(ipnet)
	allocatedIPS, err := i.storage.AllocatedIPS(ctx, *parent)
	allocatedIPsMap := make(map[string]struct{})
	for _, aIP := range allocatedIPS {
		allocatedIPsMap[aIP.IP] = struct{}{}
	}
	for ip := iprange.From(); ipnet.Contains(ip); ip = ip.Next() {
		ipstring := ip.String()
		if _, ok := allocatedIPsMap[ipstring]; ok {
			continue
		}
		if specificIP == "" || specificIPnet.Compare(ip) == 0 {
			acquired := &IPStorage{
				IP:           ip.String(),
				ParentPrefix: parent.Cidr,
				Namespace:    parent.Namespace,
			}
			_, err := i.storage.UpdatePrefix(ctx, *parent)
			i.storage.PutIPAddress(ctx, *parent, ipstring)
			if err != nil {
				return nil, fmt.Errorf("unable to persist acquired ip:%v error:%w", parent.Cidr, err)
			}
			return acquired, nil
		}
	}

	return nil, fmt.Errorf("%w: no more ips in prefix: %s left", ErrNoIPAvailable, parent.Cidr)
}

func (i *ipamer) AcquireIP(ctx context.Context, cidr string) (*IPStorage, error) {
	return i.AcquireSpecificIP(ctx, cidr, "")
}

func (i *ipamer) ReleaseIP(ctx context.Context, ip *IPStorage) (*Prefix, error) {
	err := i.ReleaseIPFromPrefix(ctx, ip.ParentPrefix, ip.IP)
	prefix := i.GetPrefixByCidr(ctx, ip.ParentPrefix)
	return prefix, err
}

func (i *ipamer) ReleaseIPFromPrefix(ctx context.Context, prefixCidr, ip string) error {
	namespace := namespaceFromContext(ctx)
	return i.releaseIPFromPrefixInternal(ctx, namespace, prefixCidr, ip)
}

// releaseIPFromPrefixInternal will release the given IP for later usage.
func (i *ipamer) releaseIPFromPrefixInternal(ctx context.Context, namespace, prefixCidr, ip string) error {
	prefix := i.GetPrefixByCidr(ctx, prefixCidr)
	if prefix == nil {
		return fmt.Errorf("%w: unable to find prefix for cidr:%s", ErrNotFound, prefixCidr)
	}
	allocated := i.storage.IPAllocated(ctx, *prefix, ip)
	if !allocated {
		return fmt.Errorf("%w: unable to release ip:%s because it is not allocated in prefix:%s", ErrNotFound, ip, prefixCidr)
	}
	err := i.storage.DeleteIPAddress(ctx, *prefix, ip)
	if err != nil {
		return fmt.Errorf("unable to release ip %v:%w", ip, err)
	}
	return nil
}

// PrefixesOverlapping will check if one ore more prefix of newPrefixes is overlapping
// with one of existingPrefixes
func PrefixesOverlapping(existingPrefixes []string, newPrefixes []string) error {
	for _, ep := range existingPrefixes {
		eip, err := netip.ParsePrefix(ep)
		if err != nil {
			return fmt.Errorf("parsing prefix %s failed:%w", ep, err)
		}
		for _, np := range newPrefixes {
			nip, err := netip.ParsePrefix(np)
			if err != nil {
				return fmt.Errorf("parsing prefix %s failed:%w", np, err)
			}
			if eip.Overlaps(nip) || nip.Overlaps(eip) {
				return fmt.Errorf("%s overlaps %s", nip, eip)
			}
		}
	}
	return nil
}

// newPrefix create a new Prefix from a string notation.
func (i *ipamer) newPrefix(cidr string, parentID uint, namespace string) (*Prefix, error) {
	ipnet, err := netip.ParsePrefix(cidr)
	if err != nil {
		return nil, fmt.Errorf("unable to parse cidr:%s %w", cidr, err)
	}

	p := &Prefix{
		Cidr:      ipnet.Masked().String(),
		ParentID:  parentID,
		IsParent:  false,
		Namespace: namespace,
	}

	// FIXME: should this be done by the user ?
	// First ip in the prefix and broadcast is blocked.
	//iprange := netipx.RangeOfPrefix(ipnet)
	//p.ips[iprange.From().String()] = true
	//if ipnet.Addr().Is4() {
	//	// broadcast is ipv4 only
	//	p.ips[iprange.To().String()] = true
	//}
	return p, nil
}

func (i *ipamer) ListAllPrefix(ctx context.Context) (Prefixes, error) {
	// FIXME must dump all namespaces
	return i.storage.ReadAllPrefixes(ctx)
}

// ReadAllPrefixCidrs retrieves all existing Prefix CIDRs from the underlying storage
func (i *ipamer) ReadAllPrefixCidrs(ctx context.Context) ([]string, error) {
	return i.ReadAllNamespacedPrefixCidrs(ctx)
}

// ReadAllNamespacedPrefixCidrs retrieves all existing Prefix CIDRs from the underlying storage
func (i *ipamer) ReadAllNamespacedPrefixCidrs(ctx context.Context) ([]string, error) {
	return i.storage.ReadAllPrefixCidrs(ctx)
}

// CreateNamespace creates a namespace with the given name.
func (i *ipamer) CreateNamespace(ctx context.Context, namespace string) error {
	return i.storage.CreateNamespace(ctx, namespace)
}

// ListNamespaces returns a list of all namespaces.
func (i *ipamer) ListNamespaces(ctx context.Context) ([]string, error) {
	return i.storage.ListNamespaces(ctx)
}

// DeleteNamespace deletes a namespace.
// Warning delete namespace will delete all ip and prefix which belongs it.
func (i *ipamer) DeleteNamespace(ctx context.Context, namespace string) error {
	prefixes, err := i.storage.ReadAllPrefixes(ctx)
	if err != nil {
		return err
	}
	if len(prefixes) > 0 {
		return fmt.Errorf("cannot delete namespace with allocated prefixes")
	}
	return i.storage.DeleteNamespace(ctx, namespace)
}

func (i *ipamer) AllocatedIPs(ctx context.Context, prefix Prefix) ([]IPStorage, error) {
	return i.storage.AllocatedIPS(ctx, prefix)
}
func (p *Prefix) String() string {
	return p.Cidr
}

func (u *Usage) String() string {
	if u.AcquiredPrefixes == 0 {
		return fmt.Sprintf("ip:%d/%d", u.AcquiredIPs, u.AvailableIPs)
	}
	return fmt.Sprintf("ip:%d/%d prefixes alloc:%d avail:%d", u.AcquiredIPs, u.AvailableIPs, u.AcquiredPrefixes, u.AvailableSmallestPrefixes)
}

// Network return the net.IP part of the Prefix
func (p *Prefix) Network() (netip.Addr, error) {
	ipprefix, err := netip.ParsePrefix(p.Cidr)
	if err != nil {
		return netip.Addr{}, err
	}
	return ipprefix.Addr(), nil
}

// availableips return the number of ips available in this Prefix
func (p *Prefix) availableips() uint64 {
	ipprefix, err := netip.ParsePrefix(p.Cidr)
	if err != nil {
		return 0
	}
	// We don't report more than 2^31 available IPs by design
	if (ipprefix.Addr().BitLen() - ipprefix.Bits()) > 31 {
		return math.MaxInt32
	}
	return 1 << (ipprefix.Addr().BitLen() - ipprefix.Bits())
}

// Usage report Prefix usage.
func (p *Prefix) Usage() Usage {
	return Usage{
		AvailableIPs: p.availableips(),
	}
}

func namespaceFromContext(ctx context.Context) string {
	raw := ctx.Value(namespaceContextKey{})
	if raw == nil {
		return defaultNamespace
	}
	if ns, ok := raw.(string); ok {
		return ns
	}
	return defaultNamespace
}
