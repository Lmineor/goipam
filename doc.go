/*
Package ipam is a ip address management library for ip's and prefixes (networks).

It uses either memory or postgresql database to store the ip's and prefixes.
You can also bring you own Storage implementation as you need.

Example usage:

	package main

	import (
		"context"
		"fmt"
		goipam "github.com/Lmineor/goipam"
		"gorm.io/driver/mysql"
		"gorm.io/gorm"
	)

	func main() {
		ctx := context.Background()
		namespace := "namespace1"
		db := getBackend()
		gormImpl := goipam.NewGormImpl(db, 50)
		ipam := goipam.NewWithStorage(gormImpl)

		err := ipam.CreateNamespace(ctx, namespace)
		if err != nil {
			panic(err)
		}

		ctx = goipam.NewContextWithNamespace(ctx, namespace)
		p, err := ipam.NewPrefix(ctx, "192.168.0.1/16")
		if err != nil {
			panic(err)
		}
		ip, err := ipam.AcquireIP(ctx, p.ID)
		if err != nil {
			panic(err)
		}
		fmt.Printf("got IP %s\n", ip.IP)
		p, err = ipam.ReleaseIP(ctx, ip)
		if err != nil {
			panic(err)
		}
		fmt.Printf("release IP %s\n", ip.IP)

	}

	func getBackend() *gorm.DB {
		dbName := "test"
		db := gormMysql(dbName)
		RegisterTables(db)
		return db
	}

	// RegisterTables register table to mysql if not exist
	func RegisterTables(db *gorm.DB) {
		db.Exec("DROP TABLE namespaces")
		db.Exec("DROP TABLE prefixes")
		err := db.AutoMigrate(
			&goipam.IPStorage{},
			&goipam.Namespace{},
			&goipam.Prefix{},
		)
		if err != nil {
			fmt.Printf("register table failed %s\n", err)
		}
	}

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
			DSN:                       dsn,
			DefaultStringSize:         191,
			SkipInitializeWithVersion: false,
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
*/
package ipam
