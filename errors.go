package log

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
	"golang.org/x/xerrors"
	"io"
	"net/url"
	"os"
	"time"
)

var (
	chl = NewChlogger()
)

// Chlogger struct used to store chlogger specific data
type Chlogger struct {
	Log     Logger
	Out     io.Writer
	tgurl   string
	service string
	reqUrl  string
	env     string
}

// Level Enum used to store error level
type Level int

const (
	LevelDebug Level = 0
	LevelInfo  Level = 1
	LevelError Level = 2
	LevelFatal Level = 3
	LevelPanic Level = 4
)

func (l Level) String() string {
	names := [...]string{
		"DEBUG",
		"INFO",
		"ERROR",
		"FATAL",
		"PANIC",
	}
	if l < LevelDebug || l > LevelPanic {
		return fmt.Sprintf("%d", int(l))
	}
	return names[l]
}

func Infof(s string, args ...interface{}) {
	chl.Infof(s, args...)
}

func Printf(s string, args ...interface{}) {
	chl.Printf(s, args...)
}

func Errorf(s string, args ...interface{}) {
	chl.Errorf(s, args...)
}

func Error(args ...interface{}) {
	chl.Error(args...)
}

func Info(args ...interface{}) {
	chl.Info(args...)
}

func Fatal(args ...interface{}) {
	chl.Fatal(args...)
}

func Panic(args ...interface{}) {
	chl.Panic(args...)
}

func Debug(args ...interface{}) {
	chl.Debug(args...)
}

func Fatalf(s string, args ...interface{}) {
	chl.Fatalf(s, args...)
}

type MyError struct {
	Message string
	frame   xerrors.Frame
}

type Logger interface {
	Debugf(string, ...interface{})
	Infof(string, ...interface{})
	Info(...interface{})
	Printf(string, ...interface{})
	Error(...interface{})
	Errorf(string, ...interface{})
	Fatal(...interface{})
	Panic(...interface{})
	Debug(...interface{})
	Fatalf(string, ...interface{})
}

func SetChlogger(chlogger Logger) {
	chl = chlogger
}

func NewChlogger() Logger {
	return &Chlogger{
		Log: logrus.New(),
		Out: os.Stdout,
	}
}

func (m *MyError) Error() string {
	return m.Message
}

func (m *MyError) Format(f fmt.State, c rune) {
	// implements fmt.Formatter
	xerrors.FormatError(m, f, c)
}

func (m *MyError) FormatError(p xerrors.Printer) error {
	// implements xerrors.Formatter
	if p.Detail() {
		m.frame.Format(p)
	}
	return nil
}

func (c *Chlogger) Debugf(s string, args ...interface{}) {
	c.Log.Debugf(s, args...)
}

func (c *Chlogger) Error(args ...interface{}) {
	var s string
	for _, arg := range args {
		s += fmt.Sprintf("%v", arg)
	}
	err := &MyError{Message: "", frame: xerrors.Caller(2)}
	SendCH(&ClickHouse{}, LevelError, s, fmt.Sprintf("%+v\n", err))
	SendTg(&TelegramBot{}, LevelError, s, fmt.Sprintf("%+v\n", err))
	c.Log.Error(s)
}

func (c *Chlogger) Info(args ...interface{}) {
	var s string
	for _, arg := range args {
		s += fmt.Sprintf("%v", arg)
	}
	err := &MyError{Message: "", frame: xerrors.Caller(2)}
	SendCH(&ClickHouse{}, LevelInfo, s, fmt.Sprintf("%+v\n", err))
	c.Log.Info(s)
}

func (c *Chlogger) Fatal(args ...interface{}) {
	var s string
	for _, arg := range args {
		s += fmt.Sprintf("%v", arg)
	}
	err := &MyError{Message: "", frame: xerrors.Caller(2)}
	SendCH(&ClickHouse{}, LevelFatal, s, fmt.Sprintf("%+v\n", err))
	SendTg(&TelegramBot{}, LevelFatal, s, fmt.Sprintf("%+v\n", err))
	c.Log.Fatal(s)
}

func (c *Chlogger) Debug(args ...interface{}) {
	var s string
	for _, arg := range args {
		s += fmt.Sprintf("%v", arg)
	}
	err := &MyError{Message: "", frame: xerrors.Caller(2)}
	SendCH(&ClickHouse{}, LevelDebug, s, fmt.Sprintf("%+v\n", err))
	c.Log.Debug(s)
}

func (c *Chlogger) Panic(args ...interface{}) {
	var s string
	for _, arg := range args {
		s += fmt.Sprintf("%v", arg)
	}
	err := &MyError{Message: "", frame: xerrors.Caller(2)}
	SendCH(&ClickHouse{}, LevelPanic, s, fmt.Sprintf("%+v\n", err))
	c.Log.Panic(s)
}

func (c *Chlogger) Infof(pattern string, args ...interface{}) {
	c.Log.Info(fmt.Sprintf(pattern, args...))
}

func (c *Chlogger) Printf(pattern string, args ...interface{}) {
	c.Log.Info(fmt.Sprintf(pattern, args...))
}

func (c *Chlogger) Errorf(pattern string, args ...interface{}) {
	c.Log.Error(fmt.Sprintf(pattern, args...))
}

func (c *Chlogger) Fatalf(pattern string, args ...interface{}) {
	c.Log.Fatal(fmt.Sprintf(pattern, args...))
}

func Run(tgurl, service, reqUrl, env string) {
	chl.(*Chlogger).tgurl, chl.(*Chlogger).service, chl.(*Chlogger).reqUrl, chl.(*Chlogger).env = tgurl, service, reqUrl, env
}

type ClickHouse struct {
}

type TelegramBot struct {
}

func (c *ClickHouse) Send(s string) {
	req := fasthttp.AcquireRequest()
	req.SetBody([]byte(s))
	req.Header.SetMethodBytes([]byte("POST"))
	req.SetRequestURIBytes([]byte(chl.(*Chlogger).reqUrl))
	res := fasthttp.AcquireResponse()
	if err := fasthttp.Do(req, res); err != nil {
		logrus.Info("Request to kittenhouse response with error - " + err.Error())
	}
	fasthttp.ReleaseRequest(req)
	fasthttp.ReleaseResponse(res)
}

func (c *TelegramBot) Send(s string) {
	req := fasthttp.AcquireRequest()
	req.Header.SetMethodBytes([]byte("POST"))
	req.SetRequestURIBytes([]byte(s))
	res := fasthttp.AcquireResponse()
	if err := fasthttp.Do(req, res); err != nil {
		Error("Tg reroute doesn't work")
	}
	fasthttp.ReleaseRequest(req)
	fasthttp.ReleaseResponse(res)
}

type Sender interface {
	Send(s string)
}

// SendCH отправляет логи в kittenhouse который отправляет в clickhouse
func SendCH(sen Sender, level Level, s string, f string) {
	go func() {
		t := time.Now().Unix() + 9*60*60
		msg := fmt.Sprintf("('%s',%v,'%s','%s','%s',,'','','','%s')",
			f,     // function name
			t,     // timestamp
			level, // level (DEBUG, INFO, etc)
			s,     // Message
			chl.(*Chlogger).service,
			chl.(*Chlogger).env,
		)
		sen.Send(msg)
	}()
}

// SendTg отправляет логи в телегу
func SendTg(sen Sender, level Level, s string, f string) {
	go func() {
		str := getTgUrl(string(level), s, f)
		sen.Send(str)
	}()
}

func getTgUrl(level string, s string, f string) string {
	return chl.(*Chlogger).tgurl + url.QueryEscape(chl.(*Chlogger).service+": "+level+" "+s+" "+f)
}

func GetPtr() *Chlogger {
	return chl.(*Chlogger)
}
