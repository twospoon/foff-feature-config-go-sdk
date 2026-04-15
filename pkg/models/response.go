package models

type GetAllConfigsForScopeResponse struct {
	OrderedHeirarchy []string `json:"ordered_heirarchy"`

	Features map[string]map[string]interface{} `json:"features"`
}
