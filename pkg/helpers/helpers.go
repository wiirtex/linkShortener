package helpers

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"time"
)

const LoggerString = "logger"
const LongLinkString = "longLink"
const ShortLinkString = "shortLink"

type config struct {
	ShortLinkBase         string `json:"shortLinkBase"`
	CacheMaxLength        int64  `json:"cacheMaxLength"`
	CacheTimeToLive       time.Duration
	CacheTimeToLive_      int64   `json:"cacheTimeToLive"`
	CacheClearCoefficient float32 `json:"cacheClearCoefficient"`
	DbConnString          string  `json:"dbConnString"`
}

var configInstance = config{}

func GetConfig() config {
	return configInstance
}

func init() {
	readAndPlace("secrets.json")
	readAndPlace("config.json")
	configInstance.CacheTimeToLive = time.Duration(configInstance.CacheTimeToLive_)
}

func readAndPlace(file string) error {
	secrets, err := os.Open(file)
	if err != nil {
		return errors.New("Can not open secrets.json: " + err.Error())
	}
	secs, err := ioutil.ReadAll(secrets)
	if err != nil {
		return errors.New("Can not read secrets.json: " + err.Error())
	}
	err = json.Unmarshal(secs, &configInstance)
	if err != nil {
		return errors.New("Can not unmarshal secrets.json: " + err.Error())
	}
	return nil
}
