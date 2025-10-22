package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	migration "nota/cmd/migrations"
	"nota/service"
	"nota/types"
	"os"
	"os/exec"
	"path/filepath"
	"slices"

	"github.com/anotik/anocore/pkg/config"
	"github.com/anotik/anocore/pkg/logger"
	"github.com/anotik/anocore/pkg/util"
	"github.com/swaggo/swag/gen"
)

func run(args []string, mod string) error {

	cmdArgs := []string{
		"compose",
		"--env-file", ".env",
		"-f",
		"docker/docker-compose.yml",
		"up",
		"-d",
		"--build",
	}
	if mod == "dev" {

		cmdArgs = []string{
			"compose",
			"--env-file", ".env",
			"-f",
			"docker/docker-compose.dev.yml",
			"up",
		}
	}
	if mod == "db" {
		cmdArgs = []string{
			"compose",
			"--env-file", ".env",
			"-f",
			"docker/docker-compose.db.yml",
			"up",
			"-d",
		}
	}

	cmdArgs = append(cmdArgs, args...)
	cmd := exec.Command("docker", cmdArgs...)
	cmd.Env = os.Environ()
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Error running docker compose up: %v", err)
	}
	return nil
}
func build(args []string, mod string) error {

	cmdArgs := []string{
		"compose",
		"--env-file", ".env",
		"-f",
		"docker/docker-compose.yml",
		"build",
	}
	if mod == "base" {
		cmdArgs = []string{
			"compose",
			"--env-file", ".env",
			"-f",
			"docker/docker-compose.base.yml",
			"build",
		}
	} else if mod == "dev" {
		cmdArgs = []string{
			"compose",
			"--env-file", ".env",
			"-f",
			"docker/docker-compose.dev.yml",
			"build",
		}
	}
	fmt.Println(cmdArgs)

	cmdArgs = append(cmdArgs, args...)
	cmd := exec.Command("docker", cmdArgs...)
	cmd.Env = os.Environ()

	cmd.Env = append(cmd.Env, "DOCKER_BUILDKIT=1")

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Error running docker compose up: %v", err)
	}
	return nil
}

func publish(args []string, name string) error {

	registryUrl := "registry.gust.edu.kw:5002/microservices"
	commandArgs := [][]string{}

	baseUrl := fmt.Sprintf("%s/custom/%s-base:%s", registryUrl, name, "latest")
	buildCmdArgs := []string{"build", "-t", baseUrl, "-f", "docker/Dockerfile.base", "."}
	pushCmdArgs := []string{"push", baseUrl}
	commandArgs = append(commandArgs, buildCmdArgs)
	commandArgs = append(commandArgs, pushCmdArgs)

	if slices.Contains(args, "full") {
		fmt.Println("full")

		mainImageUrl := fmt.Sprintf("%s/%s:%s", registryUrl, name, "latest")

		mainBuildCmd := []string{"build", "-t", mainImageUrl, "-f", "docker/Dockerfile", "."}
		mainPushCmd := []string{"push", mainImageUrl}
		commandArgs = append(commandArgs, mainBuildCmd)
		commandArgs = append(commandArgs, mainPushCmd)

	}

	for _, cmdArgs := range commandArgs {

		cmd := exec.Command("docker", cmdArgs...)
		cmd.Env = os.Environ()

		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("Error running: %v  %v", cmdArgs, err)
		}

	}
	return nil
}

func setupHooks(hookType util.HookType) error {
	if err := util.SetupGitHook(hookType); err != nil {
		return fmt.Errorf("Error setting up pre-push hook: %v", err)
	}

	if err := util.SetupGitHook(util.PreCommit); err != nil {
		return fmt.Errorf("Error setting up pre-commit hook: %v", err)
	}

	fmt.Println("✅ Git hooks installed successfully!")
	fmt.Println("The pre-push and pre-commit hooks are now configured.")
	return nil
}

func runHook(hookType util.HookType) error {
	if err := setupHooks(hookType); err != nil {
		return err
	}
	moduleRoot, err := util.FindModuleRoot()
	if err != nil {
		moduleRoot = "./"
	}

	hookPath := filepath.Join(moduleRoot, ".git", "hooks", string(hookType))
	if _, err := os.Stat(hookPath); os.IsNotExist(err) {
		return fmt.Errorf("❌ %s hook not found. Run 'setup-hooks' first.", hookType)
	}

	hookCmd := exec.Command(hookPath)
	hookCmd.Stdout = os.Stdout
	hookCmd.Stderr = os.Stderr
	hookCmd.Dir = moduleRoot

	hookCmd.Run()
	if hookCmd.ProcessState != nil {
		return fmt.Errorf("hook exited with code %d", hookCmd.ProcessState.ExitCode())
	}
	return nil
}

func generateDocs(searchPath string) error {
	err := gen.New().Build(&gen.Config{
		SearchDir:         searchPath,
		MainAPIFile:       "main.go",
		OutputDir:         "docs",
		OutputTypes:       []string{"json"},
		ParseDependency:   1,
		RequiredByDefault: false,
		ParseDepth:        1,
		ParseGoList:       false,
		ParseFuncBody:     true,
	})
	if err != nil {
		return fmt.Errorf("Error generating docs: %v", err)
	}

	swaggerFile := filepath.Join("docs", "swagger.json")
	data, err := os.ReadFile(swaggerFile)
	if err != nil {
		return fmt.Errorf("Error reading swagger.json: %v", err)
	}

	var swagger map[string]interface{}
	if err := json.Unmarshal(data, &swagger); err != nil {
		return fmt.Errorf("Error parsing swagger.json: %v", err)
	}

	if info, ok := swagger["info"].(map[string]interface{}); ok {
		log, _ := logger.NewContextLogger(context.Background())
		log.Info("APP_VERSION", "version", os.Getenv("APP_VERSION"))
		info["version"] = os.Getenv("APP_VERSION")
	}

	updatedData, err := json.MarshalIndent(swagger, "", "  ")
	if err != nil {
		return fmt.Errorf("Error marshaling updated swagger.json: %v", err)
	}

	if err := os.WriteFile(swaggerFile, updatedData, 0644); err != nil {
		return fmt.Errorf("Error writing updated swagger.json: %v", err)
	}

	fmt.Println("✅ Swagger documentation generated successfully!")
	return nil
}

func printHelp() {
	fmt.Println("Available commands:")
	fmt.Println("  start           - Start the service in production mode")
	fmt.Println("  start-db        - Start the database in development mode")
	fmt.Println("  dev             - Start the service in development mode")
	fmt.Println("  build           - Build the production Docker image")
	fmt.Println("  build-dev       - Build the development Docker image")
	fmt.Println("  publish         - Publish Docker images to registry")
	fmt.Println("  setup-hooks     - Setup git hooks (pre-commit and pre-push)")
	fmt.Println("  run-commit      - Run pre-commit hook")
	fmt.Println("  run-push        - Run pre-push hook")
	fmt.Println("  gen-docs        - Generate Swagger documentation")
	fmt.Println("  migrate             - Run database migrations (default: run all pending migrations)")
	fmt.Println("    migrate create    - Create a new migration file")
	fmt.Println("    migrate up        - Run pending migrations")
	fmt.Println("    migrate down      - Rollback the last migration")
	fmt.Println("    migrate status    - Show migration status")
	fmt.Println("    migrate version   - Show current migration version")
	fmt.Println("    migrate reset     - Reset all migrations")
	fmt.Println("  seed                - Seed functionality not implemented for this service")
	fmt.Println("  help            - Show this help message")
}

func start(name string, path string) error {
	if len(os.Args) < 2 {
		ctx := context.Background()
		log, err := logger.NewContextLogger(ctx)
		if err != nil {
			panic(fmt.Sprintf("Failed to create context logger: %v", err))
		}

		root, err := util.FindModuleRoot()
		if err != nil {
			root = "./"
		}
		envPath := filepath.Join(root, ".env")
		cfg, err := config.LoadEnv[types.AppConfig](envPath)
		if err != nil {
			log.Error("Unable to load config", "error", err)
			return err
		}
		if err := service.Run(ctx, cfg, os.Args, os.Stdin, os.Stdout, os.Stderr); err != nil {
			log.Error("Failed to start Service", "error", err)
			return err
		}
		return nil
	} else {
		command := os.Args[1]
		args := os.Args[2:]

		if len(os.Args) < 2 {
			return run(args, "prod")
		}
		switch command {
		case "start":
			return run(args, "prod")
		case "dev":
			if err := generateDocs(path); err != nil {
				return err
			}
			return run(args, "dev")
		case "build":
			return build(args, "prod")
		case "build-dev":
			return build(args, "dev")
		case "publish":
			return publish(args, name)
		case "setup-hooks":
			if err := setupHooks(util.PreCommit); err != nil {
				return err
			}
			if err := setupHooks(util.PrePush); err != nil {
				return err
			}
			return nil
		case "run-commit":
			return runHook(util.PreCommit)
		case "run-push":
			return runHook(util.PrePush)
		case "gen-docs":
			return generateDocs(path)
		case "start-db":
			return run(args, "db")
		case "migrate":
			return migrate(args)
		case "seed":
			fmt.Println("Seed functionality is not implemented for this service.")
			return nil
		case "help":
			printHelp()
			return nil
		default:
			log, _ := logger.NewContextLogger(context.Background())
			log.Error("Unknown command", "command", command)

			printHelp()
			return fmt.Errorf("unknown command: %s", command)
		}
	}
}

func migrate(args []string) error {
	root, err := util.FindModuleRoot()
	if err != nil {
		root = "./"
	}
	migrationsDir := filepath.Join(root, "service", "db", "migrations")
	envPath := filepath.Join(root, ".env")
	cfg, err := config.LoadEnv[types.AppConfig](envPath)
	if err != nil {
		return err
	}

	dbUrl := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		url.QueryEscape(cfg.DBUserAdmin),
		url.QueryEscape(cfg.DBPasswordAdmin),
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBName,
	)

	if len(args) == 0 {
		if err := migration.RunMigrations(dbUrl, migrationsDir); err != nil {
			log, _ := logger.NewContextLogger(context.Background())
			log.Error("Error running migrations", "error", err)
			return err
		}
		return nil
	}

	command := args[0]
	switch command {
	case "create":
		if len(args) < 2 {
			log, _ := logger.NewContextLogger(context.Background())
			log.Error("Migration name is required")
			return fmt.Errorf("migration name is required")
		}
		if err := migration.CreateMigration(args[1], migrationsDir); err != nil {
			log, _ := logger.NewContextLogger(context.Background())
			log.Error("Error creating migration", "error", err)
			return err
		}
	case "up":
		if err := migration.MigrateUp(dbUrl, migrationsDir); err != nil {
			log, _ := logger.NewContextLogger(context.Background())
			log.Error("Error running up migration", "error", err)
			return err
		}
	case "down":
		if err := migration.MigrateDown(dbUrl, migrationsDir); err != nil {
			log, _ := logger.NewContextLogger(context.Background())
			log.Error("Error running down migration", "error", err)
			return err
		}
	case "status":
		if err := migration.MigrateStatus(dbUrl, migrationsDir); err != nil {
			log, _ := logger.NewContextLogger(context.Background())
			log.Error("Error getting migration status", "error", err)
			return err
		}
	case "version":
		if err := migration.MigrateVersion(dbUrl, migrationsDir); err != nil {
			log, _ := logger.NewContextLogger(context.Background())
			log.Error("Error getting migration version", "error", err)
			return err
		}
	case "reset":
		if err := migration.ResetMigrations(dbUrl, migrationsDir); err != nil {
			log, _ := logger.NewContextLogger(context.Background())
			log.Error("Error resetting migrations", "error", err)
			return err
		}
	default:
		log, _ := logger.NewContextLogger(context.Background())
		log.Error("Unknown migration command", "command", command)
		printHelp()
		return fmt.Errorf("unknown migration command: %s", command)
	}
	return nil
}

func main() {
	err := start("anotify", "./service")
	if err != nil {
		os.Exit(1)
	}
}
