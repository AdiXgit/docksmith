package graph

import (
    "errors"
)

type DependencyResolver struct {
    graph    map[string][]string
    visited  map[string]bool
    tempMark map[string]bool
}

func NewDependencyResolver() *DependencyResolver {
    return &DependencyResolver{
        graph:    make(map[string][]string),
        visited:  make(map[string]bool),
        tempMark: make(map[string]bool),
    }
}

func (dr *DependencyResolver) AddDependency(dep string, item string) {
    dr.graph[dep] = append(dr.graph[dep], item)
}

func (dr *DependencyResolver) Resolve(order []string) error {
    for item := range dr.graph {
        if err := dr.visit(item, order); err != nil {
            return err
        }
    }
    return nil
}

func (dr *DependencyResolver) visit(item string, order []string) error {
    if dr.tempMark[item] {
        return errors.New("cycle detected")
    }
    if dr.visited[item] {
        return nil
    }

    dr.tempMark[item] = true
    for _, neighbor := range dr.graph[item] {
        if err := dr.visit(neighbor, order); err != nil {
            return err
        }
    }
    dr.tempMark[item] = false
    dr.visited[item] = true
    order = append(order, item)
    return nil
}