package models

type AllConfigsForScope struct {
	OrderedHeirarchy []string `json:"ordered_heirarchy"`

	Features map[string]map[string]interface{} `json:"features"`
}
