package src

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
	"time"
)

type Migrator struct {
	DB         *sql.DB
	CurrentDir string
}
type Migration struct {
	Name  string
	Batch int
}

type MigratorInterface interface {
	ScanDir(dir string) ([]Migration, error)
	Run()
	Rollback(step string) error
	Install()
}

func (migrator *Migrator) ScanDir(dir string) ([]Migration, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var migrations []Migration
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".up.sql") {
			migrations = append(migrations, Migration{
				Name: file.Name(),
			})
		}
	}

	return migrations, nil
}

func (migrator *Migrator) Run() {
	db := migrator.DB

	migrations, err := migrator.ScanDir("migrations")
	if err != nil {
		fmt.Println("Lỗi khi tải migrations: ", err)
		return
	}

	for _, migration := range migrations {

		if !strings.Contains(migration.Name, ".up.sql") {
			continue
		}

		name := strings.TrimSuffix(migration.Name, ".up.sql")
		// Kiểm tra migration đã chạy chưa
		if CheckMigration(migrator, name) {
			continue
		}

		filePath := path.Join(migrator.CurrentDir+"\\migrations", migration.Name)
		content, err := ioutil.ReadFile(filePath)
		if err != nil {
			return
		}

		_, err = db.Exec(string(content))
		if err != nil {
			return
		}

		fmt.Println("Migrate: ", filePath)
		_, err = db.Exec("INSERT INTO migrations (name) VALUES (?)", name)
		if err != nil {
			log.Fatalf("Không thể lưu migration: %v", err)
			return
		}
	}

}

func (migrator *Migrator) Rollback(step string) {
	db := migrator.DB
	// get 3 step desc from migrations table

	rows, err := db.Query("SELECT name FROM migrations ORDER BY id DESC LIMIT 3")
	if err != nil {
		log.Fatalf("Không thể lấy migrations: %v", err)
		return
	}
	defer rows.Close()

	var migrations []Migration
	for rows.Next() {
		var migration Migration
		err := rows.Scan(&migration.Name)
		if err != nil {
			log.Fatalf("Không thể quét migrations: %v", err)
			return
		}
		migrations = append(migrations, migration)
	}

	// rollback 3 step
	for _, migration := range migrations {
		filePath := path.Join("migrations", migration.Name+".down.sql")
		content, err := ioutil.ReadFile(filePath)
		if err != nil {
			continue
		}

		_, err = db.Exec(string(content))
		if err != nil {
			log.Fatalf("Không thể rollback migration: %v", err)
			return
		}

		fmt.Println("Rollback: ", filePath)
		_, err = db.Exec("DELETE FROM migrations WHERE name = ?", migration.Name)
		if err != nil {
			log.Fatalf("Không thể xóa migration: %v", err)
			return
		}

	}

}

func (migrator *Migrator) Install() {
	migrationDir := migrator.CurrentDir + "\\migrations"
	if _, err := os.Stat(migrationDir); os.IsNotExist(err) {
		err := os.Mkdir(migrationDir, os.ModePerm)
		if err != nil {
			log.Fatalf(" %v", err)
		}
		fmt.Println("Create migration directory: ", migrationDir)

	}

	db := migrator.DB

	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS migrations (
    		id INT AUTO_INCREMENT PRIMARY KEY,
    		name VARCHAR(255) NOT NULL UNIQUE,
    		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    	)`)
	if err != nil {
		log.Fatalf("Không thể tạo bảng migrations: %v", err)
	}

}

func CheckMigration(migrator *Migrator, name string) bool {
	db := migrator.DB
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM migrations WHERE name = ?", name).Scan(&count)
	if err != nil {
		log.Fatalf("Không thể kiểm tra migration: %v", err)
		return false
	}
	return count > 0
}

func (migrator *Migrator) Create(name string) {
	currentTimeStamp := time.Now().Unix()
	upFileName := fmt.Sprintf("migrations/%d_%s.up.sql", currentTimeStamp, name)
	downFileName := fmt.Sprintf("migrations/%d_%s.down.sql", currentTimeStamp, name)
	// create file
	upFile, err := os.Create(upFileName)
	if err != nil {
		log.Fatalf("Không thể tạo file: %v", err)
	}
	defer upFile.Close()

	downFile, err := os.Create(downFileName)
	if err != nil {
		log.Fatalf("Không thể tạo file: %v", err)
	}
	defer downFile.Close()

}
