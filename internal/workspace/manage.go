package workspace

import (
	"fmt"
	"gatehill.io/imposter/internal/logging"
	"regexp"
)

const namePattern = "[a-zA-Z0-9_-]+"

var logger = logging.GetLogger()

func New(dir string, name string) (*Workspace, error) {
	if match, _ := regexp.MatchString("^"+namePattern+"$", name); !match {
		return nil, fmt.Errorf("workspace name does not match pattern: %s", namePattern)
	}

	m, err := createOrLoadMetadata(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to create new workspace: %s", err)
	}
	w := getWorkspace(m.Workspaces, name)
	if w != nil {
		return w, nil
	} else {
		return createWorkspace(dir, name, m)
	}
}

func Delete(dir string, name string) error {
	m, err := createOrLoadMetadata(dir)
	if err != nil {
		return fmt.Errorf("failed to delete workspace: %s", err)
	}
	w := getWorkspace(m.Workspaces, name)
	if w == nil {
		return fmt.Errorf("workspace '%s' does not exist", name)
	}
	if m.Active == name {
		m.Active = ""
	}
	var modified []*Workspace
	for _, workspace := range m.Workspaces {
		if workspace.Name != name {
			modified = append(modified, workspace)
		}
	}
	m.Workspaces = modified
	err = SaveMetadata(dir, m)
	if err != nil {
		return fmt.Errorf("failed to delete workspace: %s", err)
	}
	logger.Tracef("deleted workspace: %s", name)
	return nil
}

func SetActive(dir string, name string) (*Workspace, error) {
	m, err := createOrLoadMetadata(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to set active workspace: %s", err)
	}
	w := getWorkspace(m.Workspaces, name)
	if w == nil {
		return nil, fmt.Errorf("no such workspace: %s", name)
	}

	logger.Tracef("setting active workspace: %s", name)
	m.Active = name
	err = SaveMetadata(dir, m)
	if err != nil {
		return nil, err
	}
	return w, nil
}

func GetActive(dir string) (*Workspace, error) {
	m, _, err := GetActiveWithMetadata(dir)
	return m, err
}

func GetActiveWithMetadata(dir string) (*Workspace, *Metadata, error) {
	m, err := createOrLoadMetadata(dir)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get active workspace: %s", err)
	}
	if m.Active == "" {
		logger.Tracef("there is no active workspace")
		return nil, nil, nil
	}
	w := getWorkspace(m.Workspaces, m.Active)
	if w == nil {
		return nil, nil, fmt.Errorf("active workspace: %s does not exist", m.Active)
	}
	logger.Tracef("active workspace is: %s [%s]", w.Name, w.RemoteType)
	return w, m, nil
}

func List(dir string) ([]*Workspace, error) {
	m, err := createOrLoadMetadata(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to list workspaces: %s", err)
	}
	return m.Workspaces, nil
}

func getWorkspace(workspaces []*Workspace, name string) *Workspace {
	for _, workspace := range workspaces {
		if workspace.Name == name {
			return workspace
		}
	}
	return nil
}

func createWorkspace(dir string, name string, m *Metadata) (*Workspace, error) {
	w := &Workspace{
		Name: name,
	}
	setDefaults(m, w)
	m.Workspaces = append(m.Workspaces, w)
	err := SaveMetadata(dir, m)
	if err != nil {
		return nil, err
	}
	logger.Tracef("created new workspace: %s", name)
	return w, nil
}

func setDefaults(m *Metadata, w *Workspace) {
	w.RemoteType = "cloudmocks"
	if len(m.Workspaces) == 0 {
		m.Active = w.Name
	}
}
