package repository

import (
	"database/sql"
	"errors"
	"k8s-golang-addons-boilerplate/pkg/db"
	"k8s-golang-addons-boilerplate/pkg/models"

	_ "github.com/lib/pq"
)

type Example struct {
	db    db.Database
	table string
}

func NewExample(db db.Database) *Example {
	return &Example{db: db, table: "example_table"}
}

// Create method
func (e *Example) Create(input *models.ExampleInput) (*models.ExampleOutput, error) {
	conn, err := e.db.GetDB()
	if err != nil {
		return &models.ExampleOutput{}, err
	}

	record := &models.ExamplePostgreSQLModel{}
	record.FromEntity(input)

	query := "INSERT INTO " + e.table + " (name, value, created_at) VALUES ($1, $2, $3) RETURNING id"
	err = conn.QueryRow(query, record.Name, record.Value, record.CreatedAt).Scan(&record.ID)
	if err != nil {
		return &models.ExampleOutput{}, err
	}
	return record.ToEntity(), nil
}

// Update method
func (e *Example) Update(id string, input *models.ExampleInput) error {
	conn, err := e.db.GetDB()
	if err != nil {
		return err
	}

	record := &models.ExamplePostgreSQLModel{}
	record.FromEntity(input)

	query := "UPDATE " + e.table + " SET name = $1, value = $2 WHERE id = $3"
	_, err = conn.Exec(query, record.Name, record.Value, id)
	if err != nil {
		return err
	}
	return nil
}

// Get method
func (e *Example) Get(id string) (*models.ExampleOutput, error) {
	conn, err := e.db.GetDB()
	if err != nil {
		return &models.ExampleOutput{}, err
	}

	query := "SELECT id, name, value, created_at FROM " + e.table + " WHERE id = $1 AND deleted_at IS NULL"
	row := conn.QueryRow(query, id)
	var record models.ExamplePostgreSQLModel
	err = row.Scan(&record.ID, &record.Name, &record.Value, &record.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("record not found")
		}
		return nil, err
	}
	return record.ToEntity(), nil
}

// GetAll method
func (e *Example) GetAll() ([]models.ExampleOutput, error) {
	conn, err := e.db.GetDB()
	if err != nil {
		return []models.ExampleOutput{}, err
	}

	query := "SELECT id, name, value FROM " + e.table + " WHERE deleted_at IS NULL"
	rows, err := conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.ExampleOutput
	for rows.Next() {
		var model models.ExamplePostgreSQLModel
		err := rows.Scan(&model.ID, &model.Name, &model.Value, &model.CreatedAt)
		if err != nil {
			return nil, err
		}
		list = append(list, *model.ToEntity())
	}
	return list, nil
}

// Delete method
func (e *Example) Delete(id string) error {
	conn, err := e.db.GetDB()
	if err != nil {
		return err
	}

	query := "UPDATE " + e.table + " SET deleted_at = NOW() WHERE id = $1"
	_, err = conn.Exec(query, id)
	if err != nil {
		return err
	}
	return nil
}
