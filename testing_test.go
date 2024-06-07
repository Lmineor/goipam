package ipam

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

func getBackend() *gorm.DB {
	dbName := "test"
	db := gormMysql(dbName)
	RegisterTables(db)
	return db
}

// RegisterTables register table to mysql if not exist
//func RegisterTables(db *gorm.DB) {
//	db.Exec("DROP TABLE namespaces")
//	db.Exec("DROP TABLE prefixes")
//	err := db.AutoMigrate(
//		&IPStorage{},
//		&Namespace{},
//		&Prefix{},
//	)
//	if err != nil {
//		fmt.Printf("register table failed %s\n", err)
//	}
//}

func gormMysql(dbName string) *gorm.DB {
	username := "root"
	password := "123456"
	path := "127.0.0.1"
	port := "3306"
	config := "charset=utf8&parseTime=True&loc=Local"
	maxIdleConns := 10
	maxOpenConns := 100
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?%s", username, password, path, port, dbName, config)
	mysqlConfig := mysql.Config{
		DSN:                       dsn,   // DSN data source name
		DefaultStringSize:         191,   // string 类型字段的默认长度
		SkipInitializeWithVersion: false, // 根据版本自动配置
	}
	if db, err := gorm.Open(mysql.New(mysqlConfig)); err != nil {
		return nil
	} else {
		sqlDB, _ := db.DB()
		sqlDB.SetMaxIdleConns(maxIdleConns)
		sqlDB.SetMaxOpenConns(maxOpenConns)
		return db
	}
}

//
//func startCockroach() (container testcontainers.Container, dn *sql, err error) {
//	ctx := context.Background()
//	crOnce.Do(func() {
//		var err error
//		req := testcontainers.ContainerRequest{
//			Image:        "cockroachdb/cockroach:" + cockroachVersion,
//			ExposedPorts: []string{"26257/tcp", "8080/tcp"},
//			Env:          map[string]string{"POSTGRES_PASSWORD": "password"},
//			WaitingFor: wait.ForAll(
//				wait.ForLog("initialized new cluster"),
//				wait.ForListeningPort("8080/tcp"),
//				wait.ForListeningPort("26257/tcp"),
//			),
//			Cmd: []string{"start-single-node", "--insecure", "--store=type=mem,size=70%"},
//		}
//		crContainer, err = testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
//			ContainerRequest: req,
//			Started:          true,
//		})
//		if err != nil {
//			panic(err.Error())
//		}
//	})
//	ip, err := crContainer.Host(ctx)
//	if err != nil {
//		return crContainer, nil, err
//	}
//	port, err := crContainer.MappedPort(ctx, "26257")
//	if err != nil {
//		return crContainer, nil, err
//	}
//	dbname := "defaultdb"
//	db, err := newPostgres(ip, port.Port(), "root", "password", dbname, SSLModeDisable)
//
//	return crContainer, db, err
//}
//
//func startRedis() (container testcontainers.Container, s *redis, err error) {
//	ctx := context.Background()
//	redisOnce.Do(func() {
//		var err error
//		req := testcontainers.ContainerRequest{
//			Image:        "redis:" + redisVersion,
//			ExposedPorts: []string{"6379/tcp"},
//			WaitingFor: wait.ForAll(
//				wait.ForLog("Ready to accept connections"),
//				wait.ForListeningPort("6379/tcp"),
//			),
//		}
//		redisContainer, err = testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
//			ContainerRequest: req,
//			Started:          true,
//		})
//		if err != nil {
//			panic(err.Error())
//		}
//	})
//	ip, err := redisContainer.Host(ctx)
//	if err != nil {
//		return redisContainer, nil, err
//	}
//	port, err := redisContainer.MappedPort(ctx, "6379")
//	if err != nil {
//		return redisContainer, nil, err
//	}
//	db, err := newRedis(ctx, ip, port.Port())
//	if err != nil {
//		return redisContainer, nil, err
//	}
//	return redisContainer, db, nil
//}
//
//func startEtcd() (container testcontainers.Container, s *etcd, err error) {
//	ctx := context.Background()
//	etcdOnce.Do(func() {
//		var err error
//		req := testcontainers.ContainerRequest{
//			Image:        "quay.io/coreos/etcd:" + etcdVersion,
//			ExposedPorts: []string{"2379:2379", "2380:2380"},
//			Cmd: []string{"etcd",
//				"--name", "etcd",
//				"--advertise-client-urls", "http://0.0.0.0:2379",
//				"--initial-advertise-peer-urls", "http://0.0.0.0:2380",
//				"--listen-client-urls", "http://0.0.0.0:2379",
//				"--listen-peer-urls", "http://0.0.0.0:2380",
//			},
//		}
//		etcdContainer, err = testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
//			ContainerRequest: req,
//			Started:          true,
//		})
//		if err != nil {
//			panic(err.Error())
//		}
//	})
//	ip, err := etcdContainer.Host(ctx)
//	if err != nil {
//		return etcdContainer, nil, err
//	}
//	port, err := etcdContainer.MappedPort(ctx, "2379")
//	if err != nil {
//		return etcdContainer, nil, err
//	}
//	db, err := newEtcd(ctx, ip, port.Port(), nil, nil, true)
//	if err != nil {
//		return etcdContainer, nil, err
//	}
//	return etcdContainer, db, nil
//}
//
//func startMongodb() (container testcontainers.Container, s *mongodb, err error) {
//	ctx := context.Background()
//
//	mdbOnce.Do(func() {
//		var err error
//		req := testcontainers.ContainerRequest{
//			Image:        `mongo:` + mdbVersion,
//			ExposedPorts: []string{`27017/tcp`},
//			Env: map[string]string{
//				`MONGO_INITDB_ROOT_USERNAME`: `testuser`,
//				`MONGO_INITDB_ROOT_PASSWORD`: `testuser`,
//			},
//			WaitingFor: wait.ForAll(
//				wait.ForLog(`Waiting for connections`),
//				wait.ForListeningPort(`27017/tcp`),
//			),
//			Cmd: []string{`mongod`},
//		}
//		mdbContainer, err = testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
//			ContainerRequest: req,
//			Started:          true,
//		})
//		if err != nil {
//			panic(err.Error())
//		}
//	})
//	ip, err := mdbContainer.Host(ctx)
//	if err != nil {
//		return mdbContainer, nil, err
//	}
//	port, err := mdbContainer.MappedPort(ctx, `27017`)
//	if err != nil {
//		return mdbContainer, nil, err
//	}
//
//	opts := options.Client()
//	opts.ApplyURI(fmt.Sprintf(`mongodb://%s:%s`, ip, port.Port()))
//	opts.Auth = &options.Credential{
//		AuthMechanism: `SCRAM-SHA-1`,
//		Username:      `testuser`,
//		Password:      `testuser`,
//	}
//
//	c := MongoConfig{
//		DatabaseName:       `go-ipam`,
//		MongoClientOptions: opts,
//	}
//	db, err := newMongo(ctx, c)
//
//	return mdbContainer, db, err
//}
//
//func startKeyDB() (container testcontainers.Container, s *redis, err error) {
//	ctx := context.Background()
//	keyDBOnce.Do(func() {
//		var err error
//		req := testcontainers.ContainerRequest{
//			Image:        "eqalpha/keydb:" + keyDBVersion,
//			ExposedPorts: []string{"6379/tcp"},
//			WaitingFor: wait.ForAll(
//				wait.ForLog("Server initialized"),
//				wait.ForListeningPort("6379/tcp"),
//			),
//		}
//		keyDBContainer, err = testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
//			ContainerRequest: req,
//			Started:          true,
//		})
//		if err != nil {
//			panic(err.Error())
//		}
//	})
//	ip, err := keyDBContainer.Host(ctx)
//	if err != nil {
//		return keyDBContainer, nil, err
//	}
//	port, err := keyDBContainer.MappedPort(ctx, "6379")
//	if err != nil {
//		return keyDBContainer, nil, err
//	}
//	db, err := newRedis(ctx, ip, port.Port())
//	if err != nil {
//		return keyDBContainer, nil, err
//	}
//	return keyDBContainer, db, nil
//}
//
//// func stopDB(c testcontainers.Container) error {
//// 	ctx := context.Background()
//// 	return c.Terminate(ctx)
//// }
//
//// func cleanUp(s *sql) {
//// 	s.db.MustExec("DROP TABLE prefixes")
//// }
//
//// cleanable interface for impls that support cleaning before each testrun
//type cleanable interface {
//	cleanup() error
//}
//
//// extendedSQL extended sql interface
//type extendedSQL struct {
//	*sql
//	c testcontainers.Container
//}
//
//// extendedSQL extended sql interface
//type kvStorage struct {
//	*redis
//	c testcontainers.Container
//}
//type kvEtcdStorage struct {
//	*etcd
//	c testcontainers.Container
//}
//
//type docStorage struct {
//	*mongodb
//	c testcontainers.Container
//}
//
//func newPostgresWithCleanup() (*extendedSQL, error) {
//	c, s, err := startPostgres()
//	if err != nil {
//		return nil, err
//	}
//
//	ext := &extendedSQL{
//		sql: s,
//		c:   c,
//	}
//
//	return ext, nil
//}
//func newCockroachWithCleanup() (*extendedSQL, error) {
//	c, s, err := startCockroach()
//	if err != nil {
//		return nil, err
//	}
//
//	ext := &extendedSQL{
//		sql: s,
//		c:   c,
//	}
//
//	return ext, nil
//}
//func newRedisWithCleanup() (*kvStorage, error) {
//	c, r, err := startRedis()
//	if err != nil {
//		return nil, err
//	}
//
//	kv := &kvStorage{
//		redis: r,
//		c:     c,
//	}
//
//	return kv, nil
//}
//func newEtcdWithCleanup() (*kvEtcdStorage, error) {
//	c, r, err := startEtcd()
//	if err != nil {
//		return nil, err
//	}
//
//	kv := &kvEtcdStorage{
//		etcd: r,
//		c:    c,
//	}
//
//	return kv, nil
//}
//
//func newKeyDBWithCleanup() (*kvStorage, error) {
//	c, r, err := startKeyDB()
//	if err != nil {
//		return nil, err
//	}
//
//	kv := &kvStorage{
//		redis: r,
//		c:     c,
//	}
//
//	return kv, nil
//}
//
//func newMongodbWithCleanup() (*docStorage, error) {
//	c, s, err := startMongodb()
//	if err != nil {
//		return nil, err
//	}
//
//	x := &docStorage{
//		mongodb: s,
//		c:       c,
//	}
//	return x, nil
//}
//
//// cleanup database before test
//func (e *extendedSQL) cleanup() error {
//	tx := e.sql.db.MustBegin()
//	_, err := e.sql.db.Exec("TRUNCATE TABLE prefixes")
//	if err != nil {
//		return err
//	}
//	return tx.Commit()
//}
//
//// cleanup database before test
//func (kv *kvStorage) cleanup() error {
//	return kv.redis.DeleteAllPrefixes(context.Background(), defaultNamespace)
//}
//
//// cleanup database before test
//func (kv *kvEtcdStorage) cleanup() error {
//	return kv.etcd.DeleteAllPrefixes(context.Background(), defaultNamespace)
//}
//
//// cleanup database before test
//func (s *sql) cleanup() error {
//	tx := s.db.MustBegin()
//	_, err := s.db.Exec("TRUNCATE TABLE prefixes")
//	if err != nil {
//		return err
//	}
//	return tx.Commit()
//}
//
//func (ds *docStorage) cleanup() error {
//	return ds.mongodb.DeleteAllPrefixes(context.Background(), defaultNamespace)
//}
//
//type benchMethod func(b *testing.B, ipam *ipamer)
//
//func benchWithBackends(b *testing.B, fn benchMethod) {
//	for _, storageProvider := range storageProviders() {
//		if backend != "" && backend != storageProvider.name {
//			continue
//		}
//		storage := storageProvider.provide()
//
//		if tp, ok := storage.(cleanable); ok {
//			err := tp.cleanup()
//			if err != nil {
//				b.Errorf("error cleaning up, %v", err)
//			}
//		}
//
//		ipamer := &ipamer{storage: storage}
//		testName := storageProvider.name
//
//		b.Run(testName, func(b *testing.B) {
//			fn(b, ipamer)
//		})
//	}
//}
//
//type testMethod func(t *testing.T, ipam *ipamer)
//
//func testWithBackends(t *testing.T, fn testMethod) {
//	t.Helper()
//	// prevent testcontainer logging mangle test and benchmark output
//	testcontainers.WithLogger(testcontainers.TestLogger(t))
//	for _, storageProvider := range storageProviders() {
//		if backend != "" && backend != storageProvider.name {
//			continue
//		}
//		storage := storageProvider.provide()
//
//		if tp, ok := storage.(cleanable); ok {
//			err := tp.cleanup()
//			if err != nil {
//				t.Errorf("error cleaning up, %v", err)
//			}
//		}
//
//		ipamer := &ipamer{storage: storage}
//		testName := storageProvider.name
//
//		t.Run(testName, func(t *testing.T) {
//			fn(t, ipamer)
//		})
//	}
//}
//
//type sqlTestMethod func(t *testing.T, sql *sql)
//
//func testWithSQLBackends(t *testing.T, fn sqlTestMethod) {
//	t.Helper()
//	// prevent testcontainer logging mangle test and benchmark output
//	testcontainers.WithLogger(testcontainers.TestLogger(t))
//	for _, storageProvider := range storageProviders() {
//		if backend != "" && backend != storageProvider.name {
//			continue
//		}
//		sqlstorage := storageProvider.providesql()
//		if sqlstorage == nil {
//			continue
//		}
//
//		err := sqlstorage.cleanup()
//		if err != nil {
//			t.Errorf("error cleaning up, %v", err)
//		}
//
//		testName := storageProvider.name
//
//		t.Run(testName, func(t *testing.T) {
//			fn(t, sqlstorage)
//		})
//	}
//}

//
//type provide func() Storage
//type providesql func() *sql
//
//// storageProvider provides different storages
//type storageProvider struct {
//	name       string
//	provide    provide
//	providesql providesql
//}
//
//func storageProviders() []storageProvider {
//	return []storageProvider{
//		{
//			name: "Memory",
//			provide: func() Storage {
//				return NewMemory(context.Background())
//			},
//			providesql: func() *sql {
//				return nil
//			},
//		},
//		{
//			name: "Postgres",
//			provide: func() Storage {
//				storage, err := newPostgresWithCleanup()
//				if err != nil {
//					panic("error getting postgres storage")
//				}
//				return storage
//			},
//			providesql: func() *sql {
//				storage, err := newPostgresWithCleanup()
//				if err != nil {
//					panic("error getting postgres storage")
//				}
//				return storage.sql
//			},
//		},
//		{
//			name: "Cockroach",
//			provide: func() Storage {
//				storage, err := newCockroachWithCleanup()
//				if err != nil {
//					panic("error getting cockroach storage")
//				}
//				return storage
//			},
//			providesql: func() *sql {
//				storage, err := newCockroachWithCleanup()
//				if err != nil {
//					panic("error getting cockroach storage")
//				}
//				return storage.sql
//			},
//		},
//		{
//			name: "Redis",
//			provide: func() Storage {
//				s, err := newRedisWithCleanup()
//				if err != nil {
//					panic(fmt.Sprintf("unable to start redis:%s", err))
//				}
//				return s
//			},
//			providesql: func() *sql {
//				return nil
//			},
//		},
//		{
//			name: "Etcd",
//			provide: func() Storage {
//				s, err := newEtcdWithCleanup()
//				if err != nil {
//					panic(fmt.Sprintf("unable to start etcd:%s", err))
//				}
//				return s
//			},
//			providesql: func() *sql {
//				return nil
//			},
//		},
//		{
//			name: "KeyDB",
//			provide: func() Storage {
//				s, err := newKeyDBWithCleanup()
//				if err != nil {
//					panic(fmt.Sprintf("unable to start keydb:%s", err))
//				}
//				return s
//			},
//			providesql: func() *sql {
//				return nil
//			},
//		},
//		{
//			name: "MongoDB",
//			provide: func() Storage {
//				storage, err := newMongodbWithCleanup()
//				if err != nil {
//					panic(fmt.Sprintf(`error getting mongodb storage, error: %s`, err))
//				}
//				return storage
//			},
//			providesql: func() *sql {
//				return nil
//			},
//		},
//	}
//}
