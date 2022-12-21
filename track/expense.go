package track

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/lib/pq"
)

type Service struct {
	db *sql.DB
}

// ErrNotFound is returned when the expense could not be found.
var ErrNotFound = errors.New("not found")

// ErrAmountInvalid is returned when the amount of expense is less than zero.
var ErrAmountInvalid = errors.New("amount must be greater than zero")

// ErrTitleEmpty is returned when the title is empty.
var ErrTitleEmpty = errors.New("empty title")

func NewService(db *sql.DB) *Service {
	return &Service{
		db: db,
	}
}

func (s *Service) Save(ctx context.Context, e *Expense) (*Expense, error) {
	if err := createExpense(ctx, s.db, e); err != nil {
		return nil, fmt.Errorf("createExpense(): %w", err)
	}
	return e, nil
}

type Expense struct {
	ID     int64    `json:"id"`
	Amount float64  `json:"amount"`
	Title  string   `json:"title"`
	Note   string   `json:"note"`
	Tags   []string `json:"tags"`
}

func (e *Expense) Validate() error {
	if e.Amount <= 0 {
		return ErrAmountInvalid
	}
	if e.Title == "" {
		return ErrTitleEmpty
	}
	return nil
}

func createExpense(ctx context.Context, db *sql.DB, e *Expense) error {
	query, args, err := sq.Insert("expenses").
		Columns(
			"amount",
			"title",
			"note",
			"tags",
		).
		Values(
			e.Amount,
			e.Title,
			e.Note,
			pq.Array(e.Tags),
		).
		Suffix(`
      RETURNING id, amount, title, note, tags
    `).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return err
	}

	row := db.QueryRowContext(ctx, query, args...)
	if err := row.Scan(
		&e.ID,
		&e.Amount,
		&e.Title,
		&e.Note,
		pq.Array(&e.Tags),
	); err != nil {
		return err
	}
	return nil
}
