package expressions

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEvaluate(t *testing.T) {
	items := []struct {
		input    string
		expected string
		extra    map[string]any
	}{
		{input: "{{url_encode('test}aaa')}}", expected: "test%7Daaa", extra: map[string]any{}},
		{input: "{{hex_encode('PING')}}", expected: "50494e47", extra: map[string]any{}},
		{input: "{{hex_encode('{{')}}", expected: "7b7b", extra: map[string]any{}},
		{input: `{{concat("{{", 123, "*", 123, "}}")}}`, expected: "{{123*123}}", extra: map[string]any{}},
		{input: `{{concat("{{", "123*123", "}}")}}`, expected: "{{123*123}}", extra: map[string]any{}},
		{input: `{{"{{" + '123*123' + "}}"}}`, expected: `{{"{{" + '123*123' + "}}"}}`, extra: map[string]any{}},
		{input: `{{a + '123*123' + b}}`, expected: `aa123*123bb`, extra: map[string]any{"a": "aa", "b": "bb"}},
		{input: `{{concat(123,'*',123)}}`, expected: "123*123", extra: map[string]any{}},
		{input: `{{1+1}}`, expected: "{{1+1}}", extra: map[string]any{}},
		{input: `{{"1"+"1"}}`, expected: `{{"1"+"1"}}`, extra: map[string]any{}},
		{input: `{{"1" + '*' + "1"}}`, expected: `{{"1" + '*' + "1"}}`, extra: map[string]any{}},
		{input: `{{"a" + 'b' + "c"}}`, expected: `{{"a" + 'b' + "c"}}`, extra: map[string]any{}},
		{input: `{{10*2}}`, expected: `{{10*2}}`, extra: map[string]any{}},
		{input: `{{10/2}}`, expected: `{{10/2}}`, extra: map[string]any{}},
		{input: `{{10-2}}`, expected: `{{10-2}}`, extra: map[string]any{}},
		{input: "test", expected: "test", extra: map[string]any{}},
		{input: "{{hex_encode(Item)}}", expected: "50494e47", extra: map[string]any{"Item": "PING"}},
		{input: "{{hex_encode(Item)}}\r\n", expected: "50494e47\r\n", extra: map[string]any{"Item": "PING"}},
		{input: "{{someTestData}}{{hex_encode('PING')}}", expected: "{{someTestData}}50494e47", extra: map[string]any{}},
		{input: `_IWP_JSON_PREFIX_{{base64("{\"iwp_action\":\"add_site\",\"params\":{\"username\":\"\"}}")}}`, expected: "_IWP_JSON_PREFIX_eyJpd3BfYWN0aW9uIjoiYWRkX3NpdGUiLCJwYXJhbXMiOnsidXNlcm5hbWUiOiIifX0=", extra: map[string]any{}},
		{input: "{{}}", expected: "{{}}", extra: map[string]any{}},
		{input: `"{{hex_encode('PING')}}"`, expected: `"50494e47"`, extra: map[string]any{}},
	}
	for _, item := range items {
		value, err := Evaluate(item.input, item.extra)
		require.Nil(t, err, "could not evaluate helper")

		require.Equal(t, item.expected, value, "could not get correct expression")
	}
}

func TestEval(t *testing.T) {
	items := []struct {
		input    string
		values   map[string]any
		expected any
	}{
		{input: "'a' + 'a'", values: nil, expected: "aa"},
		{input: "10 + to_number(b)", values: map[string]any{"b": "4"}, expected: float64(14)},
	}
	for _, item := range items {
		value, err := Eval(item.input, item.values)
		require.Nil(t, err, "could not evaluate helper")
		require.Equal(t, item.expected, value, "could not get correct expression")
	}
}
