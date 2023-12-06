package main

type Deployment struct {
	namespace string
	name      string
	depLabels map[string]string
}
