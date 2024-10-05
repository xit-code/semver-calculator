package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// SemVer represents a semantic versioning tag
type SemVer struct {
	Major int
	Minor int
	Patch int
}

func (v SemVer) String() string {
	return fmt.Sprintf("v%d.%d.%d", v.Major, v.Minor, v.Patch)
}

// checkIfPathExists verifies if the given path exists
func checkIfPathExists(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("path %s does not exist", path)
	}
	return nil
}

// checkIfGitRepo verifies if the given path is a Git repository
func checkIfGitRepo(path string) error {
	// Change working directory to the provided path
	if err := os.Chdir(path); err != nil {
		return fmt.Errorf("failed to change directory to %s", path)
	}

	// Run the 'git rev-parse --is-inside-work-tree' command to verify if it's a git repo
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	output, err := cmd.CombinedOutput()
	if err != nil || strings.TrimSpace(string(output)) != "true" {
		return fmt.Errorf("path %s is not a Git repository", path)
	}
	return nil
}

// getSemverTags retrieves the list of tags in the format vX.Y.Z and returns a sorted list of SemVer
func getSemverTags() ([]SemVer, error) {
	// Run the 'git tag' command
	cmd := exec.Command("git", "tag", "--list")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to get tags: %s", err)
	}

	// Regex to match semantic versioning tags vX.Y.Z
	semverRegex := regexp.MustCompile(`^v(\d+)\.(\d+)\.(\d+)$`)
	tags := strings.Split(string(output), "\n")
	var semverTags []SemVer

	// Parse tags that match the SemVer format
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if semverRegex.MatchString(tag) {
			matches := semverRegex.FindStringSubmatch(tag)
			major, _ := strconv.Atoi(matches[1])
			minor, _ := strconv.Atoi(matches[2])
			patch, _ := strconv.Atoi(matches[3])
			semverTags = append(semverTags, SemVer{Major: major, Minor: minor, Patch: patch})
		}
	}

	// Sort the tags by version (latest first)
	sort.Slice(semverTags, func(i, j int) bool {
		if semverTags[i].Major != semverTags[j].Major {
			return semverTags[i].Major > semverTags[j].Major
		}
		if semverTags[i].Minor != semverTags[j].Minor {
			return semverTags[i].Minor > semverTags[j].Minor
		}
		return semverTags[i].Patch > semverTags[j].Patch
	})

	if len(semverTags) == 0 {
		return nil, fmt.Errorf("no valid semver tags found")
	}

	return semverTags, nil
}

// calculateNextVersion calculates the next version based on the latest tag and command inputs
func calculateNextVersion(latestTag SemVer, majorInput, minorInput int) (SemVer, error) {
	if majorInput < latestTag.Major {
		return SemVer{}, fmt.Errorf("invalid major version: input major cannot be less than the latest major version")
	}
	if majorInput == latestTag.Major {
		if minorInput < latestTag.Minor {
			return SemVer{}, fmt.Errorf("invalid minor version: input minor cannot be less than the latest minor version")
		}
		if minorInput == latestTag.Minor {
			// Same major and minor, increment patch
			return SemVer{Major: majorInput, Minor: minorInput, Patch: latestTag.Patch + 1}, nil
		} else if minorInput == latestTag.Minor+1 {
			// Next minor version, reset patch to 0
			return SemVer{Major: majorInput, Minor: minorInput, Patch: 0}, nil
		}
		return SemVer{}, fmt.Errorf("invalid minor version: you cannot skip minor versions")
	}

	// New major version, minor should start from 0
	if majorInput == latestTag.Major+1 && minorInput == 0 {
		return SemVer{Major: majorInput, Minor: minorInput, Patch: 0}, nil
	}

	return SemVer{}, fmt.Errorf("invalid major or minor version: skipping versions is not allowed")
}

func main() {
	// Simulating inputs
	if len(os.Args) < 4 {
		log.Fatalf("Usage: %s <path> <major> <minor>", filepath.Base(os.Args[0]))
	}

	path := os.Args[1]
	majorInput, err := strconv.Atoi(os.Args[2])
	if err != nil {
		log.Fatalf("Invalid major version: %s", os.Args[2])
	}
	minorInput, err := strconv.Atoi(os.Args[3])
	if err != nil {
		log.Fatalf("Invalid minor version: %s", os.Args[3])
	}

	// Step 1: Check if the path exists
	if err := checkIfPathExists(path); err != nil {
		log.Fatal(err)
	}

	// Step 2: Check if the path is a Git repository
	if err := checkIfGitRepo(path); err != nil {
		log.Fatal(err)
	}

	// Step 3: Get the latest SemVer tag
	tags, err := getSemverTags()
	if err != nil {
		log.Fatal(err)
	}
	latestTag := tags[0]

	// Step 4: Calculate the next version based on inputs
	nextVersion, err := calculateNextVersion(latestTag, majorInput, minorInput)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Next version: %s\n", nextVersion)
}
