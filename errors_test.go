package log

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func init() {
	Run("telegram_url",
		"NOTIFY", "clickhouse_url", "STAGE")
}

type TestSender struct {
	test *testing.T
	string string
	function string
}

func (tes *TestSender) Send(s string) {
	assert.Contains(tes.test, s, tes.string, "string must be equal")
	assert.Contains(tes.test, s, tes.function, "function must be equal")
}

func Test_Print(t *testing.T) {
	fmt.Print("level")
	fmt.Print(fmt.Sprintf("%s", LevelInfo))
}

func TestClickHouse_Send(t *testing.T) {
	te := &TestSender{
		test: t,
		string: "fdsa",
		function: "function1",
	}
	SendCH(te, LevelError, te.string, te.function)
}

func TestTelegramBot_Send(t *testing.T) {
	te := &TestSender{
		test: t,
		string: "test123",
		function: "function34",
	}
	SendTg(te, LevelError, te.string, te.function)
}
