package main

import (
	"database/sql"
	"flag"
	"net/http"

	"github.com/go-playground/validator/v10"
	_ "github.com/mattn/go-sqlite3"

	"github.com/andriibeee/task/commands"
	"github.com/andriibeee/task/database"
	"github.com/andriibeee/task/rest"
	"github.com/gorilla/mux"
)

func main() {

	secretStr := flag.String("secret", "secret", "jwt secret")
	dsn := flag.String("dsn", "file::memory:?cache=shared", "sqlite dsn")

	flag.Parse()

	secret := []byte(*secretStr)
	db, err := sql.Open("sqlite3", *dsn)

	if err != nil {
		panic(err)
	}

	err = database.RunMigrations(db)
	if err != nil {
		panic(err)
	}

	answersService := database.NewAnswersService(db)
	questionsService := database.NewQuestionsService(db)
	signaturesService := database.NewSignaturesService(db)
	usersService := database.NewUsersService(db)

	transactionManager := database.NewTransactionManager(db)

	signCMD := commands.NewSignCommand(
		usersService,
		questionsService,
		answersService,
		signaturesService,
		transactionManager,
	)

	fetchCMD := commands.NewFetchCommand(
		answersService,
		signaturesService,
	)

	rest := rest.NewRESTInterface(
		signCMD,
		fetchCMD,
		secret,
		validator.New(validator.WithRequiredStructEnabled()),
	)

	r := mux.NewRouter()
	r.HandleFunc("/signature/{signature}", rest.Fetch).Methods("GET")
	r.HandleFunc("/signature", rest.Sign).Methods("POST")

	http.ListenAndServe(":3000", r)
}
