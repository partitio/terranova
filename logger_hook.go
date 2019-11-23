package terranova

import (
	"log"

	"github.com/hashicorp/terraform/addrs"
	"github.com/hashicorp/terraform/plans"
	"github.com/hashicorp/terraform/providers"
	"github.com/hashicorp/terraform/states"
	"github.com/hashicorp/terraform/terraform"
	"github.com/zclconf/go-cty/cty"
)

type logHook struct{}

func (*logHook) PreApply(addr addrs.AbsResourceInstance, gen states.Generation, action plans.Action, priorState, plannedNewState cty.Value) (terraform.HookAction, error) {
	return terraform.HookActionContinue, nil
}

func (*logHook) PostApply(addr addrs.AbsResourceInstance, gen states.Generation, newState cty.Value, err error) (terraform.HookAction, error) {
	return terraform.HookActionContinue, nil
}

func (*logHook) PreDiff(addr addrs.AbsResourceInstance, gen states.Generation, priorState, proposedNewState cty.Value) (terraform.HookAction, error) {
	return terraform.HookActionContinue, nil
}

func (*logHook) PostDiff(addr addrs.AbsResourceInstance, gen states.Generation, action plans.Action, priorState, plannedNewState cty.Value) (terraform.HookAction, error) {
	return terraform.HookActionContinue, nil
}

func (*logHook) PreProvisionInstance(addr addrs.AbsResourceInstance, state cty.Value) (terraform.HookAction, error) {
	return terraform.HookActionContinue, nil
}

func (*logHook) PostProvisionInstance(addr addrs.AbsResourceInstance, state cty.Value) (terraform.HookAction, error) {
	return terraform.HookActionContinue, nil
}

func (*logHook) PreProvisionInstanceStep(addr addrs.AbsResourceInstance, typeName string) (terraform.HookAction, error) {
	return terraform.HookActionContinue, nil
}

func (*logHook) PostProvisionInstanceStep(addr addrs.AbsResourceInstance, typeName string, err error) (terraform.HookAction, error) {
	return terraform.HookActionContinue, nil
}

func (*logHook) ProvisionOutput(addr addrs.AbsResourceInstance, typeName string, line string) {
	log.Printf("[INFO] %s: %s", typeName, line)
}

func (*logHook) PreRefresh(addr addrs.AbsResourceInstance, gen states.Generation, priorState cty.Value) (terraform.HookAction, error) {
	return terraform.HookActionContinue, nil
}

func (*logHook) PostRefresh(addr addrs.AbsResourceInstance, gen states.Generation, priorState cty.Value, newState cty.Value) (terraform.HookAction, error) {
	return terraform.HookActionContinue, nil
}

func (*logHook) PreImportState(addr addrs.AbsResourceInstance, importID string) (terraform.HookAction, error) {
	return terraform.HookActionContinue, nil
}

func (*logHook) PostImportState(addr addrs.AbsResourceInstance, imported []providers.ImportedResource) (terraform.HookAction, error) {
	return terraform.HookActionContinue, nil
}

func (*logHook) PostStateUpdate(new *states.State) (terraform.HookAction, error) {
	return terraform.HookActionContinue, nil
}
