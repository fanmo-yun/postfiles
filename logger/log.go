package logger

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func InitLogger() {
	log.Logger = zerolog.New(zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.DateTime,
		FormatLevel: func(i interface{}) string {
			return fmt.Sprintf("[%s]", strings.ToUpper(i.(string)))
		},
	}).
		Level(zerolog.InfoLevel).
		With().
		Timestamp().
		Logger()
}
