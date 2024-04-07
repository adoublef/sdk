package date_test

import (
	"context"
	"encoding/json"
	"testing"

	"go.adoublef.dev/is"
	"go.adoublef.dev/sdk/database/sql3"
	. "go.adoublef.dev/sdk/time/date"
)

func Test_Parse(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		is := is.NewRelaxed(t)

		d, err := Parse("2016-10-18")
		is.NoErr(err) // date.Parse

		is.Equal(d.Year, 2016)
		is.Equal(d.Month, October)
		is.Equal(d.Day, 18)
	})
}

func Test_Date_UnmarshalJSON(t *testing.T) {
	t.Run("Civil", func(t *testing.T) {
		is := is.NewRelaxed(t)

		var s = `"2016-10-18"`

		var d Date
		is.NoErr(json.Unmarshal([]byte(s), &d))
	})
	t.Run("RFC3339", func(t *testing.T) {
		is := is.NewRelaxed(t)

		var s = `"2016-10-18T00:00:00Z"`

		var d Date
		is.NoErr(json.Unmarshal([]byte(s), &d))
	})

	t.Run("Informal", func(t *testing.T) {
		is := is.NewRelaxed(t)

		var s = `"2016-10-18t00:00:00z"`

		var d Date
		is.NoErr(json.Unmarshal([]byte(s), &d))
	})
}

func Test_Date_MarshalJSON(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		is := is.NewRelaxed(t)

		a := Date{Year: 2016, Month: October, Day: 18}

		p, err := json.Marshal(a)
		is.NoErr(err) // json.Marshal

		var b Date
		is.NoErr(json.Unmarshal(p, &b))

		is.Equal(a, b)
	})
}

func Test_Date_Scan(t *testing.T) {
	db, err := sql3.Open(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatalf("sql3.Open: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	_, err = db.Exec(context.TODO(), `
	create table t (id int not null, d text) strict;
	insert into t (id, d) values (1, '2022-10-18');
	insert into t (id, d) values (2, null);`)
	if err != nil {
		t.Fatalf("(sql3.DB).Exec: %v", err)
	}

	t.Run("OK", func(t *testing.T) {
		is := is.NewRelaxed(t)
		var d Date
		err = db.QueryRow(context.TODO(), `select d from t where id = 1`).Scan(&d)
		is.NoErr(err) // sql3.DB).QueryRow
	})

	t.Run("Zero", func(t *testing.T) {
		is := is.NewRelaxed(t)
		var d Date
		err = db.QueryRow(context.TODO(), `select d from t where id = 2`).Scan(&d)
		is.NoErr(err) // sql3.DB).QueryRow
	})

	t.Run("Nil", func(t *testing.T) {
		is := is.NewRelaxed(t)
		var d *Date
		err = db.QueryRow(context.TODO(), `select d from t where id = 2`).Scan(&d)
		is.NoErr(err) // sql3.DB).QueryRow

		is.Equal(d, nil)
	})
}

func Test_Date_Value(t *testing.T) {
	db, err := sql3.Open(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatalf("sql3.Open: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	_, err = db.Exec(context.TODO(), `
	create table t (id int not null, d text) strict;`)
	if err != nil {
		t.Fatalf("(sql3.DB).Exec: %v", err)
	}

	t.Run("OK", func(t *testing.T) {
		is := is.NewRelaxed(t)
		a := Date{2006, January, 02}

		_, err = db.Exec(context.TODO(), `insert into t (id, d) values (1, ?)`, a)
		is.NoErr(err) // sql3.DB).Exec

		var b Date
		err = db.QueryRow(context.TODO(), `select d from t where id = 1`).Scan(&b)

		is.True(a.Compare(b) == 0)
	})
}
