package internal

import (
	"fmt"
	"slices"
	"strings"

	"github.com/haijima/stool/internal/log"
	"golang.org/x/exp/maps"
)

type ParamProfiler struct {
}

func NewParamProfiler() *ParamProfiler {
	return &ParamProfiler{}
}

func (p *ParamProfiler) Profile(reader *log.LTSVReader) (*Param, error) {
	param := &Param{
		Endpoints:             make([]string, 0),
		Count:                 make(map[string]int),
		Path:                  make(map[string][]map[string]int),
		PathName:              make(map[string][]string),
		QueryKey:              make(map[string]map[string]int),
		QueryKeyCombination:   make(map[string]map[string]int),
		QueryValue:            make(map[string]map[string]map[string]int),
		QueryValueCombination: make(map[string]map[string]int),
	}

	var entry log.LogEntry
	endpointsMap := make(map[string]interface{})
	for reader.Read() {
		_, err := reader.Parse(&entry)
		if err != nil {
			if err == log.Filtered {
				continue
			}
			return nil, err
		}

		if entry.MatchedGroup == nil {
			continue
		}

		_, uri, query := log.ParseReq(entry.Req)

		// Path param
		subMatches := entry.MatchedGroup.FindStringSubmatch(uri)
		key := fmt.Sprintf("%s %s", entry.Method, entry.Uri)
		endpointsMap[key] = nil
		if len(subMatches) > 1 { // this entry URI has path param
			if _, ok := param.Path[key]; !ok {
				param.Path[key] = make([]map[string]int, len(subMatches)-1)
				for i := range param.Path[key] {
					param.Path[key][i] = map[string]int{}
				}
				param.PathName[key] = make([]string, len(subMatches)-1)
			}
			for i, v := range subMatches[1:] {
				param.Path[key][i][v] += 1
				param.PathName[key][i] = entry.MatchedGroup.SubexpNames()[i+1]
			}
		}

		// Query param
		if query != "" {
			if _, ok := param.QueryValue[key]; !ok {
				param.QueryKey[key] = map[string]int{}
				param.QueryKeyCombination[key] = map[string]int{}
				param.QueryValue[key] = map[string]map[string]int{}
				param.QueryValueCombination[key] = map[string]int{}
			}
			qs := strings.Split(query, "&")
			slices.Sort(qs)
			qks := make([]string, 0)
			for _, q := range qs {
				if k, v, ok := strings.Cut(q, "="); ok {
					param.QueryKey[key][k] += 1
					if _, ok := param.QueryValue[key][k]; !ok {
						param.QueryValue[key][k] = map[string]int{}
					}
					param.QueryValue[key][k][v] += 1
					qks = append(qks, k)
				}
			}
			// qks is not needed to be sorted because it is already sorted (qs is sorted)
			param.QueryKeyCombination[key][strings.Join(qks, "&")] += 1
			param.QueryValueCombination[key][strings.Join(qs, "&")] += 1
		}

		param.Count[key] += 1
	}

	param.Endpoints = maps.Keys(endpointsMap)
	slices.SortFunc(param.Endpoints, func(i, j string) int {
		ii := strings.Split(i, " ")
		jj := strings.Split(j, " ")
		if ii[1] != jj[1] {
			return strings.Compare(ii[1], jj[1])
		}
		return strings.Compare(ii[0], jj[0])
	})
	return param, nil
}

type Param struct {
	Endpoints             []string
	Count                 map[string]int
	Path                  map[string][]map[string]int
	PathName              map[string][]string
	QueryKey              map[string]map[string]int
	QueryKeyCombination   map[string]map[string]int
	QueryValue            map[string]map[string]map[string]int
	QueryValueCombination map[string]map[string]int
}
