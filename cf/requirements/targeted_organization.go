package requirements

import (
	"fmt"
	"github.com/nimbus-cloud/cli/cf"
	"github.com/nimbus-cloud/cli/cf/configuration"
	"github.com/nimbus-cloud/cli/cf/models"
	"github.com/nimbus-cloud/cli/cf/terminal"
)

type TargetedOrgRequirement interface {
	Requirement
	GetOrganizationFields() models.OrganizationFields
}

type targetedOrgApiRequirement struct {
	ui     terminal.UI
	config configuration.Reader
}

func NewTargetedOrgRequirement(ui terminal.UI, config configuration.Reader) TargetedOrgRequirement {
	return targetedOrgApiRequirement{ui, config}
}

func (req targetedOrgApiRequirement) Execute() (success bool) {
	if !req.config.HasOrganization() {
		message := fmt.Sprintf("No org targeted, use '%s' to target an org.",
			terminal.CommandColor(cf.Name()+" target -o ORG"))
		req.ui.Failed(message)
		return false
	}

	return true
}

func (req targetedOrgApiRequirement) GetOrganizationFields() (org models.OrganizationFields) {
	return req.config.OrganizationFields()
}
