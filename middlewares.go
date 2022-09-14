package drouter

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-catupiry/catu"
	"github.com/go-catupiry/catu/helpers"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

// urlAliasMiddleware - Change url to handle aliased urls like /about to /content/1
func urlAliasMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set("pathBeforeAlias", c.Request().URL.Path)
			cfgs := catu.GetConfiguration()

			if cfgs.Get("URL_ALIAS_ENABLE") == "" {
				return next(c)
			}

			if !isAliasValidMethods(c) {
				return next(c)
			}

			path, err := getPathFromReq(c.Request())
			if err != nil {
				return c.String(http.StatusInternalServerError, "Error on parse url")
			}

			if isPublicRoute(path) {
				// public folders dont have alias...
				return next(c)
			}

			logrus.WithFields(logrus.Fields{
				"url":           path,
				"c.path":        c.Path(),
				"c.QueryString": c.QueryString(),
			}).Debug("urlAliasMiddleware url after alias")

			// save path before alias for reuse ...
			c.Set("pathBeforeAlias", path)

			var record UrlAliasModel
			err = UrlAliasGetByURL(path, &record)
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"url":           path,
					"c.path":        c.Path(),
					"c.QueryString": c.QueryString(),
					"error":         fmt.Sprintf("%+v\n", err),
				}).Error("urlAliasMiddleware Error on get url alias")
			}

			responseContentType := c.Get("responseContentType").(string)

			if record.Target != "" && record.Alias != "" {
				if record.Target == path && responseContentType == "text/html" {
					// redirect to alias url keeping the query string
					queryString := c.QueryString()
					if queryString != "" {
						queryString = "?" + queryString
					}

					return c.Redirect(302, record.Alias+queryString)
				} else {
					// override and continue with target url
					helpers.RewriteURL(record.Target, c)
					c.Set("currentUrl", path)
				}
			}

			return next(c)
		}
	}
}

func isAliasValidMethods(c echo.Context) bool {
	method := c.Request().Method

	if method == "GET" || method == "POST" || method == "PUT" || method == "PATH" || method == "DELETE" {
		return true
	}

	return false
}

func isPublicRoute(url string) bool {
	return strings.HasPrefix(url, "/health") || strings.HasPrefix(url, "/public")
}

func getUrlFromReq(req *http.Request) (string, error) {
	rawURI := req.RequestURI
	if rawURI != "" && rawURI[0] != '/' {
		prefix := ""
		if req.URL.Scheme != "" {
			prefix = req.URL.Scheme + "://"
		}
		if req.URL.Host != "" {
			prefix += req.URL.Host // host or host:port
		}
		if prefix != "" {
			rawURI = strings.TrimPrefix(rawURI, prefix)
		}
	}

	return rawURI, nil
}

func getPathFromReq(req *http.Request) (string, error) {
	return req.URL.Path, nil
}
