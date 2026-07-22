package postgres

import (
	"context"
	"errors"
	"testing"
)

func TestParseDumpFileName(t *testing.T) {
	tests := []struct {
		name      string
		file      string
		wantTable string
		wantType  string
		wantErr   bool
	}{
		{name: "binary", file: "DATA_orders.bin", wantTable: "orders", wantType: "binary"},
		{name: "csv", file: "DATA_orders.csv", wantTable: "orders", wantType: "csv"},
		{name: "underscore in table", file: "DATA_order_items.bin", wantTable: "order_items", wantType: "binary"},
		{name: "with directory", file: "sub/DATA_orders.csv", wantTable: "orders", wantType: "csv"},
		{name: "bad extension", file: "DATA_orders.txt", wantErr: true},
		{name: "missing prefix", file: "orders.bin", wantErr: true},
		{name: "empty table", file: "DATA_.bin", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			table, fileType, err := parseDumpFileName(tt.file)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error for %q, got none", tt.file)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if table != tt.wantTable {
				t.Errorf("table = %q, want %q", table, tt.wantTable)
			}
			if fileType != tt.wantType {
				t.Errorf("fileType = %q, want %q", fileType, tt.wantType)
			}
		})
	}
}

func TestRunParallel(t *testing.T) {
	names := []string{"a", "b", "c", "d"}
	success, failed := runParallel(context.Background(), names, 2, func(name string) task {
		return func(ctx context.Context) error {
			if name == "b" || name == "d" {
				return errors.New("boom")
			}
			return nil
		}
	})

	if success != 2 {
		t.Errorf("success = %d, want 2", success)
	}
	if failed != 2 {
		t.Errorf("failed = %d, want 2", failed)
	}
}

func TestRunParallelClampsThreads(t *testing.T) {
	// threads < 1 should be clamped to 1 rather than deadlocking.
	success, failed := runParallel(context.Background(), []string{"a"}, 0, func(name string) task {
		return func(ctx context.Context) error { return nil }
	})
	if success != 1 || failed != 0 {
		t.Errorf("success=%d failed=%d, want 1/0", success, failed)
	}
}
