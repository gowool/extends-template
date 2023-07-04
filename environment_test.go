package et_test

import (
	"context"
	et "github.com/gowool/extends-template"
	"github.com/stretchr/testify/assert"
	"html/template"
	"testing"
	"time"
)

func TestEnvironment_Debug(t *testing.T) {
	env := et.NewEnvironment(wrapLoader{})
	for _, d := range []bool{true, true, false, false} {
		assert.NotNil(t, env.Debug(d))
	}
}

func TestEnvironment_Funcs(t *testing.T) {
	env := et.NewEnvironment(wrapLoader{})

	assert.NotNil(t, env.Funcs(template.FuncMap{"test": func() {}}))
}

func TestEnvironment_Global(t *testing.T) {
	env := et.NewEnvironment(wrapLoader{})

	assert.NotNil(t, env.Global("t1.html", "t2.html"))
}

func TestEnvironment_Load(t *testing.T) {
	env := et.NewEnvironment(wrapLoader{t: time.Now().Unix()})

	scenarios := []struct {
		view    string
		isError bool
	}{
		{
			view:    "no-view.html",
			isError: true,
		},
		{
			view:    "@main/no-view.html",
			isError: true,
		},
		{
			view:    "@main/view.html",
			isError: false,
		},
		{
			view:    "@main/layout.html",
			isError: false,
		},
	}

	for _, s := range scenarios {
		for range []struct{}{{}, {}} {
			w, err := env.Load(context.TODO(), s.view)
			if s.isError {
				assert.Nil(t, w)
				assert.Error(t, err)
			} else {
				assert.NotNil(t, w)
				assert.Nil(t, err)
				if assert.NotNil(t, w.HTML) {
					assert.Equal(t, s.view, w.HTML.Name())
				}
			}
		}
	}
}
