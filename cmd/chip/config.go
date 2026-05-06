package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/collibra/chip/pkg/chip"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func Init() *Config {
	viper.SetConfigName("mcp")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME/.config/collibra")
	viper.AddConfigPath("/etc/collibra")
	viper.SetEnvPrefix("COLLIBRA_MCP")
	viper.AutomaticEnv()

	initConfigOptions()

	pflag.Usage = func() {
		printUsage(chip.Version)
	}

	showHelp := pflag.BoolP("help", "h", false, "Show help message")
	showVersion := pflag.BoolP("version", "v", false, "Show version information")
	pflag.Parse()

	if *showHelp {
		pflag.Usage()
		os.Exit(0)
	}

	if *showVersion {
		fmt.Println(chip.Version)
		os.Exit(0)
	}

	config := readConfigFile()
	validateConfigFile(config)
	return &config
}

func initConfigOptions() {
	pflag.String("api-url", "", "Collibra API URL (env: COLLIBRA_MCP_API_URL)")
	_ = viper.BindEnv("api.url", "COLLIBRA_MCP_API_URL")
	_ = viper.BindPFlag("api.url", pflag.Lookup("api-url"))

	pflag.String("api-username", "", "Collibra API username (env: COLLIBRA_MCP_API_USR)")
	_ = viper.BindEnv("api.username", "COLLIBRA_MCP_API_USR")
	_ = viper.BindPFlag("api.username", pflag.Lookup("api-username"))

	pflag.String("api-password", "", "Collibra API password (env: COLLIBRA_MCP_API_PWD)")
	_ = viper.BindEnv("api.password", "COLLIBRA_MCP_API_PWD")
	_ = viper.BindPFlag("api.password", pflag.Lookup("api-password"))

	pflag.Bool("skip-tls-verify", false, "Skip TLS certificate verification (env: COLLIBRA_MCP_API_SKIP_TLS_VERIFY)")
	_ = viper.BindEnv("api.skip-tls-verify", "COLLIBRA_MCP_API_SKIP_TLS_VERIFY")
	_ = viper.BindPFlag("api.skip-tls-verify", pflag.Lookup("skip-tls-verify"))
	viper.SetDefault("api.skip-tls-verify", false)

	pflag.String("api-proxy", "", "HTTP proxy URL for API requests (env: COLLIBRA_MCP_API_PROXY, HTTP_PROXY, HTTPS_PROXY)")
	_ = viper.BindEnv("api.proxy", "COLLIBRA_MCP_API_PROXY")
	_ = viper.BindEnv("api.proxy", "HTTP_PROXY")  // For compatibility with DefaultTransport
	_ = viper.BindEnv("api.proxy", "HTTPS_PROXY") // For compatibility with DefaultTransport
	_ = viper.BindPFlag("api.proxy", pflag.Lookup("api-proxy"))

	pflag.String("mode", "stdio", "MCP server mode: 'stdio', 'http', 'http-sse', or 'http-streamable' (env: COLLIBRA_MCP_MODE)")
	_ = viper.BindEnv("mcp.mode", "COLLIBRA_MCP_MODE")
	_ = viper.BindPFlag("mcp.mode", pflag.Lookup("mode"))
	viper.SetDefault("mcp.mode", "stdio")

	pflag.Int("port", 8080, "HTTP server port (only used in http mode) (env: COLLIBRA_MCP_HTTP_PORT)")
	_ = viper.BindEnv("mcp.http.port", "COLLIBRA_MCP_HTTP_PORT")
	_ = viper.BindPFlag("mcp.http.port", pflag.Lookup("port"))
	viper.SetDefault("mcp.http.port", 8080)

	pflag.String("host", "localhost", "HTTP server bind address (env: COLLIBRA_MCP_HTTP_HOST)")
	_ = viper.BindEnv("mcp.http.host", "COLLIBRA_MCP_HTTP_HOST")
	_ = viper.BindPFlag("mcp.http.host", pflag.Lookup("host"))
	viper.SetDefault("mcp.http.host", "localhost")

	pflag.String("tls-cert", "", "Path to TLS certificate file for HTTPS (env: COLLIBRA_MCP_HTTP_TLS_CERT)")
	_ = viper.BindEnv("mcp.http.tls-cert", "COLLIBRA_MCP_HTTP_TLS_CERT")
	_ = viper.BindPFlag("mcp.http.tls-cert", pflag.Lookup("tls-cert"))

	pflag.String("tls-key", "", "Path to TLS private key file for HTTPS (env: COLLIBRA_MCP_HTTP_TLS_KEY)")
	_ = viper.BindEnv("mcp.http.tls-key", "COLLIBRA_MCP_HTTP_TLS_KEY")
	_ = viper.BindPFlag("mcp.http.tls-key", pflag.Lookup("tls-key"))

	pflag.Bool("auth-enabled", false, "Enable OAuth 2.1 bearer token authentication (env: COLLIBRA_MCP_AUTH_ENABLED)")
	_ = viper.BindEnv("mcp.auth.enabled", "COLLIBRA_MCP_AUTH_ENABLED")
	_ = viper.BindPFlag("mcp.auth.enabled", pflag.Lookup("auth-enabled"))
	viper.SetDefault("mcp.auth.enabled", false)

	pflag.String("auth-environment-url", "", "Scalekit environment URL used as JWT issuer (env: COLLIBRA_MCP_AUTH_ENVIRONMENT_URL)")
	_ = viper.BindEnv("mcp.auth.environment-url", "COLLIBRA_MCP_AUTH_ENVIRONMENT_URL")
	_ = viper.BindPFlag("mcp.auth.environment-url", pflag.Lookup("auth-environment-url"))

	pflag.String("auth-resource-url", "", "This server's public URL used as JWT audience (env: COLLIBRA_MCP_AUTH_RESOURCE_URL)")
	_ = viper.BindEnv("mcp.auth.resource-url", "COLLIBRA_MCP_AUTH_RESOURCE_URL")
	_ = viper.BindPFlag("mcp.auth.resource-url", pflag.Lookup("auth-resource-url"))

	pflag.String("auth-client-id", "", "Scalekit client ID (env: COLLIBRA_MCP_AUTH_CLIENT_ID)")
	_ = viper.BindEnv("mcp.auth.client-id", "COLLIBRA_MCP_AUTH_CLIENT_ID")
	_ = viper.BindPFlag("mcp.auth.client-id", pflag.Lookup("auth-client-id"))

	pflag.String("auth-client-secret", "", "Scalekit client secret (env: COLLIBRA_MCP_AUTH_CLIENT_SECRET)")
	_ = viper.BindEnv("mcp.auth.client-secret", "COLLIBRA_MCP_AUTH_CLIENT_SECRET")
	_ = viper.BindPFlag("mcp.auth.client-secret", pflag.Lookup("auth-client-secret"))

	pflag.StringSlice("auth-authorization-servers", []string{}, "Authorization server URLs for discovery metadata, from Scalekit dashboard (env: COLLIBRA_MCP_AUTH_AUTHORIZATION_SERVERS)")
	_ = viper.BindEnv("mcp.auth.authorization-servers", "COLLIBRA_MCP_AUTH_AUTHORIZATION_SERVERS")
	_ = viper.BindPFlag("mcp.auth.authorization-servers", pflag.Lookup("auth-authorization-servers"))

	pflag.StringSlice("enabled-tools", []string{}, "Optional comma-separated list of tool names to enable instead of enabling all tools (cannot be used with disabled-tools) (env: COLLIBRA_MCP_ENABLED_TOOLS)")
	_ = viper.BindEnv("mcp.enabled-tools", "COLLIBRA_MCP_ENABLED_TOOLS")
	_ = viper.BindPFlag("mcp.enabled-tools", pflag.Lookup("enabled-tools"))

	pflag.StringSlice("disabled-tools", []string{}, "Optional comma-separated list of tool names to disable while enabling the remaining tools (cannot be used with enabled-tools) (env: COLLIBRA_MCP_DISABLED_TOOLS)")
	_ = viper.BindEnv("mcp.disabled-tools", "COLLIBRA_MCP_DISABLED_TOOLS")
	_ = viper.BindPFlag("mcp.disabled-tools", pflag.Lookup("disabled-tools"))
}

func printUsage(version string) {
	fmt.Fprintf(os.Stderr, `Collibra MCP Server %s

A Model Context Protocol (MCP) server that provides tools for interacting with Collibra.

USAGE:
  %s [flags]

FLAGS:
`, version, os.Args[0])
	pflag.PrintDefaults()
	fmt.Fprintf(os.Stderr, `
ENVIRONMENT VARIABLES:
  COLLIBRA_MCP_API_URL          Collibra API URL
  COLLIBRA_MCP_API_USR          Collibra API username
  COLLIBRA_MCP_API_PWD          Collibra API password
  COLLIBRA_MCP_API_SKIP_TLS_VERIFY  Skip TLS certificate verification (default: false)
  COLLIBRA_MCP_API_PROXY        HTTP proxy URL for API requests
  HTTP_PROXY                    HTTP proxy URL (alternative to COLLIBRA_MCP_API_PROXY)
  HTTPS_PROXY                   HTTPS proxy URL (alternative to COLLIBRA_MCP_API_PROXY)
  COLLIBRA_MCP_MODE             Server mode: 'stdio', 'http', 'http-sse', or 'http-streamable' (default: stdio)
  COLLIBRA_MCP_HTTP_HOST        HTTP server bind address (default: localhost)
  COLLIBRA_MCP_HTTP_PORT        HTTP server port (default: 8080)
  COLLIBRA_MCP_HTTP_TLS_CERT           Path to TLS certificate file (enables HTTPS when set with TLS_KEY)
  COLLIBRA_MCP_HTTP_TLS_KEY            Path to TLS private key file (enables HTTPS when set with TLS_CERT)
  COLLIBRA_MCP_AUTH_ENABLED            Enable OAuth 2.1 bearer token authentication (default: false)
  COLLIBRA_MCP_AUTH_ENVIRONMENT_URL    Scalekit environment URL (JWT issuer)
  COLLIBRA_MCP_AUTH_CLIENT_ID          Scalekit client ID
  COLLIBRA_MCP_AUTH_CLIENT_SECRET      Scalekit client secret
  COLLIBRA_MCP_AUTH_RESOURCE_URL       This server's public URL (JWT audience)
  COLLIBRA_MCP_AUTH_AUTHORIZATION_SERVERS  Authorization server URLs from Scalekit dashboard (comma-separated)
  COLLIBRA_MCP_ENABLED_TOOLS           Optional comma-separated list of tool names to enable instead of enabling all tools, cannot be used with disabled-tools
  COLLIBRA_MCP_DISABLED_TOOLS   Optional comma-separated list of tool names to disable while enabling the remaining tools, cannot be used with enabled-tools

CONFIGURATION:
  Configuration can be provided in the following order of precedence: command-line flags (highest), environment variables, or a YAML configuration file (lowest).
  File locations searched in order:
  - ./mcp.yaml
  - $HOME/.config/collibra/mcp.yaml
  - /etc/collibra/mcp.yaml

CONFIGURATION FILE EXAMPLE:
  api:
    url: "https://your-collibra-instance.com"
    username: "your-username"
    password: "your-password"
    skip-tls-verify: false
    proxy: "http://proxy.example.com:8080"
  mcp:
    mode: "http"  # or "stdio", "http-sse", "http-streamable"
    http:
      host: "localhost"  # bind address; use 0.0.0.0 to expose on all interfaces
      port: 8080
      tls-cert: "/path/to/cert.pem"  # optional: enables HTTPS when set with tls-key
      tls-key:  "/path/to/key.pem"   # optional: enables HTTPS when set with tls-cert
  auth:
    enabled: false
    environment-url: "https://your-env.scalekit.com"
    client-id: "your-client-id"
    client-secret: "your-client-secret"
    resource-url: "https://mcp.your-domain.com"
    authorization-servers:
      - "https://your-env.scalekit.com/resources/res_xxx"
    enabled-tools:  # Optional: list of tools to enable (cannot be used with disabled-tools)
      - "tool1"
      - "tool2"
    # disabled-tools:  # Optional: list of tools to disable (cannot be used with enabled-tools)
    #   - "tool3"
    #   - "tool4"
`)
}

func validateConfigFile(config Config) {
	if config.Mcp.Mode != "stdio" && config.Mcp.Mode != "http" && config.Mcp.Mode != "http-sse" && config.Mcp.Mode != "http-streamable" {
		slog.Error(fmt.Sprintf("Invalid server mode: %s (must be 'stdio', 'http', 'http-sse' or 'http-streamable')", config.Mcp.Mode))
		os.Exit(1)
	}

	if len(config.Mcp.EnabledTools) > 0 && len(config.Mcp.DisabledTools) > 0 {
		slog.Error("Cannot specify both enabled-tools and disabled-tools, only one can be specified")
		os.Exit(1)
	}

	certSet := config.Mcp.Http.TLSCertFile != ""
	keySet := config.Mcp.Http.TLSKeyFile != ""
	if certSet != keySet {
		slog.Error("Both --tls-cert and --tls-key must be provided together to enable HTTPS")
		os.Exit(1)
	}

	if config.Mcp.Auth.Enabled {
		if config.Mcp.Auth.EnvironmentURL == "" || config.Mcp.Auth.ClientID == "" || config.Mcp.Auth.ClientSecret == "" {
			slog.Error("--auth-environment-url, --auth-client-id, and --auth-client-secret are all required when auth is enabled")
			os.Exit(1)
		}
		if config.Mcp.Auth.ResourceURL == "" {
			slog.Error("--auth-resource-url is required when auth is enabled")
			os.Exit(1)
		}
		if config.Mcp.Mode == "stdio" {
			slog.Error("OAuth authentication requires HTTP transport, not stdio")
			os.Exit(1)
		}
	}
}

func readConfigFile() Config {
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			slog.Info("No config file found, using environment variables, command-line flags, and defaults")
		} else {
			slog.Error(fmt.Sprintf("Error reading config file: %v", err))
			os.Exit(1)
		}
	} else {
		slog.Info(fmt.Sprintf("Using config file: %s", viper.ConfigFileUsed()))
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		slog.Error(fmt.Sprintf("Unable to decode config: %v", err))
		os.Exit(1)
	}
	return config
}

type Config struct {
	Api CollibraApiConfig `mapstructure:"api"`
	Mcp McpConfig         `mapstructure:"mcp"`
}

// CollibraConfig holds Collibra-specific configuration
type CollibraApiConfig struct {
	Url           string `mapstructure:"url"`
	Username      string `mapstructure:"username"`
	Password      string `mapstructure:"password"`
	SkipTLSVerify bool   `mapstructure:"skip-tls-verify"`
	Proxy         string `mapstructure:"proxy"`
}

// ServerConfig holds server configuration
type McpConfig struct {
	Mode          string      `mapstructure:"mode"` // "stdio", "http", "http-sse", or "http-streamable"
	Http          HttpConfig  `mapstructure:"http"`
	Auth          AuthConfig  `mapstructure:"auth"`
	Stdio         StdioConfig `mapstructure:"stdio"`
	EnabledTools  []string    `mapstructure:"enabled-tools"`
	DisabledTools []string    `mapstructure:"disabled-tools"`
}

type AuthConfig struct {
	Enabled              bool     `mapstructure:"enabled"`
	EnvironmentURL       string   `mapstructure:"environment-url"`
	ClientID             string   `mapstructure:"client-id"`
	ClientSecret         string   `mapstructure:"client-secret"`
	ResourceURL          string   `mapstructure:"resource-url"`
	AuthorizationServers []string `mapstructure:"authorization-servers"`
}

type HttpConfig struct {
	Port        int    `mapstructure:"port"`
	Host        string `mapstructure:"host"`
	TLSCertFile string `mapstructure:"tls-cert"`
	TLSKeyFile  string `mapstructure:"tls-key"`
}

type StdioConfig struct {
}
