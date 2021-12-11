package log

import (
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"e.coding.net/codingcorp/carctl/pkg/util/fileutil"
)

var (
	globalLogger        Logger
	globalSugaredLogger SugaredLogger
	globalLoggerLevel   zap.AtomicLevel
)

var (
	DebugLevel = zapcore.DebugLevel
	InfoLevel  = zapcore.InfoLevel
	WarnLevel  = zapcore.WarnLevel
	ErrorLevel = zapcore.ErrorLevel
	PanicLevel = zapcore.PanicLevel
	FatalLevel = zapcore.FatalLevel
)

var levelMapping = map[string]Level{
	DebugLevel.String(): DebugLevel,
	InfoLevel.String():  InfoLevel,
	WarnLevel.String():  WarnLevel,
	ErrorLevel.String(): ErrorLevel,
	PanicLevel.String(): PanicLevel,
	FatalLevel.String(): FatalLevel,
}

type (
	// Field is an alias of zap.Field. Aliasing this type dramatically
	// improves the navigability of this package's API documentation.
	Field = zap.Field

	Level = zapcore.Level
)

type SugaredLogger interface {
	Named(name string) SugaredLogger
	With(args ...interface{}) SugaredLogger

	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	Panic(args ...interface{})
	Fatal(args ...interface{})

	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Panicf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})

	Sync()
}

// Logger defines methods of writing log
type Logger interface {
	Named(s string) Logger
	With(fields ...Field) Logger

	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)

	Clone() Logger
	Level() string
	IsDebug() bool
	Sync()

	SugaredLogger() SugaredLogger
	CoreLogger() *zap.Logger
}

type logger struct {
	level         string
	logger        *zap.Logger
	sugaredLogger *zap.SugaredLogger
}

type sugaredLogger struct {
	sugaredLogger *zap.SugaredLogger
}

type Config struct {
	Level string

	// Encoding sets the logger's encoding. Valid values are "json" and
	// "console", as well as any third-party encodings registered via
	// RegisterEncoder.
	Encoding string

	// DisableCaller configures the Logger to annotate each message with the filename
	// and line number of zap's caller, or not
	DisableCaller bool

	// OutputPaths is a list of URLs or file paths to write logging output to.
	// See Open for details.
	OutputPaths []string

	// ErrorOutputPaths is a list of URLs to write internal logger errors to.
	// The default is standard error.
	//
	// Note that this setting only affects internal errors; for sample code that
	// sends error-level logs to a different location from info- and debug-level
	// logs, see the package-level AdvancedConfiguration example.
	ErrorOutputPaths []string
}

func New(cfgs ...*Config) (Logger, zap.AtomicLevel) {
	var cfg *Config
	if len(cfgs) > 0 {
		cfg = cfgs[0]
	}

	l := &logger{
		level: getLevel(cfg),
	}

	// get encoding
	encoding := "console" // set default log format to 'console'. it could be replaced by 'json'
	if cfg != nil && cfg.Encoding != "" {
		encoding = cfg.Encoding
	}

	atomicLevel := zap.NewAtomicLevelAt(parseLevel(l.level))

	// get output paths
	outputPaths, errorOutputPaths := getOutputPaths(cfg)

	// get config
	config := getConfig(atomicLevel, encoding, outputPaths, errorOutputPaths)

	var err error
	if cfg != nil && cfg.DisableCaller {
		l.logger, err = config.Build(zap.WithCaller(false))
	} else {
		l.logger, err = config.Build(zap.AddCallerSkip(1))
	}

	if err != nil {
		panic(err)
	}

	l.sugaredLogger = l.logger.Sugar()

	return l, atomicLevel
}

func getConfig(atomicLevel zap.AtomicLevel, encoding string, outputPaths, errorOutputPaths []string) zap.Config {
	encoder := zapcore.EncoderConfig{
		TimeKey:       "time",
		LevelKey:      "level",
		NameKey:       "name",
		CallerKey:     "caller",
		MessageKey:    "msg",
		StacktraceKey: "stacktrace",
		LineEnding:    zapcore.DefaultLineEnding,
		EncodeLevel:   zapcore.LowercaseLevelEncoder,
		EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			encodeTimeLayout(t, "2006-01-02 15:04:05.000", enc)
		},
		EncodeDuration: zapcore.StringDurationEncoder,
		// EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	if encoding == "console" {
		encoder.EncodeLevel = zapcore.CapitalLevelEncoder
	}

	return zap.Config{
		Level:       atomicLevel,
		Development: false,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		Encoding:         encoding,
		EncoderConfig:    encoder,
		OutputPaths:      outputPaths,
		ErrorOutputPaths: errorOutputPaths,
	}
}

func getLevel(cfg *Config) string {
	level := InfoLevel.String()
	if cfg != nil && cfg.Level != "" && cfg.Level != level {
		level = cfg.Level
	}
	if lvl := os.Getenv("LOG_LEVEL"); lvl != "" && lvl != level {
		level = lvl
	}
	return level
}

func parseLevel(level string) Level {
	lvl, ok := levelMapping[level]
	if ok {
		return lvl
	}
	// default level: info
	return InfoLevel
}

func getOutputPaths(cfg *Config) (outputPaths, errorOutputPaths []string) {
	if cfg == nil {
		return []string{"stdout"}, []string{"stderr"}
	}

	outputPaths = cfg.OutputPaths
	errorOutputPaths = cfg.ErrorOutputPaths

	if len(cfg.OutputPaths) == 0 {
		outputPaths = []string{"stdout"}
	} else if len(cfg.OutputPaths) > 1 {
		for _, p := range outputPaths {
			if p == "stdout" || p == "stderr" {
				continue
			}
			// try to create the file if not exists
			_ = fileutil.CreateFileIfNotExists(p)
		}
	}

	if len(cfg.ErrorOutputPaths) == 0 {
		errorOutputPaths = []string{"stderr"}
	} else if len(cfg.ErrorOutputPaths) > 1 {
		for _, p := range errorOutputPaths {
			if p == "stdout" || p == "stderr" {
				continue
			}
			// try to create the file if not exists
			_ = fileutil.CreateFileIfNotExists(p)
		}
	}

	return
}

func encodeTimeLayout(t time.Time, layout string, enc zapcore.PrimitiveArrayEncoder) {
	type appendTimeEncoder interface {
		AppendTimeLayout(time.Time, string)
	}

	if enc, ok := enc.(appendTimeEncoder); ok {
		enc.AppendTimeLayout(t, layout)
		return
	}

	enc.AppendString(t.Format(layout))
}

func init() {
	// stdlog.Printf("Initiate logger ...")
	globalLogger, globalLoggerLevel = New()
	globalSugaredLogger = globalLogger.SugaredLogger()
	zap.ReplaceGlobals(globalLogger.CoreLogger())
}
