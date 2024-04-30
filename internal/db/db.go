package db

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/basedalex/effective-mobile-test/internal/types"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Postgres struct {
	db *pgxpool.Pool
}

func NewPostgres(ctx context.Context, dbConnect string) (*Postgres, error) {
	config, err := pgxpool.ParseConfig(dbConnect)
	if err != nil {
		return nil, fmt.Errorf("error parsing connection string: %w", err)
	}

	db, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("error connecting to database: %w", err)
	}

	err = db.Ping(ctx)
	if err != nil {
		return nil, fmt.Errorf("error pinging to database: %w", err)
	}

	return &Postgres{db: db}, nil
}

type Car struct {
	ID         int       `json:"id" db:"id"`
	RegNum     string    `json:"regNum" db:"regNum"`
	Mark       string    `json:"mark" db:"mark"`
	Model      string    `json:"model" db:"model"`
	Year       int       `json:"year" db:"year"`
	Owner      People    `json:"owner"`
	CreatedAt  time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt  time.Time `json:"updatedAat" db:"created_at"`
}

type People struct {
	ID		   int       `json:"id,omitempty" db:"id"`
	Name       string    `json:"name"`
	Surname    string    `json:"surname"`
	Patronymic string    `json:"patronymic"`
}

type Cars struct {
	ID         int       `json:"id" db:"id"`
	RegNum     string    `json:"regNum" db:"regNum"`
	Mark       string    `json:"mark" db:"mark"`
	Model      string    `json:"model" db:"model"`
	Year       int       `json:"year" db:"year"`
	Name       string    `json:"name"`
	Surname    string    `json:"surname"`
	Patronymic string    `json:"patronymic"`
	CreatedAt  time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt  time.Time `json:"updatedAat" db:"created_at"`
}

func (db *Postgres) CreateCar(ctx context.Context, c Car) error {
	var ownerID int = -1
	// check if owner already exists in db
	query := `
	SELECT id FROM people
	WHERE name = $1 AND surname = $2 AND patronymic = $3;`

	row := db.db.QueryRow(ctx, query, c.Owner.Name, c.Owner.Surname, c.Owner.Patronymic)

	err := row.Scan(&ownerID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("database: %w", err)
	}
	// add owner to people table if they don't exist just yet
	if errors.Is(err, pgx.ErrNoRows) {
		stmt := `
		INSERT INTO people (name, surname, patronymic)
		VALUES ($1, $2, $3)
		RETURNING id;`

		row := db.db.QueryRow(ctx, stmt, c.Owner.Name, c.Owner.Surname, c.Owner.Patronymic)

		err := row.Scan(&ownerID)
		if err != nil {
			return fmt.Errorf("database: %w", err)
		}
	}

	stmt := `
	INSERT INTO cars (reg_num, mark, model, year, owner_id, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7);`

	_, err = db.db.Exec(ctx, stmt, c.RegNum, c.Mark, c.Model, c.Year, ownerID, c.CreatedAt, c.UpdatedAt)
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}

	return nil
}

func (db *Postgres) GetCar(ctx context.Context, c *types.GetCarQuery) ([]*Car, error) {
	var sb strings.Builder

	stmt := `SELECT 
			c.car_id, 
			c.reg_num, 
			c.mark, 
			c.model, 
			c.year, 
			p.name, 
			p.surname, 
			p.patronymic, 
			c.created_at, 
			c.updated_at
			FROM cars c
			JOIN people p on c.owner_id = p.id `

	if c.RegNum != "" || c.Mark != "" || c.Model != "" || c.Year != 0 || c.Name != "" || c.Surname != "" || c.Patronymic != "" {
		stmt += ` WHERE `
	}

	sb.WriteString(stmt)

	var args []any
	
	if c.RegNum != "" {
		args = append(args, c.RegNum)
		sb.WriteString(fmt.Sprintf("reg_num = $%d AND ", len(args)))
	}
	if c.Mark != "" {
		args = append(args, c.Mark)
		sb.WriteString(fmt.Sprintf("mark = $%d AND ", len(args)))
	}
	if c.Model != "" {
		args = append(args, c.Model)
		sb.WriteString(fmt.Sprintf("model = $%d AND ", len(args)))
	}
	if c.Year != 0 {
		args = append(args, c.Year)
		sb.WriteString(fmt.Sprintf("year = $%d AND ", len(args)))
	}
	if c.Name != "" {
		args = append(args, c.Name)
		sb.WriteString(fmt.Sprintf("name = $%d AND ", len(args)))
	}
	if c.Surname != "" {
		args = append(args, c.Surname)
		sb.WriteString(fmt.Sprintf("surname = $%d AND ", len(args)))
	}
	if c.Patronymic != "" {
		args = append(args, c.Patronymic)
		sb.WriteString(fmt.Sprintf("patronymic = $%d AND ", len(args)))
	}

	sqlQuery := strings.TrimSuffix(sb.String(), "AND ")

	switch {
		case c.Limit != 0 && c.Offset == 0: sqlQuery+=fmt.Sprintf("LIMIT %d", c.Limit) 
		case c.Limit == 0 && c.Offset != 0: sqlQuery+=fmt.Sprintf("OFFSET %d", c.Offset)
		case c.Limit != 0 && c.Offset != 0: sqlQuery+=fmt.Sprintf("LIMIT %d OFFSET %d", c.Limit, c.Offset)
	}

	sqlQuery = sqlQuery + ";"

	rows, err := db.db.Query(ctx, sqlQuery, args...)
	if err != nil {
		return nil, err
	}

	cars := make([]*Car, 0)

	for rows.Next() {
		car := new(Car)

		err = rows.Scan(&car.ID, &car.RegNum, &car.Mark, &car.Model, &car.Year, &car.Owner.Name, &car.Owner.Surname, &car.Owner.Patronymic, &car.CreatedAt, &car.UpdatedAt)
		if err != nil {
			return nil, err
		}

		cars = append(cars, car)
	}

	return cars, nil
}

func (db *Postgres) UpdateCar(ctx context.Context, c Car) (Car, error) {
	var ownerID int = -1
	var sb strings.Builder

	sb.WriteString("UPDATE cars SET ")

	var args []any
	var ownerArgs []any

	if c.RegNum != "" {
		args = append(args, c.RegNum)
		sb.WriteString(fmt.Sprintf("reg_num = $%d, ", len(args)))
	}
	if c.Mark != "" {
		args = append(args, c.Mark)
		sb.WriteString(fmt.Sprintf("mark = $%d, ", len(args)))
	}
	if c.Model != "" {
		args = append(args, c.Model)
		sb.WriteString(fmt.Sprintf("model = $%d, ", len(args)))
	}
	if c.Year != 0 {
		args = append(args, c.Year)
		sb.WriteString(fmt.Sprintf("year = $%d, ", len(args)))
	}

	var ownerSB strings.Builder

	ownerSB.WriteString("UPDATE people SET ")

	if c.Owner.Name != "" {
		ownerArgs = append(ownerArgs, c.Owner.Name)
		ownerSB.WriteString(fmt.Sprintf("name = $%d, ", len(ownerArgs)))
	}
	if c.Owner.Surname != "" {
		ownerArgs = append(ownerArgs, c.Owner.Surname)
		ownerSB.WriteString(fmt.Sprintf("surname = $%d, ", len(ownerArgs)))
	}
	if c.Owner.Patronymic != "" {
		ownerArgs = append(ownerArgs, c.Owner.Patronymic)
		ownerSB.WriteString(fmt.Sprintf("patronymic = $%d, ", len(ownerArgs)))
	}

	if c.Owner.Name != "" || c.Owner.Surname != "" || c.Owner.Patronymic != ""  {
		query := "SELECT owner_id FROM cars WHERE car_id = $1;"
		row := db.db.QueryRow(ctx, query, c.ID)
		err := row.Scan(&ownerID)
		if err != nil {
			return Car{}, fmt.Errorf("database: %w", err)
		}
	}
	sqlQuery := strings.TrimSuffix(ownerSB.String(), ", ")

	ownerArgs = append(ownerArgs, ownerID)
	sqlQuery = sqlQuery + fmt.Sprintf(" WHERE id = $%d;", len(ownerArgs))
	
	if len(ownerArgs) > 1 {
		_, err := db.db.Exec(ctx, sqlQuery, ownerArgs...)
		if err != nil {
			return Car{}, fmt.Errorf("database: %w", err)
		}
	}

	now := time.Now()
	args = append(args, now)
	sb.WriteString(fmt.Sprintf("updated_at = $%d ", len(args)))

	args = append(args, c.ID)
	sb.WriteString(fmt.Sprintf("WHERE car_id = $%d;", len(args)))

	if len(ownerArgs) != 0 || len(args) != 0 {
		_, err := db.db.Query(ctx, sb.String(), args...)
		if err != nil {
			return Car{}, fmt.Errorf("database: %w", err)
		}
	}

	query := `SELECT 
	c.car_id, 
	c.reg_num, 
	c.mark, 
	c.model, 
	c.year, 
	p.name, 
	p.surname, 
	p.patronymic, 
	c.created_at, 
	c.updated_at
	FROM cars c
	JOIN people p on c.owner_id = p.id 
	WHERE car_id = $1`

	var car Car

	row := db.db.QueryRow(ctx, query, c.ID)

	err := row.Scan(
		&car.ID, &car.RegNum, &car.Mark, 
		&car.Model, &car.Year, &car.Owner.Name, 
		&car.Owner.Surname, &car.Owner.Patronymic, 
		&car.CreatedAt, &car.UpdatedAt)

	if err != nil {
		return Car{}, fmt.Errorf("database: %w", err)
	}
	
	return car, nil
}

func (db *Postgres) DeleteCar(ctx context.Context, id int) error {
	stmtDeleteCar := `DELETE FROM cars WHERE car_id = $1`

	_, err := db.db.Exec(ctx, stmtDeleteCar, id)
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}

	return nil
}