package internal

import (
	"github.com/go-redis/redis"
)

func ftIndex(c *redis.Client) (*redis.BoolCmd, error) {
	// hello darkness my old friend :'(
	cmd := redis.NewBoolCmd(
		"FT.CREATE",
		"places",

		"NOOFFSETS",
		"NOFREQS",

		"STOPWORDS",
		"0",

		"SCHEMA",
		"CountryCode",
		"TEXT",
		"NOSTEM",
		"WEIGHT",
		"1",

		"PostalCode",
		"TEXT",
		"NOSTEM",
		"WEIGHT",
		"1",

		"PlaceName",
		"TEXT",
		"NOSTEM",
		"WEIGHT",
		"1",

		"AdminCode1",
		"TEXT",
		"NOSTEM",
		"WEIGHT",
		"1",

		"Location",
		"GEO",
	)

	if err := c.Process(cmd); err != nil {
		return cmd, err
	}

	return cmd, nil
}

func NewIndexAdd(key string) *redis.BoolCmd {
	return redis.NewBoolCmd("FT.ADDHASH", "places", key, "1.0")
}
