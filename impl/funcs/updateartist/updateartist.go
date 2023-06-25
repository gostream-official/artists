package updateartist

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
//	The request body for the update artist endpoint.
type UpdateArtistRequestBody struct {

	// The name of the artist.
	Name string `json:"name,omitempty" bson:"name"`

	// The genres an artist is active in.
	Genres []string `json:"genres,omitempty" bson:"genres"`

	// The amount of followers the artist has.
	Followers uint32 `json:"followers,omitempty" bson:"followers"`

	// Some artist statistics.
	Stats UpdateArtistStatsRequestBody `json:"stats,omitempty" bson:"stats"`
}

// Description:
//
//	The request body for the artist statistics.
type UpdateArtistStatsRequestBody struct {

	// The popularity factor of the artist.
	Popularity float32 `json:"popularity,omitempty" bson:"popularity"`
}

// Description:
//
//	The error response body for the update artist endpoint.
type UpdateArtistErrorResponseBody struct {

	// The error message.
	Message string `json:"message"`
}

// Description:
//
//	Describes a validation error.
type UpdateArtistValidationError struct {

	// The JSON field which is referenced by the error message.
	FieldRef string `json:"ref"`

	// The error message.
	ErrorMessage string `json:"error"`
}

// Description:
//
//	Describes a path parameter validation error.
type UpdateArtistPathValidationError struct {

	// The JSON field which is referenced by the error message.
	PathRef string `json:"pathRef"`

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
		return nil, fmt.Errorf("updateartist: failed to deduce injector")
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
func ExtractRequestBody(request *api.APIRequest) (*UpdateArtistRequestBody, error) {
	body := &UpdateArtistRequestBody{}

	bytes := []byte(request.Body)
	err := json.Unmarshal(bytes, body)

	if err != nil {
		return nil, err
	}

	return body, nil
}

// Description:
//
//	Gets and validates the id path parameter.
//
// Parameters:
//
//	request The http request.
//
// Returns:
//
//	The id path parameter.
//	A validatior error if the id is not a valid uuid.
func GetAndValidateID(request *api.APIRequest) (string, *UpdateArtistPathValidationError) {
	id := request.PathParameters["id"]

	_, err := uuid.Parse(id)
	if err != nil {
		return "", &UpdateArtistPathValidationError{
			PathRef:      ":id",
			ErrorMessage: "value is not a valid uuid",
		}
	}

	return id, nil
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
func ValidateRequestBody(request *UpdateArtistRequestBody) *UpdateArtistValidationError {
	if request.Name != "" {
		artistName := strings.TrimSpace(request.Name)

		if len(artistName) == 0 {
			return &UpdateArtistValidationError{
				FieldRef:     "name",
				ErrorMessage: "value must not be empty",
			}
		}
	}

	if len(request.Genres) > 0 {
		artistGenres := arrays.Map[string](request.Genres, func(genre string) string {
			return strings.TrimSpace(genre)
		})

		for _, genre := range artistGenres {
			if len(genre) == 0 {
				return &UpdateArtistValidationError{
					FieldRef:     "genres",
					ErrorMessage: "array value must not be empty",
				}
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
func ValidateArtistStatsRequestBody(request *UpdateArtistStatsRequestBody) *UpdateArtistValidationError {
	return nil
}

// Description:
//
//	Searches an artist with the given id in the database.
//
// Parameters:
//
//	store 	The store to search through.
//	id 		The id to search for.
//
// Returns:
//
//	The first matched track.
//	An error if the query fails.
func FindArtistByID(store *store.MongoStore[models.ArtistInfo], id string) (*models.ArtistInfo, error) {
	filter := query.Filter{
		Root: query.FilterOperatorEq{
			Key:   "_id",
			Value: id,
		},
		Limit: 1,
	}

	items, err := store.FindItems(&filter)
	if err != nil {
		return nil, err
	}

	if len(items) == 0 {
		return nil, fmt.Errorf("update: track not found")
	}

	return &items[0], nil
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
//	An error, if the artist could not be found or an error,
//	if the database request failed, nothing if successful.
func CheckIfArtistExists(store *store.MongoStore[models.ArtistInfo], artistID string) error {
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
		return fmt.Errorf("artist not found")
	}

	return nil
}

// Description:
//
//	The router handler for track creation.
//
// Parameters:
//
//	request  The incoming request.
//	injector The injector. Contains injected dependencies.
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

	id, validationErr := GetAndValidateID(request)
	if validationErr != nil {
		log.Warnf("[%s] failed path parameter validation: %s", context.ID, validationErr.ErrorMessage)
		return &api.APIResponse{
			StatusCode: http.StatusBadRequest,
			Body:       validationErr,
		}
	}

	artistStore := store.NewMongoStore[models.ArtistInfo](injector.MongoInstance, "gostream", "artists")

	artistInfo, err := FindArtistByID(artistStore, id)
	if err != nil {
		log.Warnf("[%s] could not find artist: %s", context.ID, err)
		return &api.APIResponse{
			StatusCode: http.StatusNotFound,
		}
	}

	requestBody, err := ExtractRequestBody(request)
	if err != nil {
		log.Warnf("[%s] failed to extract request body: %s", context.ID, err)
		return &api.APIResponse{
			StatusCode: http.StatusBadRequest,
			Body: UpdateArtistErrorResponseBody{
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

	if requestBody.Name != "" {
		artistInfo.Name = requestBody.Name
	}

	if len(requestBody.Genres) > 0 {
		artistInfo.Genres = requestBody.Genres
	}

	if requestBody.Followers != 0 {
		artistInfo.Followers = requestBody.Followers
	}

	if requestBody.Stats.Popularity != 0 {
		artistInfo.Stats.Popularity = requestBody.Stats.Popularity
	}

	updateFilter := query.Filter{
		Root: query.FilterOperatorEq{
			Key:   "_id",
			Value: id,
		},
	}

	updateOperator := query.Update{
		Root: query.UpdateOperatorSet{
			Set: map[string]interface{}{
				"name":             artistInfo.Name,
				"genres":           artistInfo.Genres,
				"followers":        artistInfo.Followers,
				"stats.popularity": artistInfo.Stats.Popularity,
			},
		},
	}

	log.Tracef("[%s] attempting to update database item ...", context.ID)
	count, err := artistStore.UpdateItem(&updateFilter, &updateOperator)

	if err != nil {
		log.Errorf("[%s] failed to update database item: %s", context.ID, err)
		return &api.APIResponse{
			StatusCode: http.StatusInternalServerError,
		}
	}

	if count == 0 {
		log.Warnf("[%s] zero modified items", context.ID)
		return &api.APIResponse{
			StatusCode: http.StatusNoContent,
		}
	}

	log.Tracef("[%s] successfully completed request", context.ID)
	return &api.APIResponse{
		StatusCode: http.StatusNoContent,
	}
}
