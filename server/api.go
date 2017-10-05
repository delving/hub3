package server

import (
	"net/http"

	"bitbucket.org/delving/rapid/hub3"

	"github.com/labstack/echo"
	"github.com/labstack/gommon/log"
)

// Copyright © 2017 Delving B.V. <info@delving.eu>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

func bulkApi(c echo.Context) error {
	response, err := hub3.ReadActions(c.Request().Body)
	if err != nil {
		log.Info("Unable to read actions")
	}
	return c.JSON(http.StatusCreated, response)
}
