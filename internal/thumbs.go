package internal

import (
	"encoding/json"
)

type thumbsUpRecord struct {
	Permalink string
	Counts    map[string]bool // map of IP addresses that have given a thumbs up
}

func GetThumbsUpCount(permalink string) (int, error) {
	var thumbsUp []thumbsUpRecord

	bytes, err := GetCache(ThumbsUpCache, 0)
	if err != nil {
		return 0, err
	}
	if len(bytes) == 0 {
		return 0, nil
	}
	err = json.Unmarshal(bytes, &thumbsUp)
	if err != nil {
		return 0, err
	}

	for _, record := range thumbsUp {
		if record.Permalink == permalink {
			return len(record.Counts), nil
		}
	}
	return 0, nil
}

func ToggleThumbsUp(permalink string, ipAddress string) (int, error) {
	var thumbsUp []thumbsUpRecord

	bytes, err := GetCache(ThumbsUpCache, 0)
	if err != nil {
		return 0, err
	}
	if len(bytes) != 0 {
		err = json.Unmarshal(bytes, &thumbsUp)
		if err != nil {
			return 0, err
		}

		// Check for existing record
		for i, record := range thumbsUp {
			if record.Permalink == permalink {
				if record.Counts == nil {
					record.Counts = make(map[string]bool)
				}
				// Toggle thumbs up for the IP address
				if !record.Counts[ipAddress] {
					record.Counts[ipAddress] = true
					thumbsUp[i] = record
				} else {
					delete(record.Counts, ipAddress)
					thumbsUp[i] = record
				}
				count := len(record.Counts)
				b, err := json.Marshal(thumbsUp)
				if err != nil {
					return count, err
				}
				err = SetCache(ThumbsUpCache, b)
				if err != nil {
					return count, err
				}
				return count, nil
			}
		}
	}

	// If permalink not found, create a new record
	newRecord := thumbsUpRecord{
		Permalink: permalink,
		Counts:    map[string]bool{ipAddress: true},
	}
	thumbsUp = append(thumbsUp, newRecord)
	b, err := json.Marshal(thumbsUp)
	if err != nil {
		return 1, err
	}
	err = SetCache(ThumbsUpCache, b)
	if err != nil {
		return 1, err
	}
	return 1, nil
}
