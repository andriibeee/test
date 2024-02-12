package database

import "database/sql"

func RunMigrations(db *sql.DB) error {
	//

	_, err := db.Exec(`
	CREATE TABLE IF NOT EXISTS users  (
		uuid TEXT PRIMARY KEY,
		user_name TEXT
	);
	`)

	if err != nil {
		return err
	}

	//

	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS signatures  (
		uuid TEXT PRIMARY KEY,
		user_id INT,
		timestamp INT,
		FOREIGN KEY(user_id) REFERENCES users(uuid)
	);
	`)

	if err != nil {
		return err
	}

	//

	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS questions (
		uuid TEXT PRIMARY KEY,
		question TEXT
	);	
	`)

	if err != nil {
		return err
	}

	//

	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS answers (
		uuid TEXT PRIMARY KEY,
		question_id INT,
		answer TEXT,
		signature_id INT,
		FOREIGN KEY(question_id) REFERENCES questions(uuid),
		FOREIGN KEY(signature_id) REFERENCES signatures(uuid)
	);	
	`)

	if err != nil {
		return err
	}

	return nil
}
