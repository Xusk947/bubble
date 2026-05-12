package initcmd

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime/debug"
	"strings"
)

type Flags struct {
	Module         string
	BubbleModule   string
	BubbleVersion  string
	Dir            string
	Name           string
	DB             string
	Cache          string
	Queue          string
	S3             bool
	SkipGo         bool
	NonInteractive bool
}

const (
	dbNone     = "none"
	dbSQLite   = "sqlite"
	dbPostgres = "postgres"
)

const (
	cacheLocal = "local"
	cacheRedis = "redis"
)

const (
	queueNone  = "none"
	queueNATS  = "nats"
	queueKafka = "kafka"
)

func Run(argv []string, stdout io.Writer, stderr io.Writer) error {
	fs := flag.NewFlagSet("bubble init", flag.ContinueOnError)
	fs.SetOutput(stderr)

	var f Flags
	fs.StringVar(&f.Module, "module", "", "go module path (e.g. github.com/user/app)")
	fs.StringVar(&f.BubbleModule, "bubble-module", "", "bubble module path (e.g. github.com/user/bubble)")
	fs.StringVar(&f.Dir, "dir", "", "target directory")
	fs.StringVar(&f.Name, "name", "", "app name")
	fs.StringVar(&f.DB, "db", dbNone, "database profile: none|sqlite|postgres")
	fs.StringVar(&f.Cache, "cache", cacheLocal, "cache profile: local|redis")
	fs.StringVar(&f.Queue, "queue", queueNone, "queue profile: none|nats|kafka")
	fs.BoolVar(&f.S3, "s3", false, "enable s3 config prompts")
	fs.BoolVar(&f.SkipGo, "skip-go", false, "skip running go commands (go mod init/go get/go mod tidy)")
	fs.BoolVar(&f.NonInteractive, "non-interactive", false, "do not prompt, require flags")

	if err := fs.Parse(argv); err != nil {
		return err
	}

	if strings.TrimSpace(f.BubbleModule) == "" {
		f.BubbleModule, f.BubbleVersion = bubbleBuildInfo()
	}

	if err := resolveFlags(&f, stdout, stderr); err != nil {
		return err
	}

	targetDir := strings.TrimSpace(f.Dir)
	if targetDir == "" {
		targetDir = "./" + f.Name
	}
	targetDir, err := filepath.Abs(targetDir)
	if err != nil {
		return err
	}

	if err := ensureEmptyDir(targetDir); err != nil {
		return err
	}

	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return err
	}

	if err := writeProjectFiles(targetDir, f); err != nil {
		return err
	}

	if f.SkipGo {
		fmt.Fprintf(stdout, "created project: %s\n", targetDir)
		fmt.Fprintf(stdout, "note: go commands were skipped (--skip-go)\n")
		return nil
	}

	if err := goModInit(context.Background(), targetDir, f.Module); err != nil {
		return err
	}

	if err := goDeps(context.Background(), targetDir, f); err != nil {
		return err
	}

	fmt.Fprintf(stdout, "created project: %s\n", targetDir)
	fmt.Fprintf(stdout, "next:\n")
	fmt.Fprintf(stdout, "  cd %s\n", targetDir)
	fmt.Fprintf(stdout, "  go run ./cmd/%s\n", f.Name)
	return nil
}

func resolveFlags(f *Flags, stdout io.Writer, stderr io.Writer) error {
	if f.NonInteractive {
		if strings.TrimSpace(f.Module) == "" || strings.TrimSpace(f.Name) == "" {
			return errors.New("missing required flags: --module and --name")
		}
		if strings.TrimSpace(f.BubbleModule) == "" {
			f.BubbleModule = inferBubbleModuleFromAppModule(f.Module)
		}
		if !isFetchableModule(f.BubbleModule) {
			return errors.New("invalid bubble module value (expected a full module path, e.g. github.com/user/bubble)")
		}
		if err := validateDB(f.DB); err != nil {
			return err
		}
		if err := validateCache(f.Cache); err != nil {
			return err
		}
		if err := validateQueue(f.Queue); err != nil {
			return err
		}
		return nil
	}

	if canUseTUI() {
		res, err := runTUI(*f)
		if err == nil {
			f.Name = res.Name
			f.Module = res.Module
			f.DB = res.DB
			f.BubbleModule = res.BubbleModule
			f.Cache = res.Cache
			f.Queue = res.Queue
			f.S3 = res.S3
			if err := validateDB(f.DB); err != nil {
				return err
			}
			if err := validateCache(f.Cache); err != nil {
				return err
			}
			if err := validateQueue(f.Queue); err != nil {
				return err
			}
			if strings.TrimSpace(f.BubbleModule) == "" {
				return errors.New("bubble module is empty")
			}
			if !isFetchableModule(f.BubbleModule) {
				return errors.New("bubble module is invalid")
			}
			return nil
		}
		if errors.Is(err, errInitCanceled) {
			fmt.Fprintln(stderr, "init canceled")
			return errors.New("invalid init input")
		}
	}

	reader := bufio.NewReader(os.Stdin)

	if strings.TrimSpace(f.Name) == "" {
		fmt.Fprint(stdout, "enter app name: ")
		name, _ := reader.ReadString('\n')
		f.Name = strings.TrimSpace(name)
	}
	if strings.TrimSpace(f.Module) == "" {
		fmt.Fprint(stdout, "enter go module (e.g. github.com/user/app): ")
		module, _ := reader.ReadString('\n')
		f.Module = strings.TrimSpace(module)
	}
	if strings.TrimSpace(f.BubbleModule) == "" {
		f.BubbleModule = inferBubbleModuleFromAppModule(f.Module)
	}
	if strings.TrimSpace(f.BubbleModule) == "" || !isFetchableModule(f.BubbleModule) {
		fmt.Fprint(stdout, "enter bubble module (e.g. github.com/user/bubble): ")
		mod, _ := reader.ReadString('\n')
		f.BubbleModule = strings.TrimSpace(mod)
	}
	if strings.TrimSpace(f.DB) == "" || strings.TrimSpace(f.DB) == dbNone {
		fmt.Fprint(stdout, "enter db (none/sqlite/postgres) [none]: ")
		db, _ := reader.ReadString('\n')
		dbValue := strings.TrimSpace(db)
		if dbValue != "" {
			f.DB = dbValue
		}
	}

	if strings.TrimSpace(f.Name) == "" || strings.TrimSpace(f.Module) == "" {
		fmt.Fprintln(stderr, "init canceled: empty input")
		return errors.New("invalid init input")
	}
	if strings.TrimSpace(f.BubbleModule) == "" || !isFetchableModule(f.BubbleModule) {
		return errors.New("invalid bubble module")
	}
	if err := validateDB(f.DB); err != nil {
		return err
	}
	if err := validateCache(f.Cache); err != nil {
		return err
	}
	if err := validateQueue(f.Queue); err != nil {
		return err
	}
	return nil
}

func goModInit(ctx context.Context, dir string, module string) error {
	cmd := exec.CommandContext(ctx, "go", "mod", "init", module)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func goDeps(ctx context.Context, dir string, f Flags) error {
	bubbleRef := strings.TrimSpace(f.BubbleModule)
	if !isFetchableModule(bubbleRef) {
		return errors.New("bubble module is not fetchable")
	}
	if strings.TrimSpace(f.BubbleVersion) != "" && f.BubbleVersion != "(devel)" {
		bubbleRef = bubbleRef + "@" + f.BubbleVersion
	} else {
		bubbleRef = bubbleRef + "@latest"
	}

	pkgs := make([]string, 0, 4)
	pkgs = append(pkgs, "github.com/gofiber/fiber/v3@latest", bubbleRef)
	switch normalizeDB(f.DB) {
	case dbSQLite:
		pkgs = append(pkgs, "github.com/mattn/go-sqlite3@latest")
	case dbPostgres:
		pkgs = append(pkgs, "github.com/jackc/pgx/v5/stdlib@latest")
	default:
	}

	args := make([]string, 0, 2+len(pkgs))
	args = append(args, "get")
	args = append(args, pkgs...)
	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	tidy := exec.CommandContext(ctx, "go", "mod", "tidy")
	tidy.Dir = dir
	tidy.Stdout = os.Stdout
	tidy.Stderr = os.Stderr
	return tidy.Run()
}

func bubbleBuildInfo() (string, string) {
	if bi, ok := debug.ReadBuildInfo(); ok {
		if strings.TrimSpace(bi.Main.Path) != "" {
			if isFetchableModule(bi.Main.Path) {
				return bi.Main.Path, bi.Main.Version
			}
			return "", bi.Main.Version
		}
	}
	return "", ""
}

func ensureEmptyDir(dir string) error {
	stat, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if stat.IsDir() {
		return errors.New("target directory already exists")
	}
	return errors.New("target path exists and is not a directory")
}

func normalizeDB(value string) string {
	return strings.TrimSpace(strings.ToLower(value))
}

func validateDB(value string) error {
	switch normalizeDB(value) {
	case "", dbNone, dbSQLite, dbPostgres:
		return nil
	default:
		return errors.New("invalid --db value (expected: none|sqlite|postgres)")
	}
}

func normalizeCache(value string) string {
	return strings.TrimSpace(strings.ToLower(value))
}

func validateCache(value string) error {
	switch normalizeCache(value) {
	case "", cacheLocal, cacheRedis:
		return nil
	default:
		return errors.New("invalid --cache value (expected: local|redis)")
	}
}

func normalizeQueue(value string) string {
	return strings.TrimSpace(strings.ToLower(value))
}

func validateQueue(value string) error {
	switch normalizeQueue(value) {
	case "", queueNone, queueNATS, queueKafka:
		return nil
	default:
		return errors.New("invalid --queue value (expected: none|nats|kafka)")
	}
}

func isFetchableModule(path string) bool {
	p := strings.TrimSpace(path)
	if p == "" {
		return false
	}
	first := p
	if i := strings.IndexByte(p, '/'); i >= 0 {
		first = p[:i]
	}
	return strings.Contains(first, ".")
}

func inferBubbleModuleFromAppModule(appModule string) string {
	m := strings.TrimSpace(appModule)
	if m == "" {
		return ""
	}
	parts := strings.Split(m, "/")
	if len(parts) < 2 {
		return ""
	}
	if !strings.Contains(parts[0], ".") {
		return ""
	}
	if parts[0] == "github.com" && strings.EqualFold(parts[1], "xusk947") {
		return "github.com/Xusk947/bubble"
	}
	return parts[0] + "/" + parts[1] + "/bubble"
}
