package createartist

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gostream-official/artists/impl/inject"
	"github.com/gostream-official/artists/impl/models"
	"github.com/gostream-official/artists/pkg/api"
	"github.com/gostream-official/artists/pkg/arrays"
	"github.com/gostream-official/artists/pkg/marshal"
	"github.com/gostream-official/artists/pkg/parallel"
	"github.com/gostream-official/artists/pkg/store"
	"github.com/gostream-official/artists/pkg/store/query"
	"github.com/revx-official/output/log"

	"github.com/google/uuid"
)

// Description:
//
//	The request body for the create artist endpoint.
type CreateArtistRequestBody struct {

	// The name of the artist.
	Name string `json:"name" bson:"name"`

	// The genres an artist is active in.
	Genres []string `json:"genres" bson:"genres"`

	// The amount of followers the artist has.
	Followers uint32 `json:"followers" bson:"followers"`

	// Some artist statistics.
	Stats CreateArtistStatsRequestBody `json:"stats" bson:"stats"`
}

// Description:
//
//	The request body for the artist statistics.
type CreateArtistStatsRequestBody struct {

	// The popularity factor of the artist.
	Popularity float32 `json:"popularity" bson:"popularity"`
}

// Description:
//
//	The error response body for the create artist endpoint.
type CreateArtistErrorResponseBody struct {

	// The error message.
	Message string `json:"message"`
}

// Description:
//
//	Describes a validation error.
type CreateArtistValidationError struct {

	// The JSON field which is referenced by the error message.
	FieldRef string `json:"ref"`

	// The error message.
	ErrorMessage string `json:"error"`
}

// Description:
//
//	Attempts to cast the input object to the endpoint injector.
//	If this cast fails, we cannot proceed to process this request.
//
// Parameters:
//
//	object 	The injector object.
//
// Returns:
//
//	The injector if the cast is successful, an error otherwise.
func GetSafeInjector(object interface{}) (*inject.Injector, error) {
	injector, ok := object.(inject.Injector)

	if !ok {
		return nil, fmt.Errorf("createartist: failed to deduce injector")
	}

	return &injector, nil
}

// Description:
//
//	Unmarshals the request body for this endpoint.
//
// Parameters:
//
//	request The original request.
//
// Returns:
//
//	The unmarshalled request body, or an error when unmarshalling fails.
func ExtractRequestBody(request *api.APIRequest) (*CreateArtistRequestBody, error) {
	body := &CreateArtistRequestBody{}

	bytes := []byte(request.Body)
	err := json.Unmarshal(bytes, body)

	if err != nil {
		return nil, err
	}

	return body, nil
}

// Description:
//
//	Validates the request body for this endpoint.
//
// Parameters:
//
//	request The request body.
//
// Returns:
//
//	An error if the validation fails.
func ValidateRequestBody(request *CreateArtistRequestBody) *CreateArtistValidationError {
	artistName := strings.TrimSpace(request.Name)
	artistGenres := arrays.Map[string](request.Genres, func(genre string) string {
		return strings.TrimSpace(genre)
	})

	if len(artistName) == 0 {
		return &CreateArtistValidationError{
			FieldRef:     "name",
			ErrorMessage: "value must not be empty",
		}
	}

	for _, genre := range artistGenres {
		if len(genre) == 0 {
			return &CreateArtistValidationError{
				FieldRef:     "genres",
				ErrorMessage: "array value must not be empty",
			}
		}
	}

	return ValidateArtistStatsRequestBody(&request.Stats)
}

// Description:
//
//	Validates the statistics request body for this endpoint.
//
// Parameters:
//
//	request The statistics request body.
//
// Returns:
//
//	An error if the validation fails.
func ValidateArtistStatsRequestBody(request *CreateArtistStatsRequestBody) *CreateArtistValidationError {
	return nil
}

// Description:
//
//	Checks whether the given artist id exists in the mongo store.
//
// Parameters:
//
//	store 		The mongo store to search.
//	artistID 	The artist id to search.
//
// Returns:
//
//	An error, if the artist does exist already,
//	if the database request failed, nothing if successful.
func EnsureArtistDoesNotExist(store *store.MongoStore[models.ArtistInfo], artistID string) error {
	filter := query.Filter{
		Root: query.FilterOperatorEq{
			Key:   "_id",
			Value: artistID,
		},
		Limit: 1,
	}

	items, err := store.FindItems(&filter)
	if err != nil {
		return err
	}

	if len(items) == 0 {
		return nil
	}

	return fmt.Errorf("artist already exists")
}

// Description:
//
//	The router handler for artist creation.
//
// Parameters:
//
//	request 	The incoming request.
//	injector 	The injector. Contains injected dependencies.
//
// Returns:
//
//	An API response object.
func Handler(request *api.APIRequest, object interface{}) *api.APIResponse {
	context := parallel.NewContext()

	log.Infof("[%s] %s: %s", context.ID, request.Method, request.Path)
	log.Tracef("[%s] request: %s", context.ID, marshal.Quick(request))

	injector, err := GetSafeInjector(object)
	if err != nil {
		log.Warnf("[%s] failed to get endpoint injector: %s", context.ID, err)
		return &api.APIResponse{
			StatusCode: http.StatusInternalServerError,
		}
	}

	requestBody, err := ExtractRequestBody(request)
	if err != nil {
		log.Warnf("[%s] failed to extract request body: %s", context.ID, err)
		return &api.APIResponse{
			StatusCode: http.StatusBadRequest,
			Body: CreateArtistErrorResponseBody{
				Message: "invalid request body",
			},
		}
	}

	validationError := ValidateRequestBody(requestBody)
	if validationError != nil {
		log.Warnf("[%s] failed request body validation: %s", context.ID, validationError.ErrorMessage)
		return &api.APIResponse{
			StatusCode: http.StatusBadRequest,
			Body:       validationError,
		}
	}

	artistStore := store.NewMongoStore[models.ArtistInfo](injector.MongoInstance, "gostream", "artists")

	artist := models.ArtistInfo{
		ID:        uuid.New().String(),
		Name:      requestBody.Name,
		Genres:    requestBody.Genres,
		Followers: requestBody.Followers,
		Stats: models.ArtistStats{
			Popularity: requestBody.Stats.Popularity,
		},
	}

	err = EnsureArtistDoesNotExist(artistStore, artist.ID)
	if err != nil {
		log.Warnf("[%s] artist already exists: %s", context.ID, err)
		return &api.APIResponse{
			StatusCode: http.StatusConflict,
			Body: CreateArtistErrorResponseBody{
				Message: "artist already exists",
			},
		}
	}

	log.Tracef("[%s] attempting to create database item ...", context.ID)
	err = artistStore.CreateItem(artist)

	if err != nil {
		log.Errorf("[%s] failed to create database item: %s", context.ID, err)
		return &api.APIResponse{
			StatusCode: http.StatusInternalServerError,
		}
	}

	log.Tracef("[%s] successfully completed request", context.ID)
	return &api.APIResponse{
		StatusCode: http.StatusOK,
		Body:       artist,
	}
}
