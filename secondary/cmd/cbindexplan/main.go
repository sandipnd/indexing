// Copyright (c) 2014 Couchbase, Inc.
// Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file
// except in compliance with the License. You may obtain a copy of the License at
//   http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software distributed under the
// License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
// either express or implied. See the License for the specific language governing permissions
// and limitations under the License.

package main

import (
	"flag"
	"fmt"
	"github.com/couchbase/cbauth"
	"github.com/couchbase/cbauth/service"
	"github.com/couchbase/indexing/secondary/common"
	"github.com/couchbase/indexing/secondary/logging"
	"github.com/couchbase/indexing/secondary/planner"
	"os"
	"strings"
)

// - document that planning tool should be run on an indexer node or manual enter cpu quota
// - document that planner will not size array index from a live cluster.  It will be treated as
//   sec index (1) per key sizing estimate will be higher (46 more per key), (2) protobuf overhead
//   sizing will be smaller than actual.
// - note that replica support is not avail

func usage() {
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Usage: cbindexplan [options]")
	flag.PrintDefaults()
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, `Examples:
- Plan 
    cbindexplan -command=plan -indexes="indexes.json" -memQuota="10G" -cpuQuota=16
    cbindexplan -command=plan -cluster="127.0.0.1:8091" -username="<user>" -password="<pwd>" -indexes="indexes.json"
    cbindexplan -command=plan -cluster="127.0.0.1:8091" -username="<user>" -password="<pwd>" -indexes="indexes.json" -allowUnpin
    cbindexplan -command=plan -indexes="indexes.json" -memQuota="10G" -cpuQuota=16 -ddl="saved-ddl.txt"
    cbindexplan -command=plan -indexes="indexes.json" -memQuota="10G" -cpuQuota=16 -output="saved-plan.json"
    cbindexplan -command=plan -plan="saved-plan.json" -indexes="indexes.json"
    cbindexplan -command=plan -plan="saved-plan.json" -indexes="indexes.json" -memQuota="10G" -cpuQuota=16 -output="newplan.json"
- Rebalance 
    cbindexplan -command=rebalance-cluster="127.0.0.1:8091" -username="<user>" -password="<pwd>"
    cbindexplan -command=rebalance-cluster="127.0.0.1:8091" -username="<user>" -password="<pwd>" -addNode=3 -output="saved-plan.json"
    cbindexplan -command=rebalance -plan="saved-plan.json"
    cbindexplan -command=rebalance -plan="saved-plan.json" -output="newplan.json"
    cbindexplan -command=rebalance -plan="saved-plan.json" -addNode=1
    `)
	fmt.Fprintln(os.Stderr, `Usage Note:
1) cbindexplan should only be used with MOI clsuter.
2) When running cbindexplan, it may complain that the memory quota or cpu quota is not sufficient when pointing to a live cluster.
   This is because cbindexplan can recalculate index size using the MOI sizing formula.   In this case, use -memQuota and -cpuQuota
   to override the clsuter setup during planning.
    `)
	fmt.Fprintln(os.Stderr, `Placement Note:
1) cbindexplan is a planning recommendation tool for index placement and rebalancing. This does not actual create or rebalance index,
   but it provides recommednation to user on how to place index, or how to move the index to get a better resource utilization.
2) New indexes (to be replaced) are specified using a json file.   Example of index json file is under
   https://github.com/couchbase/indexing/blob/master/secondary/cmd/cbindexplan/sample/index.json
3) cbindexplan can recommend placement with just given an initial set of indexes (using -indexes option).  This will provide
   produce a layout of indexes which optimzie resource distribution. User can optionally save the outcome into a plan file
   (when specifying -output option).
4) cbindexplan can recommend placement of new indexes on a live clsuter (when using the -cluster option).  User can optionally save
   the outcome into a plan file (when specifying -output option).
5) cbindexplan can recommend placement of new indexes on top of a saved plan (when using the -plan option).
6) For placement, cbindexplan can generate create-index and build-index statmeents for new indexes when using -ddl option.
7) For placement, cbindexplan will recalculate the size for all indexes using MOI sizing equation.   Besides new indexes to be replaced,
   cbindexplan will also recaculate size for indexes retrived from a saved plan or live cluster before placement algorithm is run.
    `)
	fmt.Fprintln(os.Stderr, `Rebalancing Note:
1) cbindex can be used to simulate index rebalancing by using the rebalance command.   When rebalancing from a live cluster, cbindexplan
   will estimate index size from indexer stats.  The estimate live index size will be used for rebalancing algorithm.
2) For rebalancing, if an index is pinned to a node (when index is created with with-nodes option), rebalancing algorithm will not move
   those index.  Use 'unpin' option to instruct the rebalance algorithm to rebalance pinned indexes.
    `)
}

//////////////////////////////////////////////////////////////
// Global Variable
/////////////////////////////////////////////////////////////

var gHelp bool
var gDetail bool
var gGenStmt string
var gPlan string
var gIndexSpecs string
var gClusterUrl string
var gUsername string
var gPassword string
var gOutput string
var gLogLevel string
var gAllowUnpin bool
var gCommand string
var gAddNode int
var gMemQuota string
var gCpuQuota int
var gEjectedNode string

//////////////////////////////////////////////////////////////
// Initialization
/////////////////////////////////////////////////////////////

func init() {
	flag.BoolVar(&gHelp, "help", false, "print usage")
	flag.BoolVar(&gDetail, "layout", false, "print index layout plan to console after planning")
	flag.StringVar(&gLogLevel, "logLevel", "INFO", "log level")
	flag.StringVar(&gOutput, "output", "", "save index layout plan to a file after planning")
	flag.StringVar(&gGenStmt, "ddl", "", "generate DDL statement after planning for new/moved indexes")

	// command + index specification
	flag.StringVar(&gCommand, "command", "", "command = {plan | rebalance}")
	flag.StringVar(&gClusterUrl, "cluster", "", "fetch existing index layout plan from cluster url")
	flag.StringVar(&gUsername, "username", "", "admin user for the cluster")
	flag.StringVar(&gPassword, "password", "", "admin password for the cluster")
	flag.StringVar(&gIndexSpecs, "indexes", "", "list of indexes for placement")
	flag.StringVar(&gPlan, "plan", "", "fetch existing index layout from a saved plan file  (in place of specifying cluster url)")

	// quota
	flag.StringVar(&gMemQuota, "memQuota", "", "memory quota per indexer node (e.g. 100M, 1G)")
	flag.IntVar(&gCpuQuota, "cpuQuota", -1, "cpu quota per indexer node")

	// cluster size
	flag.IntVar(&gAddNode, "addNode", 0, "number of indexer to add before running the planner")

	// placement
	flag.BoolVar(&gAllowUnpin, "allowUnpin", false, "flag to tell if planner should allow existing index to move during placement.")

	// swap
	flag.StringVar(&gEjectedNode, "ejectNode", "", "node to be ejected from cluster")
}

func main() {
	flag.Parse()
	logging.SetLogLevel(logging.Level(strings.ToUpper(gLogLevel)))

	if gHelp {
		usage()
		return
	}

	if gPlan != "" && gClusterUrl != "" {
		logging.Fatalf("Invalid argument: Cannot specify both 'plan' and 'cluster'.")
		usage()
		return
	}

	plan, err := planner.ReadPlan(gPlan)
	if err != nil {
		logging.Fatalf("Error in reading plan: %v", err)
		return
	}

	if gClusterUrl != "" {

		_, err := cbauth.InternalRetryDefaultInit(gClusterUrl, gUsername, gPassword)
		if err != nil {
			logging.Fatalf("cbauth initialization fails. err = %s", gClusterUrl, err)
			return
		}

		plan, err = planner.RetrievePlanFromCluster(gClusterUrl)
		if err != nil {
			logging.Fatalf("Unable to read index layout from cluster %v. err = %s", gClusterUrl, err)
			return
		}
	}

	if gCommand == planner.CommandRebalance && plan == nil {
		logging.Fatalf("Unable to get index layout from either argument 'plan' or 'cluster'.")
		usage()
		return
	}

	memQuota, err := planner.ParseMemoryStr(gMemQuota)
	if err != nil {
		logging.Fatalf("%v", err)
		return
	}

	if gCommand == string(planner.CommandPlan) {

		indexSpecs, err := planner.ReadIndexSpecs(gIndexSpecs)
		if err != nil {
			logging.Fatalf("%v", err)
			return
		}

		if indexSpecs == nil {
			logging.Fatalf("Invalid argument: argument 'indexes' is required to specify indexes to be placed.")
			usage()
			return
		}

		_, err = planner.ExecutePlanWithOptions(plan, indexSpecs, gDetail, gGenStmt, gOutput, gAddNode, gCpuQuota, memQuota, gAllowUnpin)
		if err != nil {
			logging.Fatalf("Planner error: %v.", err)
			return
		}

	} else if gCommand == string(planner.CommandRebalance) {

		if gGenStmt != "" {
			logging.Fatalf("Invalid argument: option 'ddl' is not supported for rebalancing.")
		}

		_, err := planner.ExecuteRebalanceWithOptions(plan, nil, gDetail, gGenStmt, gOutput, gAddNode, gCpuQuota, memQuota, gAllowUnpin, nil)
		if err != nil {
			logging.Fatalf("Planner error: %v.", err)
			return
		}

	} else if gCommand == string(planner.CommandSwap) {

		logging.Infof("CommandSwap is used.  This is for internal testing only.  Some optional arguments could be ignored.")

		if gClusterUrl == "" {
			logging.Fatalf("Missing argument: option 'clusterUrl' is required.")
			return
		}

		if gEjectedNode == "" {
			logging.Fatalf("Missing argument: option 'ejectNode' is required.")
			return
		}

		str, _ := common.NewUUID()
		tcid := fmt.Sprintf("TopologyChangeID%s", str.Str())

		change := service.TopologyChange{
			ID:         tcid,
			EjectNodes: make([]service.NodeInfo, 1),
		}

		var masterId string
		for _, indexer := range plan.Placement {
			masterId = indexer.NodeUUID

			if indexer.NodeId == gEjectedNode {
				node := service.NodeInfo{
					NodeID: service.NodeID(indexer.NodeUUID),
				}
				change.EjectNodes[0] = node
				break
			}
		}

		if len(change.EjectNodes) == 0 {
			logging.Fatalf("Invalid argument: Cannot find matching node %v from cluster %v", gEjectedNode, gClusterUrl)
			return
		}

		tokens, err := planner.ExecuteRebalanceInternal(gClusterUrl, change, masterId, true, gDetail, true)
		if err != nil {
			logging.Fatalf("Planner error: %v.", err)
			return
		}

		for _, token := range tokens {
			logging.Infof("----------------------")
			logging.Infof("Transfer Token ID: %v", token.RebalId)
			logging.Infof("Transfer Token Master ID: %v", token.MasterId)
			logging.Infof("Transfer Token Source ID: %v", token.SourceId)
			logging.Infof("Transfer Token Dest ID: %v", token.DestId)
			logging.Infof("Transfer Token State: %v", token.State)
			logging.Infof("Transfer Token Index Name: %v", token.IndexInst.Defn.Name)
			logging.Infof("Transfer Token Index Bucket: %v", token.IndexInst.Defn.Bucket)
			logging.Infof("Transfer Token Index DefnId : %v", token.IndexInst.Defn.DefnId)
			logging.Infof("Transfer Token Index InstId : %v", token.InstId)
		}

	} else {
		logging.Fatalf("Invalid argument: Invalid value for 'command' : %v", gCommand)
		usage()
		return
	}
}
