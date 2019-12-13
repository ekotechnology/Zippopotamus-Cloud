package internal

import (
	"bufio"
	"github.com/sirupsen/logrus"
	"os"
	"strings"
	"time"
)

type tsvParser struct {
	log      *logrus.Logger
	routines int
}

func (p *tsvParser) read(from *os.File, output chan []string) {
	scanner := bufio.NewScanner(from)

	for scanner.Scan() {
		if err := scanner.Err(); err != nil {
			p.log.WithError(err).Fatal("unable to properly read")
		}

		line := scanner.Text()

		// Skip comment lines
		if strings.Index(line, "#") == 0 {
			p.log.Debug("Skip comment line", line)
			continue
		}

		output <- strings.Split(line, "\t")
	}

	close(output)
}

func (p *tsvParser) Run(file *os.File, runner func([]string)) {
	start := time.Now()

	output := make(chan []string)

	go func() {
		p.read(file, output)
	}()

	done := make(chan int)

	for id := 0; id < p.routines; id++ {
		go func(id int) {
			for line := range output {
				runner(line)
			}

			done <- id
		}(id)

	}

	var doneIds []int
	for {
		select {
		case id := <-done:
			if len(doneIds) == p.routines-1 {
				p.log.Infof("Parsing %s took %s\n", file.Name(), time.Since(start))
				return
			}

			doneIds = append(doneIds, id)
		}
	}
}

type TsvParserOptions struct {
	Logger   *logrus.Logger
	Routines int
}

func NewTsvParser(options *TsvParserOptions) *tsvParser {
	parser := &tsvParser{}

	parser.log = options.Logger

	parser.routines = options.Routines

	return parser
}
