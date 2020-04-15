package gorethink

import (
	p "gopkg.in/gorethink/gorethink.v3/ql2"
)

// Config can be used to read and/or update the configurations for individual
// tables or databases.
func (t Term) Config() Term {
	return constructMethodTerm(t, "Config", p.Term_CONFIG, []interface{}{}, map[string]interface{}{})
}

// Rebalance rebalances the shards of a table. When called on a database, all
// the tables in that database will be rebalanced.
func (t Term) Rebalance() Term {
	return constructMethodTerm(t, "Rebalance", p.Term_REBALANCE, []interface{}{}, map[string]interface{}{})
}

// ReconfigureOpts contains the optional arguments for the Reconfigure term.
type ReconfigureOpts struct {
	Shards               interface{} `gorethink:"shards,omitempty"`
	Replicas             interface{} `gorethink:"replicas,omitempty"`
	DryRun               interface{} `gorethink:"dry_run,omitempty"`
	EmergencyRepair      interface{} `gorethink:"emergency_repair,omitempty"`
	NonVotingReplicaTags interface{} `gorethink:"nonvoting_replica_tags,omitempty"`
	PrimaryReplicaTag    interface{} `gorethink:"primary_replica_tag,omitempty"`
}

func (o ReconfigureOpts) toMap() map[string]interface{} {
	return optArgsToMap(o)
}

// Reconfigure a table's sharding and replication.
func (t Term) Reconfigure(optArgs ...ReconfigureOpts) Term {
	opts := map[string]interface{}{}
	if len(optArgs) >= 1 {
		opts = optArgs[0].toMap()
	}
	return constructMethodTerm(t, "Reconfigure", p.Term_RECONFIGURE, []interface{}{}, opts)
}

// Status return the status of a table
func (t Term) Status() Term {
	return constructMethodTerm(t, "Status", p.Term_STATUS, []interface{}{}, map[string]interface{}{})
}

// WaitOpts contains the optional arguments for the Wait term.
type WaitOpts struct {
	WaitFor interface{} `gorethink:"wait_for,omitempty"`
	Timeout interface{} `gorethink:"timeout,omitempty"`
}

func (o WaitOpts) toMap() map[string]interface{} {
	return optArgsToMap(o)
}

// Wait for a table or all the tables in a database to be ready. A table may be
// temporarily unavailable after creation, rebalancing or reconfiguring. The
// wait command blocks until the given table (or database) is fully up to date.
//
// Deprecated: This function is not supported by RethinkDB 2.3 and above.
func Wait(optArgs ...WaitOpts) Term {
	opts := map[string]interface{}{}
	if len(optArgs) >= 1 {
		opts = optArgs[0].toMap()
	}
	return constructRootTerm("Wait", p.Term_WAIT, []interface{}{}, opts)
}

// Wait for a table or all the tables in a database to be ready. A table may be
// temporarily unavailable after creation, rebalancing or reconfiguring. The
// wait command blocks until the given table (or database) is fully up to date.
func (t Term) Wait(optArgs ...WaitOpts) Term {
	opts := map[string]interface{}{}
	if len(optArgs) >= 1 {
		opts = optArgs[0].toMap()
	}
	return constructMethodTerm(t, "Wait", p.Term_WAIT, []interface{}{}, opts)
}

// Grant modifies access permissions for a user account, globally or on a
// per-database or per-table basis.
func (t Term) Grant(args ...interface{}) Term {
	return constructMethodTerm(t, "Grant", p.Term_GRANT, args, map[string]interface{}{})
}
