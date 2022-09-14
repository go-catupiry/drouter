package drouter

import (
	"net/http"

	"github.com/go-catupiry/catu"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type ListJSONResponse struct {
	catu.BaseListReponse
	Records *[]UrlAliasModel `json:"url-alia"`
}

type CountJSONResponse struct {
	catu.BaseMetaResponse
}

type FindOneJSONResponse struct {
	Record *UrlAliasModel `json:"url-alia"`
}

type BodyRequest struct {
	Record *UrlAliasModel `json:"url-alia"`
}

type TeaserTPL struct {
	Ctx    *catu.RequestContext
	Record *UrlAliasModel
}

// Http content controller | struct with http handlers
type UrlAliasController struct {
}

func (ctl *UrlAliasController) Query(c echo.Context) error {
	var err error

	RequestContext := c.(*catu.RequestContext)

	var count int64
	var records []UrlAliasModel
	err = QueryAndCountReq(&QueryOpts{
		Records: &records,
		Count:   &count,
		Limit:   RequestContext.GetLimit(),
		Offset:  RequestContext.GetOffset(),
		C:       c,
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Debug("Query Error on find url-alias")
	}

	RequestContext.Pager.Count = count

	logrus.WithFields(logrus.Fields{
		"count":             count,
		"len_records_found": len(records),
	}).Debug("Query count result")

	for i := range records {
		records[i].LoadData()
	}

	resp := ListJSONResponse{
		Records: &records,
	}

	resp.Meta.Count = count

	return c.JSON(200, &resp)
}

func (ctl *UrlAliasController) Create(c echo.Context) error {
	logrus.Debug("UrlAliasController.Create running")
	var err error
	ctx := c.(*catu.RequestContext)

	can := ctx.Can("create_url-alias")
	if !can {
		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	var body BodyRequest

	if err := c.Bind(&body); err != nil {
		if _, ok := err.(*echo.HTTPError); ok {
			return err
		}
		return c.NoContent(http.StatusNotFound)
	}

	record := body.Record
	record.ID = 0

	if err := c.Validate(record); err != nil {
		if _, ok := err.(*echo.HTTPError); ok {
			return err
		}
		return err
	}

	logrus.WithFields(logrus.Fields{
		"body": body,
	}).Info("UrlAliasController.Create params")

	err = record.Save()
	if err != nil {
		return err
	}

	err = record.LoadData()
	if err != nil {
		return err
	}

	resp := FindOneJSONResponse{
		Record: record,
	}

	return c.JSON(http.StatusCreated, &resp)
}

func (ctl *UrlAliasController) Count(c echo.Context) error {
	var err error
	RequestContext := c.(*catu.RequestContext)

	var count int64
	err = B3NewsCountReq(&QueryOpts{
		Count:  &count,
		Limit:  RequestContext.GetLimit(),
		Offset: RequestContext.GetOffset(),
		C:      c,
	})

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Debug("UrlAliasController.Count Error on find contents")
	}

	RequestContext.Pager.Count = count

	resp := CountJSONResponse{}
	resp.Count = count

	return c.JSON(200, &resp)
}

func (ctl *UrlAliasController) FindOne(c echo.Context) error {
	id := c.Param("id")

	logrus.WithFields(logrus.Fields{
		"id": id,
	}).Debug("UrlAliasController.FindOne id from params")

	record := UrlAliasModel{}
	err := FindOne(id, &record)
	if err != nil {
		return err
	}

	if record.ID == 0 {
		logrus.WithFields(logrus.Fields{
			"id": id,
		}).Debug("UrlAliasController.FindOne id record not found")

		return echo.NotFoundHandler(c)
	}

	record.LoadData()

	resp := FindOneJSONResponse{
		Record: &record,
	}

	return c.JSON(200, &resp)
}

func (ctl *UrlAliasController) Update(c echo.Context) error {
	var err error

	id := c.Param("id")

	RequestContext := c.(*catu.RequestContext)

	logrus.WithFields(logrus.Fields{
		"id":    id,
		"roles": RequestContext.GetAuthenticatedRoles(),
	}).Debug("UrlAliasController.Update id from params")

	can := RequestContext.Can("update_url-alias")
	if !can {
		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	record := UrlAliasModel{}
	err = FindOne(id, &record)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"id":    id,
			"error": err,
		}).Debug("UrlAliasController.Update error on find one")
		return errors.Wrap(err, "UrlAliasController.Update error on find one")
	}

	record.LoadData()

	body := FindOneJSONResponse{Record: &record}

	if err := c.Bind(&body); err != nil {
		logrus.WithFields(logrus.Fields{
			"id":    id,
			"error": err,
		}).Debug("UrlAliasController.Update error on bind")

		if _, ok := err.(*echo.HTTPError); ok {
			return err
		}
		return c.NoContent(http.StatusNotFound)
	}

	err = record.Save()
	if err != nil {
		return err
	}
	resp := FindOneJSONResponse{
		Record: &record,
	}

	return c.JSON(http.StatusOK, &resp)
}

func (ctl *UrlAliasController) Delete(c echo.Context) error {
	var err error

	id := c.Param("id")

	logrus.WithFields(logrus.Fields{
		"id": id,
	}).Debug("UrlAliasController.Delete id from params")

	RequestContext := c.(*catu.RequestContext)

	can := RequestContext.Can("delete_content")
	if !can {
		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	record := UrlAliasModel{}
	err = FindOne(id, &record)
	if err != nil {
		return err
	}

	err = record.Delete()
	if err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}

type ControllerConfiguration struct {
}

func NewUrlAliasController(cfg *ControllerConfiguration) *UrlAliasController {
	ctx := UrlAliasController{}

	return &ctx
}
