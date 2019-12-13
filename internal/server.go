package internal

import (
	"fmt"
	"github.com/go-chi/chi"
	"github.com/go-redis/redis"
	"github.com/segmentio/encoding/json"
	"github.com/sirupsen/logrus"
	"github.com/umahmood/haversine"
	"net/http"
	"net/url"
	"sort"
)

type httpServerHandlers struct {
	redis *redis.Client
	log   *logrus.Logger
	names *AdminCodeNames
}

func NewHttpServerHandlers(redis *redis.Client, log *logrus.Logger, names *AdminCodeNames) *httpServerHandlers {
	return &httpServerHandlers{redis, log, names}
}

func (s *httpServerHandlers) hydrateStaticData(places []Place) []Place {
	for i, _ := range places {
		s.names.ExpandPlace(&places[i])
	}

	return places
}

func (s *httpServerHandlers) HandleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Location", "https://docs.zippopotam.us")
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func (s *httpServerHandlers) HandleCheckCountryAvailable(w http.ResponseWriter, r *http.Request) {
	if r.Method != "HEAD" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	countryCode := chi.URLParam(r, "countryCode")

	count, err := CountPlacesByCountryCode(s.redis, s.log, countryCode)

	if err != nil {
		s.log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if count == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *httpServerHandlers) HandleGetPlacesByCountryAndPostCode(w http.ResponseWriter, r *http.Request) {
	countryCode := chi.URLParam(r, "countryCode")
	postalCode := chi.URLParam(r, "postalCode")

	places, err := QueryPlaces(s.redis, s.log, fmt.Sprintf("@CountryCode:'%s' @PostalCode:'%s'", countryCode, postalCode))

	if err != nil {
		s.log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if len(places) == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var res interface{}

	v := GetVersion(r)

	places = s.hydrateStaticData(places)

	switch v {
	case "v1":
		placeData := make([]map[string]string, 0, len(places))

		for i, p := range places {
			data := make(map[string]string)

			data["place name"] = p.PlaceName
			data["longitude"] = fmt.Sprintf("%.4f", p.Longitude)
			data["state"] = places[i].AdminName1
			data["state abbreviation"] = p.AdminCode1
			data["latitude"] = fmt.Sprintf("%.4f", p.Latitude)
			data["county"] = places[i].AdminName2

			placeData = append(placeData, data)
		}

		res = struct {
			PostCode            string              `json:"post code"`
			Country             string              `json:"country"`
			CountryAbbreviation string              `json:"country abbreviation"`
			Places              []map[string]string `json:"places"`
		}{
			PostCode:            places[0].PostalCode,
			Country:             places[0].Country,
			CountryAbbreviation: places[0].CountryCode,
			Places:              placeData,
		}
	}

	jsonRes, err := json.Marshal(res)
	if err != nil {
		s.log.WithError(err).Error("Unable to marshal response json")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	if _, err := w.Write(jsonRes); err != nil {
		s.log.WithError(err).Error("Unable to write response json body")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (s *httpServerHandlers) HandleGetPlacesByCountryAreaAndPlaceName(w http.ResponseWriter, r *http.Request) {
	countryCode := chi.URLParam(r, "countryCode")
	area := chi.URLParam(r, "area")
	place, err := url.QueryUnescape(chi.URLParam(r, "place"))

	if err != nil {
		s.log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	places, err := QueryPlaces(s.redis, s.log, fmt.Sprintf("@CountryCode:'%s' @AdminCode1:'%s' @PlaceName:'%s'", countryCode, area, place))

	if err != nil {
		s.log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if len(places) == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	places = s.hydrateStaticData(places)

	var res interface{}

	v := GetVersion(r)

	switch v {
	case "v1":
		placeData := make([]map[string]string, 0, len(places))

		for _, p := range places {
			data := make(map[string]string)

			data["place name"] = p.PlaceName
			data["post code"] = p.PostalCode
			data["longitude"] = fmt.Sprintf("%.4f", p.Longitude)
			data["latitude"] = fmt.Sprintf("%.4f", p.Latitude)

			placeData = append(placeData, data)
		}

		res = struct {
			Country             string              `json:"country"`
			CountryAbbreviation string              `json:"country abbreviation"`
			PlaceName           string              `json:"place name"`
			State               string              `json:"state"`
			StateAbbreviation   string              `json:"state abbreviation"`
			Places              []map[string]string `json:"places"`
		}{
			Country:             places[0].Country,
			CountryAbbreviation: places[0].CountryCode,
			PlaceName:           places[0].PlaceName,
			State:               places[0].AdminName1,
			StateAbbreviation:   places[0].AdminCode1,
			Places:              placeData,
		}
	}

	jsonRes, err := json.Marshal(res)

	w.Header().Add("Content-Type", "application/json")

	if err != nil {
		s.log.WithError(err).Error("Unable to marshal response json")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if _, err := w.Write(jsonRes); err != nil {
		s.log.WithError(err).Error("Unable to write response json body")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

}

func (s *httpServerHandlers) HandleGetNearbyPlaces(w http.ResponseWriter, r *http.Request) {
	countryCode := chi.URLParam(r, "countryCode")
	postalCode := chi.URLParam(r, "postalCode")

	basisPlaces, err := QueryPlaces(s.redis, s.log, fmt.Sprintf("@CountryCode:'%s' @PostalCode:'%s'", countryCode, postalCode))

	if err != nil {
		s.log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if len(basisPlaces) == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	dist := 10

	places, err := QueryPlaces(s.redis, s.log, fmt.Sprintf("@Location:[%.3f %.3f %d mi]", basisPlaces[0].Latitude, basisPlaces[0].Longitude, dist))

	if err != nil {
		s.log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	places = s.hydrateStaticData(places)

	var res interface{}

	v := GetVersion(r)

	switch v {
	case "v1":
		placeData := make([]map[string]interface{}, 0, len(places))

		for i, p := range places {
			if p.PostalCode == basisPlaces[0].PostalCode {
				continue
			}
			data := make(map[string]interface{})

			data["place name"] = p.PlaceName
			data["post code"] = p.PostalCode
			data["state"] = places[i].AdminName1
			data["state abbreviation"] = p.AdminCode1

			distMi, _ := haversine.Distance(haversine.Coord{
				Lat: basisPlaces[0].Latitude,
				Lon: basisPlaces[0].Longitude,
			}, haversine.Coord{
				Lat: p.Latitude,
				Lon: p.Longitude,
			})

			data["distance"] = distMi

			placeData = append(placeData, data)
		}

		sort.Slice(placeData, func(i, j int) bool {
			return placeData[i]["distance"].(float64) < placeData[j]["distance"].(float64)
		})

		res = struct {
			NearLatitude  float64                  `json:"near latitude"`
			NearLongitude float64                  `json:"near longitude"`
			Places        []map[string]interface{} `json:"nearby"`
		}{
			NearLatitude:  basisPlaces[0].Latitude,
			NearLongitude: basisPlaces[0].Longitude,
			Places:        placeData,
		}
	}

	jsonRes, err := json.Marshal(res)

	w.Header().Add("Content-Type", "application/json")

	if err != nil {
		s.log.WithError(err).Error("Unable to marshal response json")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if _, err := w.Write(jsonRes); err != nil {
		s.log.WithError(err).Error("Unable to write response json body")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

}
