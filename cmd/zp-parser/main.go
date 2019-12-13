package main

import (
	"flag"
	"github.com/sirupsen/logrus"
	"github.com/zippopotamus/zippopotamus/internal"
)

func main() {
	opts := &internal.PostcodeParserOptions{}

	flag.StringVar(&opts.RedisAddr, "redis-addr", "127.0.0.1:6379", "the location at which Redis server can be found")
	flag.StringVar(&opts.SkipCountries, "skip-countries", "", "a comma separated list of countries that may be skipped while processing")

	flag.Parse()

	l := logrus.StandardLogger()

	l.SetLevel(logrus.TraceLevel)

	opts.Logger = l

	parser := internal.NewPostcodeParser(opts)

	parser.WaitUntilRedisReadyOrFail()

	parser.Run()
}
