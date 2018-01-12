package storagetests

import (
	"encoding/json"
	"testing"

	"github.com/anothermemory/storage"
	"github.com/anothermemory/unit"
	"github.com/stretchr/testify/assert"
)

// CreateFunc represents function which must return created storage object
type CreateFunc func() storage.Interface

// LoadFromConfigFunc represents test function which return storage configured from JSON config
type LoadFromConfigFunc func(b []byte) (storage.Interface, error)

// Func represents test function for single test-case
type Func func(t *testing.T, c CreateFunc, l LoadFromConfigFunc, is *assert.Assertions)

// RunStorageTests performs full test run for all test-cases for given storage
func RunStorageTests(t *testing.T, c CreateFunc, l LoadFromConfigFunc) {
	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) { test.testFunc(t, c, l, assert.New(t)) })
	}

}

var tests = []struct {
	title    string
	testFunc Func
}{
	{"Storage is not created initially when initialized first time with given arguments", func(t *testing.T, c CreateFunc, l LoadFromConfigFunc, is *assert.Assertions) {
		is.False(c().IsCreated())
	}},
	{"Storage can be successfully created", func(t *testing.T, c CreateFunc, l LoadFromConfigFunc, is *assert.Assertions) {
		s := c()
		is.NoError(s.Create())
		is.True(s.IsCreated())
	}},
	{"Storage can not be used before it will be created", func(t *testing.T, c CreateFunc, l LoadFromConfigFunc, is *assert.Assertions) {
		s := c()
		u := unit.NewUnit()
		is.Error(s.SaveUnit(u))
		is.Error(s.RemoveUnit(u))
		u, e := s.LoadUnit("123")
		is.Error(e)
		is.Nil(u)
	}},
	{"Storage can be removed if not created before", func(t *testing.T, c CreateFunc, l LoadFromConfigFunc, is *assert.Assertions) {
		s := c()
		is.NoError(s.Remove())
	}},
	{"Storage can be removed if was created before", func(t *testing.T, c CreateFunc, l LoadFromConfigFunc, is *assert.Assertions) {
		s := c()
		is.NoError(s.Create())
		is.NoError(s.Remove())
	}},
	{"Storage is not created when removed", func(t *testing.T, c CreateFunc, l LoadFromConfigFunc, is *assert.Assertions) {
		s := c()
		is.NoError(s.Create())
		is.NoError(s.Remove())
		is.False(c().IsCreated())
	}},
	{"Storage can handle all supported simple unit types", func(t *testing.T, c CreateFunc, l LoadFromConfigFunc, is *assert.Assertions) {
		unitUnit := unit.NewUnit(unit.OptionTitle("MyUnit"))
		unitTextPlain := unit.NewTextPlain(unit.OptionTitle("MyUnit"), unit.OptionTextPlainData("MyData"))
		unitTextMarkdown := unit.NewTextMarkdown(unit.OptionTitle("MyUnit"), unit.OptionTextMarkdownData("MyData"))
		unitTextCode := unit.NewTextCode(unit.OptionTitle("MyUnit"), unit.OptionTextCodeData("MyData"), unit.OptionTextCodeLanguage("MyLang"))

		unitTodo := unit.NewTodo(unit.OptionTitle("MyUnit"))
		t1 := unitTodo.NewItem()
		t1.SetData("Data1")
		t1.SetDone(true)
		t2 := unitTodo.NewItem()
		t2.SetData("Data2")
		t2.SetDone(false)
		unitTodo.SetItems([]unit.TodoItem{t1, t2})

		unitsTests := []unit.Unit{
			unitUnit,
			unitTextPlain,
			unitTextMarkdown,
			unitTextCode,
			unitTodo,
		}

		for _, u := range unitsTests {
			t.Run(u.Type().String(), func(t *testing.T) {
				is := assert.New(t)
				s := c()
				is.NoError(s.Create())
				is.NoError(s.SaveUnit(u))
				l, e := s.LoadUnit(u.ID())
				is.NoError(e)
				is.True(unit.Equal(u, l))
				is.NoError(s.RemoveUnit(l))
				r, e := s.LoadUnit(l.ID())
				is.Error(e)
				is.Nil(r)
			})
		}
	}},
	{"Storage can handle list unit", func(t *testing.T, c CreateFunc, l LoadFromConfigFunc, is *assert.Assertions) {
		unitUnit := unit.NewUnit(unit.OptionTitle("MyUnit"))
		unitTextPlain := unit.NewTextPlain(unit.OptionTitle("MyUnit"), unit.OptionTextPlainData("MyData"))
		unitTextMarkdown := unit.NewTextMarkdown(unit.OptionTitle("MyUnit"), unit.OptionTextMarkdownData("MyData"))
		unitTextCode := unit.NewTextCode(unit.OptionTitle("MyUnit"), unit.OptionTextCodeData("MyData"), unit.OptionTextCodeLanguage("MyLang"))

		unitTodo := unit.NewTodo(unit.OptionTitle("MyUnit"))
		t1 := unitTodo.NewItem()
		t1.SetData("Data1")
		t1.SetDone(true)
		t2 := unitTodo.NewItem()
		t2.SetData("Data2")
		t2.SetDone(false)
		unitTodo.SetItems([]unit.TodoItem{t1, t2})

		unitList := unit.NewList(unit.OptionTitle("MyUnit"))
		unitList.SetItems([]unit.Unit{
			unitUnit,
			unitTextPlain,
			unitTextMarkdown,
			unitTextCode,
			unitTodo,
		})

		s := c()
		is.NoError(s.Create())
		is.NoError(s.SaveUnit(unitUnit))
		is.NoError(s.SaveUnit(unitTextPlain))
		is.NoError(s.SaveUnit(unitTextMarkdown))
		is.NoError(s.SaveUnit(unitTextCode))
		is.NoError(s.SaveUnit(unitTodo))
		is.NoError(s.SaveUnit(unitList))
		lu, e := s.LoadUnit(unitList.ID())
		is.NoError(e)
		is.True(unit.Equal(unitList, lu))
		is.NoError(s.RemoveUnit(lu))
		r, e := s.LoadUnit(lu.ID())
		is.Error(e)
		is.Nil(r)
	}},
	{"Nil unit cannot be saved", func(t *testing.T, c CreateFunc, l LoadFromConfigFunc, is *assert.Assertions) {
		s := c()
		is.NoError(s.Create())
		is.Error(s.SaveUnit(nil))
	}},
	{"Nil unit cannot be removed", func(t *testing.T, c CreateFunc, l LoadFromConfigFunc, is *assert.Assertions) {
		s := c()
		is.NoError(s.Create())
		is.Error(s.RemoveUnit(nil))
	}},
	{"Empty ID cannot be used to load unit", func(t *testing.T, c CreateFunc, l LoadFromConfigFunc, is *assert.Assertions) {
		s := c()
		is.NoError(s.Create())
		lu, e := s.LoadUnit("")
		is.Error(e)
		is.Nil(lu)
	}},
	{"Directory config can be JSON serialized/deserialized ", func(t *testing.T, c CreateFunc, l LoadFromConfigFunc, is *assert.Assertions) {
		if nil == l {
			t.SkipNow()
			return
		}
		u := unit.NewTextPlain(unit.OptionTitle("MyUnit"), unit.OptionTextPlainData("MyData"))
		s := c()
		is.NoError(s.Create())
		is.NoError(s.SaveUnit(u))

		config, err := json.Marshal(s)
		is.NoError(err)
		is.NotNil(config)

		sl, err := l(config)
		is.NoError(err)
		is.NotNil(sl)

		lu, err := sl.LoadUnit(u.ID())
		is.NoError(err)
		is.NotNil(lu)
		is.True(unit.Equal(u, lu))
	}},
}
