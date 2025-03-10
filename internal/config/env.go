package config

import (
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/soa-rs/fit/internal/config/logger"
)

// Environment variables
const (
	EnvBackendPrefix    = "FIT_SOARS_BACKEND_"
	EnvBackendProfile   = EnvBackendPrefix + "PROFILE"
	EnvBackendHost      = EnvBackendPrefix + "HOST"
	EnvBackendPort      = EnvBackendPrefix + "PORT"
	EnvBackendLogLevel  = EnvBackendPrefix + "LOG_LEVEL"
	EnvBackendLogFormat = EnvBackendPrefix + "LOG_FORMAT"
	EnvBackendLogFile   = EnvBackendPrefix + "LOG_FILE"
	EnvBackendLogOutput = EnvBackendPrefix + "LOG_OUTPUT"
)

// Default values
const (
	// DefaultPort is the default port to which the server will bind.
	// It is a string for compatibility with environment variables.
	DefaultPort = "1369"
	// DefaultHost is the default host to which the server will bind.
	DefaultHost = "127.0.0.1"
	// DefaultProfile is the default profile in which the backend will
	// run.
	DefaultProfile = "production"
	// DefaultGinMode is the default mode for the HTTP server.
	// While `gin` itself prefers to default to `DebugMode`, we want to
	// default to `ReleaseMode` because the latter is unnecessarily
	// verbose.
	DefaultGinMode = gin.ReleaseMode
	// DefaultLogLevel is the default log level for the server.
	DefaultLogLevel = "info"
	// DefaultLogFormat is the default log format for the server.
	DefaultLogFormat = "pretty"
	// DefaultLogOutput is the default log output for the server.
	DefaultLogOutput = "console"
	// DefaultLogFile is the default log file for the server.
	DefaultLogFile = "backend.log"
)

// Defaults is a map of environment variables to their default values.
var (
	EnvBackendDefaults = map[string]string{
		EnvBackendProfile:   DefaultProfile,
		EnvBackendHost:      DefaultHost,
		EnvBackendPort:      DefaultPort,
		gin.EnvGinMode:      DefaultGinMode,
		EnvBackendLogLevel:  DefaultLogLevel,
		EnvBackendLogFormat: DefaultLogFormat,
		EnvBackendLogFile:   DefaultLogFile,
		EnvBackendLogOutput: DefaultLogOutput,
	}
)

// GetEnvOrDefault returns the environment variable with the given key.
// If the environment variable is not set, the default value is
// returned.
func GetEnvOrDefault(key string) (res string) {
	defer func() {
		logger.LogTrace(
			"GetEnvOrDefault(%s) = %s",
			key, res,
		)
	}()
	value, ok := os.LookupEnv(key)
	if !ok || value == "" {
		logger.LogInfo(
			"Environment variable %s not set, using default value %s",
			key, EnvBackendDefaults[key],
		)
		// If the environment variable is not set, return the default.
		// If the default is not set, the function will panic.
		// Since that indicates a misconfiguration in the code, the
		// panic is an acceptable response.
		return EnvBackendDefaults[key]
	}
	return value
}

// LoadEnvs loads the environment variables for the backend.
func LoadEnvs() {
	files := GetEnvFilesList()
	for _, file := range files {
		// not all files may exist, so we don't care if Load fails
		godotenv.Load(file)
	}
	// `gin` sets the mode in init(), which does not capture the .env
	// files, so we set it manually here.
	ginMode := GetEnvOrDefault(gin.EnvGinMode)
	gin.SetMode(ginMode)
}

// GetEnvFilesList returns the list of environment files to load,
// depending on the profile. The order of the files is such that
// the environment variables in the earlier files take precedence
// over those in the later files.
func GetEnvFilesList() []string {
	profile := GetEnvOrDefault(EnvBackendProfile)
	files := make([]string, 0, 4)
	if profile != "production" {
		files = append(files, fmt.Sprintf(".env.%s.local", profile))
		if profile != "test" {
			files = append(files, ".env.local")
		}
		files = append(files, fmt.Sprintf(".env.%s", profile))
	}
	files = append(files, ".env")
	return files
}
