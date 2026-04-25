package cliArgs

import (
	"flag"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

func ParseCliArgs() (string, string, string, string, string, string, string, bool, time.Duration, []string, bool, int) {
	spec := flag.String("spec", "", "OpenAPI file specification (JSON or YAML format, Required)")
	targetURL := flag.String("url", "", "target URL")
	username := flag.String("user", "", "username (Basic auth)")
	password := flag.String("password", "", "password (Basic auth)")
	apikey := flag.String("apikey", "", "api key for auth")
	token := flag.String("token", "", "token (Bearer auth)")
	outputDir := flag.String("output", "fuzzer_output", "output directory")
	detailedOutput := flag.Bool("detailed", false, "include successful test cases")
	duration := flag.Duration("duration", 3600000000000, "time duration of fuzzing")
	enableDebug := flag.Bool("debug", false, "enable debug logs")
	rps := flag.Int("rps", 0, "requests per second limit (0 means no limit)")

	flag.Parse()

	if *spec == "" {
		flag.PrintDefaults()
		log.Fatal("Provide OpenAPI file")
	}

	if *targetURL == "" {
		log.Info("No target URL provided via CLI. Will attempt to use server URLs from the OpenAPI specification.")
	}

	validateSpecFileExtension(*spec)

	re := regexp.MustCompile("/$")
	if re.FindStringSubmatch(*targetURL) != nil {
		*targetURL = re.Split(*targetURL, 2)[0]
	}

	extraArgs := flag.Args()

	log.Info("[+++] cli arguments are parsed")

	return *spec, *targetURL, *username, *password, *apikey, *token, *outputDir, *detailedOutput, *duration, extraArgs, *enableDebug, *rps
}

// validateSpecFileExtension checks if the spec file has a valid extension
func validateSpecFileExtension(specFile string) {
	ext := strings.ToLower(filepath.Ext(specFile))
	if ext != ".json" && ext != ".yaml" && ext != ".yml" {
		log.Fatalf("Invalid spec file extension: %s. Supported formats are JSON (.json), YAML (.yaml, .yml)", ext)
	}
}
