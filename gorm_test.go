package ipam

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_gorm_prefixExists(t *testing.T) {
	ctx := context.Background()
	namespace := "%u6c^qi$u%tSqhQTcjR!zZHNvMB$3XJd"
	ctx = NewContextWithNamespace(ctx, namespace)
	db := getBackend()
	g := NewGormImpl(db, 50)

	require.NotNil(t, db)

	if !g.checkNamespaceExists(namespace) {
		err := g.CreateNamespace(ctx, namespace)
		require.NoError(t, err)
	}
	// Existing Prefix
	prefix := Prefix{Cidr: "10.0.0.0/16", Namespace: namespace}
	p, err := g.CreatePrefix(ctx, prefix)
	require.NoError(t, err)
	require.NotNil(t, p)
	require.Equal(t, prefix.Cidr, p.Cidr)
	got, exists := g.prefixExists(ctx, prefix)
	require.True(t, exists)
	require.Equal(t, got.Cidr, prefix.Cidr)

	// NonExisting Prefix
	notExistingPrefix := Prefix{Cidr: "10.0.0.0/8"}
	got, exists = g.prefixExists(ctx, notExistingPrefix)
	fmt.Println(exists)
	require.False(t, exists)
	require.Nil(t, got)

	// Delete Existing Prefix
	_, err = g.DeletePrefix(ctx, prefix)
	require.NoError(t, err)
	got, exists = g.prefixExists(ctx, prefix)
	require.False(t, exists)
	require.Nil(t, got)
}

func Test_sql_CreatePrefix(t *testing.T) {
	ctx := context.Background()
	namespace := "%u6c^qi$u%tSqhQTcjR!zZHNvMB$3XJd"
	ctx = NewContextWithNamespace(ctx, namespace)
	db := getBackend()
	g := NewGormImpl(db, 50)

	require.NotNil(t, db)
	if !g.checkNamespaceExists(namespace) {
		err := g.CreateNamespace(ctx, namespace)
		require.NoError(t, err)
	}

	// Existing Prefix
	prefix := Prefix{Cidr: "11.0.0.0/16"}
	got, exists := g.prefixExists(ctx, prefix)
	require.False(t, exists)
	require.Nil(t, got)
	p, err := g.CreatePrefix(ctx, prefix)
	require.NoError(t, err)
	require.NotNil(t, p)
	require.Equal(t, prefix.Cidr, p.Cidr)
	got, exists = g.prefixExists(ctx, prefix)
	require.True(t, exists)
	require.Equal(t, got.Cidr, prefix.Cidr)

	// Duplicate Prefix
	p, err = g.CreatePrefix(ctx, prefix)
	require.NoError(t, err)
	require.NotNil(t, p)
	require.Equal(t, prefix.Cidr, p.Cidr)

	ps, err := g.ReadAllPrefixCidrs(ctx)
	require.NoError(t, err)
	require.NotNil(t, ps)
	require.Equal(t, 1, len(ps))
}

func Test_sql_ReadPrefix(t *testing.T) {
	ctx := context.Background()
	db := getBackend()
	g := NewGormImpl(db, 50)
	require.NotNil(t, db)

	// Prefix
	ctx1 := NewContextWithNamespace(ctx, "a")
	p, err := g.ReadPrefix(ctx1, "12.0.0.0/8")
	require.Error(t, err)
	require.ErrorIs(t, err, ErrNamespaceDoesNotExist)
	require.Empty(t, p)

	prefix := Prefix{Cidr: "12.0.0.0/16"}

	// Create Namespace
	err = g.CreateNamespace(ctx, "a")
	require.NoError(t, err)

	ctx1 = NewContextWithNamespace(ctx, "a")
	p, err = g.CreatePrefix(ctx1, prefix)
	require.NoError(t, err)
	require.NotNil(t, p)

	p, err = g.ReadPrefix(ctx1, "12.0.0.0/16")
	require.NoError(t, err)
	require.NotNil(t, p)
	require.Equal(t, "12.0.0.0/16", p.Cidr)
}

func Test_sql_ReadAllPrefix(t *testing.T) {
	ctx := context.Background()
	namespace := "%u6c^qi$u%tSqhQTcjR!zZHNvMB$3XJd"
	ctx = NewContextWithNamespace(ctx, namespace)
	db := getBackend()
	g := NewGormImpl(db, 50)

	require.NotNil(t, db)
	if !g.checkNamespaceExists(namespace) {
		err := g.CreateNamespace(ctx, namespace)
		require.NoError(t, err)
	}
	// no Prefixes
	ps, err := g.ReadAllPrefixCidrs(ctx)
	require.NoError(t, err)
	require.NotNil(t, ps)
	require.Equal(t, 0, len(ps))

	// One Prefix
	prefix := Prefix{Cidr: "12.0.0.0/16"}
	p, err := g.CreatePrefix(ctx, prefix)
	require.NoError(t, err)
	require.NotNil(t, p)
	ps, err = g.ReadAllPrefixCidrs(ctx)
	require.NoError(t, err)
	require.NotNil(t, ps)
	require.Equal(t, 1, len(ps))

	// no Prefixes again
	_, err = g.DeletePrefix(ctx, prefix)
	require.NoError(t, err)
	ps, err = g.ReadAllPrefixCidrs(ctx)
	require.NoError(t, err)
	require.NotNil(t, ps)
	require.Equal(t, 0, len(ps))
}

func Test_sql_CreateNamespace(t *testing.T) {
	ctx := context.Background()
	ctx = NewContextWithNamespace(ctx, "/root/")
	db := getBackend()
	g := NewGormImpl(db, 50)

	require.NotNil(t, db)
	{
		// Create a namespace with special characters in name
		namespace := "%u6c^qi$u%tSqhQTcjR!zZHNvMB$3XJd"
		err := g.CreateNamespace(ctx, namespace)
		require.NoError(t, err)

		err = g.DeleteNamespace(ctx, namespace)
		require.NoError(t, err)
	}
	{
		// Create a long namespace name
		namespace := "d4546731-6056-4b48-80e9-ef924ca7f651"
		err := g.CreateNamespace(ctx, namespace)
		require.NoError(t, err)

		err = g.DeleteNamespace(ctx, namespace)
		require.NoError(t, err)
	}
	{
		// Create a namespace with a name that is too long
		namespace := "d4546731-6056-4b48-80e9-ef924ca7f651d4546731-6056-4b48-80e9-ef924ca7f651d4546731-6056-4b48-80e9-ef924ca7f651d4546731-6056-4b48-80e9-ef924ca7f651"
		err := g.CreateNamespace(ctx, namespace)
		require.Error(t, err)
		require.ErrorIs(t, err, ErrNameTooLong)
	}
}

func Test_AcquirePrefixIPv4(t *testing.T) {
	ctx := context.Background()
	namespace := "%u6c^qi$u%tSqhQTcjR!zZHNvMB$3XJd"
	ctx = NewContextWithNamespace(ctx, namespace)
	db := getBackend()
	g := NewGormImpl(db, 50)

	require.NotNil(t, db)
	if !g.checkNamespaceExists(namespace) {
		err := g.CreateNamespace(ctx, namespace)
		require.NoError(t, err)
	}
	ipamer := NewWithStorage(&g)

	const parentCidr = "2.1.0.0/16"
	parent, err := ipamer.NewPrefix(ctx, parentCidr)
	require.NoError(t, err)
	for i := 0; i < 4; i++ {
		acquirePrefix(t, ctx, &g, parent.ID)
	}
}

func Test_AcquirePrefixIPv6(t *testing.T) {
	ctx := context.Background()
	namespace := "%u6c^qi$u%tSqhQTcjR!zZHNvMB$3XJd"
	ctx = NewContextWithNamespace(ctx, namespace)
	db := getBackend()
	g := NewGormImpl(db, 50)

	require.NotNil(t, db)
	if !g.checkNamespaceExists(namespace) {
		err := g.CreateNamespace(ctx, namespace)
		require.NoError(t, err)
	}
	ipamer := NewWithStorage(&g)

	const parentCidr = "1::0/64"
	parent, err := ipamer.NewPrefix(ctx, parentCidr)
	require.NoError(t, err)
	for i := 0; i < 1000; i++ {
		acquirePrefix(t, ctx, &g, parent.ID)
	}
}

func TestIpamer_ReleaseChildPrefix(t *testing.T) {
	ctx := context.Background()
	namespace := "%u6c^qi$u%tSqhQTcjR!zZHNvMB$3XJd"
	ctx = NewContextWithNamespace(ctx, namespace)
	db := getBackend()
	g := NewGormImpl(db, 50)

	require.NotNil(t, db)
	if !g.checkNamespaceExists(namespace) {
		err := g.CreateNamespace(ctx, namespace)
		require.NoError(t, err)
	}
	ipamer := NewWithStorage(&g)

	const parentCidr = "1::0/64"
	parent, err := ipamer.NewPrefix(ctx, parentCidr)
	require.NoError(t, err)
	for i := 0; i < 1000; i++ {
		releasePrefix(t, ctx, &g, parent.ID)
	}
}
func releasePrefix(t *testing.T, ctx context.Context, g *gormImplement, parentID uint) {
	require.NotNil(t, g)
	ipamer := NewWithStorage(g)
	childPrefix, err := ipamer.AcquireChildPrefix(ctx, parentID, 128)
	err = ipamer.ReleaseChildPrefix(ctx, childPrefix)
	require.NoError(t, err)
	fmt.Println(childPrefix.Cidr)
}

func acquirePrefix(t *testing.T, ctx context.Context, g *gormImplement, parentID uint) {
	require.NotNil(t, g)
	ipamer := NewWithStorage(g)
	childPrefix, err := ipamer.AcquireChildPrefix(ctx, parentID, 24)
	require.NoError(t, err)
	fmt.Println(childPrefix.Cidr)
}
func Test_AcquireIP(t *testing.T) {
	ctx := context.Background()
	namespace := "%u6c^qi$u%tSqhQTcjR!zZHNvMB$3XJd"
	ctx = NewContextWithNamespace(ctx, namespace)
	db := getBackend()
	g := NewGormImpl(db, 50)

	require.NotNil(t, db)
	if !g.checkNamespaceExists(namespace) {
		err := g.CreateNamespace(ctx, namespace)
		require.NoError(t, err)
	}
	ipamer := NewWithStorage(&g)
	const parentCidr = "2001:db8:85a3::/124"
	parent, err := ipamer.NewPrefix(ctx, parentCidr)
	require.NoError(t, err)
	require.NotNil(t, db)
	{
		// Create a namespace with special characters in name
		ip, err := ipamer.AcquireIP(ctx, parent.ID)
		require.NoError(t, err)
		fmt.Println(ip.IP)
	}
	{
		// Create a long namespace name
		ip, err := ipamer.AcquireIP(ctx, parent.ID)
		require.NoError(t, err)
		fmt.Println(ip.IP)
	}
	{
		// Create a namespace with a name that is too long
		ip, err := ipamer.AcquireIP(ctx, parent.ID)
		require.NoError(t, err)
		fmt.Println(ip.IP)
	}
	//{
	//	// Create a namespace with a name that is too long
	//	ip, err := ipamer.AcquireSpecificIP(ctx, parent.ID, "1.0.0.100")
	//	require.ErrorIs(t, err, ErrAlreadyAllocated)
	//	fmt.Println(ip.IP)
	//}
}

//func Test_ConcurrentAcquireIP(t *testing.T) {
//	ctx := context.Background()
//	testWithSQLBackends(t, func(t *testing.T, db *sql) {
//		require.NotNil(t, db)
//
//		ipamer := NewWithStorage(db)
//
//		const parentCidr = "2.7.0.0/16"
//		_, err := ipamer.NewPrefix(ctx, parentCidr)
//		require.NoError(t, err)
//
//		count := 30
//		ips := make(chan string)
//		for i := 0; i < count; i++ {
//			go acquireIP(t, ctx, db, parentCidr, ips)
//		}
//
//		ipMap := make(map[string]bool)
//		for i := 0; i < count; i++ {
//			p := <-ips
//			_, duplicate := ipMap[p]
//			if duplicate {
//				t.Errorf("prefix:%s already acquired", p)
//			}
//			ipMap[p] = true
//		}
//	})
//}
//
//func acquireIP(t *testing.T, ctx context.Context, db *sql, prefix string, ips chan string) {
//	require.NotNil(t, db)
//	ipamer := NewWithStorage(db)
//
//	var ip *IP
//	var err error
//	for ip == nil {
//		ip, err = ipamer.AcquireIP(ctx, prefix)
//		if err != nil {
//			t.Error(err)
//		}
//		time.Sleep(100 * time.Millisecond)
//	}
//	ips <- ip.IP.String()
//}
