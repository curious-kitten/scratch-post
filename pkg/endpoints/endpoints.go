package endpoints

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	"github.com/curious-kitten/scratch-post/internal/http/authorization"
	"github.com/curious-kitten/scratch-post/internal/http/helpers"
	"github.com/curious-kitten/scratch-post/internal/store"
	"github.com/curious-kitten/scratch-post/pkg/metadata"
)

type create func(ctx context.Context, author string, body io.Reader) (interface{}, error)
type list func(ctx context.Context) ([]interface{}, error)
type get func(ctx context.Context, id string) (interface{}, error)
type updateItem func(ctx context.Context, author string, id string, body io.Reader) (interface{}, error)
type deleteItem func(ctx context.Context, id string) error

// Creator reponds to a HTTP Post request to a collection
func Creator(ctx context.Context, createFunc create, r *mux.Router) {
	c := func(w http.ResponseWriter, r *http.Request) {
		user, err := authorization.GetUserIDFromRequest(r)
		if err != nil {
			helpers.FormatError(w, err.Error(), http.StatusBadRequest)
			return
		}
		toctx, cancel := context.WithTimeout(ctx, time.Second*10)
		defer cancel()

		item, err := createFunc(toctx, user, r.Body)
		if err != nil {
			handleError(err, w)
			return
		}
		helpers.FormatResponse(w, item, http.StatusCreated)
	}
	r.HandleFunc("", c).Methods(http.MethodPost)
	r.HandleFunc("/", c).Methods(http.MethodPost)
}

// Lister reponds to a HTTP Get request for a collection
func Lister(ctx context.Context, listFunc list, r *mux.Router) {
	l := func(w http.ResponseWriter, r *http.Request) {
		toctx, cancel := context.WithTimeout(ctx, time.Second*10)
		defer cancel()
		items, err := listFunc(toctx)
		if err != nil {
			helpers.FormatError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		itemList := &ItemList{
			Count:      len(items),
			TotalCount: len(items),
			StartIndex: 0,
			EndIndex:   len(items) - 1,
			Items:      items,
		}
		helpers.FormatResponse(w, itemList, http.StatusOK)
	}
	r.HandleFunc("", l).Methods(http.MethodGet)
	r.HandleFunc("/", l).Methods(http.MethodGet)
}

// ItemList formats the collection get response to a list
type ItemList struct {
	Count      int           `json:"count"`
	TotalCount int           `json:"totalCount"`
	StartIndex int           `json:"startIndex"`
	EndIndex   int           `json:"endIndex"`
	Items      []interface{} `json:"items"`
}

// Getter returns a single instance of an item based on the ID in the path
func Getter(ctx context.Context, getterFunc get, r *mux.Router) {
	i := func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		id := params["id"]
		toctx, cancel := context.WithTimeout(ctx, time.Second*10)
		defer cancel()
		item, err := getterFunc(toctx, id)
		if err != nil {
			handleError(err, w)
			return
		}
		helpers.FormatResponse(w, item, http.StatusOK)
	}
	r.HandleFunc("/{id}", i).Methods(http.MethodGet)
}

// Deleter provides an API endpoint used to delete an intem
func Deleter(ctx context.Context, deleterFunc deleteItem, r *mux.Router) {
	d := func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		id := params["id"]
		toctx, cancel := context.WithTimeout(ctx, time.Second*10)
		defer cancel()
		if err := deleterFunc(toctx, id); err != nil {
			handleError(err, w)
			return
		}
		helpers.FormatResponse(w, struct {
			Item string `json:"item"`
		}{Item: id}, http.StatusOK)
	}
	r.HandleFunc("/{id}", d).Methods(http.MethodDelete)
}

// Updater provides an API endpoint used to update an intem
func Updater(ctx context.Context, updateFunc updateItem, r *mux.Router) {
	u := func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		id := params["id"]
		user, err := authorization.GetUserIDFromRequest(r)
		if err != nil {
			helpers.FormatError(w, err.Error(), http.StatusBadRequest)
			return
		}
		toctx, cancel := context.WithTimeout(ctx, time.Second*10)
		defer cancel()
		item, err := updateFunc(toctx, user, id, r.Body)
		if err != nil {
			handleError(err, w)
			return
		}
		helpers.FormatResponse(w, item, http.StatusOK)
	}
	r.HandleFunc("/{id}", u).Methods(http.MethodPut)
}

func handleError(err error, w http.ResponseWriter) {
	switch {
	case store.IsNotFoundError(err):
		helpers.FormatError(w, "could not find requested item", http.StatusNotFound)
	case metadata.IsValidationError(err):
		helpers.FormatError(w, err.Error(), http.StatusBadRequest)
	case store.IsDuplicateError(err):
		helpers.FormatError(w, "item already exists", http.StatusBadRequest)
	default:
		helpers.FormatError(w, err.Error(), http.StatusInternalServerError)
	}
}
