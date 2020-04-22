package tools

import "testing"

func Test_CreateScript(t *testing.T) {
	result, err := CreateExecutingScript([]string{
		"ls -la",
		`echo "Hello world"`,
	})
	if err != nil {
		t.Error("error while compiling script")
	}
	t.Log(string(result))
}
