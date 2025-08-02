package service

import "github.com/rs/zerolog"

var log = zerolog.New(zerolog.NewConsoleWriter()).With().Timestamp().Logger()
