package internal

import (
	"bufio"
	"github.com/go-redis/redis"
	"github.com/sirupsen/logrus"
	"go.uber.org/atomic"
	"os"
	"strconv"
	"strings"
	"time"
)

type postcodeParser struct {
	r       *redis.Client
	counter *atomic.Uint64
	log     *logrus.Logger
}

func (p *postcodeParser) readFromStdin(places chan *Place) {
	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		if err := scanner.Err(); err != nil {
			p.log.WithError(err).Fatal("unable to properly read from stdin")
		}

		line := scanner.Text()

		row := strings.Split(line, "\t")

		logContext := p.log.WithFields(logrus.Fields{
			"CountryCode": row[countryCode],
			"PostalCode":  row[postalCode],
			"PlaceName":   row[placeName],
		})

		lat, err := strconv.ParseFloat(row[latitude], 64)

		if err != nil {
			lat = float64(0)
			logContext.Debugf("'%s' for latitude failed to parse as float", row[latitude])
		}

		long, err := strconv.ParseFloat(row[longitude], 64)

		if err != nil {
			long = float64(0)
			logContext.Debugf("'%s' for longitude failed to parse as float", row[longitude])
		}

		var accuracyValue int64
		if len(row) == 12 {
			accuracyValue, err = strconv.ParseInt(row[accuracy], 10, 64)

			if err != nil {
				accuracyValue = int64(0)
				logContext.Debugf("'%s' for accuracy failed to parse as int", row[accuracy])
			}
		} else {
			logContext.Debugf("%+v contains len %d", row, len(row))
			accuracyValue = 0
		}

		places <- &Place{
			CountryCode: row[countryCode],
			PostalCode:  row[postalCode],
			PlaceName:   row[placeName],
			AdminName1:  row[adminName1],
			AdminCode1:  row[adminCode1],
			AdminName2:  row[adminName2],
			AdminCode2:  row[adminCode2],
			AdminName3:  row[adminName3],
			AdminCode3:  row[adminCode3],
			Latitude:    lat,
			Longitude:   long,
			Accuracy:    accuracyValue,
		}
	}

	close(places)
}

func (p *postcodeParser) WaitUntilRedisReadyOrFail() {
	attemptLimit := 3
	for attempts := 1; attempts <= attemptLimit; attempts++ {
		_, err := p.r.Ping().Result()

		if err != nil {
			if attempts == attemptLimit {
				p.log.WithError(err).Fatalf("Failed to connect to Redisearch instance after %d tries", attemptLimit)
			} else {
				delay := 5 * attempts
				p.log.WithError(err).Warnf("Failed to connect to Redisearch, will try again in %d seconds", delay)
				time.Sleep(time.Duration(delay) * time.Second)
			}
		} else {
			break
		}
	}
	p.log.Infof("Redis is ready!")
}

func (p *postcodeParser) Run() {
	start := time.Now()

	p.log.Info("Creating FT index...")
	_, err := ftIndex(p.r)

	if err != nil {
		p.log.Warn(err)
	}

	places := make(chan *Place)

	go func() {
		p.readFromStdin(places)
	}()

	done := make(chan int)

	for id := 0; id < 10; id++ {
		go func(id int) {
			for place := range places {

				_, err := p.r.Pipelined(func(pipe redis.Pipeliner) error {
					c := p.counter.Inc()
					tries := 0
					for {
						err := pipe.HMSet(place.Key(c), place.HashFields()).Err()
						if err != nil {
							c = p.counter.Inc()
							tries = tries + 1
						} else {
							if tries > 1 {
								p.log.Printf("run: Saved %s after %d tries\n", place.Key(0), tries)
							}
							break
						}
					}

					idx := NewIndexAdd(place.Key(c))

					pipe.Process(idx)

					_, err := idx.Result()

					if err != nil {
						p.log.WithError(err).Error("failed to add to index")
					}

					return nil
				})

				if err != nil {
					p.log.WithError(err).Error("failed to run pipelined commands")
				}
			}

			done <- id
		}(id)

	}

	var doneIds []int
	for {
		select {
		case id := <-done:
			if len(doneIds) == 9 {
				p.log.Infof("Parsing took %s\n", time.Since(start))
				return
			}

			doneIds = append(doneIds, id)
		}
	}
}

type PostcodeParserOptions struct {
	RedisAddr       string
	SkipCreateIndex bool
	SkipCountries   string
	Logger          *logrus.Logger
}

func NewPostcodeParser(options *PostcodeParserOptions) *postcodeParser {
	parser := &postcodeParser{}

	parser.r = redis.NewClient(&redis.Options{
		Addr:     options.RedisAddr,
		Password: "",
		DB:       0,
	})

	parser.counter = new(atomic.Uint64)

	parser.log = options.Logger

	return parser
}