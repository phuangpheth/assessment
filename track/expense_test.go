package track

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExpenseValidate(t *testing.T) {
	t.Run("ErrAmountInvalid", func(t *testing.T) {
		exp := &Expense{
			ID:     1,
			Amount: -1,
			Title:  "Hot Tea",
			Note:   "Buy tea in the market",
			Tags:   []string{"drinks", "juices"},
		}

		want := ErrAmountInvalid
		err := exp.Validate()

		assert.EqualError(t, err, want.Error())
	})

	t.Run("ErrTitleEmpty", func(t *testing.T) {
		exp := &Expense{
			ID:     1,
			Amount: 10,
			Title:  "",
			Note:   "Buy tea in the market",
			Tags:   []string{"drinks", "juices"},
		}

		want := ErrTitleEmpty
		err := exp.Validate()

		assert.EqualError(t, err, want.Error())
	})

	t.Run("Validate No Error", func(t *testing.T) {
		exp := &Expense{
			ID:     1,
			Amount: 10,
			Title:  "Hot Tea",
			Note:   "Buy tea in the market",
			Tags:   []string{"drinks", "juices"},
		}

		err := exp.Validate()

		assert.NoError(t, err)
	})
}
