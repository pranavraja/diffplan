package main

import (
	"bufio"
	"fmt"
	"io"
	"reflect"
	"sort"
	"strings"

	"github.com/vrischmann/logfmt"
)

type Resource struct {
	ID     string            `json:"id"`
	Fields map[string]string `json:"fields"`
}

type Change string

const (
	Create Change = "CREATE"
	Update Change = "UPDATE"
	Delete Change = "DELETE"
)

type Plan struct {
	Change Change `json:"change,omitempty"` // CREATE,UPDATE,DELETE
	Resource
	Old map[string]string `json:"old,omitempty"`
}

func (p Plan) String() string {
	var pairs logfmt.Pairs
	for k, v := range p.Fields {
		pairs = append(pairs, logfmt.Pair{Key: k, Value: v})
	}
	sort.Sort(pairs)
	switch p.Change {
	case Create:
		return fmt.Sprintf("%v", pairs.Format())
	default:
		return fmt.Sprintf("#%s: %s", p.ID, pairs.Format())
	}
}

type Plans []Plan

func (plans Plans) String() string {
	w := new(strings.Builder)
	for i, p := range plans {
		fmt.Fprintf(w, " - %s\n", p)
		if i >= 9 {
			fmt.Fprintf(w, " ...and %d more\n", len(plans)-i)
			break
		}
	}
	return w.String()
}

func diff(original, current []Resource) (plans Plans) {
	og := make(map[string]Resource)
	for _, v := range original {
		og[v.ID] = v
	}

	stillaround := make(map[string]struct{})
	for _, v := range current {
		stillaround[v.ID] = struct{}{}
		existing, ok := og[v.ID]
		if !ok {
			plans = append(plans, Plan{Create, v, nil})
			continue
		}
		if !reflect.DeepEqual(existing.Fields, v.Fields) {
			v.Fields = fieldChange(existing.Fields, v.Fields)
			plans = append(plans, Plan{Update, v, existing.Fields})
		}
	}

	for _, v := range original {
		if _, ok := stillaround[v.ID]; !ok {
			plans = append(plans, Plan{Delete, v, nil})
		}
	}
	return plans
}

func fieldChange(from, to map[string]string) map[string]string {
	changes := make(map[string]string)
	for k, v := range to {
		if from[k] != v {
			changes[k] = v
		}
	}
	return changes
}

func parse(r io.Reader) ([]Resource, error) {
	scanner := bufio.NewScanner(r)
	var resources []Resource
	for scanner.Scan() {
		pairs := logfmt.Split(strings.TrimSpace(scanner.Text()))
		var resource Resource
		resource.Fields = make(map[string]string)
		for _, pair := range pairs {
			if pair.Key == "id" {
				resource.ID = pair.Value
			} else {
				resource.Fields[pair.Key] = pair.Value
			}
		}
		resources = append(resources, resource)
	}
	return resources, nil
}
