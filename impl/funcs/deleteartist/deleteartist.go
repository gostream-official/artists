package deleteartist

import (
	"fmt"
	"net/http"

	"github.com/gostream-official/artists/impl/inject"
	"github.com/gostream-official/artists/impl/models"
	"github.com/gostream-official/artists/pkg/api"
	"github.com/gostream-official/artists/pkg/marshal"
	"github.com/gostream-official/artists/pkg/parallel"
	"github.com/gostream-official/artists/pkg/store"
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
		return nil, fmt.Errorf("deleteartist: failed to deduce injector")
	}

	return &injector, nil
}

// Description:
//
//	The router handler for: Get Track By ID
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
		log.Errorf("[%s] failed to get endpoint injector: %s", context.ID, err)
		return &api.APIResponse{
			StatusCode: http.StatusInternalServerError,
		}
	}

	idToDelete := request.PathParameters["id"]

	store := store.NewMongoStore[models.ArtistInfo](injector.MongoInstance, "gostream", "artists")
	count, err := store.DeleteItem(idToDelete)

	if err != nil {
		log.Errorf("[%s] failed to delete database items: %s", context.ID, err)
		return &api.APIResponse{
			StatusCode: http.StatusInternalServerError,
		}
	}

	if count == 0 {
		return &api.APIResponse{
			StatusCode: http.StatusNoContent,
		}
	}

	return &api.APIResponse{
		StatusCode: http.StatusAccepted,
	}
}
