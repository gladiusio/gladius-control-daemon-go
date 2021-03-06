package handlers

import (
	"encoding/json"
	"github.com/gladiusio/gladius-application-server/pkg/controller"
	"github.com/gladiusio/gladius-controld/pkg/utils"
	"github.com/jinzhu/gorm"
	"net/http"

	"github.com/gladiusio/gladius-controld/pkg/blockchain"
	"github.com/gladiusio/gladius-controld/pkg/routing/response"
	"github.com/gorilla/mux"
)

func PoolPublicDataHandler(ga *blockchain.GladiusAccountManager) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		err := AccountErrorHandler(w, r, ga)
		if err != nil {
			return
		}

		vars := mux.Vars(r)
		poolAddress := vars["poolAddress"]

		poolResponse, err := PoolResponseForAddress(poolAddress, ga)

		if err != nil {
			ErrorHandler(w, r, "Pool data could not be found for Pool: "+poolAddress, err, http.StatusBadRequest)
			return
		}

		poolInformationResponse, err := utils.SendRequest(http.MethodGet, poolResponse.Url+"server/info", nil)
		var defaultResponse response.DefaultResponse
		json.Unmarshal([]byte(poolInformationResponse), &defaultResponse)

		ResponseHandler(w, r, "null", true, nil, defaultResponse.Response, nil)
	}
}

func PoolRetrievePendingPoolConfirmationApplicationsHandler(db *gorm.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		profiles, err := controller.NodesPendingPoolConfirmation(db)

		if err != nil {
			ErrorHandler(w, r, "Could not retrieve applications", err, http.StatusNotFound)
			return
		}

		ResponseHandler(w, r, "null", true, nil, profiles, nil)
	}
}

func PoolRetrievePendingNodeConfirmationApplicationsHandler(db *gorm.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		profiles, err := controller.NodesPendingNodeConfirmation(db)

		if err != nil {
			ErrorHandler(w, r, "Could not retrieve applications", err, http.StatusNotFound)
			return
		}

		ResponseHandler(w, r, "null", true, nil, profiles, nil)
	}
}

func PoolRetrieveApprovedApplicationsHandler(db *gorm.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		profiles, err := controller.NodesAccepted(db)

		if err != nil {
			ErrorHandler(w, r, "Could not retrieve applications", err, http.StatusNotFound)
			return
		}

		ResponseHandler(w, r, "null", true, nil, profiles, nil)
	}
}

func PoolRetrieveRejectedApplicationsHandler(db *gorm.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		profiles, err := controller.NodesRejected(db)

		if err != nil {
			ErrorHandler(w, r, "Could not retrieve applications", err, http.StatusNotFound)
			return
		}

		ResponseHandler(w, r, "null", true, nil, profiles, nil)
	}
}
