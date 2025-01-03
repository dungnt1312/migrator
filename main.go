package main

import (
	"bufio"
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"migrator/src"
	"os"
	"strings"
)

func ConnectDB(env map[string]string) *sql.DB {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		env["DB_USERNAME"], env["DB_PASSWORD"], env["DB_HOST"], env["DB_PORT"], env["DB_DATABASE"],
	)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("%v", err)
	}
	if err := db.Ping(); err != nil {
		log.Fatalf("%v", err)
	}

	return db
}

func LoadEnv(currentDir string) map[string]string {
	envMap := make(map[string]string)

	envFile := currentDir + "\\.env"
	file, err := os.Open(envFile)
	if err != nil {
		log.Fatalf(" %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			envMap[key] = value
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("%v", err)
	}

	return envMap
}

func main() {
	// get the command line arguments
	args := os.Args

	params := args[1:]
	if len(params) == 0 {
		fmt.Println("Please provide a command")
		return
	}
	currentDir, err := os.Getwd()
	if err != nil {
		log.Fatalf(" %v", err)
	}

	env := LoadEnv(currentDir)

	db := ConnectDB(env)
	defer db.Close()
	migrator := &src.Migrator{
		DB:         db,
		CurrentDir: currentDir,
	}

	typeCommand := params[0]
	switch typeCommand {
	case "run":
		migrator.Run()
		break
	case "rollback":
		if len(params) < 2 {
			fmt.Println("Please provide a command")
			return
		}
		step := params[1]
		migrator.Rollback(step)
		break
	case "install":
		migrator.Install()
		break
	case "create":
		if len(params) < 2 {
			fmt.Println("Please provide a command")
			return
		}
		name := params[1]
		migrator.Create(name)
		break

	default:
		fmt.Println("Command not found")
	}

}
