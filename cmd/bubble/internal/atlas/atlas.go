package atlas

import (
	"context"
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/Xusk947/bubble/cmd/bubble/internal/usage"
	"github.com/Xusk947/bubble/config"
	"github.com/Xusk947/bubble/database"
)

const defaultEntToURL = "ent://./ent/schema"

type commonFlags struct {
	DotenvPath string
	AtlasBin   string
}

func Run(argv []string, stdout io.Writer, stderr io.Writer) error {
	if len(argv) < 1 {
		usage.Print(stderr)
		return nil
	}

	switch argv[0] {
	case "diff":
		return runDiff(context.Background(), argv[1:], stdout, stderr)
	case "apply":
		return runApply(context.Background(), argv[1:], stdout, stderr)
	case "help", "-h", "--help":
		usage.Print(stdout)
		return nil
	default:
		fmt.Fprintf(stderr, "unknown atlas command: %s\n", argv[0])
		usage.Print(stderr)
		return nil
	}
}

type diffFlags struct {
	Common  commonFlags
	Name    string
	ToURL   string
	DevURL  string
	DirPath string
	Format  string
}

func runDiff(ctx context.Context, argv []string, stdout io.Writer, stderr io.Writer) error {
	fs := flag.NewFlagSet("bubble atlas diff", flag.ContinueOnError)
	fs.SetOutput(stderr)

	var f diffFlags
	addCommonFlags(fs, &f.Common)
	fs.StringVar(&f.Name, "name", "", "migration name")
	fs.StringVar(&f.ToURL, "to", defaultEntToURL, "to url (e.g. ent://./ent/schema)")
	fs.StringVar(&f.DevURL, "dev-url", "", "dev database url (required)")
	fs.StringVar(&f.DirPath, "dir", "", "migrations directory")
	fs.StringVar(&f.Format, "format", "", "atlas --format template")

	if err := fs.Parse(argv); err != nil {
		return err
	}
	if strings.TrimSpace(f.Name) == "" || strings.TrimSpace(f.DevURL) == "" {
		usage.Print(stderr)
		return nil
	}

	cfg, err := loadConfig(f.Common.DotenvPath)
	if err != nil {
		return err
	}

	dir := resolveMigrationsDir(strings.TrimSpace(f.DirPath), cfg)
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return err
	}

	args := make([]string, 0, 12)
	args = append(args, "migrate", "diff", f.Name)
	args = append(args, "--to", f.ToURL)
	args = append(args, "--dev-url", f.DevURL)
	args = append(args, "--dir", "file://"+absDir)
	if strings.TrimSpace(f.Format) != "" {
		args = append(args, "--format", f.Format)
	}

	return execAtlas(ctx, f.Common.AtlasBin, args, stdout, stderr)
}

type applyFlags struct {
	Common  commonFlags
	URL     string
	DirPath string
}

func runApply(ctx context.Context, argv []string, stdout io.Writer, stderr io.Writer) error {
	fs := flag.NewFlagSet("bubble atlas apply", flag.ContinueOnError)
	fs.SetOutput(stderr)

	var f applyFlags
	addCommonFlags(fs, &f.Common)
	fs.StringVar(&f.URL, "url", "", "database url for atlas migrate apply")
	fs.StringVar(&f.DirPath, "dir", "", "migrations directory")

	if err := fs.Parse(argv); err != nil {
		return err
	}

	cfg, err := loadConfig(f.Common.DotenvPath)
	if err != nil {
		return err
	}

	dir := resolveMigrationsDir(strings.TrimSpace(f.DirPath), cfg)
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return err
	}

	url := strings.TrimSpace(f.URL)
	if url == "" {
		url, err = database.ResolveAtlasURL(cfg.Database)
		if err != nil {
			return err
		}
	}

	args := make([]string, 0, 10)
	args = append(args, "migrate", "apply")
	args = append(args, "--url", url)
	args = append(args, "--dir", "file://"+absDir)

	return execAtlas(ctx, f.Common.AtlasBin, args, stdout, stderr)
}

func addCommonFlags(fs *flag.FlagSet, common *commonFlags) {
	fs.StringVar(&common.DotenvPath, "dotenv", "", "dotenv file path")
	fs.StringVar(&common.AtlasBin, "atlas-bin", "atlas", "atlas binary path")
}

func loadConfig(dotenvPath string) (config.Config, error) {
	if strings.TrimSpace(dotenvPath) != "" {
		return config.Load(config.WithDotenvPath(dotenvPath))
	}
	return config.Load()
}

func resolveMigrationsDir(flagValue string, cfg config.Config) string {
	dir := strings.TrimSpace(flagValue)
	if dir == "" {
		dir = strings.TrimSpace(cfg.Database.Migrations.Dir)
	}
	if dir == "" {
		dir = "./migrations"
	}
	return dir
}
