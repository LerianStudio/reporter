package http

import (
	"fmt"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"k8s-golang-addons-boilerplate/pkg"
	"k8s-golang-addons-boilerplate/pkg/constant"
	"strconv"
	"strings"
	"time"
)

// QueryHeader entity from query parameter from get apis
type QueryHeader struct {
	Metadata     *bson.M
	Limit        int
	Page         int
	Cursor       string
	SortOrder    string
	StartDate    time.Time
	EndDate      time.Time
	Alias        string
	UseMetadata  bool
	PortfolioID  string
	ToAssetCodes []string
}

// Pagination entity from query parameter from get apis
type Pagination struct {
	Limit     int
	Page      int
	Cursor    string
	SortOrder string
	StartDate time.Time
	EndDate   time.Time
	Alias     string
}

func (qh *QueryHeader) ToOffsetPagination() Pagination {
	return Pagination{
		Limit:     qh.Limit,
		Page:      qh.Page,
		SortOrder: qh.SortOrder,
		StartDate: qh.StartDate,
		EndDate:   qh.EndDate,
		Alias:     qh.Alias,
	}
}

// ValidateParameters validate and return struct of default parameters
func ValidateParameters(params map[string]string) (*QueryHeader, error) {
	var (
		metadata     *bson.M
		portfolioID  string
		toAssetCodes []string
		startDate    time.Time
		endDate      time.Time
		cursor       string
		alias        string
		limit        = 10
		page         = 1
		sortOrder    = "desc"
		useMetadata  = false
	)

	for key, value := range params {
		switch {
		case strings.Contains(key, "metadata."):
			metadata = &bson.M{key: value}
			useMetadata = true
		case strings.Contains(key, "limit"):
			limit, _ = strconv.Atoi(value)
		case strings.Contains(key, "page"):
			page, _ = strconv.Atoi(value)
		case strings.Contains(key, "cursor"):
			cursor = value
		case strings.Contains(key, "sort_order"):
			sortOrder = strings.ToLower(value)
		case strings.Contains(key, "start_date"):
			fmt.Println("teste")

			startDate, _ = time.Parse("2006-01-02", value)
		case strings.Contains(key, "end_date"):
			endDate, _ = time.Parse("2006-01-02", value)
		case strings.Contains(key, "portfolio_id"):
			portfolioID = value
		case strings.Contains(key, "to"):
			toAssetCodes = strings.Split(value, ",")
		case key == "alias":
			alias = value
		}
	}

	err := validateDates(&startDate, &endDate)
	if err != nil {
		return nil, err
	}

	err = validatePagination(cursor, sortOrder, limit)
	if err != nil {
		return nil, err
	}

	if !pkg.IsNilOrEmpty(&portfolioID) {
		_, err := uuid.Parse(portfolioID)
		if err != nil {
			return nil, pkg.ValidateBusinessError(constant.ErrInvalidQueryParameter, "", "portfolio_id")
		}
	}

	query := &QueryHeader{
		Metadata:     metadata,
		Limit:        limit,
		Page:         page,
		Cursor:       cursor,
		SortOrder:    sortOrder,
		StartDate:    startDate,
		EndDate:      endDate,
		Alias:        alias,
		UseMetadata:  useMetadata,
		PortfolioID:  portfolioID,
		ToAssetCodes: toAssetCodes,
	}

	return query, nil
}

func validateDates(startDate, endDate *time.Time) error {
	maxDateRangeMonths := pkg.SafeInt64ToInt(pkg.GetenvIntOrDefault("MAX_PAGINATION_MONTH_DATE_RANGE", 1))

	defaultStartDate := time.Now().AddDate(0, -maxDateRangeMonths, 0)
	defaultEndDate := time.Now()

	if !startDate.IsZero() && !endDate.IsZero() {
		if !pkg.IsValidDate(pkg.NormalizeDate(*startDate, nil)) || !pkg.IsValidDate(pkg.NormalizeDate(*endDate, nil)) {
			return pkg.ValidateBusinessError(constant.ErrInvalidDateFormat, "")
		}

		if !pkg.IsInitialDateBeforeFinalDate(*startDate, *endDate) {
			return pkg.ValidateBusinessError(constant.ErrInvalidFinalDate, "")
		}

		if !pkg.IsDateRangeWithinMonthLimit(*startDate, *endDate, maxDateRangeMonths) {
			return pkg.ValidateBusinessError(constant.ErrDateRangeExceedsLimit, "", maxDateRangeMonths)
		}
	}

	if startDate.IsZero() && endDate.IsZero() {
		*startDate = defaultStartDate
		*endDate = defaultEndDate
	}

	if (!startDate.IsZero() && endDate.IsZero()) ||
		(startDate.IsZero() && !endDate.IsZero()) {
		return pkg.ValidateBusinessError(constant.ErrInvalidDateRange, "")
	}

	return nil
}

func validatePagination(cursor, sortOrder string, limit int) error {
	maxPaginationLimit := pkg.SafeInt64ToInt(pkg.GetenvIntOrDefault("MAX_PAGINATION_LIMIT", 100))

	if limit > maxPaginationLimit {
		return pkg.ValidateBusinessError(constant.ErrPaginationLimitExceeded, "", maxPaginationLimit)
	}

	if (sortOrder != string(constant.Asc)) && (sortOrder != string(constant.Desc)) {
		return pkg.ValidateBusinessError(constant.ErrInvalidSortOrder, "")
	}

	if !pkg.IsNilOrEmpty(&cursor) {
		_, err := DecodeCursor(cursor)
		if err != nil {
			return pkg.ValidateBusinessError(constant.ErrInvalidQueryParameter, "", "cursor")
		}
	}

	return nil
}
