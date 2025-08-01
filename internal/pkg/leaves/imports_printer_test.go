package leaves

import (
	"bytes"
	"testing"
	"time"

	"github.com/dmytrogajewski/hercules/internal/app/core"
	"github.com/dmytrogajewski/hercules/internal/pkg/importmodel"
	"github.com/dmytrogajewski/hercules/internal/pkg/plumbing"
	"github.com/dmytrogajewski/hercules/internal/pkg/plumbing/identity"
	"github.com/dmytrogajewski/hercules/internal/pkg/plumbing/imports"
	"github.com/dmytrogajewski/hercules/internal/pkg/test"
	gitplumbing "github.com/go-git/go-git/v6/plumbing"
	"github.com/stretchr/testify/assert"
)

func fixtureImportsPerDev() *ImportsPerDeveloper {
	d := ImportsPerDeveloper{}
	d.Initialize(test.Repository)
	people := [...]string{"one@srcd", "two@srcd"}
	d.reversedPeopleDict = people[:]
	return &d
}

func TestImportsPerDeveloperMeta(t *testing.T) {
	ipd := fixtureImportsPerDev()
	ass := assert.New(t)
	ass.Equal(ipd.Name(), "ImportsPerDeveloper")
	ass.Equal(len(ipd.Provides()), 0)
	ass.Equal(len(ipd.Requires()), 3)
	ass.Equal(ipd.Requires()[0], imports.DependencyImports)
	ass.Equal(ipd.Requires()[1], identity.DependencyAuthor)
	ass.Equal(ipd.Requires()[2], plumbing.DependencyTick)
	ass.Equal(ipd.Flag(), "imports-per-dev")
	assert.Len(t, ipd.ListConfigurationOptions(), 0)
	assert.True(t, len(ipd.Description()) > 0)
	logger := core.GetLogger()
	assert.NoError(t, ipd.Configure(map[string]interface{}{
		core.ConfigLogger: logger,
		identity.FactIdentityDetectorReversedPeopleDict: []string{"1", "2"},
		plumbing.FactTickSize:                           time.Hour,
	}))
	ass.Equal(logger, ipd.l)
	ass.Equal([]string{"1", "2"}, ipd.reversedPeopleDict)
	ass.Equal(time.Hour, ipd.TickSize)
}

func TestImportsPerDeveloperRegistration(t *testing.T) {
	summoned := core.Registry.Summon((&ImportsPerDeveloper{}).Name())
	assert.Len(t, summoned, 1)
	assert.Equal(t, summoned[0].Name(), "ImportsPerDeveloper")
	leaves := core.Registry.GetLeaves()
	matched := false
	for _, tp := range leaves {
		if tp.Flag() == (&ImportsPerDeveloper{}).Flag() {
			matched = true
			break
		}
	}
	assert.True(t, matched)
}

func TestImportsPerDeveloperInitialize(t *testing.T) {
	ipd := fixtureImportsPerDev()
	assert.NotNil(t, ipd.imports)
	assert.Equal(t, time.Hour*24, ipd.TickSize)
}

func TestImportsPerDeveloperConsumeFinalize(t *testing.T) {
	deps := map[string]interface{}{}
	deps[core.DependencyIsMerge] = false
	deps[identity.DependencyAuthor] = 0
	deps[plumbing.DependencyTick] = 1
	imps := map[gitplumbing.Hash]importmodel.File{}
	imps[gitplumbing.NewHash("291286b4ac41952cbd1389fda66420ec03c1a9fe")] =
		importmodel.File{Lang: "Go", Imports: []string{"sys"}}
	imps[gitplumbing.NewHash("c29112dbd697ad9b401333b80c18a63951bc18d9")] =
		importmodel.File{Lang: "Python", Imports: []string{"sys"}}
	deps[imports.DependencyImports] = imps
	ipd := fixtureImportsPerDev()
	ipd.reversedPeopleDict = []string{"1", "2"}
	_, err := ipd.Consume(deps)
	assert.NoError(t, err)
	assert.Equal(t, ImportsMap{
		0: {"Go": {"sys": {1: 1}}, "Python": {"sys": {1: 1}}},
	}, ipd.imports)
	_, err = ipd.Consume(deps)
	assert.NoError(t, err)
	assert.Equal(t, ImportsMap{
		0: {"Go": {"sys": {1: 2}}, "Python": {"sys": {1: 2}}},
	}, ipd.imports)
	deps[identity.DependencyAuthor] = 1
	_, err = ipd.Consume(deps)
	assert.NoError(t, err)
	assert.Equal(t, ImportsMap{
		0: {"Go": {"sys": {1: 2}}, "Python": {"sys": {1: 2}}},
		1: {"Go": {"sys": {1: 1}}, "Python": {"sys": {1: 1}}},
	}, ipd.imports)
	deps[core.DependencyIsMerge] = true
	_, err = ipd.Consume(deps)
	assert.NoError(t, err)
	assert.Equal(t, ImportsMap{
		0: {"Go": {"sys": {1: 2}}, "Python": {"sys": {1: 2}}},
		1: {"Go": {"sys": {1: 1}}, "Python": {"sys": {1: 1}}},
	}, ipd.imports)
	result := ipd.Finalize().(ImportsPerDeveloperResult)
	assert.Equal(t, ipd.reversedPeopleDict, result.reversedPeopleDict)
	assert.Equal(t, ipd.imports, result.Imports)
}

func TestImportsPerDeveloperSerializeText(t *testing.T) {
	ipd := fixtureImportsPerDev()
	res := ImportsPerDeveloperResult{Imports: ImportsMap{
		0: {"Go": {"sys": {1: 2}}, "Python": {"sys": {1: 2}}},
		1: {"Go": {"sys": {1: 1}}, "Python": {"sys": {1: 1}}},
	}, reversedPeopleDict: []string{"one", "two"}}
	buffer := &bytes.Buffer{}
	assert.NoError(t, ipd.Serialize(res, false, buffer))
	assert.Equal(t, `  tick_size: 0
  imports:
    "one": {"Go":{"sys":{"1":2}},"Python":{"sys":{"1":2}}}
    "two": {"Go":{"sys":{"1":1}},"Python":{"sys":{"1":1}}}
`, buffer.String())
}

func TestImportsPerDeveloperSerializeBinary(t *testing.T) {
	ipd := fixtureImportsPerDev()
	ass := assert.New(t)
	res := ImportsPerDeveloperResult{Imports: ImportsMap{
		0: {"Go": {"sys": {1: 2}}, "Python": {"sys": {1: 2}}},
		1: {"Go": {"sys": {1: 1}}, "Python": {"sys": {1: 1}}},
	}, reversedPeopleDict: []string{"one", "two"}}
	buffer := &bytes.Buffer{}
	ass.NoError(ipd.Serialize(res, true, buffer))
	back, err := ipd.Deserialize(buffer.Bytes())
	ass.NoError(err)
	ass.Equal(res, back)
}
