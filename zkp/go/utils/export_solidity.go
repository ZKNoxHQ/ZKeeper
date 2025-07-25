package utils

import (
	"fmt"
	"os"

	"github.com/consensys/gnark/backend/plonk"
	plonk_bn254 "github.com/consensys/gnark/backend/plonk/bn254"
	"github.com/consensys/gnark/backend/witness"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

func WriteSolidityTestFile(proof plonk.Proof, publicWitness witness.Witness) error {

	// 9. Export the Solidity verifier test
	fmt.Println("\n--- Exporting Solidity Verifier Test ---")
	verifierTestFile, err := os.Create("../solidity/test/Verifier.t.sol")
	if err != nil {
		return fmt.Errorf("Error while creating the file Verifier.t.sol: %w\n", err)
	}
	defer verifierTestFile.Close()

	// header
	verifierTestFile.Write([]byte(`// SPDX-License-Identifier: UNLICENSED
	pragma solidity ^0.8.25;

	import {Test, console} from "forge-std/Test.sol";
	import {PlonkVerifier} from "../src/Verifier.sol";

	contract VerifierTest is Test {
	    PlonkVerifier ZkK1;

	    function setUp() public {
	        ZkK1 = new PlonkVerifier();
	    }

	    function test_k1Plonk() public view {
	`))

	Proof := proof.(*plonk_bn254.Proof)
	verifierTestFile.Write([]byte(`bytes memory proof = hex"` + hexutil.Encode(Proof.MarshalSolidity())[2:] + `";`))
	verifierTestFile.Write([]byte("\n"))

	PI := fmt.Sprintf("%v", publicWitness.Vector())

	verifierTestFile.Write([]byte("uint256[5] memory public_inputs = " + PI + ";\n"))

	// footer
	verifierTestFile.Write([]byte(`
	        uint256[] memory inputs = new uint256[](5);
	        for (uint i = 0; i < 5; i++) inputs[i] = uint256(public_inputs[i]);

	        bool res = ZkK1.Verify(proof, inputs);
	        assertTrue(res);
	        console.log(res);
	    }
	}
	`))
	fmt.Println("Successfully exported solidty/test/Verifier.t.sol")

	fmt.Print("\n\n\n=======================\nPROOF and PUBLIC INPUTS\n=======================\n0x", hexutil.Encode(Proof.MarshalSolidity())[2:], " \"", publicWitness.Vector(), "\"\n")

	return nil
}
