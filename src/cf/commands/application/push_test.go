package application_test

import (
	"cf"
	"cf/api"
	. "cf/commands/application"
	"cf/configuration"
	"github.com/stretchr/testify/assert"
	"os"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
	"testing"
)

func TestPushingRequirements(t *testing.T) {
	starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo := getPushDependencies()
	fakeUI := new(testterm.FakeUI)
	config := &configuration.Configuration{}
	cmd := NewPush(fakeUI, config, starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo)
	ctxt := testcmd.NewContext("push", []string{})

	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true}
	assert.True(t, testcmd.CommandDidPassRequirements)

	reqFactory = &testreq.FakeReqFactory{LoginSuccess: false, TargetedSpaceSuccess: true}
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	assert.False(t, testcmd.CommandDidPassRequirements)

	testcmd.CommandDidPassRequirements = true

	reqFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: false}
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	assert.False(t, testcmd.CommandDidPassRequirements)
}

func TestPushingAppWhenItDoesNotExist(t *testing.T) {
	starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo := getPushDependencies()
	domain_Auto := cf.Domain{}
	domain_Auto.Name = "foo.cf-app.com"
	domain_Auto.Guid = "foo-domain-guid"
	domains := []cf.Domain{domain_Auto}

	domainRepo.DefaultAppDomain = domains[0]
	routeRepo.FindByHostAndDomainErr = true
	appRepo.FindByNameNotFound = true

	fakeUI := callPush(t, []string{"my-new-app"}, starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo)

	assert.Contains(t, fakeUI.Outputs[0], "Creating app")
	assert.Contains(t, fakeUI.Outputs[0], "my-new-app")
	assert.Contains(t, fakeUI.Outputs[0], "my-org")
	assert.Contains(t, fakeUI.Outputs[0], "my-space")
	assert.Contains(t, fakeUI.Outputs[0], "my-user")
	assert.Equal(t, appRepo.CreatedApp.Name, "my-new-app")
	assert.Equal(t, appRepo.CreatedApp.InstanceCount, 1)
	assert.Equal(t, appRepo.CreatedApp.Memory, uint64(128))
	assert.Equal(t, appRepo.CreatedApp.BuildpackUrl, "")
	assert.Contains(t, fakeUI.Outputs[1], "OK")

	assert.Contains(t, fakeUI.Outputs[3], "my-new-app.foo.cf-app.com")
	assert.Equal(t, routeRepo.FindByHostAndDomainHost, "my-new-app")
	assert.Equal(t, routeRepo.CreatedHost, "my-new-app")
	assert.Equal(t, routeRepo.CreatedDomainGuid, "foo-domain-guid")
	assert.Contains(t, fakeUI.Outputs[4], "OK")

	assert.Contains(t, fakeUI.Outputs[6], "my-new-app.foo.cf-app.com")
	assert.Equal(t, routeRepo.BoundAppGuid, "my-new-app")
	assert.Equal(t, routeRepo.BoundRouteGuid, "my-new-app")
	assert.Contains(t, fakeUI.Outputs[7], "OK")

	expectedAppDir, err := os.Getwd()
	assert.NoError(t, err)

	assert.Contains(t, fakeUI.Outputs[9], "my-new-app")
	assert.Equal(t, appBitsRepo.UploadedAppGuid, "my-new-app-guid")
	assert.Equal(t, appBitsRepo.UploadedDir, expectedAppDir)
	assert.Contains(t, fakeUI.Outputs[10], "OK")

	assert.Equal(t, stopper.AppToStop.Name, "my-new-app")
	assert.Equal(t, starter.AppToStart.Name, "my-stopped-app")
}

func TestPushingAppWhenItDoesNotExistButRouteExists(t *testing.T) {
	starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo := getPushDependencies()

	domain_Auto1 := cf.Domain{}
	domain_Auto1.Name = "foo.cf-app.com"
	domain_Auto1.Guid = "foo-domain-guid"

	domain_Auto2 := cf.DomainFields{}
	domain_Auto2.Name = "foo.cf-app.com"
	domain_Auto2.Guid = "foo-domain-guid"

	route := cf.Route{}
	route.Guid = "my-route-guid"
	route.Host = "my-new-app"
	route.Domain = domain_Auto2

	domainRepo.DefaultAppDomain = domain_Auto1
	routeRepo.FindByHostAndDomainRoute = route
	appRepo.FindByNameNotFound = true

	fakeUI := callPush(t, []string{"my-new-app"}, starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo)

	assert.Empty(t, routeRepo.CreatedHost)
	assert.Empty(t, routeRepo.CreatedDomainGuid)
	assert.Contains(t, fakeUI.Outputs[3], "my-new-app.foo.cf-app.com")
	assert.Equal(t, routeRepo.FindByHostAndDomainHost, "my-new-app")

	assert.Contains(t, fakeUI.Outputs[4], "my-new-app.foo.cf-app.com")
	assert.Equal(t, routeRepo.BoundAppGuid, "my-new-app-guid")
	assert.Equal(t, routeRepo.BoundRouteGuid, "my-route-guid")
	assert.Contains(t, fakeUI.Outputs[5], "OK")
}

func TestPushingAppWithCustomFlags(t *testing.T) {
	starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo := getPushDependencies()
	domain := cf.Domain{}
	domain.Name = "bar.cf-app.com"
	domain.Guid = "bar-domain-guid"
	stack := cf.Stack{}
	stack.Name = "customLinux"
	stack.Guid = "custom-linux-guid"

	domainRepo.FindByNameDomain = domain
	routeRepo.FindByHostAndDomainErr = true
	stackRepo.FindByNameStack = stack
	appRepo.FindByNameNotFound = true

	fakeUI := callPush(t, []string{
		"-c", "unicorn -c config/unicorn.rb -D",
		"-d", "bar.cf-app.com",
		"-n", "my-hostname",
		"-i", "3",
		"-m", "2G",
		"-b", "https://github.com/heroku/heroku-buildpack-play.git",
		"-p", "/Users/pivotal/workspace/my-new-app",
		"-s", "customLinux",
		"--no-start",
		"my-new-app",
	}, starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo)

	assert.Contains(t, fakeUI.Outputs[0], "customLinux")
	assert.Equal(t, stackRepo.FindByNameName, "customLinux")

	assert.Contains(t, fakeUI.Outputs[1], "my-new-app")
	assert.Equal(t, appRepo.CreatedApp.Name, "my-new-app")
	assert.Equal(t, appRepo.CreatedApp.Command, "unicorn -c config/unicorn.rb -D")
	assert.Equal(t, appRepo.CreatedApp.InstanceCount, 3)
	assert.Equal(t, appRepo.CreatedApp.Memory, uint64(2048))
	assert.Equal(t, appRepo.CreatedApp.Stack.Guid, "custom-linux-guid")
	assert.Equal(t, appRepo.CreatedApp.BuildpackUrl, "https://github.com/heroku/heroku-buildpack-play.git")
	assert.Contains(t, fakeUI.Outputs[2], "OK")

	assert.Contains(t, fakeUI.Outputs[4], "my-hostname.bar.cf-app.com")
	assert.Equal(t, domainRepo.FindByNameInCurrentSpaceName, "bar.cf-app.com")
	assert.Equal(t, routeRepo.CreatedHost, "my-hostname")
	assert.Equal(t, routeRepo.CreatedDomainGuid, "bar-domain-guid")
	assert.Contains(t, fakeUI.Outputs[5], "OK")

	assert.Contains(t, fakeUI.Outputs[7], "my-hostname.bar.cf-app.com")
	assert.Contains(t, fakeUI.Outputs[7], "my-new-app")
	assert.Equal(t, routeRepo.BoundAppGuid, "my-new-app-guid")
	assert.Equal(t, routeRepo.BoundRouteGuid, "my-hostname")
	assert.Contains(t, fakeUI.Outputs[8], "OK")

	assert.Contains(t, fakeUI.Outputs[10], "my-new-app")
	assert.Equal(t, appBitsRepo.UploadedAppGuid, "my-new-app-guid")
	assert.Equal(t, appBitsRepo.UploadedDir, "/Users/pivotal/workspace/my-new-app")
	assert.Contains(t, fakeUI.Outputs[11], "OK")

	assert.Equal(t, starter.AppToStart.Name, "")
}

func TestPushingAppWithNoRoute(t *testing.T) {
	starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo := getPushDependencies()
	domain := cf.Domain{}
	domain.Name = "bar.cf-app.com"
	domain.Guid = "bar-domain-guid"
	stack := cf.Stack{}
	stack.Name = "customLinux"
	stack.Guid = "custom-linux-guid"

	domainRepo.FindByNameDomain = domain
	routeRepo.FindByHostErr = true
	stackRepo.FindByNameStack = stack
	appRepo.FindByNameNotFound = true

	callPush(t, []string{
		"--no-route",
		"my-new-app",
	}, starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo)

	assert.Equal(t, appRepo.CreatedApp.Name, "my-new-app")
	assert.Equal(t, routeRepo.CreatedHost, "")
	assert.Equal(t, routeRepo.CreatedDomainGuid, "")
}

func TestPushingAppWithNoHostname(t *testing.T) {
	starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo := getPushDependencies()
	domain := cf.Domain{}
	domain.Name = "bar.cf-app.com"
	domain.Guid = "bar-domain-guid"
	stack := cf.Stack{}
	stack.Name = "customLinux"
	stack.Guid = "custom-linux-guid"

	domainRepo.DefaultAppDomain = domain
	routeRepo.FindByHostAndDomainErr = true
	stackRepo.FindByNameStack = stack
	appRepo.FindByNameNotFound = true

	callPush(t, []string{
		"--no-hostname",
		"my-new-app",
	}, starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo)

	assert.Equal(t, appRepo.CreatedApp.Name, "my-new-app")
	assert.Equal(t, routeRepo.CreatedHost, "")
	assert.Equal(t, routeRepo.CreatedDomainGuid, "bar-domain-guid")
}

func TestPushingAppWithMemoryInMegaBytes(t *testing.T) {
	starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo := getPushDependencies()
	domain := cf.Domain{}
	domain.Name = "bar.cf-app.com"
	domain.Guid = "bar-domain-guid"
	domainRepo.FindByNameDomain = domain
	appRepo.FindByNameNotFound = true

	callPush(t, []string{
		"-m", "256M",
		"my-new-app",
	}, starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo)

	assert.Equal(t, appRepo.CreatedApp.Memory, uint64(256))
}

func TestPushingAppWithMemoryWithoutUnit(t *testing.T) {
	starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo := getPushDependencies()
	domain := cf.Domain{}
	domain.Name = "bar.cf-app.com"
	domain.Guid = "bar-domain-guid"
	domainRepo.FindByNameDomain = domain
	appRepo.FindByNameNotFound = true

	callPush(t, []string{
		"-m", "512",
		"my-new-app",
	}, starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo)

	assert.Equal(t, appRepo.CreatedApp.Memory, uint64(512))
}

func TestPushingAppWithInvalidMemory(t *testing.T) {
	starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo := getPushDependencies()
	domain := cf.Domain{}
	domain.Name = "bar.cf-app.com"
	domain.Guid = "bar-domain-guid"
	domainRepo.FindByNameDomain = domain
	appRepo.FindByNameNotFound = true

	callPush(t, []string{
		"-m", "abcM",
		"my-new-app",
	}, starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo)

	assert.Equal(t, appRepo.CreatedApp.Memory, uint64(128))
}

func TestPushingAppWhenItAlreadyExistsAndNothingIsSpecified(t *testing.T) {
	starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo := getPushDependencies()

	existingRoute := cf.RouteSummary{}
	existingRoute.Host = "existing-app"

	existingApp := cf.Application{}
	existingApp.Name = "existing-app"
	existingApp.Guid = "existing-app-guid"
	existingApp.Routes = []cf.RouteSummary{existingRoute}

	appRepo.FindByNameApp = existingApp

	domain_Auto := cf.DomainFields{}
	domain_Auto.Name = "example.com"

	foundRoute := cf.Route{}
	foundRoute.RouteFields = existingRoute.RouteFields
	foundRoute.Domain = domain_Auto

	routeRepo.FindByHostAndDomainRoute = foundRoute
	fakeUI := callPush(t, []string{"existing-app"}, starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo)

	assert.Equal(t, stopper.AppToStop.Name, "existing-app")
	assert.Contains(t, fakeUI.Outputs[0], "Using route")
	assert.Contains(t, fakeUI.Outputs[0], "existing-app.example.com")
	assert.Equal(t, appBitsRepo.UploadedAppGuid, "existing-app-guid")
}

func TestPushingAppWhenItAlreadyExistsAndDomainIsSpecifiedIsAlreadyBound(t *testing.T) {
	starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo := getPushDependencies()

	existingRoute := cf.RouteSummary{}
	existingRoute.Host = "existing-app"

	existingApp := cf.Application{}
	existingApp.Name = "existing-app"
	existingApp.Guid = "existing-app-guid"
	existingApp.Routes = []cf.RouteSummary{existingRoute}

	domain_Auto := cf.DomainFields{}
	domain_Auto.Name = "example.com"

	foundRoute := cf.Route{}
	foundRoute.RouteFields = existingRoute.RouteFields
	foundRoute.Domain = domain_Auto

	appRepo.FindByNameApp = existingApp
	routeRepo.FindByHostAndDomainRoute = foundRoute

	fakeUI := callPush(t, []string{"-d", "example.com", "existing-app"}, starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo)

	assert.Contains(t, fakeUI.Outputs[0], "Using route")
	assert.Contains(t, fakeUI.Outputs[0], "existing-app")
	assert.Equal(t, appBitsRepo.UploadedAppGuid, "existing-app-guid")
}

func TestPushingAppWhenItAlreadyExistsAndDomainSpecifiedIsNotBound(t *testing.T) {
	starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo := getPushDependencies()

	domain_Auto := cf.DomainFields{}
	domain_Auto.Name = "example.com"

	existingRoute := cf.RouteSummary{}
	existingRoute.Host = "existing-app"
	existingRoute.Domain = domain_Auto

	existingApp := cf.Application{}
	existingApp.Name = "existing-app"
	existingApp.Guid = "existing-app-guid"
	existingApp.Routes = []cf.RouteSummary{existingRoute}

	domain := cf.Domain{}
	domain.Guid = "domain-guid"
	domain.Name = "newdomain.com"

	appRepo.FindByNameApp = existingApp
	routeRepo.FindByHostAndDomainNotFound = true
	domainRepo.FindByNameDomain = domain

	fakeUI := callPush(t, []string{"-d", "newdomain.com", "existing-app"}, starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo)

	assert.Contains(t, fakeUI.Outputs[0], "Creating route")
	assert.Contains(t, fakeUI.Outputs[0], "existing-app.newdomain.com")
	assert.Contains(t, fakeUI.Outputs[1], "OK")
	assert.Contains(t, fakeUI.Outputs[3], "Binding")
	assert.Contains(t, fakeUI.Outputs[3], "existing-app.newdomain.com")

	assert.Equal(t, appBitsRepo.UploadedAppGuid, "existing-app-guid")
	assert.Equal(t, domainRepo.FindByNameInCurrentSpaceName, "newdomain.com")
	assert.Equal(t, routeRepo.FindByHostAndDomainDomain, "newdomain.com")
	assert.Equal(t, routeRepo.FindByHostAndDomainHost, "existing-app")
	assert.Equal(t, routeRepo.CreatedHost, "existing-app")
	assert.Equal(t, routeRepo.CreatedDomainGuid, "domain-guid")
}

func TestPushingAppWhenItAlreadyExistsAndHostIsSpecified(t *testing.T) {
	starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo := getPushDependencies()

	domain_Auto := cf.DomainFields{}
	domain_Auto.Name = "example.com"

	existingRoute := cf.RouteSummary{}
	existingRoute.Host = "existing-app"
	existingRoute.Domain = domain_Auto

	existingApp := cf.Application{}
	existingApp.Name = "existing-app"
	existingApp.Guid = "existing-app-guid"
	existingApp.Routes = []cf.RouteSummary{existingRoute}

	appRepo.FindByNameApp = existingApp
	routeRepo.FindByHostAndDomainNotFound = true

	fakeUI := callPush(t, []string{"-n", "new-host", "existing-app"}, starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo)

	assert.Contains(t, fakeUI.Outputs[0], "Creating route")
	assert.Contains(t, fakeUI.Outputs[0], "new-host.example.com")
	assert.Contains(t, fakeUI.Outputs[1], "OK")
	assert.Contains(t, fakeUI.Outputs[3], "Binding")
	assert.Contains(t, fakeUI.Outputs[3], "new-host.example.com")

	assert.Equal(t, routeRepo.FindByHostAndDomainDomain, "example.com")
	assert.Equal(t, routeRepo.FindByHostAndDomainHost, "new-host")
	assert.Equal(t, routeRepo.CreatedHost, "new-host")
	assert.Equal(t, routeRepo.CreatedDomainGuid, "domain-guid")
}

func TestPushingAppWhenItAlreadyExistsAndNoRouteFlagIsPresent(t *testing.T) {
	starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo := getPushDependencies()
	existingApp := cf.Application{}
	existingApp.Name = "existing-app"
	existingApp.Guid = "existing-app-guid"

	appRepo.FindByNameApp = existingApp

	fakeUI := callPush(t, []string{"--no-route", "existing-app"}, starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo)

	assert.Contains(t, fakeUI.Outputs[0], "Uploading")
	assert.Contains(t, fakeUI.Outputs[0], "existing-app")
	assert.Contains(t, fakeUI.Outputs[1], "OK")

	assert.Equal(t, appBitsRepo.UploadedAppGuid, "existing-app-guid")
	assert.Equal(t, domainRepo.FindByNameInCurrentSpaceName, "")
	assert.Equal(t, routeRepo.FindByHostAndDomainDomain, "")
	assert.Equal(t, routeRepo.FindByHostAndDomainHost, "")
	assert.Equal(t, routeRepo.CreatedHost, "")
	assert.Equal(t, routeRepo.CreatedDomainGuid, "")
}

func TestPushingAppWhenItAlreadyExistsAndNoHostFlagIsPresent(t *testing.T) {
	starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo := getPushDependencies()

	domain_Auto := cf.DomainFields{}
	domain_Auto.Name = "example.com"

	existingRoute := cf.RouteSummary{}
	existingRoute.Host = "existing-app"
	existingRoute.Domain = domain_Auto

	existingApp := cf.Application{}
	existingApp.Name = "existing-app"
	existingApp.Guid = "existing-app-guid"
	existingApp.Routes = []cf.RouteSummary{existingRoute}

	appRepo.FindByNameApp = existingApp
	routeRepo.FindByHostAndDomainNotFound = true

	fakeUI := callPush(t, []string{"--no-hostname", "existing-app"}, starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo)

	assert.Contains(t, fakeUI.Outputs[0], "Creating route")
	assert.Contains(t, fakeUI.Outputs[0], "example.com")
	assert.NotContains(t, fakeUI.Outputs[0], "existing-app.example.com")
	assert.Contains(t, fakeUI.Outputs[1], "OK")
	assert.Contains(t, fakeUI.Outputs[3], "Binding")
	assert.Contains(t, fakeUI.Outputs[3], "example.com")
	assert.NotContains(t, fakeUI.Outputs[3], "existing-app.example.com")

	assert.Equal(t, routeRepo.FindByHostAndDomainDomain, "example.com")
	assert.Equal(t, routeRepo.FindByHostAndDomainHost, "")
	assert.Equal(t, routeRepo.CreatedHost, "")
	assert.Equal(t, routeRepo.CreatedDomainGuid, "domain-guid")
}

func TestPushingAppWhenItAlreadyExistsWithoutARouteAndARouteIsNotProvided(t *testing.T) {
	starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo := getPushDependencies()
	existingApp := cf.Application{}
	existingApp.Name = "existing-app"
	existingApp.Guid = "existing-app-guid"

	appRepo.FindByNameApp = existingApp

	fakeUI := callPush(t, []string{"existing-app"}, starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo)

	testassert.SliceContains(t, fakeUI.Outputs, []string{
		"skipping route creation",
		"Uploading",
		"OK",
	})

	assert.Equal(t, routeRepo.FindByHostAndDomainDomain, "")
	assert.Equal(t, routeRepo.FindByHostAndDomainHost, "")
	assert.Equal(t, routeRepo.CreatedHost, "")
	assert.Equal(t, routeRepo.CreatedDomainGuid, "")
}

func TestPushingAppWithInvalidPath(t *testing.T) {
	starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo := getPushDependencies()
	appBitsRepo.UploadAppErr = true

	fakeUI := callPush(t, []string{"app"}, starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo)

	testassert.SliceContains(t, fakeUI.Outputs, []string{"Uploading", "FAILED"})
}

func getPushDependencies() (starter *testcmd.FakeAppStarter,
	stopper *testcmd.FakeAppStopper,
	appRepo *testapi.FakeApplicationRepository,
	domainRepo *testapi.FakeDomainRepository,
	routeRepo *testapi.FakeRouteRepository,
	stackRepo *testapi.FakeStackRepository,
	appBitsRepo *testapi.FakeApplicationBitsRepository) {

	starter = &testcmd.FakeAppStarter{}
	stopper = &testcmd.FakeAppStopper{}
	appRepo = &testapi.FakeApplicationRepository{}
	domainRepo = &testapi.FakeDomainRepository{}
	routeRepo = &testapi.FakeRouteRepository{}
	stackRepo = &testapi.FakeStackRepository{}
	appBitsRepo = &testapi.FakeApplicationBitsRepository{}

	return
}

func callPush(t *testing.T,
	args []string,
	starter ApplicationStarter,
	stopper ApplicationStopper,
	appRepo api.ApplicationRepository,
	domainRepo api.DomainRepository,
	routeRepo api.RouteRepository,
	stackRepo api.StackRepository,
	appBitsRepo *testapi.FakeApplicationBitsRepository) (fakeUI *testterm.FakeUI) {

	fakeUI = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("push", args)

	token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
		Username: "my-user",
	})
	assert.NoError(t, err)
	org_Auto := cf.OrganizationFields{}
	org_Auto.Name = "my-org"
	space_Auto := cf.SpaceFields{}
	space_Auto.Name = "my-space"
	config := &configuration.Configuration{
		Space:        space_Auto,
		Organization: org_Auto,
		AccessToken:  token,
	}

	cmd := NewPush(fakeUI, config, starter, stopper, appRepo, domainRepo, routeRepo, stackRepo, appBitsRepo)
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true}
	testcmd.RunCommand(cmd, ctxt, reqFactory)

	return
}
