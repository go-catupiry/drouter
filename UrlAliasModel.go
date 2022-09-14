package drouter

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/go-catupiry/catu"
	"github.com/go-catupiry/catu/helpers"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type UrlAliasModel struct {
	ID     uint64 `gorm:"primary_key" json:"id" filter:"param:id;type:number"`
	Alias  string `gorm:"column:alias;type:TEXT" json:"alias" filter:"param:alias;type:string"`
	Target string `gorm:"column:target;type:TEXT" json:"target" filter:"param:target;type:string"`
	Locale string `gorm:"column:locale;" json:"locale"`

	CreatedAt time.Time `gorm:"column:createdAt;" json:"createdAt"`
	UpdatedAt time.Time `gorm:"column:updatedAt;" json:"updatedAt"`

	LinkPermanent string `gorm:"-" json:"linkPermanent"`
}

func (r *UrlAliasModel) TableName() string {
	return "urlAlias"
}

func UrlAliasGetByURL(url string, record *UrlAliasModel) error {
	db := catu.GetDefaultDatabaseConnection()

	if err := db.
		Select([]string{"alias", "target", "locale"}).
		Where("alias = ? OR target = ?", url, url).
		Order("createdAt DESC").
		Limit(1).
		Find(record).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			return err
		}
		return err
	}

	return nil
}

func (r *UrlAliasModel) LoadTeaserData() error {
	return nil
}

func (r *UrlAliasModel) LoadData() error {
	r.LoadPath()
	return nil
}

func (r *UrlAliasModel) GetPath() string {
	path := ""

	if r.ID != 0 {
		path += "/url-alias/" + r.GetIDString()
	}

	return path
}

func (r *UrlAliasModel) LoadPath() error {
	app := catu.GetApp()
	r.LinkPermanent = app.GetConfiguration().Get("APP_ORIGIN") + r.GetPath()
	return nil
}

// Save - Create if is new or update
func (m *UrlAliasModel) Save() error {
	db := catu.GetDefaultDatabaseConnection()

	if m.ID == 0 {
		// create ....
		err := db.Create(&m).Error
		if err != nil {
			return err
		}
	} else {
		// update ...
		err := db.Save(&m).Error
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *UrlAliasModel) GetIDString() string {
	return strconv.FormatUint(m.ID, 10)
}

func (m *UrlAliasModel) ToJSON() string {
	jsonString, _ := json.Marshal(m)
	return string(jsonString)
}

// FindOne - Find one record
func FindOne(id string, record *UrlAliasModel) error {
	db := catu.GetDefaultDatabaseConnection()
	return db.First(&record, id).Error
}

func URLAliasCreateIfNotExists(alias, target, locale string, r *UrlAliasModel) error {
	if locale == "" {
		locale = "pt_BR"
	}

	err := URLAliasFindOneByTarget(target, r)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	if r.ID != 0 {
		// already exists, skip
		return nil
	}

	r.Alias = alias
	r.Target = target
	r.Locale = locale

	return r.Save()
}

func URLAliasUpsert(alias, target, locale string, r *UrlAliasModel) error {
	if locale == "" {
		locale = "pt_BR"
	}

	err := URLAliasFindOneByTarget(target, r)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	if r.ID != 0 {
		// already exists, update

		if r.Alias == alias {
			return nil
		}

		r.Alias = alias
		return r.Save()
	}

	r.Alias = alias
	r.Target = target
	r.Locale = locale

	return r.Save()
}

// URLAliasFindOne - Find one URLAlias record by ID
func URLAliasFindOne(id string, r *UrlAliasModel) error {
	db := catu.GetDefaultDatabaseConnection()

	return db.First(&r, id).Error
}

// URLAliasFindOne - Find one URLAlias record by target
func URLAliasFindOneByTarget(target string, r *UrlAliasModel) error {
	db := catu.GetDefaultDatabaseConnection()

	return db.
		Where("target = ?", target).
		First(&r).Error
}

func URLAliasDeleteByTarget(target string) error {
	db := catu.GetDefaultDatabaseConnection()
	return db.Where("target = ?", target).Delete(&UrlAliasModel{}).Error
}

func (r *UrlAliasModel) Delete() error {
	db := catu.GetDefaultDatabaseConnection()
	return db.Unscoped().Delete(&r).Error
}

type QueryOpts struct {
	Records *[]UrlAliasModel
	Count   *int64
	Limit   int
	Offset  int
	C       echo.Context
	IsHTML  bool
}

func QueryAndCountReq(opts *QueryOpts) error {
	db := catu.GetDefaultDatabaseConnection()

	c := opts.C

	q := c.QueryParam("q")

	query := db

	ctx := c.(*catu.RequestContext)

	queryI, err := ctx.Query.SetDatabaseQueryForModel(query, &UrlAliasModel{})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": fmt.Sprintf("%+v\n", err),
		}).Error("QueryAndCountReq error")
	}
	query = queryI.(*gorm.DB)

	if q != "" {
		query = query.Where(
			db.Where("alias LIKE ?", "%"+q+"%").Or(db.Where("target LIKE ?", "%"+q+"%")),
		)
	}

	orderColumn, orderIsDesc, orderValid := helpers.ParseUrlQueryOrder(c.QueryParam("order"))

	if orderValid {
		query = query.Order(clause.OrderByColumn{
			Column: clause.Column{Table: clause.CurrentTable, Name: orderColumn},
			Desc:   orderIsDesc,
		})
	} else {
		query = query.
			Order("createdAt DESC").
			Order("id DESC")
	}

	query = query.Limit(opts.Limit).
		Offset(opts.Offset)

	r := query.Find(opts.Records)
	if r.Error != nil {
		return r.Error
	}

	return B3NewsCountReq(opts)
}

func B3NewsCountReq(opts *QueryOpts) error {
	db := catu.GetDefaultDatabaseConnection()

	c := opts.C

	q := c.QueryParam("q")

	ctx := c.(*catu.RequestContext)

	// Count ...
	queryCount := db

	if q != "" {
		queryCount = queryCount.Or(
			db.Where("alias LIKE ?", "%"+q+"%"),
			db.Where("target LIKE ?", "%"+q+"%"),
		)
	}

	queryICount, err := ctx.Query.SetDatabaseQueryForModel(queryCount, &UrlAliasModel{})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": fmt.Sprintf("%+v\n", err),
		}).Error("B3NewsCountReq count error")
	}
	queryCount = queryICount.(*gorm.DB)

	return queryCount.
		Table("urlAlias").
		Count(opts.Count).Error
}
