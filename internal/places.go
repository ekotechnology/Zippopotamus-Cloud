package internal

import (
	"errors"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/sirupsen/logrus"
	"strconv"
	"strings"
)

type PlaceIndex int

const (
	countryCode PlaceIndex = iota
	postalCode
	placeName
	adminName1
	adminCode1
	adminName2
	adminCode2
	adminName3
	adminCode3
	latitude
	longitude
	accuracy
)

type Place struct {
	PlaceName   string  `json:"place_name"`
	CountryCode string  `json:"country_code"`
	Country     string  `json:"country"`
	PostalCode  string  `json:"postal_code"`
	AdminName1  string  `json:"admin_name_1"`
	AdminCode1  string  `json:"admin_code_1"`
	AdminName2  string  `json:"admin_name_2"`
	AdminCode2  string  `json:"admin_code_2"`
	AdminName3  string  `json:"admin_name_3"`
	AdminCode3  string  `json:"admin_code_3"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	Accuracy    int64   `json:"accuracy"`
	hashkey     string
	admin1key   string
	admin2key   string
}

func (p *Place) Admin1Key() string {
	if p.admin1key != "" {
		return p.admin1key
	}

	key := fmt.Sprintf("%s.%s", p.CountryCode, p.AdminCode1)

	p.admin1key = key

	return key
}

func (p *Place) Admin2Key() string {
	if p.admin2key != "" {
		return p.admin2key
	}

	key := fmt.Sprintf("%s.%s.%s", p.CountryCode, p.AdminCode1, p.AdminCode2)

	p.admin2key = key

	return key
}

func (p *Place) Key(diff uint64) string {
	return fmt.Sprintf("%s:%s.%d", p.CountryCode, p.PostalCode, diff)
}

func (p *Place) HashFields() map[string]interface{} {
	fields := make(map[string]interface{})

	fields["PlaceName"] = p.PlaceName
	fields["CountryCode"] = p.CountryCode
	fields["PostalCode"] = p.PostalCode
	fields["AdminCode1"] = p.AdminCode1
	fields["AdminCode2"] = p.AdminCode2
	fields["AdminCode3"] = p.AdminCode3
	fields["AdminName3"] = p.AdminName3
	fields["Location"] = fmt.Sprintf("%f %f", p.Longitude, p.Latitude)
	fields["Accuracy"] = p.Accuracy

	return fields
}

func locationToLatLngPair(location string) (float64, float64) {
	parts := strings.Split(location, " ")

	lat, err := strconv.ParseFloat(parts[0], 64)

	if err != nil {
		lat = float64(0)
	}

	long, err := strconv.ParseFloat(parts[1], 64)

	if err != nil {
		long = float64(0)
	}

	return lat, long
}

func PlaceFromRedisHashInterfaceSlice(raw []interface{}) Place {
	place := Place{}

	props := make(map[string]interface{}, len(raw)/2)

	for i := 0; i+1 <= len(raw); i = i + 2 {
		key := raw[i].(string)
		value := raw[i+1]

		props[key] = value
	}

	place.CountryCode = props["CountryCode"].(string)
	place.PlaceName = props["PlaceName"].(string)
	place.PostalCode = props["PostalCode"].(string)
	place.AdminCode1 = props["AdminCode1"].(string)
	place.AdminCode2 = props["AdminCode2"].(string)
	place.AdminCode3 = props["AdminCode3"].(string)
	place.AdminName3 = props["AdminName3"].(string)

	lat, long := locationToLatLngPair(props["Location"].(string))

	place.Latitude = lat
	place.Longitude = long

	return place
}

func CountPlacesByCountryCode(r *redis.Client, l *logrus.Logger, countryCode string) (int64, error) {
	cmd := r.Do("FT.SEARCH", "places", fmt.Sprintf("@CountryCode:%s", countryCode), "NOCONTENT", "LIMIT", 0, 0)

	res, err := cmd.Result()

	if err != nil {
		return 0, err
	}

	data, ok := res.([]interface{})

	if !ok {
		return 0, errors.New("unable to cast response to interface slice")
	}

	count := data[0].(int64)

	return count, nil
}

func QueryPlaces(r *redis.Client, l *logrus.Logger, query string) ([]Place, error) {
	cmd := r.Do("FT.SEARCH", "places", query, "VERBATIM")

	data, err := cmd.Result()

	if err != nil {
		return []Place{}, err
	}

	d, ok := data.([]interface{})

	if !ok {
		return []Place{}, fmt.Errorf("converting data to slice of interface didn't work")
	}

	resultCount, ok := d[0].(int64)

	if !ok {
		return []Place{}, fmt.Errorf("converting data[0] to uint64 to get result count failed")
	}

	places := make([]Place, 0, resultCount)

	for _, val := range d[1:] {
		switch v := val.(type) {
		case []interface{}:
			places = append(places, PlaceFromRedisHashInterfaceSlice(v))
		}
	}

	return places, nil
}
