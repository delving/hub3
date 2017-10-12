package server

import (
	"fmt"
	"net/http"

	"bitbucket.org/delving/rapid/hub3"

	"github.com/labstack/echo"
	"github.com/labstack/gommon/log"
	"github.com/renevanderark/goharvest/oai"
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

// bulkApi receives bulkActions in JSON form (1 per line) and processes them in
// ingestion pipeline.
func bulkAPI(c echo.Context) error {
	response, err := hub3.ReadActions(c.Request().Body)
	if err != nil {
		log.Info("Unable to read actions")
	}
	return c.JSON(http.StatusCreated, response)
}

// bindPMHRequest the query parameters to the OAI-Request
func bindPMHRequest(c echo.Context) oai.Request {
	r := c.Request()
	baseURL := fmt.Sprintf("http://%s%s", r.Host, r.URL.Path)
	req := oai.Request{
		Verb:            c.QueryParam("verb"),
		MetadataPrefix:  c.QueryParam("metadataPrefix"),
		Set:             c.QueryParam("set"),
		From:            c.QueryParam("from"),
		Until:           c.QueryParam("until"),
		Identifier:      c.QueryParam("identifier"),
		ResumptionToken: c.QueryParam("resumptionToken"),
		BaseURL:         baseURL,
	}
	return req
}

// oaiPmhEndpoint processed OAI-PMH request and returns the results
func oaiPmhEndpoint(c echo.Context) (err error) {
	req := bindPMHRequest(c)
	return c.JSON(http.StatusOK, req)
}
