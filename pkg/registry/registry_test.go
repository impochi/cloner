//nolint:testpackage
package registry

import (
	"fmt"
	"os"
	"testing"
)

const (
	username = "foo"
	password = "bar"
	provider = "test"
)

func TestFetchCredentials(t *testing.T) {
	cases := []struct {
		provider  string
		username  string
		password  string
		expectErr bool
	}{
		{
			provider: provider,
			username: username,
			password: password,
		},
		{
			provider: "",
			username: username,
			password: password,
		},
		{
			provider:  provider,
			username:  "",
			password:  password,
			expectErr: true,
		},
		{
			provider:  provider,
			username:  username,
			password:  "",
			expectErr: true,
		},
		{
			provider:  provider,
			username:  "",
			password:  "",
			expectErr: true,
		},
		{
			provider:  "",
			username:  "",
			password:  "",
			expectErr: true,
		},
	}

	for _, testcase := range cases {
		if err := os.Setenv("REGISTRY_PROVIDER", testcase.provider); err != nil {
			t.Fatalf("Failed to set env variable `REGISTRY_PROVIDER`: %q", err)
		}

		if err := os.Setenv("REGISTRY_USERNAME", testcase.username); err != nil {
			t.Fatalf("Failed to set env variable `REGISTRY_PROVIDER`: %q", err)
		}

		if err := os.Setenv("REGISTRY_PASSWORD", testcase.password); err != nil {
			t.Fatalf("Failed to set env variable `REGISTRY_PROVIDER`: %q", err)
		}

		_, err := fetchCredentials()
		if err != nil && !testcase.expectErr {
			t.Errorf("Failed to fetch credentials: %v", err)
		}
	}
}

func TestGetDestinationImage(t *testing.T) { //nolint:funlen
	cases := []struct {
		provider string
		username string
		input    string
		output   string
	}{
		{
			provider: provider,
			username: username,
			input:    "ubuntu",
			output:   fmt.Sprintf("%s/%s/ubuntu", provider, username),
		},
		{
			provider: "",
			username: username,
			input:    "ubuntu",
			output:   fmt.Sprintf("%s/ubuntu", username),
		},
		{
			provider: provider,
			username: username,
			input:    "quay.io/busybox",
			output:   fmt.Sprintf("%s/%s/busybox", provider, username),
		},
		{
			provider: provider,
			username: username,
			input:    "ubuntu:1.0",
			output:   fmt.Sprintf("%s/%s/ubuntu:1.0", provider, username),
		},
		{
			provider: provider,
			username: username,
			input:    "quay.io/testrepo:v2.0",
			output:   fmt.Sprintf("%s/%s/testrepo:v2.0", provider, username),
		},
		{
			provider: "",
			username: username,
			input:    "quay.io/testrepo:v2.0",
			output:   fmt.Sprintf("%s/testrepo:v2.0", username),
		},
	}

	for _, testcase := range cases {
		if err := os.Setenv("REGISTRY_PROVIDER", testcase.provider); err != nil {
			t.Fatalf("Failed to set env variable `REGISTRY_PROVIDER`: %q", err)
		}

		if err := os.Setenv("REGISTRY_USERNAME", testcase.username); err != nil {
			t.Fatalf("Failed to set env variable `REGISTRY_PROVIDER`: %q", err)
		}

		if err := os.Setenv("REGISTRY_PASSWORD", password); err != nil {
			t.Fatalf("Failed to set env variable `REGISTRY_PROVIDER`: %q", err)
		}

		dst, err := GetDestinationImage(testcase.input)
		if err != nil {
			t.Errorf("Failed to get destination image: %v", err)
		}

		if testcase.output != dst {
			t.Errorf("Expected destination image as %q, got %q", testcase.output, dst)
		}
	}
}

func TestGetRepoAndTagFromImage(t *testing.T) {
	cases := []struct {
		input string
		repo  string
		tag   string
	}{
		{
			input: "ubuntu",
			repo:  "ubuntu",
			tag:   "",
		},
		{
			input: "quay.io/busybox",
			repo:  "busybox",
			tag:   "",
		},
		{
			input: "ubuntu:1.0",
			repo:  "ubuntu",
			tag:   "1.0",
		},
		{
			input: "quay.io/testrepo:v2.0",
			repo:  "testrepo",
			tag:   "v2.0",
		},
	}

	for _, testcase := range cases {
		imagerepo, imagetag := getRepoAndTagFromImage(testcase.input)

		if testcase.repo != imagerepo {
			t.Errorf("Expected repository name as %q, got %q", testcase.repo, imagerepo)
		}

		if testcase.tag != imagetag {
			t.Errorf("Expected repository name as %q, got %q", testcase.tag, imagetag)
		}
	}
}
