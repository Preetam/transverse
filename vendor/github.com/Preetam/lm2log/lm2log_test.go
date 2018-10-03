package lm2log

import (
	"reflect"
	"testing"
)

func TestPrepare(t *testing.T) {
	lg, err := New("/tmp/lm2log_Prepare.lm2")
	if err != nil {
		t.Fatal(err)
	}
	defer lg.Destroy()

	err = lg.Prepare("a")
	if err != nil {
		t.Fatal(err)
	}

	prepared, err := lg.Prepared()
	if err != nil {
		t.Fatal(err)
	}

	if prepared != 1 {
		t.Errorf("expected prepared to be %d, got %d", 1, prepared)
	}
}

func TestPrepareCommit(t *testing.T) {
	lg, err := New("/tmp/lm2log_PrepareCommit.lm2")
	if err != nil {
		t.Fatal(err)
	}
	defer lg.Destroy()

	err = lg.Prepare("a")
	if err != nil {
		t.Fatal(err)
	}

	err = lg.Commit()
	if err != nil {
		t.Fatal(err)
	}

	committed, err := lg.Committed()
	if err != nil {
		t.Fatal(err)
	}

	if committed != 1 {
		t.Errorf("expected committed to be %d, got %d", 1, committed)
	}

	err = lg.Rollback()
	if err != nil {
		t.Fatal(err)
	}

	committed, err = lg.Committed()
	if err != nil {
		t.Fatal(err)
	}

	if committed != 1 {
		t.Errorf("expected committed to be %d, got %d", 1, committed)
	}
}

func TestPrepareRollback(t *testing.T) {
	lg, err := New("/tmp/lm2log_PrepareRollback.lm2")
	if err != nil {
		t.Fatal(err)
	}
	defer lg.Destroy()

	err = lg.Prepare("a")
	if err != nil {
		t.Fatal(err)
	}

	prepared, err := lg.Prepared()
	if err != nil {
		t.Fatal(err)
	}

	if prepared != 1 {
		t.Errorf("expected prepared to be %d, got %d", 1, prepared)
	}

	err = lg.Rollback()
	if err != nil {
		t.Fatal(err)
	}

	prepared, err = lg.Prepared()
	if err == nil {
		t.Errorf("expected an error getting prepared data, but got %d", prepared)
	}
}

func TestCompact(t *testing.T) {
	expected := [][2]string{
		{"100", "x"},
		{"90", "x"},
		{"91", "x"},
		{"92", "x"},
		{"93", "x"},
		{"94", "x"},
		{"95", "x"},
		{"96", "x"},
		{"97", "x"},
		{"98", "x"},
		{"99", "x"},
		{"committed", "100"},
	}

	lg, err := New("/tmp/lm2log_Compact.lm2")
	if err != nil {
		t.Fatal(err)
	}
	defer lg.Destroy()

	for i := 0; i < 100; i++ {
		err = lg.Prepare("x")
		if err != nil {
			t.Fatal(err)
		}
		err = lg.Commit()
		if err != nil {
			t.Fatal(err)
		}
	}

	err = lg.Compact(10)
	if err != nil {
		t.Fatal(err)
	}

	cur, err := lg.col.NewCursor()
	if err != nil {
		t.Fatal(err)
	}

	records := [][2]string{}
	for cur.Next() {
		records = append(records, [2]string{cur.Key(), cur.Value()})
	}

	if !reflect.DeepEqual(expected, records) {
		t.Errorf("expected records %v, got %v", expected, records)
	}
}
