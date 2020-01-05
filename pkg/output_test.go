package pkg

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func compare(t *testing.T, output strings.Builder) {
	path, err := filepath.Abs(fmt.Sprintf("../testdata/%s.golden", t.Name()))
	assert.Nil(t, err)

	goldenOutput, err := ioutil.ReadFile(path)
	assert.Nil(t, err)

	assert.Equal(t, string(goldenOutput), output.String())
	assert.Nil(t, err)
}

func TestBasicTableOutput(t *testing.T) {
	var output strings.Builder
	table := NewTable()
	table.SetOutput(&output)
	table.SetHeader([]interface{}{"FIRST", "SECOND"})
	table.AppendRow([]interface{}{1, 2})
	table.Render()

	compare(t, output)
}

func TestBasicJsonOutput(t *testing.T) {
	t.Skip()
	// var output strings.Builder
}
