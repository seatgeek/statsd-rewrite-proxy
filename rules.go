package main

import (
	"fmt"
	"regexp"
	"strings"
)

const (
	ruleActionMatch = "match"
	ruleActionDrop  = "drop"
	ruleActionRelay = "relay"
	ruleActionMiss  = "miss"
)

// Rule ...
type Rule struct {
	*regexp.Regexp
	name   string
	action string
}

// RuleResult ...
type RuleResult struct {
	Captures map[string]string
	Tags     []string
	name     string
	action   string
}

// Rules ...
type Rules struct {
	list []*Rule
}

func (r *Rules) Match(ruleString, newPath string) {
	r.list = append(r.list, NewMatchRule(ruleString, newPath))
}

func (r *Rules) Relay(ruleString string) {
	r.list = append(r.list, NewRelayRule(ruleString))
}

func (r *Rules) Drop(ruleString string) {
	r.list = append(r.list, NewDropRule(ruleString))
}

// NewMatchRule ...
func NewMatchRule(ruleString string, newPath string) *Rule {
	return &Rule{
		action: ruleActionMatch,
		Regexp: buildRegexp(ruleString),
		name:   newPath,
	}
}

// NewDropRule ..
func NewDropRule(ruleString string) *Rule {
	return &Rule{
		action: ruleActionDrop,
		Regexp: buildRegexp(ruleString),
	}
}

// NewRelayRule ...
func NewRelayRule(rulestring string) *Rule {
	return &Rule{
		action: ruleActionRelay,
		Regexp: buildRegexp(rulestring),
	}
}

// FindStringSubmatchMap add a new method to our new regular expression type
func (r *Rule) FindStringSubmatchMap(s string) *RuleResult {
	result := &RuleResult{
		action: r.action,
	}

	match := r.FindStringSubmatch(s)
	if match == nil {
		result.action = ruleActionMiss
		return result
	}

	if r.action == ruleActionDrop || r.action == ruleActionRelay {
		return result
	}

	result.Captures = make(map[string]string, 0)
	result.name = r.name

	for i, name := range r.SubexpNames() {
		if i == 0 {
			continue
		}

		result.Captures[name] = match[i]
		result.Tags = append(result.Tags, fmt.Sprintf("%s:%s", name, match[i]))
		result.name = strings.Replace(result.name, "{"+name+"}", match[i], -1)
	}

	return result
}

func createRules() {

	/*********************************************************************************************************************************************************
	 * Vault Metrics
	 *********************************************************************************************************************************************************/

	// internal metrics
	rules.Match("vault.runtime.{vault_runtime_type}", "vault.runtime")
	rules.Match("vault.audit.{vault_audit_type}", "vault.audit")
	rules.Match("vault.barrier.{vault_barrier_type}", "vault.barrier")
	rules.Match("vault.consul.{vault_consul_type}", "vault.consul")
	rules.Match("vault.core.{vault_core_type}", "vault.core")

	// policy and token metrics
	rules.Match("vault.expire.{vault_expire_type}", "vault.expire")
	rules.Match("vault.policy.{vault_policy_type}", "vault.policy")
	rules.Match("vault.token.{vault_token_type}", "vault.token")

	// authentication
	rules.Match("vault.rollback.attempt.{vault_auth_backend}", "vault.authentication.attempt")
	rules.Match("vault.route.read.{vault_auth_backend}", "vault.authentication.read")
	rules.Match("vault.route.renew.{vault_auth_backend}", "vault.authentication.renew")
	rules.Match("vault.route.revoke.{vault_auth_backend}", "vault.authentication.revoke")
	rules.Match("vault.route.rollback.{vault_auth_backend}", "vault.authentication.rollback")
	rules.Match("vault.route.update.{vault_auth_backend}", "vault.authentication.update")

	// storage backends
	rules.Match("vault.azure.{vault_storage_action}", "vault.storage.azure")
	rules.Match("vault.dynamodb.{vault_storage_action}", "vault.storage.storage")
	rules.Match("vault.etcd.{vault_storage_action}", "vault.storage.etcd")
	rules.Match("vault.gcs.{vault_storage_action}", "vault.storage.gcs")
	rules.Match("vault.mysql.{vault_storage_action}", "vault.storage.mysql")
	rules.Match("vault.postgres.{vault_storage_action}", "vault.storage.postgres")
	rules.Match("vault.s3.{vault_storage_action}", "vault.storage.s3")
	rules.Match("vault.swift.{vault_storage_action}", "vault.storage.swift")
	rules.Match("vault.zookeeper.{vault_storage_action}", "vault.storage.zookeeper")

	// Drop anything we didn't match
	rules.Relay("vault.*")

	/*********************************************************************************************************************************************************
	 * Nomad Key Metrics
	 *********************************************************************************************************************************************************/

	// nomad.runtime.*
	rules.Relay("nomad.runtime.*")

	// nomad.raft.*
	rules.Relay("nomad.raft.*")

	// nomad.broker.*
	rules.Relay("nomad.broker.*")

	// nomad.plan.*
	rules.Relay("nomad.plan.*")

	// nomad.uptime
	rules.Relay("nomad.uptime")

	// nomad.worker.wait_for_index
	rules.Relay("nomad.worker.wait_for_index")

	// nomad.worker.invoke_scheduler.<type>
	rules.Match("nomad.worker.invoke_scheduler.{nomad_scheduler}", "nomad.worker.invoke_scheduler")

	// nomad.heartbeat.*
	rules.Relay("nomad.heartbeat.*")

	// nomad.rpc.*
	rules.Relay("nomad.rpc.*")

	/*********************************************************************************************************************************************************
	 * Nomad Host Metrics
	 *********************************************************************************************************************************************************/

	// nomad.client.uptime.<HostID>
	rules.Match("nomad.client.uptime.{nomad_client}", "nomad.client.uptime")

	// nomad.client.host.cpu.<HostID>.<CPU-Core>.total
	// nomad.client.host.cpu.<HostID>.<CPU-Core>.user
	// nomad.client.host.cpu.<HostID>.<CPU-Core>.system
	// nomad.client.host.cpu.<HostID>.<CPU-Core>.idle
	rules.Match("nomad.client.host.cpu.{nomad_client}.{nomad_client_cpu_core}.{nomad_cpu_metric}", "nomad.client.cpu.{nomad_cpu_metric}")

	// nomad.client.host.disk.<HostID>.<Device-Name>.size
	// nomad.client.host.disk.<HostID>.<Device-Name>.used
	// nomad.client.host.disk.<HostID>.<Device-Name>.available
	// nomad.client.host.disk.<HostID>.<Device-Name>.used_percent
	// nomad.client.host.disk.<HostID>.<Device-Name>.inodes_percent
	rules.Match("nomad.client.host.disk.{nomad_client}.{nomad_client_device}.{nomad_disk_metric}", "nomad.client.disk.{nomad_disk_metric}")

	// nomad.client.host.memory.<HostID>.total
	// nomad.client.host.memory.<HostID>.available
	// nomad.client.host.memory.<HostID>.used
	// nomad.client.host.memory.<HostID>.free
	rules.Match("nomad.client.host.memory.{nomad_client}.{nomad_client_memory_metric}", "nomad.client.host.memory.{nomad_client_memory_metric}")

	// nomad.client.allocated.cpu.<HostID>
	rules.Match("nomad.client.allocated.cpu.{nomad_client}", "nomad.client.allocated.cpu")

	// nomad.client.allocated.memory.<HostID>
	rules.Match("nomad.client.allocated.memory.{nomad_client}", "nomad.client.allocated.memory")

	// nomad.client.allocated.disk.<HostID>
	rules.Match("nomad.client.allocated.disk.{nomad_client}", "nomad.client.allocated.disk")

	// nomad.client.allocated.iops.<HostID>
	rules.Match("nomad.client.allocated.iops.{nomad_client}", "nomad.client.allocated.iops")

	// nomad.client.allocated.network.<Device-Name>.<HostID>
	rules.Match("nomad.client.allocated.network.{nomad_device_name}.{nomad_client}", "nomad.client.allocated.network")

	// nomad.client.unallocated.cpu.<HostID>
	rules.Match("nomad.client.unallocated.cpu.{nomad_client}", "nomad.client.unallocated.cpu")

	// nomad.client.unallocated.memory.<HostID>
	rules.Match("nomad.client.unallocated.memory.{nomad_client}", "nomad.client.unallocated.memory")

	// nomad.client.unallocated.disk.<HostID>
	rules.Match("nomad.client.unallocated.disk.{nomad_client}", "nomad.client.unallocated.disk")

	// nomad.client.unallocated.iops.<HostID>
	rules.Match("nomad.client.unallocated.iops.{nomad_client}", "nomad.client.unallocated.iops")

	// nomad.client.unallocated.network.<Device-Name>.<HostID>
	rules.Match("nomad.client.unallocated.network.{nomad_device_name}.{nomad_client}", "nomad.client.unallocated.network")

	/*********************************************************************************************************************************************************
	 * Nomad Allocation Metrics
	 *********************************************************************************************************************************************************/

	// nomad.client.allocations.migrating.<HostID>
	rules.Match("nomad.client.allocations.migrating.{nomad_client}", "nomad.client.allocations.migrating")

	// nomad.client.allocations.blocked.<HostID>
	rules.Match("nomad.client.allocations.blocked.{nomad_client}", "nomad.client.allocations.blocked")

	// nomad.client.allocations.pending.<HostID>
	rules.Match("nomad.client.allocations.pending.{nomad_client}", "nomad.client.allocations.pending")

	// nomad.client.allocations.running.<HostID>
	rules.Match("nomad.client.allocations.running.{nomad_client}", "nomad.client.allocations.running")

	// nomad.client.allocations.terminal.<HostID>
	rules.Match("nomad.client.allocations.terminal.{nomad_client}", "nomad.client.allocations.terminal")

	// nomad.client.allocs.<Job>.<TaskGroup>.<AllocID>.<Task>.memory.rss
	// nomad.client.allocs.<Job>.<TaskGroup>.<AllocID>.<Task>.memory.cache
	// nomad.client.allocs.<Job>.<TaskGroup>.<AllocID>.<Task>.memory.swap
	// nomad.client.allocs.<Job>.<TaskGroup>.<AllocID>.<Task>.memory.max_usage
	// nomad.client.allocs.<Job>.<TaskGroup>.<AllocID>.<Task>.memory.kernel_usage
	// nomad.client.allocs.<Job>.<TaskGroup>.<AllocID>.<Task>.memory.kernel_max_usage
	rules.Match("nomad.client.allocs.{nomad_job}.{nomad_task_group}.{nomad_allocation_id}.{nomad_task}.memory.{nomad_job_memory_metric}", "nomad.allocation.memory.{nomad_job_memory_metric}")

	// nomad.client.allocs.<Job>.<TaskGroup>.<AllocID>.<Task>.cpu.total_percent
	// nomad.client.allocs.<Job>.<TaskGroup>.<AllocID>.<Task>.cpu.system
	// nomad.client.allocs.<Job>.<TaskGroup>.<AllocID>.<Task>.cpu.user
	// nomad.client.allocs.<Job>.<TaskGroup>.<AllocID>.<Task>.cpu.throttled_time
	// nomad.client.allocs.<Job>.<TaskGroup>.<AllocID>.<Task>.cpu.total_ticks
	rules.Match("nomad.client.allocs.{nomad_job}.{nomad_task_group}.{nomad_allocation_id}.{nomad_task}.cpu.{nomad_job_cpu_metric}", "nomad.allocation.cpu.{nomad_job_cpu_metric}")

	rules.Drop("nomad.*")

	/*********************************************************************************************************************************************************
	 * Fabio Metrics
	 *********************************************************************************************************************************************************/

	rules.Match("fabio.{fabio_service}.*.{fabio_path}.*.count", "fabio.requests.count")
	rules.Match("fabio.{fabio_service}.*.{fabio_path}.*.min", "fabio.requests.min")
	rules.Match("fabio.{fabio_service}.*.{fabio_path}.*.max", "fabio.requests.max")
	rules.Match("fabio.{fabio_service}.*.{fabio_path}.*.95_percentile", "fabio.requests.95_percentile")
	rules.Match("fabio.{fabio_service}.*.{fabio_path}.*.99_percentile", "fabio.requests.99_percentile")
	rules.Match("fabio.{fabio_service}.*.{fabio_path}.*.999_percentile", "fabio.requests.999_percentile")

	rules.Match("fabio.http.status.{fabio_response_code}.count", "fabio.http.response_code.count")
	rules.Match("fabio.http.status.{fabio_response_code}.min", "fabio.http.response_code.min")
	rules.Match("fabio.http.status.{fabio_response_code}.max", "fabio.http.response_code.max")
	rules.Match("fabio.http.status.{fabio_response_code}.95_percentile", "fabio.http.response_code.95_percentile")
	rules.Match("fabio.http.status.{fabio_response_code}.99_percentile", "fabio.http.response_code.99_percentile")
	rules.Match("fabio.http.status.{fabio_response_code}.999_percentile", "fabio.http.response_code.999_percentile")

	rules.Drop("fabio.*")
}
