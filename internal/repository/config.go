package repository

import (
	"context"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Config defines and returns configuration of pgxpool based on DB_URI provided in env
func Config() *pgxpool.Config {

	dbConfig, err := pgxpool.ParseConfig(os.Getenv("DB_URI"))
	if err != nil {
		log.Printf("error: failed while configuring: %v", err)
		return nil
	}

	dbConfig.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		var modeOid uint32
		err := conn.QueryRow(ctx, "SELECT 'mode'::regtype::oid").Scan(&modeOid)
		if err != nil {
			return err
		}

		// OIDs are primarily for internal system use, especially in the system catalog tables
		var modeArrayOid uint32
		err = conn.QueryRow(ctx, "SELECT '_mode'::regtype::oid").Scan(&modeArrayOid)
		if err != nil {
			return err
		}

		conn.TypeMap().RegisterType(&pgtype.Type{
			Name:  "mode",
			OID:   modeOid,
			Codec: &pgtype.EnumCodec{},
		})

		conn.TypeMap().RegisterType(&pgtype.Type{
			Name:  "_mode",
			OID:   modeArrayOid,
			Codec: &pgtype.ArrayCodec{ElementType: &pgtype.Type{Name: "mode", OID: modeOid, Codec: &pgtype.EnumCodec{}}},
		})

		return nil
	}

	dbConfig.PrepareConn = func(ctx context.Context, conn *pgx.Conn) (bool, error) {
		log.Println("Acquiring connection")
		return true, nil
	}

	dbConfig.AfterRelease = func(conn *pgx.Conn) bool {
		log.Println("Releasing connection")
		return true
	}

	dbConfig.BeforeClose = func(conn *pgx.Conn) {
		log.Println("Closing connection")
	}
	return dbConfig
}
