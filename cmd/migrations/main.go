package main

import (
	"github.com/alpha-omega-corp/auth-svc/pkg/models"
	"github.com/alpha-omega-corp/services/database"
	"github.com/alpha-omega-corp/services/server"
	"github.com/spf13/viper"
	_ "github.com/spf13/viper/remote"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dbfixture"
	"github.com/uptrace/bun/migrate"
	"github.com/urfave/cli/v2"
	"log"
	"os"
)

func main() {
	v := viper.New()
	cManager := server.NewConfigManager(v)

	c, err := cManager.HostsConfig()
	if err != nil {
		panic(err)
	}

	dbHandler := database.NewHandler(c.Auth.Dsn)

	defer func(db *bun.DB) {
		err := db.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(dbHandler.Database())

	appCli := &cli.App{
		Name:  "authentication-svc",
		Usage: "bootstrap the service",
		Commands: []*cli.Command{
			migrateCommand(dbHandler.Database()),
		},
	}

	if err := appCli.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func migrateCommand(db *bun.DB) *cli.Command {
	db.RegisterModel(
		(*models.UserToRole)(nil),
	)

	return &cli.Command{
		Name:  "db",
		Usage: "manage database migrations",
		Subcommands: []*cli.Command{
			{
				Name:  "init",
				Usage: "create migration tables",
				Action: func(c *cli.Context) error {
					migrator := migrate.NewMigrator(db, migrate.NewMigrations())
					return migrator.Init(c.Context)
				},
			},
			{
				Name:  "reset",
				Usage: "migrate database",
				Action: func(c *cli.Context) error {
					if err := db.ResetModel(c.Context,
						(*models.User)(nil),
						(*models.Role)(nil),
						(*models.UserToRole)(nil),
						(*models.Service)(nil),
						(*models.Permission)(nil),
					); err != nil {
						return err
					}

					fixture := dbfixture.New(db)
					if err := fixture.Load(c.Context, os.DirFS("cmd/migrations/fixtures"), "fixture.yml"); err != nil {
						panic(err)
					}

					return nil
				},
			},
		},
	}
}
