package getartists

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gostream-official/artists/impl/inject"
	"github.com/gostream-official/artists/impl/models"
	"github.com/gostream-official/artists/pkg/api"
	"github.com/gostream-official/artists/pkg/marshal"
	"github.com/gostream-official/artists/pkg/parallel"
	"github.com/gostream-official/artists/pkg/store"
	"github.com/gostream-official/artists/pkg/store/query"
	"github.com/revx-official/output/log"
)

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
		return nil, fmt.Errorf("getartists: failed to deduce injector")
	}

	return &injector, nil
}

// Description:
//
//	Creates a query filter from the given API request.
//
// Parameters:
//
//	request The API request.
//
// Returns:
//
//	The created filter.
func CreateFilterFromQueryParameters(request *api.APIRequest) query.Filter {
	andFilter := query.FilterOperatorAnd{
		And: make([]query.IQuery, 0),
	}

	var realLimit int
	var realLimitErr error

	limit, limitOk := request.QueryParameters["limit"]
	if limitOk {
		realLimit, realLimitErr = strconv.Atoi(limit)
	}

	artistName, artistsName := request.QueryParameters["name"]
	if artistsName {
		andFilter.And = append(andFilter.And, query.FilterOperatorEq{
			Key:   "name",
			Value: artistName,
		})
	}

	resultFilter := query.Filter{}

	if limitOk && realLimitErr == nil {
		resultFilter.Limit = uint32(realLimit)
	}

	if len(andFilter.And) > 0 {
		resultFilter.Root = andFilter
	}

	return resultFilter
}

// Description:
//
//	The router handler for retrieving artists by given search parameters.
//
// Parameters:
//
//	request The incoming request.
//	object 	The injector. Contains injected dependencies.
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
		log.Errorf("[%s] failed to get endpoint injector: %s", context.ID, err)
		return &api.APIResponse{
			StatusCode: http.StatusInternalServerError,
		}
	}

	store := store.NewMongoStore[models.ArtistInfo](injector.MongoInstance, "gostream", "artists")
	filter := CreateFilterFromQueryParameters(request)

	items, err := store.FindItems(&filter)

	if err != nil {
		log.Errorf("[%s] failed to retrieve database items: %s", context.ID, err)
		return &api.APIResponse{
			StatusCode: http.StatusInternalServerError,
		}
	}

	return &api.APIResponse{
		StatusCode: http.StatusOK,
		Body:       items,
	}
}
