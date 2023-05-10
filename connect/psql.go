package connect

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

func ConnectPSQL(ctx context.Context, user string, password string, dbname string, sslmode string) (*pgxpool.Pool, error) {
	fmt.Print("Connecting to postgresql...\n")
	connectionString := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=%s", user, password, dbname, sslmode)
	fmt.Println(connectionString)
	dbpool, err := pgxpool.New(ctx, connectionString)
	if err != nil {
		log.Printf("error during intial connection: %v\n", err)
		return nil, err
	}
	// need to remove line below to prevent
	// error occuring in main func
	// defer dbpool.Close()
	return dbpool, nil
}

// func ConnectGTpsql(ctx context.Context, user string, password string, dbname string, sslmode string) (*pgxpool.Pool, error) {
// 	fmt.Print("Connecting to postgresql...\n")
// 	connectionString := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=%s", user, password, dbname, sslmode)
// 	fmt.Println(connectionString)
// 	dbpool, err := pgxpool.New(ctx, connectionString)
// 	if err != nil {
// 		log.Printf("error during intial connection: %v\n", err)
// 		return nil, err
// 	}
// 	// need to remove line below to prevent
// 	// error occuring in main func
// 	// defer dbpool.Close()
// 	return dbpool, nil
// }
