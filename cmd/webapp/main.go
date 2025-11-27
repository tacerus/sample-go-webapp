/*
   Copyright (C) 2025  SUSE LLC <georg.pfuetzenreuter@suse.com>

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU General Public License for more details.

   You should have received a copy of the GNU General Public License
   along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"

	"github.com/tacerus/sample-go-webapp/core"
	"github.com/tacerus/sample-go-webapp/web"
)

func main() {
	var (
		configArg   string
		logLevelArg string
	)

	flag.StringVar(&configArg, "config", "./config.json", "Configuration file")
	flag.StringVar(&logLevelArg, "loglevel", "info", "Logging level")

	flag.Parse()

	slog.SetDefault(newSlog(newLogLevel(logLevelArg)))

	slog.Info("Booting web app ...")

	app := web.NewWebApp(core.NewConfig(configArg))
	if app == nil {
		os.Exit(1)
	}

	cs := make(chan os.Signal, 1)
	signal.Notify(cs, os.Interrupt)

	srv := app.Start()
	defer srv.Shutdown(context.Background())

main:
	for {
		select {
		case <-cs:
			slog.Debug("Received interrupt")
			break main
		}
	}

	slog.Info("Shutting down ...")
}
