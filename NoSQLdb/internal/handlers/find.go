package handlers

import (
	"nosql_db/internal/api"
	"nosql_db/internal/index"
	"nosql_db/internal/operators"
	"nosql_db/internal/storage"
)

func handleFind(coll *storage.Collection, req api.Request) api.Response {
	var results []map[string]any
	usedIndex := false

	if len(req.Query) == 1 && !hasLogicalOperators(req.Query) {
		for field, condition := range req.Query {
			if coll.HasIndex(field) {
				results = findWithIndex(coll, field, condition)
				usedIndex = true
				break
			}
		}
	}

	if !usedIndex {
		results = findFullScan(coll, req.Query)
	}

	return api.Response{
		Status: api.StatusSuccess,
		Data:   results,
		Count:  len(results),
	}
}

func hasLogicalOperators(conditions map[string]any) bool {
	_, hasOr := conditions["$or"]
	_, hasAnd := conditions["$and"]
	return hasOr || hasAnd
}

func findFullScan(coll *storage.Collection, queryMap map[string]any) []map[string]any {
	var results []map[string]any
	allDocs := coll.All()

	for _, doc := range allDocs {
		if operators.MatchDocument(doc, queryMap) {
			results = append(results, doc)
		}
	}
	return results
}

func findWithIndex(coll *storage.Collection, field string, condition any) []map[string]any {
	btree, ok := coll.GetIndex(field)
	if !ok {
		return nil
	}

	var docIDs []string
	switch v := condition.(type) {
	case float64, int, int64, string, bool:
		key := index.ValueToKey(v)
		values := btree.Search(key)
		docIDs = index.ValuesToStrings(values)
	case map[string]any:
		if gtValue, exists := v["$gt"]; exists {
			key := index.ValueToKey(gtValue)
			values := btree.SearchGreaterThan(key)
			docIDs = index.ValuesToStrings(values)
		} else if ltValue, exists := v["$lt"]; exists {
			key := index.ValueToKey(ltValue)
			values := btree.SearchLessThan(key)
			docIDs = index.ValuesToStrings(values)
		} else if eqValue, exists := v["$eq"]; exists {
			key := index.ValueToKey(eqValue)
			values := btree.Search(key)
			docIDs = index.ValuesToStrings(values)
		} else if inValues, exists := v["$in"]; exists {
			if inArray, ok := inValues.([]any); ok {
				var keys []index.Key
				for _, val := range inArray {
					keys = append(keys, index.ValueToKey(val))
				}
				values := btree.SearchIn(keys)
				docIDs = index.ValuesToStrings(values)
			}
		}
	}

	var results []map[string]any
	for _, id := range docIDs {
		if doc, ok := coll.GetByID(id); ok {
			results = append(results, doc)
		}
	}
	return results
}
