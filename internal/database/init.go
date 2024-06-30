package database

import (
	"bursa-alert/lib/alerts"
	"encoding/gob"
	"os"
)

func init() {
	// Create the config directory if it doesn't exist
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		panic(err)
	}
	err = os.MkdirAll(cfgDir+"/bursa", 0755)
	if err != nil {
		panic(err)
	}
}

// We don't actually need full SQL and stuff. Just gobs are fine

func LoadAlerts() []alerts.Alert {
	// Get cross platform config path
	f, err := os.Open(cfgPath())
	if err != nil {
		return []alerts.Alert{}
	}
	defer f.Close()
	dec := gob.NewDecoder(f)
	var al []alerts.Alert
	err = dec.Decode(&al)
	if err != nil {
		// Delete the file
		os.Remove(cfgPath())
		return []alerts.Alert{}
	}
	return al
}

func SaveAlerts(al []alerts.Alert) {
	// Get cross platform config path
	f, err := os.Create(cfgPath())
	if err != nil {
		panic(err)
	}
	defer f.Close()
	enc := gob.NewEncoder(f)
	err = enc.Encode(al)
	if err != nil {
		panic(err)
	}
}

func cfgPath() string {
	// Get cross platform config path
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		panic(err)
	}
	return cfgDir + "/bursa/" + "alerts.gob"
}
