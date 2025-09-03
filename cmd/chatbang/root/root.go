package root

import (
    "fmt"
    "io"
    "os"
    "path/filepath"
    "strings"

    "github.com/joho/godotenv"
    "github.com/sirupsen/logrus"
    "github.com/spf13/cobra"

    "gg/pkg/app"
)

var (
    flagConfigLogin bool
)

// rootCmd defines the base command for chatbang
var rootCmd = &cobra.Command{
    Use:   "chatbang [prompt]",
    Short: "Chat with ChatGPT from your terminal",
    Long:  "Chatbang opens a Chromium session to ChatGPT and lets you chat from the terminal. Configure browser path in ~/.config/chatbang/chatbang.",
    Args:  cobra.ArbitraryArgs,
    RunE: func(cmd *cobra.Command, args []string) error {
        // Optional: if a prompt is provided as args, join into a single prompt
        var prompt string
        if len(args) > 0 {
            prompt = "" + args[0]
            if len(args) > 1 {
                for _, s := range args[1:] {
                    prompt += " " + s
                }
            }
        }
        a := app.New()
        if flagConfigLogin {
            return a.Login()
        }
        return a.Run(prompt)
    },
}

// Execute runs the Cobra root command.
func Execute() {
    // Load environment from .env if present and configure logger
    _ = godotenv.Load()
    level := strings.ToLower(strings.TrimSpace(os.Getenv("LOG_LEVEL")))
    if level == "" && (os.Getenv("DEBUG") == "1" || strings.EqualFold(os.Getenv("DEBUG"), "true")) {
        level = "debug"
    }
    switch level {
    case "trace": logrus.SetLevel(logrus.TraceLevel)
    case "debug": logrus.SetLevel(logrus.DebugLevel)
    case "warn": logrus.SetLevel(logrus.WarnLevel)
    case "error": logrus.SetLevel(logrus.ErrorLevel)
    default: logrus.SetLevel(logrus.InfoLevel)
    }
    logrus.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})

    // Optional file logging via LOG_FILE. If set, duplicate output to file.
    var logFile *os.File
    if lf := strings.TrimSpace(os.Getenv("LOG_FILE")); lf != "" {
        // Expand ~/ paths
        if strings.HasPrefix(lf, "~") {
            if home, err := os.UserHomeDir(); err == nil {
                lf = filepath.Join(home, strings.TrimPrefix(lf, "~"))
            }
        }
        if err := os.MkdirAll(filepath.Dir(lf), 0o755); err == nil {
            f, err := os.OpenFile(lf, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
            if err == nil {
                logFile = f
                logrus.SetOutput(io.MultiWriter(os.Stderr, f))
                logrus.WithField("file", lf).Info("logging to file enabled")
            } else {
                logrus.WithError(err).Warn("failed to open LOG_FILE; using stderr only")
            }
        } else {
            logrus.WithError(err).Warn("failed to create directory for LOG_FILE; using stderr only")
        }
    }

    if err := rootCmd.Execute(); err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }

    if logFile != nil { _ = logFile.Close() }
}

func init() {
    rootCmd.Flags().BoolVar(&flagConfigLogin, "config", false, "Open ChatGPT and grant clipboard permission (login/profile setup)")
}
