// Package api collects the data of the Hetzner Robot API
package api

import (
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"
)

// LiveData is the response of the Hetzner Robot API
type LiveData struct {
	Hash         string `json:"hash"`
	MinMaxValues struct {
		MinPrice     int     `json:"minPrice"`
		MaxPrice     float64 `json:"maxPrice"`
		MinRAM       int     `json:"minRam"`
		MaxRAM       int     `json:"maxRam"`
		MinHDDSize   int     `json:"minHDDSize"`
		MaxHDDSize   int     `json:"maxHDDSize"`
		MinHDDCount  int     `json:"minHDDCount"`
		MaxHDDCount  int     `json:"maxHDDCount"`
		MinBenchmark int     `json:"minBenchmark"`
		MaxBenchmark int     `json:"maxBenchmark"`
	} `json:"minMaxValues"`
	Server []struct {
		Key          int      `json:"key"`
		Name         string   `json:"name"`
		Description  []string `json:"description"`
		CPU          string   `json:"cpu"`
		CPUBenchmark int      `json:"cpu_benchmark"`
		CPUCount     int      `json:"cpu_count"`
		IsHighio     bool     `json:"is_highio"`
		IsEcc        bool     `json:"is_ecc"`
		Traffic      string   `json:"traffic"`
		Dist         []string `json:"dist"`
		Bandwith     int      `json:"bandwith"`
		RAM          int      `json:"ram"`
		Price        string   `json:"price"`
		PriceV       string   `json:"price_v"`
		RAMHr        string   `json:"ram_hr"`
		SetupPrice   string   `json:"setup_price"`
		HddSize      int      `json:"hdd_size"`
		HddCount     int      `json:"hdd_count"`
		HddHr        string   `json:"hdd_hr"`
		FixedPrice   bool     `json:"fixed_price"`
		NextReduce   int      `json:"next_reduce"`
		NextReduceHr string   `json:"next_reduce_hr"`
		Datacenter   []string `json:"datacenter"`
		Specials     []string `json:"specials"`
		SpecialHdd   string   `json:"specialHdd"`
		Freetext     string   `json:"freetext"`
	} `json:"server"`
}

// GetLiveData returns the live data of the Hetzner Robot API
func GetLiveData() (*LiveData, error) {
	resp, err := http.Get("https://www.hetzner.de/a_hz_serverboerse/live_data.json")
	if err != nil {
		return nil, errors.Wrap(err, "could not get live data")
	}
	liveData := &LiveData{}
	if err := json.NewDecoder(resp.Body).Decode(&liveData); err != nil {
		return nil, errors.Wrap(err, "could not decode response")
	}
	return liveData, nil
}
