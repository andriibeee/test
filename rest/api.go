package rest

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/andriibeee/task/commands"
	"github.com/andriibeee/task/database"
	"github.com/andriibeee/task/entities"
	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type RESTInterface struct {
	signCmd  commands.SignCommand
	fetchCmd commands.FetchCommand

	validate *validator.Validate

	secret []byte
}

func NewRESTInterface(
	signCmd commands.SignCommand,
	fetchCmd commands.FetchCommand,
	secret []byte,
	validate *validator.Validate,
) RESTInterface {
	return RESTInterface{
		signCmd:  signCmd,
		fetchCmd: fetchCmd,
		secret:   secret,
		validate: validate,
	}
}

type Question struct {
	ID       string `json:"id" validate:"required,uuid"`
	Question string `json:"question" validate:"required"`
}

type Answer struct {
	Question Question `json:"question" validate:"required"`
	Answer   string   `json:"answer" validate:"required"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}

var unauthorized = errors.New("unauthorized")

func (api *RESTInterface) GetUser(r *http.Request) (*entities.User, error) {
	authorization := strings.TrimSpace(r.Header.Get("Authorization"))
	fragments := strings.SplitN(authorization, " ", 2)

	if len(fragments) != 2 {
		return nil, unauthorized
	}

	if strings.ToLower(fragments[0]) != "bearer" {
		return nil, unauthorized
	}

	token, err := jwt.Parse(fragments[1], func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return api.secret, nil
	})

	if err != nil {
		return nil, unauthorized
	}

	claims := token.Claims.(jwt.MapClaims)

	return &entities.User{
		ID:       uuid.MustParse(claims["sub"].(string)),
		UserName: claims["username"].(string),
	}, nil
}

func (api *RESTInterface) Sign(w http.ResponseWriter, r *http.Request) {
	user, err := api.GetUser(r)

	enc := json.NewEncoder(w)

	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		enc.Encode(ErrorResponse{
			Message: err.Error(),
		})
		return
	}

	var body []Answer

	err = json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		enc.Encode(ErrorResponse{
			Message: err.Error(),
		})
		return
	}

	if len(body) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		enc.Encode(ErrorResponse{
			Message: "empty answers",
		})
		return
	}

	q := map[uuid.UUID]entities.Question{}
	a := map[uuid.UUID]entities.Answer{}

	for _, question := range body {
		err = api.validate.Struct(question)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			enc.Encode(ErrorResponse{
				Message: err.Error(),
			})
			return
		}
		id := uuid.MustParse(question.Question.ID)
		q[id] = entities.Question{
			ID:       id,
			Question: question.Question.Question,
		}
		a[id] = entities.Answer{
			ID:       id,
			Question: q[id],
			Answer:   question.Answer,
		}
	}
	output, err := api.signCmd.Handle(
		r.Context(),
		commands.SignInput{
			User:      *user,
			Questions: q,
			Answers:   a,
		},
	)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		enc.Encode(ErrorResponse{
			Message: err.Error(),
		})
		return
	}

	enc.Encode(output)
}

func (api *RESTInterface) Fetch(w http.ResponseWriter, r *http.Request) {

	user, err := api.GetUser(r)

	enc := json.NewEncoder(w)

	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		enc.Encode(ErrorResponse{
			Message: err.Error(),
		})
		return
	}

	vars := mux.Vars(r)
	signature, err := uuid.Parse(vars["signature"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		enc.Encode(ErrorResponse{
			Message: err.Error(),
		})
		return
	}

	output, err := api.fetchCmd.Handle(
		r.Context(),
		commands.FetchInput{
			User:      *user,
			Signature: signature,
		},
	)

	if err != nil && errors.Is(err, database.SignatureNotFound) {
		w.WriteHeader(http.StatusNotFound)
		enc.Encode(ErrorResponse{
			Message: "not found",
		})
		return
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		enc.Encode(ErrorResponse{
			Message: err.Error(),
		})
		return
	}

	enc.Encode(output)
}
