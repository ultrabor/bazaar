package database

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
)

// fakeTx implements pgx.Tx through the embedded interface; only Commit and
// Rollback are called by withinTx, any other method would panic.
type fakeTx struct {
	pgx.Tx
	committed  bool
	rolledBack bool
}

func (f *fakeTx) Commit(ctx context.Context) error {
	f.committed = true
	return nil
}

func (f *fakeTx) Rollback(ctx context.Context) error {
	f.rolledBack = true
	return nil
}

type fakeBeginner struct {
	tx    *fakeTx
	begun int
}

func (f *fakeBeginner) Begin(ctx context.Context) (pgx.Tx, error) {
	f.begun++
	return f.tx, nil
}

func TestWithinTxRollsBackWhenFnFails(t *testing.T) {
	b := &fakeBeginner{tx: &fakeTx{}}
	fnErr := errors.New("insert failed")

	err := withinTx(context.Background(), b, func(ctx context.Context) error {
		return fnErr
	})

	if !errors.Is(err, fnErr) {
		t.Fatalf("withinTx error = %v, want %v", err, fnErr)
	}
	if !b.tx.rolledBack || b.tx.committed {
		t.Fatalf("rolledBack = %v, committed = %v, want true, false", b.tx.rolledBack, b.tx.committed)
	}
}

func TestWithinTxCommitsWhenFnSucceeds(t *testing.T) {
	b := &fakeBeginner{tx: &fakeTx{}}

	err := withinTx(context.Background(), b, func(ctx context.Context) error {
		return nil
	})

	if err != nil {
		t.Fatalf("withinTx error = %v", err)
	}
	if !b.tx.committed || b.tx.rolledBack {
		t.Fatalf("committed = %v, rolledBack = %v, want true, false", b.tx.committed, b.tx.rolledBack)
	}
}

func TestWithinTxRollsBackAndRepanics(t *testing.T) {
	b := &fakeBeginner{tx: &fakeTx{}}

	defer func() {
		if p := recover(); p != "boom" {
			t.Fatalf("recovered %v, want boom", p)
		}
		if !b.tx.rolledBack || b.tx.committed {
			t.Fatalf("rolledBack = %v, committed = %v, want true, false", b.tx.rolledBack, b.tx.committed)
		}
	}()

	_ = withinTx(context.Background(), b, func(ctx context.Context) error {
		panic("boom")
	})
}

func TestWithinTxNestedReusesOuterTx(t *testing.T) {
	b := &fakeBeginner{tx: &fakeTx{}}

	err := withinTx(context.Background(), b, func(ctx context.Context) error {
		outer, _ := txFromContext(ctx)
		return withinTx(ctx, b, func(ctx context.Context) error {
			inner, ok := txFromContext(ctx)
			if !ok || inner != outer {
				t.Fatal("nested call did not get the outer tx")
			}
			return nil
		})
	})

	if err != nil {
		t.Fatalf("withinTx error = %v", err)
	}
	if b.begun != 1 {
		t.Fatalf("Begin calls = %d, want 1", b.begun)
	}
}
