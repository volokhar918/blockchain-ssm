// ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ----
// Copyright Luc Yriarte <luc.yriarte@thingagora.org> 2018 
// License: Apache-2.0
//
// Signing State Machines chaincode
//
// ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ----

package main

import (
	"fmt"
	"encoding/json"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)


// ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ----
//
// chaincode interface
//
// ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ----

type SSMChaincode struct {
}


//
// chaincode initialization
//

func (self *SSMChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	_, args := stub.GetFunctionAndParameters()
	if len(args) != 1 {
		return shim.Error("Incorrect arg count. Expecting 1")
	}

	// "init", admins: <Agent>* 
	var admins []Agent
	// Create admins array from JSON string
	err := json.Unmarshal([]byte(args[0]), &admins)
	if (err != nil) {
		return shim.Error(err.Error())
	}
	// Verify admins public key before storing
	for i := 0; i < len(admins); i++ {
		_, err = admins[i].PublicKey()
		if err != nil {
			return shim.Error(err.Error())
		}
	}
	for i := 0; i < len(admins); i++ {
		// Store every admin
		err = admins[i].Put(stub, "ADMIN_" + admins[i].Name)
		if (err != nil) {
			return shim.Error(err.Error())
		}
	}
	return shim.Success(nil)
}


//
// chaincode invocation for transactions and queries
//

func (self *SSMChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	
	errmsg := "Incorrect arg count."
	var err error	
	
	//
	//	transactions
	//
	
	// "register", user:Agent, admin_name:string, signature:b64	
	if function == "register" {
		if len(args) != 3 {
			return shim.Error(errmsg)
		}
		return self.Register(stub, args)
	}

	// "create", ssm:SigningStateMachine, admin_name:string, signature:b64
	if function == "create" {
		if len(args) != 3 {
			return shim.Error(errmsg)
		}
		return self.Create(stub, args)
	}
	
	// "start", init:State, admin_name:string, signature:b64
	if function == "start" {
		if len(args) != 3 {
			return shim.Error(errmsg)
		}
		return self.Start(stub, args)
	}
	
	// "perform", action:string, context:State, user_name:string, signature:b64
	if function == "perform" {
		if len(args) != 4 {
			return shim.Error(errmsg)
		}
		return self.Perform(stub, args)
	}
	
	//
	//	queries
	//

	if len(args) != 1 {
		return shim.Error(errmsg)
	}
	errmsg = "Unknown operation."
	var dat []byte
	
	if function == "session" {
	// "session", <session id> 
		dat, err = stub.GetState("STATE_" + args[0])
	} else if function == "ssm" {
	// "ssm", <ssm name> 
		dat, err = stub.GetState("SSM_" + args[0])
	} else if function == "user" {
	// "user", <user name>
		dat, err = stub.GetState("USER_" + args[0])
	} else if function == "admin" {
	// "admin", <admin name>
		dat, err = stub.GetState("ADMIN_" + args[0])
	} else {
		return shim.Error(errmsg)
	}
	
	if (err != nil) {
		return shim.Error(err.Error())
	}
	
	return shim.Success(dat)
}


// ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ----
//
// transactions API implementation
//
// ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ----

// "register", user:Agent, admin_name:string, signature:b64	
func (self *SSMChaincode) Register(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var admin Agent
	var user Agent
	// Get validating admin
	err := admin.Get(stub, "ADMIN_" + args[1])
	if (err != nil) {
		return shim.Error(err.Error())
	}
	// TODO Validate admin signature on user JSON string
	// Create user from JSON string
	err = user.Deserialize([]byte(args[0]))
	if (err != nil) {
		return shim.Error(err.Error())
	}
	// Verify user public key before storing
	_, err = user.PublicKey()
	if err != nil {
		return shim.Error(err.Error())
	}
	// Store user if not alreay existing
	err = user.Put(stub, "USER_" + user.Name)
	if (err != nil) {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}


// "create", ssm:SigningStateMachine, admin_name:string, signature:b64
func (self *SSMChaincode) Create(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var admin Agent
	var ssm SigningStateMachine
	// Get validating admin
	err := admin.Get(stub, "ADMIN_" + args[1])
	if (err != nil) {
		return shim.Error(err.Error())
	}
	// TODO Validate admin signature on ssm JSON string
	// Create ssm from JSON string
	err = ssm.Deserialize([]byte(args[0]))
	if (err != nil) {
		return shim.Error(err.Error())
	}
	// Store ssm if not alreay existing
	err = ssm.Put(stub, "SSM_" + ssm.Name)
	if (err != nil) {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}


// "start", init:State, admin_name:string, signature:b64
func (self *SSMChaincode) Start(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var admin Agent
	var state State
	// Get validating admin
	err := admin.Get(stub, "ADMIN_" + args[1])
	if (err != nil) {
		return shim.Error(err.Error())
	}
	// TODO Validate admin signature on state JSON string
	// Create state from JSON string
	err = state.Deserialize([]byte(args[0]))
	if (err != nil) {
		return shim.Error(err.Error())
	}
	// Store state if not alreay existing
	err = state.Put(stub, "STATE_" + state.Session)
	if (err != nil) {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}


// "perform", action:string, context:State, user_name:string, signature:b64
func (self *SSMChaincode) Perform(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var user Agent
	var state State
	// Get validating user
	err := user.Get(stub, "USER_" + args[2])
	if (err != nil) {
		return shim.Error(err.Error())
	}
	// TODO Validate user signature on state JSON string
	// Create state from JSON string
	err = state.Get(stub, "STATE_" + args[1])
	if (err != nil) {
		return shim.Error(err.Error())
	}
	// TODO Update state if validated
	return shim.Success(nil)
}




// ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ----
//
// main function
//
// ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ----

func main() {
	err := shim.Start(new(SSMChaincode))
	if err != nil {
		fmt.Println("Signing State Machines Chaincode init error:", err)
	}
}
