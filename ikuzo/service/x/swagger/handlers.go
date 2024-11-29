package swagger

import (
	"fmt"
	"net/http"

	"github.com/swaggo/swag"

	"github.com/delving/hub3/ikuzo/domain"
	"github.com/delving/hub3/ikuzo/render"
)

func (s *Service) HandleJSON() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		orgID := domain.GetOrganizationID(r)
		cfg, found := s.orgs.RetrieveConfig(orgID.String())
		if !found {
			render.Error(w, r, domain.ErrOrgNotFound, &render.ErrorConfig{
				StatusCode:    http.StatusNotFound,
				OrgID:         orgID.String(),
				PreventBubble: false,
			})
			return
		}

		swagCfg := cfg.Swagger
		if swagCfg == nil {
			swagCfg = &domain.SwaggerConfig{}
			render.Error(w, r, domain.ErrServiceNotEnabled, &render.ErrorConfig{
				StatusCode:    http.StatusForbidden,
				OrgID:         orgID.String(),
				PreventBubble: false,
			})
			return
		}

		swaggerInfo := &swag.Spec{
			Version:          swagCfg.Info.Version,
			Host:             swagCfg.Info.Host,
			BasePath:         "/api",
			Title:            swagCfg.Info.Title,
			Description:      swagCfg.Info.Description,
			InfoInstanceName: "swagger-" + orgID.String(),
			SwaggerTemplate:  docTemplate,
			LeftDelim:        "{{",
			RightDelim:       "}}",
		}

		// FIXME: add support for contact and license
		// swaggerInfo.Info.Contact = &swag.Contact{
		// 	Name:  "Your Name",
		// 	URL:   "https://yourwebsite.com",
		// 	Email: "your.email@domain.com",
		// }
		// swag.Spec.Info.License = &swag.License{
		// 	Name: "MIT",
		// 	URL:  "https://opensource.org/licenses/MIT",
		// }

		doc := swaggerInfo.ReadDoc()
		_, writeErr := fmt.Fprint(w, doc)
		if writeErr != nil {
			render.Error(w, r, writeErr, &render.ErrorConfig{
				StatusCode:    http.StatusInternalServerError,
				OrgID:         orgID.String(),
				PreventBubble: false,
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
	}
}

func (s *Service) HandleHTML() http.HandlerFunc {
	template := `
<!DOCTYPE html>
<html>
  <head>
    <title>Redoc</title>
    <!-- needed for adaptive design -->
    <meta charset="utf-8"/>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <link href="https://fonts.googleapis.com/css?family=Montserrat:300,400,700|Roboto:300,400,700" rel="stylesheet">

    <!--
    Redoc doesn't change outer page styles
    -->
    <style>
      body {
        margin: 0;
        padding: 0;
      }
    </style>
  </head>
  <body>
    <redoc spec-url='/api-docs/swagger.json'></redoc>
    <script src="https://cdn.redoc.ly/redoc/latest/bundles/redoc.standalone.js"> </script>
  </body>
</html>
	`
	return func(w http.ResponseWriter, r *http.Request) {
		render.HTML(w, r, template)
	}
}
