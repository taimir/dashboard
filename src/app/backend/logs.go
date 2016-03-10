// Copyright 2015 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"io/ioutil"
	"log"
	"strings"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"
	client "k8s.io/kubernetes/pkg/client/unversioned"
)

// LogQuery specifies which part of the logs should be returned.
// StartIndex and Count specify the range from the log list to be returned.
type LogQuery struct {
	Namespace  string
	PodID      string
	Container  string
	StartIndex int64
	Count      int64
}

// Logs is a representation of logs response structure.
type Logs struct {
	// Pod name.
	PodID string `json:"PodID"`

	// Specific time when logs started.
	SinceTime unversioned.Time `json:"sinceTime"`

	// Logs string lines.
	Logs []string `json:"logs"`

	// The name of the container the logs are for.
	Container string `json:"container"`

	// The current total number of logs for this Container
	Total int64 `json:"total"`

	// The index of the first returned log
	StartIndex int64 `json:"startIndex"`
}

// GetPodLogs returns logs for particular pod and container or error when occurred. When container
// is null, logs for the first one are returned.
func GetPodLogs(client *client.Client, query *LogQuery) (*Logs, error) {
	log.Printf("Getting logs from %s container from %s pod in %s namespace", query.Container, query.PodID,
		query.Namespace)

	pod, err := client.Pods(query.Namespace).Get(query.PodID)
	if err != nil {
		return nil, err
	}

	if query.Container == "" {
		query.Container = pod.Spec.Containers[0].Name
	}

	logOptions := &api.PodLogOptions{
		Container:  query.Container,
		Follow:     false,
		Previous:   false,
		Timestamps: true,
	}

	rawLogs, err := getRawPodLogs(client, query, logOptions)
	if err != nil {
		return nil, err
	}

	return constructLogs(pod.CreationTimestamp, rawLogs, query), nil
}

// Construct a request for getting the logs for a pod and retrieves the logs.
func getRawPodLogs(client *client.Client, query *LogQuery, logOptions *api.PodLogOptions) (
	string, error) {
	req := client.RESTClient.Get().
		Namespace(query.Namespace).
		Name(query.PodID).
		Resource("pods").
		SubResource("log").
		VersionedParams(logOptions, api.ParameterCodec)

	readCloser, err := req.Stream()
	if err != nil {
		return "", err
	}

	defer readCloser.Close()

	result, err := ioutil.ReadAll(readCloser)
	if err != nil {
		return "", err
	}

	return string(result), nil
}

// Return Logs structure for given parameters.
func constructLogs(sinceTime unversioned.Time, rawLogs string, query *LogQuery) *Logs {
	logs := strings.Split(rawLogs, "\n")
	startIndex, selectedLogs := selectLogs(logs, query.StartIndex, query.Count)
	res := &Logs{
		PodID:      query.PodID,
		SinceTime:  sinceTime,
		Logs:       selectedLogs,
		Container:  query.Container,
		Total:      int64(len(logs)),
		StartIndex: startIndex,
	}
	return res
}

// Selects the logs for a specific page, count specifies the logs count per page.
// Start index specifies the index of the first log on the page.
func selectLogs(logs []string, startIndex, count int64) (int64, []string) {
	if startIndex > (int64(len(logs))) {
		startIndex = -1
	}
	// if startIndex negative, return the last logs page
	if startIndex < 0 {
		// integer division
		startIndex = (int64(len(logs)) / count) * count
	}
	endIndex := startIndex + count
	if endIndex > int64(len(logs)) {
		endIndex = int64(len(logs))
	}
	return startIndex, logs[startIndex:endIndex]
}
