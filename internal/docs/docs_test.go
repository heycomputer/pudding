package docs

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/heycomputer/pudding/internal/parser"
)

// -----------------------------------------------------------------------------
// Mocks
// -----------------------------------------------------------------------------

type CommandRunnerMock struct {
	mock.Mock
}

// Run adapts testify/mock to the CommandRunner type:
//   func(name string, args ...string) ([]byte, error)
func (m *CommandRunnerMock) Run(name string, args ...string) ([]byte, error) {
	callArgs := make([]interface{}, 0, 1+len(args))
	callArgs = append(callArgs, name)
	for _, a := range args {
		callArgs = append(callArgs, a)
	}

	ret := m.Called(callArgs...)

	var out []byte
	if b, ok := ret.Get(0).([]byte); ok {
		out = b
	}
	return out, ret.Error(1)
}

type BrowserOpenerMock struct {
	mock.Mock
}

// Open adapts testify/mock to the BrowserOpener type:
//   func(url string) error
func (m *BrowserOpenerMock) Open(url string) error {
	args := m.Called(url)
	return args.Error(0)
}

// Helper to keep call sites clean.
func callFetchAndOpen(
	dep *parser.Dependency,
	projectType parser.ProjectType,
	keywords string,
	cmd *CommandRunnerMock,
	browser *BrowserOpenerMock,
) error {
	return fetchAndOpenWithFuncs(dep, projectType, keywords, cmd.Run, browser.Open)
}

// -----------------------------------------------------------------------------
// Tests
// -----------------------------------------------------------------------------

func TestFetchAndOpen_UnsupportedProjectType(t *testing.T) {
	cmdMock := &CommandRunnerMock{}
	browserMock := &BrowserOpenerMock{}

	dep := &parser.Dependency{
		Name:    "test",
		Version: "1.0.0",
		Type:    "unknown",
	}

	err := callFetchAndOpen(dep, parser.ProjectTypeUnknown, "", cmdMock, browserMock)
	require.Error(t, err, "Expected error for unsupported project type")
	assert.Contains(t, strings.ToLower(err.Error()), "unsupported project type")

	// Should not run any commands or open browser
	assert.Len(t, cmdMock.Calls, 0, "expected no commands to be run")
	assert.Len(t, browserMock.Calls, 0, "expected browser not to be opened")
}

func TestFetchElixirDocs_Success_NoKeywords(t *testing.T) {
	cmdMock := &CommandRunnerMock{}
	browserMock := &BrowserOpenerMock{}

	dep := &parser.Dependency{
		Name:    "phoenix",
		Version: "1.7.0",
		Type:    "elixir",
	}

	// Simulate "mix hex.docs fetch phoenix 1.7.0" output with a path.
	docPath := "/tmp/phoenix-docs"
	cmdMock.
		On("Run",
			"mix",
			"hex.docs", "fetch", "phoenix", "1.7.0",
		).
		Return([]byte("Docs fetched: "+docPath+"\n"), nil).
		Once()

	expectedURL := "file://" + docPath + "/index.html"

	browserMock.
		On("Open", expectedURL).
		Return(nil).
		Once()

	err := callFetchAndOpen(dep, parser.ProjectTypeElixir, "", cmdMock, browserMock)
	require.NoError(t, err)

	cmdMock.AssertExpectations(t)
	browserMock.AssertExpectations(t)
}

func TestFetchElixirDocs_Success_WithKeywords(t *testing.T) {
	cmdMock := &CommandRunnerMock{}
	browserMock := &BrowserOpenerMock{}

	dep := &parser.Dependency{
		Name:    "phoenix",
		Version: "1.7.0",
		Type:    "elixir",
	}

	keywords := "live view"
	docPath := "/tmp/phoenix-docs"
	cmdMock.
		On("Run",
			"mix",
			"hex.docs", "fetch", "phoenix", "1.7.0",
		).
		Return([]byte("Docs fetched: "+docPath+"\n"), nil).
		Once()

	// search.html?q=live+view
	expectedURL := "file://" + docPath + "/search.html?q=live+view"

	browserMock.
		On("Open", expectedURL).
		Return(nil).
		Once()

	err := callFetchAndOpen(dep, parser.ProjectTypeElixir, keywords, cmdMock, browserMock)
	require.NoError(t, err)

	cmdMock.AssertExpectations(t)
	browserMock.AssertExpectations(t)
}

func TestFetchElixirDocs_CommandFailure(t *testing.T) {
	cmdMock := &CommandRunnerMock{}
	browserMock := &BrowserOpenerMock{}

	dep := &parser.Dependency{
		Name:    "phoenix",
		Version: "1.7.0",
		Type:    "elixir",
	}

	cmdErr := errors.New("mock command error")

	cmdMock.
		On("Run",
			"mix",
			"hex.docs", "fetch", "phoenix", "1.7.0",
		).
		Return([]byte(nil), cmdErr).
		Once()

	err := callFetchAndOpen(dep, parser.ProjectTypeElixir, "", cmdMock, browserMock)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to fetch docs for phoenix")

	cmdMock.AssertExpectations(t)
	browserMock.AssertNotCalled(t, "Open", mock.Anything)
}

func TestFetchElixirDocs_InvalidOutput_NoPath(t *testing.T) {
	cmdMock := &CommandRunnerMock{}
	browserMock := &BrowserOpenerMock{}

	dep := &parser.Dependency{
		Name:    "phoenix",
		Version: "1.7.0",
		Type:    "elixir",
	}

	// No "/" in output => extractDocPath returns empty and should error.
	cmdMock.
		On("Run",
			"mix",
			"hex.docs", "fetch", "phoenix", "1.7.0",
		).
		Return([]byte("no docs path here"), nil).
		Once()

	err := callFetchAndOpen(dep, parser.ProjectTypeElixir, "", cmdMock, browserMock)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to extract docs path for phoenix")

	cmdMock.AssertExpectations(t)
	browserMock.AssertNotCalled(t, "Open", mock.Anything)
}

func TestFetchRubyDocs_Success_NoKeywords(t *testing.T) {
	cmdMock := &CommandRunnerMock{}
	browserMock := &BrowserOpenerMock{}

	dep := &parser.Dependency{
		Name:    "rails",
		Version: "7.0.0",
		Type:    "gem",
	}

	gemHome := "/home/user/.gem"

	// 1. rdoc rails --rdoc --version 7.0.0
	cmdMock.
		On("Run",
			"rdoc", "rails", "--rdoc", "--version", "7.0.0",
		).
		Return([]byte("rdoc ok"), nil).
		Once()

	// 2. sh -c "gem env home"
	cmdMock.
		On("Run",
			"sh", "-c", "gem env home",
		).
		Return([]byte(gemHome+"\n"), nil).
		Once()

	expectedURL := "file://" + gemHome + "/doc/rails-7.0.0/rdoc/table_of_contents.html"

	browserMock.
		On("Open", expectedURL).
		Return(nil).
		Once()

	err := callFetchAndOpen(dep, parser.ProjectTypeRuby, "", cmdMock, browserMock)
	require.NoError(t, err)

	cmdMock.AssertExpectations(t)
	browserMock.AssertExpectations(t)
}

func TestFetchRubyDocs_Success_WithKeywords(t *testing.T) {
	cmdMock := &CommandRunnerMock{}
	browserMock := &BrowserOpenerMock{}

	dep := &parser.Dependency{
		Name:    "rails",
		Version: "7.0.0",
		Type:    "gem",
	}

	keywords := "active record"
	gemHome := "/home/user/.gem"

	cmdMock.
		On("Run",
			"rdoc", "rails", "--rdoc", "--version", "7.0.0",
		).
		Return([]byte("rdoc ok"), nil).
		Once()

	cmdMock.
		On("Run",
			"sh", "-c", "gem env home",
		).
		Return([]byte(gemHome+"\n"), nil).
		Once()

	expectedURL := "file://" + gemHome + "/doc/rails-7.0.0/rdoc/index.html?q=active+record"

	browserMock.
		On("Open", expectedURL).
		Return(nil).
		Once()

	err := callFetchAndOpen(dep, parser.ProjectTypeRuby, keywords, cmdMock, browserMock)
	require.NoError(t, err)

	cmdMock.AssertExpectations(t)
	browserMock.AssertExpectations(t)
}

func TestFetchRubyDocs_RdocFailure(t *testing.T) {
	cmdMock := &CommandRunnerMock{}
	browserMock := &BrowserOpenerMock{}

	dep := &parser.Dependency{
		Name:    "rails",
		Version: "7.0.0",
		Type:    "gem",
	}

	cmdErr := errors.New("mock rdoc error")

	cmdMock.
		On("Run",
			"rdoc", "rails", "--rdoc", "--version", "7.0.0",
		).
		Return([]byte(nil), cmdErr).
		Once()

	err := callFetchAndOpen(dep, parser.ProjectTypeRuby, "", cmdMock, browserMock)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to generate rdoc for rails")

	cmdMock.AssertExpectations(t)
	browserMock.AssertNotCalled(t, "Open", mock.Anything)
}

func TestFetchRubyDocs_GemEnvFailure(t *testing.T) {
	cmdMock := &CommandRunnerMock{}
	browserMock := &BrowserOpenerMock{}

	dep := &parser.Dependency{
		Name:    "rails",
		Version: "7.0.0",
		Type:    "gem",
	}

	cmdMock.
		On("Run",
			"rdoc", "rails", "--rdoc", "--version", "7.0.0",
		).
		Return([]byte("rdoc ok"), nil).
		Once()

	envErr := errors.New("mock gem env error")

	cmdMock.
		On("Run",
			"sh", "-c", "gem env home",
		).
		Return([]byte(nil), envErr).
		Once()

	err := callFetchAndOpen(dep, parser.ProjectTypeRuby, "", cmdMock, browserMock)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get gem env home")

	cmdMock.AssertExpectations(t)
	browserMock.AssertNotCalled(t, "Open", mock.Anything)
}

func TestFetchRubyDocs_BrowserFailure(t *testing.T) {
	cmdMock := &CommandRunnerMock{}
	browserMock := &BrowserOpenerMock{}

	dep := &parser.Dependency{
		Name:    "rails",
		Version: "7.0.0",
		Type:    "gem",
	}

	gemHome := "/home/user/.gem"

	cmdMock.
		On("Run",
			"rdoc", "rails", "--rdoc", "--version", "7.0.0",
		).
		Return([]byte("rdoc ok"), nil).
		Once()

	cmdMock.
		On("Run",
			"sh", "-c", "gem env home",
		).
		Return([]byte(gemHome+"\n"), nil).
		Once()

	expectedURL := "file://" + gemHome + "/doc/rails-7.0.0/rdoc/table_of_contents.html"

	browserErr := errors.New("mock browser error")

	browserMock.
		On("Open", expectedURL).
		Return(browserErr).
		Once()

	err := callFetchAndOpen(dep, parser.ProjectTypeRuby, "", cmdMock, browserMock)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to open rdoc for rails")

	cmdMock.AssertExpectations(t)
	browserMock.AssertExpectations(t)
}
