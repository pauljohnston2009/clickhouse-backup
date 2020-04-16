package main

import (
	"fmt"
	"log"
	"os"
	"io"

	"github.com/AlexAkulov/clickhouse-backup/pkg/chbackup"

	"github.com/urfave/cli"

	"net/http"
	"github.com/julienschmidt/httprouter"
)

const (
	defaultConfigPath = "/etc/clickhouse-backup/config.yml"
)

var (
	version   = "unknown"
	gitCommit = "unknown"
	buildDate = "unknown"
)

func attachConfig(h func(c *cli.Context, w http.ResponseWriter, r *http.Request, ps httprouter.Params) error, c *cli.Context) httprouter.Handle {
  return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
        err := h(c, w, r, ps)

        fmt.Println(err)
        if err != nil {
            str := err.Error()
            http.Error(w, str, 500)
            fmt.Println(str)
        }
  	}
}

func create(c *cli.Context, w http.ResponseWriter, r *http.Request, ps httprouter.Params) error {
    var backupName = ps.ByName("backupName")
    return chbackup.CreateBackup(*getConfig(c), backupName, c.String("t"), false)
}

func restore(c *cli.Context, w http.ResponseWriter, r *http.Request, ps httprouter.Params) error {
    var backupName = ps.ByName("backupName")
    return chbackup.Restore(*getConfig(c), backupName, c.String("t"), c.Bool("s"), c.Bool("d"))
}

func delete(c *cli.Context, w http.ResponseWriter, r *http.Request, ps httprouter.Params) error {
    var serverType = ps.ByName("serverType")
    var backupName = ps.ByName("backupName")
    return deleteBackup(c, serverType, backupName)
}

func upload(c *cli.Context, w http.ResponseWriter, r *http.Request, ps httprouter.Params) error {
    var backupName = ps.ByName("backupName")
    return chbackup.Upload(*getConfig(c), backupName, c.String("diff-from"))
}

func download(c *cli.Context, w http.ResponseWriter, r *http.Request, ps httprouter.Params) error {
    var backupName = ps.ByName("backupName")
    fmt.Println(backupName)
    return chbackup.Download(*getConfig(c), backupName)
}

func uploadWithDiff(c *cli.Context, w http.ResponseWriter, r *http.Request, ps httprouter.Params) error {
    var backupName = ps.ByName("backupName")
    var diffFrom = ps.ByName("diffFrom")
    return chbackup.Upload(*getConfig(c), backupName, diffFrom)
}

func freeze(c *cli.Context, w http.ResponseWriter, r *http.Request, ps httprouter.Params) error {
    return chbackup.Freeze(*getConfig(c), c.String("t"))
}

func tables(c *cli.Context, w http.ResponseWriter, r *http.Request, ps httprouter.Params) error {
    return chbackup.PrintTables(*getConfig(c), w)
}

func clean(c *cli.Context, w http.ResponseWriter, r *http.Request, ps httprouter.Params) error {
    return chbackup.Clean(*getConfig(c))
}

func isClean(c *cli.Context, w http.ResponseWriter, r *http.Request, ps httprouter.Params) error {
    return chbackup.IsClean(*getConfig(c), w)
}

func list(c *cli.Context, w http.ResponseWriter, r *http.Request, ps httprouter.Params) error {
    var serverType = ps.ByName("serverType")
    var format = ps.ByName("format")
    return listBackups(c, serverType, format, w)
}

func listAll(c *cli.Context, w http.ResponseWriter, r *http.Request, ps httprouter.Params) error {
    var serverType = ps.ByName("serverType")
    return listBackups(c, serverType, "", w)
}

func listBackups(c *cli.Context, serverType string, format string, w io.Writer) error {
    config := getConfig(c)
    switch serverType {
    case "local":
        return chbackup.PrintLocalBackups(*config, format, w)
    case "remote":
        return chbackup.PrintRemoteBackups(*config, format, w)
    case "all", "":
        fmt.Println("Local backups:")
        if err := chbackup.PrintLocalBackups(*config, format, w); err != nil {
            return err
        }
        fmt.Println("Remote backups:")
        if err := chbackup.PrintRemoteBackups(*config, format, w); err != nil {
            return err
        }
    default:
        fmt.Fprintf(os.Stderr, "Unknown command '%s'\n", serverType)
        cli.ShowCommandHelpAndExit(c, c.Command.Name, 1)
    }
    return nil
}

func deleteBackup(c *cli.Context, serverType string, backupName string) error {
    config := getConfig(c)
    switch serverType {
    case "local":
        return chbackup.RemoveBackupLocal(*config, backupName)
    case "remote":
        return chbackup.RemoveBackupRemote(*config, backupName)
    default:
        fmt.Fprintf(os.Stderr, "Unknown command '%s'\n", serverType)
        cli.ShowCommandHelpAndExit(c, c.Command.Name, 1)
    }
    return nil
}

func getConfigAndRun(c *cli.Context) error {
	router := httprouter.New()
    router.POST("/create/:backupName", attachConfig(create, c))
    router.POST("/upload/:backupName", attachConfig(upload, c))
    router.POST("/download/:backupName", attachConfig(download, c))
    router.POST("/upload/:backupName/:diffFrom", attachConfig(uploadWithDiff, c))
    router.POST("/freeze", attachConfig(freeze, c))
    router.GET("/tables", attachConfig(tables, c))
    router.GET("/list/:serverType/:format", attachConfig(list, c))
    router.GET("/list/:serverType", attachConfig(listAll, c))
    router.POST("/restore/:backupName", attachConfig(restore, c))
    router.POST("/delete/:serverType/:backupName", attachConfig(delete, c))
    router.POST("/clean", attachConfig(clean, c))
    router.GET("/is-clean", attachConfig(isClean, c))
    // todo check for empty shadow dir so we can check that the last backup ran fine, and someone else is not in teh middle of making one

    return http.ListenAndServe(":8123", router)
}

func main() {

    log.SetOutput(os.Stdout)

    cliapp := cli.NewApp()
	cliapp.Name = "clickhouse-backup"
	cliapp.Usage = "Tool for easy backup of ClickHouse with cloud support"
	cliapp.UsageText = "clickhouse-backup <command> [-t, --tables=<db>.<table>] <backup_name>"
	cliapp.Description = "Run as 'root' or 'clickhouse' user"
	cliapp.Version = version

	cliapp.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "config, c",
			Value:  defaultConfigPath,
			Usage:  "Config `FILE` name.",
			EnvVar: "CLICKHOUSE_BACKUP_CONFIG",
		},
	}
	cliapp.CommandNotFound = func(c *cli.Context, command string) {
		fmt.Printf("Error. Unknown command: '%s'\n\n", command)
		cli.ShowAppHelpAndExit(c, 1)
	}

	cli.VersionPrinter = func(c *cli.Context) {
		fmt.Println("Version:\t", c.App.Version)
		fmt.Println("Git Commit:\t", gitCommit)
		fmt.Println("Build Date:\t", buildDate)
	}

	cliapp.Commands = []cli.Command{
		{
			Name:      "tables",
			Usage:     "Print list of tables",
			UsageText: "clickhouse-backup tables",
			Action: func(c *cli.Context) error {
				return chbackup.PrintTables(*getConfig(c), os.Stdout)
			},
			Flags: cliapp.Flags,
		},
		{
			Name:        "create",
			Usage:       "Create new backup",
			UsageText:   "clickhouse-backup create [-t, --tables=<db>.<table>] <backup_name>",
			Description: "Create new backup",
			Action: func(c *cli.Context) error {
				return chbackup.CreateBackup(*getConfig(c), c.Args().First(), c.String("t"), false)
			},
			Flags: append(cliapp.Flags,
				cli.StringFlag{
					Name:   "table, tables, t",
					Hidden: false,
				},
			),
		},
		{
			Name:      "upload",
			Usage:     "Upload backup to remote storage",
			UsageText: "clickhouse-backup upload [--diff-from=<backup_name>] <backup_name>",
			Action: func(c *cli.Context) error {
				return chbackup.Upload(*getConfig(c), c.Args().First(), c.String("diff-from"))
			},
			Flags: append(cliapp.Flags,
				cli.StringFlag{
					Name:   "diff-from",
					Hidden: false,
				},
			),
		},
		{
			Name:      "list",
			Usage:     "Print list of backups",
			UsageText: "clickhouse-backup list [all|local|remote] [latest|penult]",
			Action: func(c *cli.Context) error {
				return listBackups(c, c.Args().Get(0), c.Args().Get(1), os.Stdout)
			},
			Flags: cliapp.Flags,
		},
		{
			Name:      "download",
			Usage:     "Download backup from remote storage",
			UsageText: "clickhouse-backup download <backup_name>",
			Action: func(c *cli.Context) error {
				return chbackup.Download(*getConfig(c), c.Args().First())
			},
			Flags: cliapp.Flags,
		},
		{
			Name:      "restore",
			Usage:     "Create schema and restore data from backup",
			UsageText: "clickhouse-backup restore [--schema] [--data] [-t, --tables=<db>.<table>] <backup_name>",
			Action: func(c *cli.Context) error {
				return chbackup.Restore(*getConfig(c), c.Args().First(), c.String("t"), c.Bool("s"), c.Bool("d"))
			},
			Flags: append(cliapp.Flags,
				cli.StringFlag{
					Name:   "table, tables, t",
					Hidden: false,
				},
				cli.BoolFlag{
					Name:   "schema, s",
					Hidden: false,
					Usage:  "Restore schema only",
				},
				cli.BoolFlag{
					Name:   "data, d",
					Hidden: false,
					Usage:  "Restore data only",
				},
			),
		},
		{
			Name:      "delete",
			Usage:     "Delete specific backup",
			UsageText: "clickhouse-backup delete <local|remote> <backup_name>",
			Action: func(c *cli.Context) error {
				if c.Args().Get(1) == "" {
					fmt.Fprintln(os.Stderr, "Backup name must be defined")
					cli.ShowCommandHelpAndExit(c, c.Command.Name, 1)
				}
				return deleteBackup(c, c.Args().Get(0), c.Args().Get(1))
			},
			Flags: cliapp.Flags,
		},
		{
			Name:  "default-config",
			Usage: "Print default config",
			Action: func(*cli.Context) {
				chbackup.PrintDefaultConfig()
			},
			Flags: cliapp.Flags,
		},
		{
			Name:        "freeze",
			Usage:       "Freeze tables",
			UsageText:   "clickhouse-backup freeze [-t, --tables=<db>.<table>] <backup_name>",
			Description: "Freeze tables",
			Action: func(c *cli.Context) error {
				return chbackup.Freeze(*getConfig(c), c.String("t"))
			},
			Flags: append(cliapp.Flags,
				cli.StringFlag{
					Name:   "table, tables, t",
					Hidden: false,
				},
			),
		},
		{
			Name:  "clean",
			Usage: "Remove data in 'shadow' folder",
			Action: func(c *cli.Context) error {
				return chbackup.Clean(*getConfig(c))
			},
			Flags: cliapp.Flags,
		},
		{
			Name:  "isclean",
			Usage: "Checks if the shadow dir is clean",
			Action: func(c *cli.Context) error {
				return chbackup.IsClean(*getConfig(c), os.Stdout)
			},
			Flags: cliapp.Flags,
		},
		{
			Name:  "serve",
			Usage: "Starts http server for handling backup commands",
			Action: func(c *cli.Context) error {
				return getConfigAndRun(c)
			},
			Flags: cliapp.Flags,
		},
	}
	if err := cliapp.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func getConfig(ctx *cli.Context) *chbackup.Config {
	configPath := ctx.String("config")
	if configPath == defaultConfigPath {
		configPath = ctx.GlobalString("config")
	}

	config, err := chbackup.LoadConfig(configPath)
	if err != nil {
		log.Fatal(err)
	}

	return config
}
