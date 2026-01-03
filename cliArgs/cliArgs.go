package cliArgs

import (
	"flag"
	"regexp"
	"time"

	log "github.com/sirupsen/logrus"
)

func ParseCliArgs() (string, string, string, string, string, string, string, bool, time.Duration, []string, bool) {
	spec := flag.String("spec", "", "openapi file specification (Required)")
	targetURL := flag.String("url", "", "target URL (Required)")
	username := flag.String("user", "", "username (Basic auth)")
	password := flag.String("password", "", "password (Basic auth)")
	apikey := flag.String("apikey", "", "api key for auth")
	token := flag.String("token", "", "token (Bearer auth)")
	outputDir := flag.String("output", "fuzzer_output", "output directory")
	detailedOutput := flag.Bool("detailed", false, "include successful test cases")
	duration := flag.Duration("duration", 3600000000000, "time duration of fuzzing")
	enableDebug := flag.Bool("debug", false, "enable debug logs (default: true)")

	flag.Parse()

	if *spec == "" || *targetURL == "" {
		flag.PrintDefaults()
		log.Fatal("Provide OpenAPI file and target URL")
	}

	re := regexp.MustCompile("/$")
	if re.FindStringSubmatch(*targetURL) != nil {
		*targetURL = re.Split(*targetURL, 2)[0]
	}

	extraArgs := flag.Args()

	log.Info("[+++] cli arguments are parsed")

	return *spec, *targetURL, *username, *password, *apikey, *token, *outputDir, *detailedOutput, *duration, extraArgs, *enableDebug
}
