// An attempt to make the Config a bit more object-oriented.
// Eventually this might become a package of its own, we'll see.
// For now, the main issue is dealing with the logging...
package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

// Shamelessly copied from zerolog,
// The idea is to make a switch *easy*

// Level defines log levels.
type Level int8

const (
	// DebugLevel defines debug log level.
	DebugLevel Level = iota
	// InfoLevel defines info log level.
	InfoLevel
	// WarnLevel defines warn log level.
	WarnLevel
	// ErrorLevel defines error log level.
	ErrorLevel
	// FatalLevel defines fatal log level.
	FatalLevel
	// PanicLevel defines panic log level.
	PanicLevel
	// NoLevel defines an absent log level.
	NoLevel
	// Disabled disables the logger.
	Disabled

	// TraceLevel defines trace log level.
	TraceLevel Level = -1
	// Values less than TraceLevel are handled as numbers.
)

func (l Level) String() string {
	switch l {
	case TraceLevel:
		return LevelTraceValue
	case DebugLevel:
		return LevelDebugValue
	case InfoLevel:
		return LevelInfoValue
	case WarnLevel:
		return LevelWarnValue
	case ErrorLevel:
		return LevelErrorValue
	case FatalLevel:
		return LevelFatalValue
	case PanicLevel:
		return LevelPanicValue
	case Disabled:
		return "disabled"
	case NoLevel:
		return ""
	}
	return strconv.Itoa(int(l))
}

// ParseLevel converts a level string into a zerolog Level value.
// returns an error if the input string does not match known values.
func ParseLevel(levelStr string) (Level, error) {
	switch {
	case strings.EqualFold(levelStr, LevelFieldMarshalFunc(TraceLevel)):
		return TraceLevel, nil
	case strings.EqualFold(levelStr, LevelFieldMarshalFunc(DebugLevel)):
		return DebugLevel, nil
	case strings.EqualFold(levelStr, LevelFieldMarshalFunc(InfoLevel)):
		return InfoLevel, nil
	case strings.EqualFold(levelStr, LevelFieldMarshalFunc(WarnLevel)):
		return WarnLevel, nil
	case strings.EqualFold(levelStr, LevelFieldMarshalFunc(ErrorLevel)):
		return ErrorLevel, nil
	case strings.EqualFold(levelStr, LevelFieldMarshalFunc(FatalLevel)):
		return FatalLevel, nil
	case strings.EqualFold(levelStr, LevelFieldMarshalFunc(PanicLevel)):
		return PanicLevel, nil
	case strings.EqualFold(levelStr, LevelFieldMarshalFunc(Disabled)):
		return Disabled, nil
	case strings.EqualFold(levelStr, LevelFieldMarshalFunc(NoLevel)):
		return NoLevel, nil
	}
	i, err := strconv.Atoi(levelStr)
	if err != nil {
		return NoLevel, fmt.Errorf("Unknown Level String: '%s', defaulting to NoLevel", levelStr)
	}
	if i > 127 || i < -128 {
		return NoLevel, fmt.Errorf("Out-Of-Bounds Level: '%d', defaulting to NoLevel", i)
	}
	return Level(i), nil
}

// UnmarshalText implements encoding.TextUnmarshaler to allow for easy reading from toml/yaml/json formats
func (l *Level) UnmarshalText(text []byte) error {
	if l == nil {
		return errors.New("can't unmarshal a nil *Level")
	}
	var err error
	*l, err = ParseLevel(string(text))
	return err
}

// MarshalText implements encoding.TextMarshaler to allow for easy writing into toml/yaml/json formats
func (l Level) MarshalText() ([]byte, error) {
	return []byte(LevelFieldMarshalFunc(l)), nil
}

// Having fun with colours!
const (
	colorBlack = iota + 30
	colorRed
	colorGreen
	colorYellow
	colorBlue
	colorMagenta
	colorCyan
	colorWhite

	colorBold     = 1
	colorDarkGray = 90

	unknownLevel = "???"
)

// Now, more variables:
var (
	// LevelFieldName is the field name used for the level field.
	LevelFieldName = "level"

	// LevelTraceValue is the value used for the trace level field.
	LevelTraceValue = "trace"
	// LevelDebugValue is the value used for the debug level field.
	LevelDebugValue = "debug"
	// LevelInfoValue is the value used for the info level field.
	LevelInfoValue = "info"
	// LevelWarnValue is the value used for the warn level field.
	LevelWarnValue = "warn"
	// LevelErrorValue is the value used for the error level field.
	LevelErrorValue = "error"
	// LevelFatalValue is the value used for the fatal level field.
	LevelFatalValue = "fatal"
	// LevelPanicValue is the value used for the panic level field.
	LevelPanicValue = "panic"

	// LevelFieldMarshalFunc allows customization of global level field marshaling.
	LevelFieldMarshalFunc = func(l Level) string {
		return l.String()
	}

	// LevelColors are used by ConsoleWriter's consoleDefaultFormatLevel to color
	// log levels.
	LevelColors = map[Level]int{
		TraceLevel: colorBlue,
		DebugLevel: 0,
		InfoLevel:  colorGreen,
		WarnLevel:  colorYellow,
		ErrorLevel: colorRed,
		FatalLevel: colorRed,
		PanicLevel: colorRed,
	}

	// FormattedLevels are used by ConsoleWriter's consoleDefaultFormatLevel
	// for a short level name.
	FormattedLevels = map[Level]string{
		TraceLevel: "TRC",
		DebugLevel: "DBG",
		InfoLevel:  "INF",
		WarnLevel:  "WRN",
		ErrorLevel: "ERR",
		FatalLevel: "FTL",
		PanicLevel: "PNC",
	}
)

// colorize returns the string s wrapped in ANSI code c, unless disabled is true or c is 0.
func colorize(s interface{}, c int, disabled bool) string {
	e := os.Getenv("NO_COLOR")
	if e != "" || c == 0 {
		disabled = true
	}

	if disabled {
		return fmt.Sprintf("%s", s)
	}
	return fmt.Sprintf("\x1b[%dm%v\x1b[0m", c, s)
}

// My code starts here!

// Fast lookup map for colours etc.
var levelLookupMap = map[string]Level{
	LevelTraceValue: TraceLevel,
	LevelDebugValue: DebugLevel,
	LevelInfoValue:  InfoLevel,
	LevelWarnValue:  WarnLevel,
	LevelErrorValue: ErrorLevel,
	LevelFatalValue: FatalLevel,
	LevelPanicValue: PanicLevel,
}

// Config now becomes a type, because we wish to use a simple logging system based on it.
// Probably this will be placed in a separate package for good measure!
type Config map[string]*string

// Singleton for config:
var config Config

// Our first sequence of methods will be just to deal with pretty-printing errors to the log!

func (config Config) LogTrace(thingies ...any) {
	if levelLookupMap[*config["ginMode"]] <= TraceLevel {
		log.Println("["+colorize(FormattedLevels[TraceLevel], LevelColors[TraceLevel], false)+"]", thingies)
	}
}

func (config Config) LogDebug(thingies ...any) {
	if levelLookupMap[*config["ginMode"]] <= DebugLevel {
		log.Println("["+colorize(FormattedLevels[DebugLevel], LevelColors[DebugLevel], false)+"]", thingies)
	}
}

func (config Config) LogInfo(thingies ...any) {
	if levelLookupMap[*config["ginMode"]] <= InfoLevel {
		log.Println("["+colorize(FormattedLevels[InfoLevel], LevelColors[InfoLevel], false)+"]", thingies)
	}
}

func (config Config) LogWarn(thingies ...any) {
	if levelLookupMap[*config["ginMode"]] <= WarnLevel {
		log.Println("["+colorize(FormattedLevels[WarnLevel], LevelColors[WarnLevel], false)+"]", thingies)
	}
}

func (config Config) LogError(thingies ...any) {
	if levelLookupMap[*config["ginMode"]] <= ErrorLevel {
		log.Println("["+colorize(FormattedLevels[ErrorLevel], LevelColors[ErrorLevel], false)+"]", thingies)
	}
}

func (config Config) LogFatal(thingies ...any) {
	if levelLookupMap[*config["ginMode"]] <= FatalLevel {
		log.Fatal("["+colorize(FormattedLevels[FatalLevel], LevelColors[FatalLevel], false)+"]", thingies)
	}
}

func (config Config) LogPanic(thingies ...any) {
	if levelLookupMap[*config["ginMode"]] <= PanicLevel {
		log.Panic("["+colorize(FormattedLevels[PanicLevel], LevelColors[PanicLevel], false)+"]", thingies)
	}
}

// Now with variable parameters

func (config Config) LogTracef(format string, thingies ...any) {
	if levelLookupMap[*config["ginMode"]] <= TraceLevel {
		log.Printf("["+colorize(FormattedLevels[TraceLevel], LevelColors[TraceLevel], false)+"] "+format, thingies)
	}
}

func (config Config) LogDebugf(format string, thingies ...any) {
	if levelLookupMap[*config["ginMode"]] <= DebugLevel {
		log.Println("["+colorize(FormattedLevels[DebugLevel], LevelColors[DebugLevel], false)+"] "+format, thingies)
	}
}

func (config Config) LogInfof(format string, thingies ...any) {
	if levelLookupMap[*config["ginMode"]] <= InfoLevel {
		log.Println("["+colorize(FormattedLevels[InfoLevel], LevelColors[InfoLevel], false)+"] "+format, thingies)
	}
}

func (config Config) LogWarnf(format string, thingies ...any) {
	if levelLookupMap[*config["ginMode"]] <= WarnLevel {
		log.Println("["+colorize(FormattedLevels[WarnLevel], LevelColors[WarnLevel], false)+"] "+format, thingies)
	}
}

func (config Config) LogErrorf(format string, thingies ...any) {
	if levelLookupMap[*config["ginMode"]] <= ErrorLevel {
		log.Println("["+colorize(FormattedLevels[ErrorLevel], LevelColors[ErrorLevel], false)+"] "+format, thingies)
	}
}

func (config Config) LogFatalf(format string, thingies ...any) {
	if levelLookupMap[*config["ginMode"]] <= FatalLevel {
		log.Fatalf("["+colorize(FormattedLevels[FatalLevel], LevelColors[FatalLevel], false)+"] "+format, thingies)
	}
}

func (config Config) LogPanicf(format string, thingies ...any) {
	if levelLookupMap[*config["ginMode"]] <= PanicLevel {
		log.Panicf("["+colorize(FormattedLevels[PanicLevel], LevelColors[PanicLevel], false)+"] "+format, thingies)
	}
}
