/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
)

type stageStrategy struct {
	bucket string
	ci     bool
	suffix string
}

// Return something like gs://bucket/ci/suffix
func (s *stageStrategy) String() string {
	p := "devel"
	if s.ci {
		p = "ci"
	}
	return fmt.Sprintf("%v%v%v", s.bucket, p, s.suffix)
}

// Parse bucket, ci, suffix from gs://BUCKET/ci/SUFFIX
func (s *stageStrategy) Set(value string) error {
	re := regexp.MustCompile(`^(gs://[\w-]+)/(devel|ci)(/.*)?`)
	mat := re.FindStringSubmatch(value)
	if mat == nil {
		return fmt.Errorf("Invalid stage location: %v. Use gs://bucket/ci/optional-suffix", value)
	}
	s.bucket = mat[1]
	s.ci = mat[2] == "ci"
	s.suffix = mat[3]
	return nil
}

// True when this kubetest invocation wants to stage the release
func (s *stageStrategy) Enabled() bool {
	return s.bucket != ""
}

// Stage the release build to GCS.
// Essentially release/push-build.sh --bucket=B --ci? --gcs-suffix=S --federation?
func (s *stageStrategy) Stage() error {
	name := "../release/push-build.sh"
	args := []string{
		"--nomock",
		"--verbose",
		fmt.Sprintf("--bucket=%v", s.bucket),
	}
	if s.ci {
		args = append(args, "--ci")
	}
	if len(s.suffix) > 0 {
		args = append(args, fmt.Sprintf("--gcs-suffix=%v", s.suffix))
	}
	if os.Getenv("FEDERATION") == "true" {
		args = append(args, "--federation")
	}

	return finishRunning(exec.Command(name, args...))
}
