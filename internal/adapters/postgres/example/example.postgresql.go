package example

import (
	"context"
	"database/sql"
	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
	"k8s-golang-addons-boilerplate/internal/services"
	"k8s-golang-addons-boilerplate/pkg"
	"k8s-golang-addons-boilerplate/pkg/constant"
	"k8s-golang-addons-boilerplate/pkg/example_model/model"
	"k8s-golang-addons-boilerplate/pkg/net/http"
	"k8s-golang-addons-boilerplate/pkg/opentelemetry"
	"k8s-golang-addons-boilerplate/pkg/pointers"
	"k8s-golang-addons-boilerplate/pkg/postgres"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// Repository provides an interface for operations related to organization entities.
//
//go:generate mockgen --destination=../../../mocks/postgres/example_mock.go --package=example . Repository
type Repository interface {
	Create(ctx context.Context, input *model.Example) (*model.ExampleOutput, error)
	Find(ctx context.Context, id uuid.UUID) (*model.ExampleOutput, error)
	FindAll(ctx context.Context, filter http.Pagination) ([]*model.ExampleOutput, error)
	Update(ctx context.Context, id uuid.UUID, example *model.Example) (*model.ExampleOutput, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type ExamplePostgresSQLRepository struct {
	conn      *postgres.PostgresConnection
	tableName string
}

func NewExamplePostgresSQLRepository(conn *postgres.PostgresConnection) *ExamplePostgresSQLRepository {
	c := &ExamplePostgresSQLRepository{
		conn:      conn,
		tableName: "example",
	}

	_, err := c.conn.GetDB()
	if err != nil {
		panic("Failed to connect database")
	}

	return c
}

// Create method to insert an example in database
func (exp *ExamplePostgresSQLRepository) Create(ctx context.Context, input *model.Example) (*model.ExampleOutput, error) {
	tracer := pkg.NewTracerFromContext(ctx)

	ctx, span := tracer.Start(ctx, "postgres.create_example")
	defer span.End()

	conn, err := exp.conn.GetDB()
	if err != nil {
		opentelemetry.HandleSpanError(&span, "Failed to get database connection", err)

		return nil, err
	}

	record := &ExamplePostgreSQLModel{}
	record.FromEntity(input)

	ctx, spanExec := tracer.Start(ctx, "postgres.create.exec")

	err = opentelemetry.SetSpanAttributesFromStruct(&spanExec, "example_repository_input", record)
	if err != nil {
		opentelemetry.HandleSpanError(&spanExec, "Failed to convert example record from entity to JSON string", err)

		return nil, err
	}

	result, err := conn.ExecContext(ctx, `
		INSERT INTO example (id, name, age, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, $5)
		RETURNING *`,
		record.ID,
		record.Name,
		record.Age,
		record.CreatedAt,
		record.UpdatedAt,
	)

	if err != nil {
		opentelemetry.HandleSpanError(&spanExec, "Failed to execute query", err)

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			return nil, err
		}

		return nil, err
	}

	spanExec.End()

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		opentelemetry.HandleSpanError(&span, "Failed to get rows affected", err)

		return nil, err
	}

	if rowsAffected == 0 {
		err := pkg.ValidateBusinessError(constant.ErrEntityNotFound, reflect.TypeOf(model.Example{}).Name())

		opentelemetry.HandleSpanError(&span, "Failed to create example", err)

		return nil, err
	}

	return record.ToEntity(), nil
}

// Find method to get an example in database by id
func (exp *ExamplePostgresSQLRepository) Find(ctx context.Context, id uuid.UUID) (*model.ExampleOutput, error) {
	tracer := pkg.NewTracerFromContext(ctx)

	ctx, span := tracer.Start(ctx, "postgres.find_example")
	defer span.End()

	conn, err := exp.conn.GetDB()
	if err != nil {
		opentelemetry.HandleSpanError(&span, "Failed to get database connection", err)

		return nil, err
	}

	example := &ExamplePostgreSQLModel{}
	ctx, spanQuery := tracer.Start(ctx, "postgres.find.query")

	row := conn.QueryRowContext(ctx, `SELECT * FROM example WHERE id = $1`, id)

	spanQuery.End()

	if err := row.Scan(&example.ID, &example.Name, &example.Age, &example.CreatedAt, &example.UpdatedAt, &example.DeletedAt); err != nil {
		opentelemetry.HandleSpanError(&span, "Failed to scan row", err)

		if errors.Is(err, sql.ErrNoRows) {
			return nil, pkg.ValidateBusinessError(constant.ErrEntityNotFound, reflect.TypeOf(model.Example{}).Name())
		}

		return nil, err
	}

	return example.ToEntity(), nil
}

// FindAll retrieves Example entities from the database.
func (exp *ExamplePostgresSQLRepository) FindAll(ctx context.Context, filter http.Pagination) ([]*model.ExampleOutput, error) {
	tracer := pkg.NewTracerFromContext(ctx)

	ctx, span := tracer.Start(ctx, "postgres.find_all_examples")
	defer span.End()

	conn, err := exp.conn.GetDB()
	if err != nil {
		opentelemetry.HandleSpanError(&span, "Failed to get database connection", err)

		return nil, err
	}

	var examples []*model.ExampleOutput

	findAll := squirrel.Select("*").
		From(exp.tableName).
		Where(squirrel.Eq{"deleted_at": nil}).
		Where(squirrel.GtOrEq{"created_at": pkg.NormalizeDate(filter.StartDate, pointers.Int(-1))}).
		Where(squirrel.LtOrEq{"created_at": pkg.NormalizeDate(filter.EndDate, pointers.Int(1))}).
		OrderBy("id " + strings.ToUpper(filter.SortOrder)).
		Limit(pkg.SafeIntToUint64(filter.Limit)).
		Offset(pkg.SafeIntToUint64((filter.Page - 1) * filter.Limit)).
		PlaceholderFormat(squirrel.Dollar)

	query, args, err := findAll.ToSql()
	if err != nil {
		opentelemetry.HandleSpanError(&span, "Failed to build query", err)

		return nil, err
	}

	ctx, spanQuery := tracer.Start(ctx, "postgres.find_all.query")

	rows, err := conn.QueryContext(ctx, query, args...)
	if err != nil {
		opentelemetry.HandleSpanError(&spanQuery, "Failed to execute query", err)

		return nil, err
	}

	spanQuery.End()

	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			opentelemetry.HandleSpanError(&span, "Failed to close row", err)
			return
		}
	}(rows)

	for rows.Next() {
		var example ExamplePostgreSQLModel

		if err := rows.Scan(&example.ID, &example.Name, &example.Age,
			&example.CreatedAt, &example.UpdatedAt, &example.DeletedAt); err != nil {
			opentelemetry.HandleSpanError(&span, "Failed to scan row", err)

			return nil, err
		}

		examples = append(examples, example.ToEntity())
	}

	if err := rows.Err(); err != nil {
		opentelemetry.HandleSpanError(&span, "Failed to get rows", err)

		return nil, err
	}

	return examples, nil
}

// Update an example by id
func (exp *ExamplePostgresSQLRepository) Update(ctx context.Context, id uuid.UUID, example *model.Example) (*model.ExampleOutput, error) {
	tracer := pkg.NewTracerFromContext(ctx)

	ctx, span := tracer.Start(ctx, "postgres.update_example")
	defer span.End()

	conn, err := exp.conn.GetDB()
	if err != nil {
		opentelemetry.HandleSpanError(&span, "Failed to get database connection", err)

		return nil, err
	}

	record := &ExamplePostgreSQLModel{}
	record.FromEntity(example)

	var updates []string

	var args []any

	if !pkg.IsNilOrEmpty(&example.Name) {
		updates = append(updates, "name = $"+strconv.Itoa(len(args)+1))
		args = append(args, record.Name)
	}

	if !(example.Age == 0) {
		updates = append(updates, "age = $"+strconv.Itoa(len(args)+1))
		args = append(args, record.Age)
	}

	record.UpdatedAt = time.Now()

	updates = append(updates, "updated_at = $"+strconv.Itoa(len(args)+1))

	args = append(args, record.UpdatedAt, id)
	query := `UPDATE example SET ` + strings.Join(updates, ", ") +
		` WHERE id = $` + strconv.Itoa(len(args)) +
		` AND deleted_at IS NULL`

	ctx, spanExec := tracer.Start(ctx, "postgres.update.exec")

	err = opentelemetry.SetSpanAttributesFromStruct(&spanExec, "example_repository_input", record)
	if err != nil {
		opentelemetry.HandleSpanError(&spanExec, "Failed to convert example record from entity to JSON string", err)

		return nil, err
	}

	result, err := conn.ExecContext(ctx, query, args...)
	if err != nil {
		opentelemetry.HandleSpanError(&spanExec, "Failed to execute query", err)

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			return nil, services.ValidatePGError(pgErr, reflect.TypeOf(model.Example{}).Name())
		}

		return nil, err
	}

	spanExec.End()

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		opentelemetry.HandleSpanError(&span, "Failed to get rows affected", err)

		return nil, err
	}

	if rowsAffected == 0 {
		err := pkg.ValidateBusinessError(constant.ErrEntityNotFound, reflect.TypeOf(model.Example{}).Name())

		opentelemetry.HandleSpanError(&span, "Failed to update example. Rows affected is 0", err)

		return nil, err
	}

	return record.ToEntity(), nil
}

// Delete removes an Example entity from the database using the provided ID.
func (exp *ExamplePostgresSQLRepository) Delete(ctx context.Context, id uuid.UUID) error {
	tracer := pkg.NewTracerFromContext(ctx)

	ctx, span := tracer.Start(ctx, "postgres.delete_example")
	defer span.End()

	conn, err := exp.conn.GetDB()
	if err != nil {
		opentelemetry.HandleSpanError(&span, "Failed to get database connection", err)

		return err
	}

	ctx, spanExec := tracer.Start(ctx, "postgres.delete.exec")

	result, err := conn.ExecContext(ctx, `UPDATE example SET deleted_at = now() WHERE id = $1 AND deleted_at IS NULL`, id)
	if err != nil {
		opentelemetry.HandleSpanError(&spanExec, "Failed to execute query", err)

		return err
	}

	spanExec.End()

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		opentelemetry.HandleSpanError(&span, "Failed to get rows affected", err)

		return err
	}

	if rowsAffected == 0 {
		err := pkg.ValidateBusinessError(constant.ErrEntityNotFound, reflect.TypeOf(model.Example{}).Name())

		opentelemetry.HandleSpanError(&span, "Failed to delete example. Rows affected is 0", err)

		return err
	}

	return nil
}
