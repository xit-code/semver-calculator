package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
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

func main() {
	// Parse command-line flags
	path := flag.String("path", "", "Path to the Git repository")
	major := flag.Int("major", -1, "Major version number")
	minor := flag.Int("minor", -1, "Minor version number")
	flag.Parse()

	// Validate inputs
	if *path == "" || *major == -1 || *minor == -1 {
		log.Fatal("All parameters (--path, --major, --minor) must be provided")
	}

	if err := run(*path, *major, *minor); err != nil {
		log.Fatal(err)
	}
}

func run(path string, majorInput, minorInput int) error {
	// Step 1: Check if the path exists
	if err := checkIfPathExists(path); err != nil {
		return err
	}

	// Step 2: Check if the path is a Git repository
	if err := checkIfGitRepo(path); err != nil {
		return err
	}

	// Step 3: Get the latest SemVer tag
	tags, err := getSemverTags()
	if err != nil {
		return err
	}
	latestTag := tags[0]

	// Step 4: Calculate the next version based on inputs
	nextVersion, err := calculateNextVersion(latestTag, majorInput, minorInput)
	if err != nil {
		return err
	}

	fmt.Print(nextVersion)
	return nil
}

func checkIfPathExists(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("path %s does not exist", path)
	}
	return nil
}

func checkIfGitRepo(path string) error {
	if err := os.Chdir(path); err != nil {
		return fmt.Errorf("failed to change directory to %s: %w", path, err)
	}

	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	output, err := cmd.CombinedOutput()
	if err != nil || strings.TrimSpace(string(output)) != "true" {
		return fmt.Errorf("path %s is not a Git repository", path)
	}
	return nil
}

func getSemverTags() ([]SemVer, error) {
	cmd := exec.Command("git", "tag", "--list")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to get tags: %w", err)
	}

	semverRegex := regexp.MustCompile(`^v(\d+)\.(\d+)\.(\d+)$`)
	tags := strings.Split(string(output), "\n")
	var semverTags []SemVer

	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			continue
		}
		if matches := semverRegex.FindStringSubmatch(tag); matches != nil {
			major, _ := strconv.Atoi(matches[1])
			minor, _ := strconv.Atoi(matches[2])
			patch, _ := strconv.Atoi(matches[3])
			semverTags = append(semverTags, SemVer{Major: major, Minor: minor, Patch: patch})
		}
	}

	if len(semverTags) == 0 {
		// No existing semver tags found; start from v0.0.0
		semverTags = append(semverTags, SemVer{Major: 0, Minor: 0, Patch: 0})
	}

	sort.Slice(semverTags, func(i, j int) bool {
		if semverTags[i].Major != semverTags[j].Major {
			return semverTags[i].Major > semverTags[j].Major
		}
		if semverTags[i].Minor != semverTags[j].Minor {
			return semverTags[i].Minor > semverTags[j].Minor
		}
		return semverTags[i].Patch > semverTags[j].Patch
	})

	return semverTags, nil
}

func calculateNextVersion(latestTag SemVer, majorInput, minorInput int) (SemVer, error) {
	if majorInput < latestTag.Major {
		return SemVer{}, fmt.Errorf("invalid major version: input major (%d) cannot be less than the latest major version (%d)", majorInput, latestTag.Major)
	}
	if majorInput == latestTag.Major {
		if minorInput < latestTag.Minor {
			return SemVer{}, fmt.Errorf("invalid minor version: input minor (%d) cannot be less than the latest minor version (%d)", minorInput, latestTag.Minor)
		}
		if minorInput == latestTag.Minor {
			return SemVer{Major: majorInput, Minor: minorInput, Patch: latestTag.Patch + 1}, nil
		} else if minorInput == latestTag.Minor+1 {
			return SemVer{Major: majorInput, Minor: minorInput, Patch: 0}, nil
		}
		return SemVer{}, fmt.Errorf("invalid minor version: you cannot skip minor versions (latest: %d, input: %d)", latestTag.Minor, minorInput)
	}

	if majorInput == latestTag.Major+1 && minorInput == 0 {
		return SemVer{Major: majorInput, Minor: minorInput, Patch: 0}, nil
	}

	return SemVer{}, fmt.Errorf("invalid version: skipping versions is not allowed (latest: %s, input: v%d.%d.x)", latestTag, majorInput, minorInput)
}
