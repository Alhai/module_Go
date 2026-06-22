package main

import (
	"database/sql"
	"log"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// User représente un utilisateur dans la base de données
type User struct {
	ID    int    `db:"id" gorm:"primaryKey;autoIncrement"`
	Name  string `db:"name"`
	Email string `db:"email"`
}

func initSQL() *sql.DB {
	db, err := sql.Open("sqlite3", "./test.db")
	if err != nil {
		log.Fatal(err)
	}

	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}
	log.Println("Connecté à SQLite (database/sql)")

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS users (
		id    INTEGER PRIMARY KEY AUTOINCREMENT,
		name  TEXT NOT NULL,
		email TEXT UNIQUE NOT NULL
	)`)
	if err != nil {
		log.Fatal("Création table users:", err)
	}
	log.Println("Table users prête (database/sql)")
	return db
}

func CreateUserSQL(db *sql.DB, user User) (int64, error) {
	result, err := db.Exec(
		"INSERT INTO users (name, email) VALUES (?, ?)",
		user.Name, user.Email,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func GetUsersSQL(db *sql.DB) ([]User, error) {
	rows, err := db.Query("SELECT id, name, email FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Name, &u.Email); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func GetUserByIDSQL(db *sql.DB, id int) (*User, error) {
	var u User
	err := db.QueryRow(
		"SELECT id, name, email FROM users WHERE id = ?", id,
	).Scan(&u.ID, &u.Name, &u.Email)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func UpdateUserSQL(db *sql.DB, user User) error {
	_, err := db.Exec(
		"UPDATE users SET name = ?, email = ? WHERE id = ?",
		user.Name, user.Email, user.ID,
	)
	return err
}

func DeleteUserSQL(db *sql.DB, id int) error {
	_, err := db.Exec("DELETE FROM users WHERE id = ?", id)
	return err
}

func CreateUserSQLX(db *sqlx.DB, user User) (int64, error) {
	result, err := db.Exec(
		"INSERT INTO users (name, email) VALUES (?, ?)",
		user.Name, user.Email,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func GetUsersSQLX(db *sqlx.DB) ([]User, error) {
	var users []User
	err := db.Select(&users, "SELECT id, name, email FROM users")
	return users, err
}

func GetUserByIDSQLX(db *sqlx.DB, id int) (*User, error) {
	var u User
	err := db.Get(&u, "SELECT id, name, email FROM users WHERE id = ?", id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func UpdateUserSQLX(db *sqlx.DB, user User) error {
	_, err := db.Exec(
		"UPDATE users SET name = ?, email = ? WHERE id = ?",
		user.Name, user.Email, user.ID,
	)
	return err
}

func DeleteUserSQLX(db *sqlx.DB, id int) error {
	_, err := db.Exec("DELETE FROM users WHERE id = ?", id)
	return err
}

func CreateUserGORM(db *gorm.DB, user User) error {
	return db.Create(&user).Error
}

func GetUsersGORM(db *gorm.DB) ([]User, error) {
	var users []User
	result := db.Find(&users)
	return users, result.Error
}

func GetUserByIDGORM(db *gorm.DB, id int) (*User, error) {
	var u User
	result := db.First(&u, id)
	if result.Error == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if result.Error != nil {
		return nil, result.Error
	}
	return &u, nil
}

func UpdateUserGORM(db *gorm.DB, user User) error {
	return db.Save(&user).Error
}

func DeleteUserGORM(db *gorm.DB, id int) error {
	return db.Delete(&User{}, id).Error
}

func main() {
	// --- database/sql ---
	db := initSQL()
	defer db.Close()

	// --- CRUD database/sql ---
	id1, err := CreateUserSQL(db, User{Name: "Alice", Email: "alice@example.com"})
	if err != nil {
		log.Println("CreateUserSQL:", err)
	} else {
		log.Println("CreateUserSQL - nouvel ID:", id1)
	}

	id2, err := CreateUserSQL(db, User{Name: "Bob", Email: "bob@example.com"})
	if err != nil {
		log.Println("CreateUserSQL:", err)
	} else {
		log.Println("CreateUserSQL - nouvel ID:", id2)
	}

	users, err := GetUsersSQL(db)
	if err != nil {
		log.Println("GetUsersSQL:", err)
	} else {
		log.Println("GetUsersSQL:", users)
	}

	user, err := GetUserByIDSQL(db, int(id1))
	if err != nil {
		log.Println("GetUserByIDSQL:", err)
	} else {
		log.Println("GetUserByIDSQL:", user)
	}

	err = UpdateUserSQL(db, User{ID: int(id1), Name: "Alice Modifiée", Email: "alice.mod@example.com"})
	if err != nil {
		log.Println("UpdateUserSQL:", err)
	} else {
		log.Println("UpdateUserSQL: OK")
	}

	err = DeleteUserSQL(db, int(id2))
	if err != nil {
		log.Println("DeleteUserSQL:", err)
	} else {
		log.Println("DeleteUserSQL: OK")
	}

	users, err = GetUsersSQL(db)
	if err != nil {
		log.Println("GetUsersSQL final:", err)
	} else {
		log.Println("GetUsersSQL final:", users)
	}

	// --- sqlx ---
	dbx, err := sqlx.Open("sqlite3", "./test.db")
	if err != nil {
		log.Fatal(err)
	}
	defer dbx.Close()
	log.Println("Connecté à SQLite (sqlx)")

	idx1, err := CreateUserSQLX(dbx, User{Name: "Charlie", Email: "charlie@example.com"})
	if err != nil {
		log.Println("CreateUserSQLX:", err)
	} else {
		log.Println("CreateUserSQLX - nouvel ID:", idx1)
	}

	idx2, err := CreateUserSQLX(dbx, User{Name: "Diana", Email: "diana@example.com"})
	if err != nil {
		log.Println("CreateUserSQLX:", err)
	} else {
		log.Println("CreateUserSQLX - nouvel ID:", idx2)
	}

	usersx, err := GetUsersSQLX(dbx)
	if err != nil {
		log.Println("GetUsersSQLX:", err)
	} else {
		log.Println("GetUsersSQLX:", usersx)
	}

	userx, err := GetUserByIDSQLX(dbx, int(idx1))
	if err != nil {
		log.Println("GetUserByIDSQLX:", err)
	} else {
		log.Println("GetUserByIDSQLX:", userx)
	}

	err = UpdateUserSQLX(dbx, User{ID: int(idx1), Name: "Charlie Modifié", Email: "charlie.mod@example.com"})
	if err != nil {
		log.Println("UpdateUserSQLX:", err)
	} else {
		log.Println("UpdateUserSQLX: OK")
	}

	err = DeleteUserSQLX(dbx, int(idx2))
	if err != nil {
		log.Println("DeleteUserSQLX:", err)
	} else {
		log.Println("DeleteUserSQLX: OK")
	}

	usersx, err = GetUsersSQLX(dbx)
	if err != nil {
		log.Println("GetUsersSQLX final:", err)
	} else {
		log.Println("GetUsersSQLX final:", usersx)
	}

	// --- GORM ---
	gormDB, err := gorm.Open(sqlite.Open("gorm_test.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("GORM connexion:", err)
	}
	log.Println("Connecté à SQLite (GORM)")

	if err = gormDB.AutoMigrate(&User{}); err != nil {
		log.Fatal("AutoMigrate:", err)
	}
	log.Println("AutoMigrate OK")

	userGorm := User{Name: "Eve", Email: "eve@example.com"}
	if err = CreateUserGORM(gormDB, userGorm); err != nil {
		log.Println("CreateUserGORM:", err)
	} else {
		log.Println("CreateUserGORM: OK")
	}

	userGorm2 := User{Name: "Frank", Email: "frank@example.com"}
	if err = CreateUserGORM(gormDB, userGorm2); err != nil {
		log.Println("CreateUserGORM:", err)
	} else {
		log.Println("CreateUserGORM: OK")
	}

	usersGorm, err := GetUsersGORM(gormDB)
	if err != nil {
		log.Println("GetUsersGORM:", err)
	} else {
		log.Println("GetUsersGORM:", usersGorm)
	}

	var firstGormUser User
	gormDB.First(&firstGormUser)

	fetchedGorm, err := GetUserByIDGORM(gormDB, firstGormUser.ID)
	if err != nil {
		log.Println("GetUserByIDGORM:", err)
	} else {
		log.Println("GetUserByIDGORM:", fetchedGorm)
	}

	firstGormUser.Name = "Eve Modifiée"
	firstGormUser.Email = "eve.mod@example.com"
	if err = UpdateUserGORM(gormDB, firstGormUser); err != nil {
		log.Println("UpdateUserGORM:", err)
	} else {
		log.Println("UpdateUserGORM: OK")
	}

	var secondGormUser User
	gormDB.Last(&secondGormUser)
	if err = DeleteUserGORM(gormDB, secondGormUser.ID); err != nil {
		log.Println("DeleteUserGORM:", err)
	} else {
		log.Println("DeleteUserGORM: OK")
	}

	usersGorm, err = GetUsersGORM(gormDB)
	if err != nil {
		log.Println("GetUsersGORM final:", err)
	} else {
		log.Println("GetUsersGORM final:", usersGorm)
	}
}
