package methods

import (
	"context"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"

	"github.com/curious-kitten/scratch-post/internal/decoder"
	"github.com/curious-kitten/scratch-post/internal/http/response"
	"github.com/curious-kitten/scratch-post/internal/logger"
	"github.com/curious-kitten/scratch-post/internal/store"
)

type create func(ctx context.Context, author string, body io.Reader) (interface{}, error)
type list func(ctx context.Context, filter map[string][]string, sortBy string, reverse bool, count int, previousLastValue string) ([]interface{}, error)
type get func(ctx context.Context, id string) (interface{}, error)
type updateItem func(ctx context.Context, author string, id string, body io.Reader) (interface{}, error)
type deleteItem func(ctx context.Context, id string) error
type extractUserName func(r *http.Request) (string, error)

// Post reponds to a HTTP Post request to a collection
func Post(ctx context.Context, createFunc create, getUser extractUserName, r *mux.Router, log logger.Logger) {
	c := func(w http.ResponseWriter, r *http.Request) {
		user, err := getUser(r)
		if err != nil {
			response.SendError(w, err.Error(), http.StatusBadRequest)
			return
		}
		toctx, cancel := context.WithTimeout(ctx, time.Second*10)
		defer cancel()

		item, err := createFunc(toctx, user, r.Body)
		if err != nil {
			handleError(err, w)
			return
		}
		response.Send(w, item, http.StatusCreated)
	}
	route := r.HandleFunc("", c).Methods(http.MethodPost)
	path, _ := route.GetPathTemplate()
	log.Infow("added endpoint", "path", path, "method", http.MethodPost)
}

// List reponds to a HTTP Get request for a collection
func List(ctx context.Context, listFunc list, r *mux.Router, log logger.Logger) {
	l := func(w http.ResponseWriter, r *http.Request) {
		var err error
		toctx, cancel := context.WithTimeout(ctx, time.Second*10)
		defer cancel()
		queries := r.URL.Query()
		sortBy := ""
		reverse := false
		lastFoundValue := ""
		if sorting := queries.Get("sortBy"); sorting != "" {
			queries.Del("sortBy")
			sortValues := strings.Split(sorting, ":")
			sortBy = sortValues[0]
			if len(sortValues) == 2 {
				switch strings.ToLower(sortValues[1]) {
				case "asc":
					reverse = false
				case "desc":
					reverse = true
				}
			}
			if val := queries.Get("lastValue"); val != "" {
				lastFoundValue = val
				queries.Del("lastValue")
			}
		}
		count := 0
		if cnt := queries.Get("count"); cnt != "" {
			count, _ = strconv.Atoi(cnt)
			queries.Del("count")
		}
		items, err := listFunc(toctx, queries, sortBy, reverse, count, lastFoundValue)
		if err != nil {
			response.SendError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		itemList := &ItemList{
			Count: len(items),
			Items: items,
		}
		response.Send(w, itemList, http.StatusOK)
	}
	route := r.HandleFunc("", l).Methods(http.MethodGet)
	path, _ := route.GetPathTemplate()
	log.Infow("added endpoint", "path", path, "method", http.MethodGet)
}

// ItemList formats the collection get response to a list
type ItemList struct {
	Count int           `json:"count"`
	Items []interface{} `json:"items"`
}

// Get returns a single instance of an item based on the ID in the path
func Get(ctx context.Context, getterFunc get, r *mux.Router, log logger.Logger) {
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
		response.Send(w, item, http.StatusOK)
	}
	route := r.HandleFunc("/{id}", i).Methods(http.MethodGet)
	path, _ := route.GetPathTemplate()
	log.Infow("added endpoint", "path", path, "method", http.MethodGet)
}

// Delete provides an API endpoint used to delete an intem
func Delete(ctx context.Context, deleterFunc deleteItem, r *mux.Router, log logger.Logger) {
	d := func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		id := params["id"]
		toctx, cancel := context.WithTimeout(ctx, time.Second*10)
		defer cancel()
		if err := deleterFunc(toctx, id); err != nil {
			handleError(err, w)
			return
		}
		response.Send(w, struct {
			Item string `json:"item"`
		}{Item: id}, http.StatusOK)
	}
	route := r.HandleFunc("/{id}", d).Methods(http.MethodDelete)
	path, _ := route.GetPathTemplate()
	log.Infow("added endpoint", "path", path, "method", http.MethodDelete)
}

// Put provides an API endpoint used to update an intem
func Put(ctx context.Context, updateFunc updateItem, getUser extractUserName, r *mux.Router, log logger.Logger) {
	u := func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		id := params["id"]
		user, err := getUser(r)
		if err != nil {
			response.SendError(w, err.Error(), http.StatusBadRequest)
			return
		}
		toctx, cancel := context.WithTimeout(ctx, time.Second*10)
		defer cancel()
		item, err := updateFunc(toctx, user, id, r.Body)
		if err != nil {
			handleError(err, w)
			return
		}
		response.Send(w, item, http.StatusOK)
	}
	route := r.HandleFunc("/{id}", u).Methods(http.MethodPut)
	path, _ := route.GetPathTemplate()
	log.Infow("added endpoint", "path", path, "method", http.MethodPut)
}

func handleError(err error, w http.ResponseWriter) {
	switch {
	case store.IsNotFoundError(err):
		response.SendError(w, "could not find requested item", http.StatusNotFound)
	case decoder.IsValidationError(err):
		response.SendError(w, err.Error(), http.StatusBadRequest)
	case store.IsDuplicateError(err):
		response.SendError(w, "item already exists", http.StatusBadRequest)
	default:
		response.SendError(w, err.Error(), http.StatusInternalServerError)
	}
}
