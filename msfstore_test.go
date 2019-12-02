package MSFStore

import (
	"math/rand"
	"reflect"
	"strconv"
	"testing"
)

func TestMSFS(t *testing.T) {
	t.Run("creation", func(t *testing.T) {
		x := New(3)
		if (x.Resolution != 3) || (x.Registers == nil) {
			t.Fail()
		}
	})
	x := New(2)
	z := New(5)
	t.Run("projection sanity", func(t *testing.T) {
		y := x.project(345.234212312)
		w := z.project(345.234212312)
		if (y != "3.45e+02") || (w != "3.45234e+02") {
			t.Log(y, w)
			t.Fail()
		}
	})
	t.Run("projection roundtrip sanity", func(t *testing.T) {
		w := x.project(456.32)
		z, err := strconv.ParseFloat(w, 64)
		if err != nil {
			t.Log(w, z, err)
			t.Fail()
		}
		w2 := x.project(z)
		if w2 != w {
			t.Log(w, z, w2, err)
			t.Fail()
		}
	})

	t.Run("Insert and read", func(t *testing.T) {
		x = x.Insert(123.3, 23.32)
		y := x.Read(123.3)
		if y != 23.32 {
			t.Log(y, "!= 23.32")
			t.Fail()
		}
	})
	t.Run("Total", func(t *testing.T) {
		x = x.Insert(23456.43, 11)
		y := x.Total()
		if y != 34.32 {
			t.Log(y)
			t.Fail()
		}
	})
	t.Run("(De)Serialization roundtrip equality", func(t *testing.T) {
		xs := x.Serialize()
		xd, err := Deserialize(xs, 2)
		xs2 := xd.Serialize()
		if len(xs2) != len(xs) {
			t.Log(xs, xs2, err)
			t.Fail()
		}
		for ii, xx := range xs2 {
			if xs2[ii] != xx {
				t.Log(xs2[ii], xx, err)
				t.Fail()
			}
		}
	})

	t.Run("Interchange", func(t *testing.T) {
		l := x.ToRawHist()
		z := FromRawHist(l, 2)
		if !reflect.DeepEqual(z, x) {
			t.Log(z, "!=", x)
			t.Fail()
		}
	})
	t.Run("Combine", func(t *testing.T) {
		y := x.Combine(x)
		if 2*x.Total() != y.Total() {
			t.Log(x, y)
			t.Fail()
		}
	})
	t.Run("Cancel", func(t *testing.T) {
		y := x.Cancel(x)
		if 0 != y.Total() {
			t.Log(x, y)
			t.Fail()
		}
	})
	t.Run("Min", func(t *testing.T) {
		q := New(2)
		z := x.Min(q)
		if z.Total() == x.Total() {
			t.Log(z, x)
			t.Fail()
		}
		z = x.Min(x)
		if z.Total() != x.Total() {
			t.Log(z, x)
			t.Fail()
		}
		q.Insert(123., 6.)
		q.Insert(23456., 1000.)
		z = x.Min(q)
		if z.Total() != 6.+11. {
			t.Log(z, x)
			t.Fail()
		}
	})

}

var (
	x = New(3)
)

func BenchmarkProject(b *testing.B) {
	for i := 0; i < b.N; i++ {
		x.project(1.23134253452)
	}
}

func BenchmarkInsert(b *testing.B) {
	for i := 0; i < b.N; i++ {
		x.Insert(rand.Float64(), rand.Float64())
	}
}

func BenchmarkRead(b *testing.B) {
	for i := 0; i < b.N; i++ {
		x.Read(rand.Float64())
	}
}

func BenchmarkTotal(b *testing.B) {
	for i := 0; i < b.N; i++ {
		x.Total()
	}
}
