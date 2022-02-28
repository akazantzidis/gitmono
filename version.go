package gitmono

import (
	"fmt"
	"strings"

	"github.com/hashicorp/go-version"
)

type Versioner struct {
	mono *GitMono
}

type VersionedCommit struct {
	CommitID      string
	Version       *version.Version
	VersionPrefix string
	Project       string
}

func (vc *VersionedCommit) GetTag() string {
	var projectPrefix string
	if vc.Project != "." {
		projectPrefix = fmt.Sprintf("%s/", vc.Project)
	}
	return fmt.Sprintf("%s%s%s", projectPrefix, vc.VersionPrefix, vc.Version.String())
}

func NewVersioner(mono *GitMono) *Versioner {
	return &Versioner{
		mono: mono,
	}
}

func (v *Versioner) CurrentVersion() (*VersionedCommit, error) {
	if len(v.mono.projects) != 1 {
		return nil, fmt.Errorf("expected single project")
	}
	givenProject := v.mono.projects[0]

	tagger := &Tagger{mono: v.mono}
	tags, err := tagger.Tags()
	if err != nil {
		return nil, err
	}

	for _, tag := range tags {
		project, version := v.parseProjectVersion(tag)
		if givenProject != "." && !strings.EqualFold(project, givenProject) {
			continue
		}

		logger := &Logger{v.mono}
		commitHash, err := logger.CommitHashByRevision(tag)
		if err != nil {
			return nil, err
		}

		parsedVersion, err := v.parseVersion(version)
		if err != nil {
			return nil, err
		}

		currentVersion := VersionedCommit{
			Version:       parsedVersion,
			VersionPrefix: v.mono.versionPrefix,
			Project:       project,
			CommitID:      commitHash,
		}

		return &currentVersion, nil
	}

	return nil, nil
}

func (v *Versioner) parseProjectVersion(tag string) (string, string) {
	idx := strings.LastIndex(tag, "/")
	if idx == -1 {
		return ".", tag
	}

	return tag[0:idx], tag[idx+1:]
}

func (v *Versioner) parseVersion(vv string) (*version.Version, error) {
	var (
		versionPrefix = v.mono.versionPrefix
		versionValue  = vv
	)
	if versionPrefix != "" && strings.HasPrefix(vv, versionPrefix) {
		versionValue = vv[len(versionPrefix):]
	}

	parsedVersion, err := version.NewSemver(versionValue)
	if err != nil {
		return nil, err
	}

	return parsedVersion, nil
}

func (v *Versioner) NewVersion() (*VersionedCommit, error) {
	currentVersion, err := v.CurrentVersion()
	if err != nil {
		return nil, err
	}
	if currentVersion == nil {
		return nil, nil
	}

	logger := NewLogger(v.mono)
	newCommits, err := logger.Log(currentVersion.CommitID, "HEAD")
	if err != nil {
		return nil, err
	}
	if len(newCommits) == 0 {
		return nil, fmt.Errorf("no new commits were found")
	}

	var (
		commitParser = CommitParser{scheme: v.mono.commitScheme}
		bump         bumper
	)
	for _, cm := range newCommits {
		bump = commitParser.parseCommit(cm)
		if bump != nil {
			break
		}
	}

	if bump == nil {
		bump = patchBumper
	}
	newVersion, err := bump.bump(currentVersion.Version)
	if err != nil {
		return nil, err
	}

	newVersionedCommit := VersionedCommit{
		CommitID:      "HEAD",
		Version:       newVersion,
		VersionPrefix: currentVersion.VersionPrefix,
		Project:       currentVersion.Project,
	}

	if !v.mono.dryRun {
		tagger := &Tagger{mono: v.mono}

		err := tagger.WriteTag(&newVersionedCommit)
		if err != nil {
			return nil, err
		}
	}

	return &newVersionedCommit, nil
}

func (v *Versioner) InitVersion() ([]*VersionedCommit, error) {
	projectsMap := make(map[string]struct{}, len(v.mono.projects))
	for _, project := range v.mono.projects {
		projectsMap[project] = struct{}{}
	}

	tagger := &Tagger{mono: v.mono}
	tags, err := tagger.Tags()
	if err != nil {
		return nil, err
	}

	for _, tag := range tags {
		project, _ := v.parseProjectVersion(tag)
		delete(projectsMap, project)
	}

	initVersion, _ := version.NewSemver("0.1.0")
	newVersionedCommits := make([]*VersionedCommit, 0, len(projectsMap))

	for project := range projectsMap {
		newVersionedCommit := VersionedCommit{
			CommitID:      "HEAD",
			Project:       project,
			Version:       initVersion,
			VersionPrefix: v.mono.versionPrefix,
		}
		newVersionedCommits = append(newVersionedCommits, &newVersionedCommit)

		if !v.mono.dryRun {
			err := tagger.WriteTag(&newVersionedCommit)
			if err != nil {
				return nil, err
			}
		}
	}

	return newVersionedCommits, nil
}
