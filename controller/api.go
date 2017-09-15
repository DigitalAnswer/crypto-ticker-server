// Copyright 2017 Axel Etcheverry. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package controller

import (
	"net/http"

	"github.com/DigitalAnswer/crypto-ticker-server/server"
	"github.com/DigitalAnswer/crypto-ticker-server/services"
)

// APIController struct
type APIController struct {
	authenticationService *services.AuthenticationService
	userService           *services.UserService
}

// NewAPIController func
func NewAPIController(authenticationService *services.AuthenticationService, userService *services.UserService) (*APIController, error) {
	return &APIController{
		authenticationService: authenticationService,
		userService:           userService,
	}, nil
}

// Mount endpoints
func (c APIController) Mount(r *server.Router) {
	r.AddRouteFunc("/api/authenticate", c.PostAuthHandler).Methods(http.MethodPost)

	r.AddRouteFunc("/api/users", c.PostUserHandler).Methods(http.MethodPost)
	r.AddRouteFunc("/api/me", c.GetMeHandler).Methods(http.MethodGet)
}
