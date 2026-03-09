package tools

import "github.com/on2itsecurity/go-auxo/v2/zerotrust"

// FlattenMeasureGroups flattens measure groups into a flat slice of maps
func FlattenMeasureGroups(groups zerotrust.MeasureGroups) []map[string]interface{} {
	var allMeasures []map[string]interface{}
	for _, group := range groups.Groups {
		for _, measure := range group.Measures {
			measureData := map[string]interface{}{
				"name":          measure.Name,
				"caption":       measure.Caption,
				"explanation":   measure.Explanation,
				"mappings":      measure.Mappings,
				"group_name":    group.Name,
				"group_label":   group.Label,
				"group_caption": group.Caption,
			}
			allMeasures = append(allMeasures, measureData)
		}
	}
	return allMeasures
}
