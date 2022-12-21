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

func (s *Service) Update(ctx context.Context, e *Expense) (*Expense, error) {
	exp, err := getExpenseByID(ctx, s.db, e.ID)
	if err != nil {
		return nil, fmt.Errorf("getExpenseByID(%d): %w", e.ID, err)
	}
	exp.Amount = e.Amount
	exp.Title = e.Title
	exp.Note = e.Note
	exp.Tags = e.Tags
	if err := updateExpense(ctx, s.db, exp); err != nil {
		return nil, fmt.Errorf("updateExpense(): %w", err)
	}
	return exp, nil
}

func (s *Service) GetByID(ctx context.Context, id int64) (*Expense, error) {
	exp, err := getExpenseByID(ctx, s.db, id)
	if err != nil {
		return nil, fmt.Errorf("getExpenseByID(%d): %w", id, err)
	}
	return exp, nil
}

func (s *Service) List(ctx context.Context) ([]Expense, error) {
	exps, err := listExpenses(ctx, s.db)
	if err != nil {
		return nil, fmt.Errorf("listExpenses(): %w", err)
	}
	return exps, nil
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

func updateExpense(ctx context.Context, db *sql.DB, e *Expense) error {
	query, args, err := sq.Update("expenses").
		Set("amount", e.Amount).
		Set("title", e.Title).
		Set("note", e.Note).
		Set("tags", pq.Array(e.Tags)).
		Where(sq.Eq{"id": e.ID}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return err
	}
	if _, err := db.ExecContext(ctx, query, args...); err != nil {
		return err
	}
	return nil
}

func getExpenseByID(ctx context.Context, db *sql.DB, id int64) (*Expense, error) {
	query, args, err := sq.Select(expenseColumns...).
		From("expenses").
		Where(sq.Eq{"id": id}).
		Limit(1).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, err
	}

	row := db.QueryRowContext(ctx, query, args...)
	e, err := scanExpense(row.Scan)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func listExpenses(ctx context.Context, db *sql.DB) ([]Expense, error) {
	query, args, err := sq.Select(expenseColumns...).
		From("expenses").
		OrderBy("id DESC").
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	exps := make([]Expense, 0)
	for rows.Next() {
		e, err := scanExpense(rows.Scan)
		if err != nil {
			return nil, err
		}
		exps = append(exps, e)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return exps, nil
}

var expenseColumns = []string{
	"id",
	"amount",
	"title",
	"note",
	"tags",
}

func scanExpense(scan func(...any) error) (e Expense, _ error) {
	return e, scan(
		&e.ID,
		&e.Amount,
		&e.Title,
		&e.Note,
		pq.Array(&e.Tags),
	)
}
