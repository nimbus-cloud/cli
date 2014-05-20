package api

import "github.com/nimbus-cloud/cli/cf/models"

type FakeServiceSummaryRepo struct {
	GetSummariesInCurrentSpaceInstances []models.ServiceInstance
}

func (repo *FakeServiceSummaryRepo) GetSummariesInCurrentSpace() (instances []models.ServiceInstance, apiErr error) {
	instances = repo.GetSummariesInCurrentSpaceInstances
	return
}
