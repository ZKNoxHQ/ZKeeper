/**
 *
 */
/*ZZZZZZZZZZZZZZZZZZZKKKKKKKKK    KKKKKKKNNNNNNNN        NNNNNNNN     OOOOOOOOO     XXXXXXX       XXXXXXX                         ..../&@&#.       .###%@@@#, ..                         
/*Z:::::::::::::::::ZK:::::::K    K:::::KN:::::::N       N::::::N   OO:::::::::OO   X:::::X       X:::::X                      ...(@@* .... .           &#//%@@&,.                       
/*Z:::::::::::::::::ZK:::::::K    K:::::KN::::::::N      N::::::N OO:::::::::::::OO X:::::X       X:::::X                    ..*@@.........              .@#%%(%&@&..                    
/*Z:::ZZZZZZZZ:::::Z K:::::::K   K::::::KN:::::::::N     N::::::NO:::::::OOO:::::::OX::::::X     X::::::X                   .*@( ........ .  .&@@@@.      .@%%%%%#&@@.                  
/*ZZZZZ     Z:::::Z  KK::::::K  K:::::KKKN::::::::::N    N::::::NO::::::O   O::::::OXXX:::::X   X::::::XX                ...&@ ......... .  &.     .@      /@%%%%%%&@@#                  
/*        Z:::::Z      K:::::K K:::::K   N:::::::::::N   N::::::NO:::::O     O:::::O   X:::::X X:::::X                   ..@( .......... .  &.     ,&      /@%%%%&&&&@@@.              
/*       Z:::::Z       K::::::K:::::K    N:::::::N::::N  N::::::NO:::::O     O:::::O    X:::::X:::::X                   ..&% ...........     .@%(#@#      ,@%%%%&&&&&@@@%.               
/*      Z:::::Z        K:::::::::::K     N::::::N N::::N N::::::NO:::::O     O:::::O     X:::::::::X                   ..,@ ............                 *@%%%&%&&&&&&@@@.               
/*     Z:::::Z         K:::::::::::K     N::::::N  N::::N:::::::NO:::::O     O:::::O     X:::::::::X                  ..(@ .............             ,#@&&&&&&&&&&&&@@@@*               
/*    Z:::::Z          K::::::K:::::K    N::::::N   N:::::::::::NO:::::O     O:::::O    X:::::X:::::X                   .*@..............  . ..,(%&@@&&&&&&&&&&&&&&&&@@@@,               
/*   Z:::::Z           K:::::K K:::::K   N::::::N    N::::::::::NO:::::O     O:::::O   X:::::X X:::::X                 ...&#............. *@@&&&&&&&&&&&&&&&&&&&&@@&@@@@&                
/*ZZZ:::::Z     ZZZZZKK::::::K  K:::::KKKN::::::N     N:::::::::NO::::::O   O::::::OXXX:::::X   X::::::XX               ...@/.......... *@@@@. ,@@.  &@&&&&&&@@@@@@@@@@@.               
/*Z::::::ZZZZZZZZ:::ZK:::::::K   K::::::KN::::::N      N::::::::NO:::::::OOO:::::::OX::::::X     X::::::X               ....&#..........@@@, *@@&&&@% .@@@@@@@@@@@@@@@&                  
/*Z:::::::::::::::::ZK:::::::K    K:::::KN::::::N       N:::::::N OO:::::::::::::OO X:::::X       X:::::X                ....*@.,......,@@@...@@@@@@&..%@@@@@@@@@@@@@/                   
/*Z:::::::::::::::::ZK:::::::K    K:::::KN::::::N        N::::::N   OO:::::::::OO   X:::::X       X:::::X                   ...*@,,.....%@@@,.........%@@@@@@@@@@@@(                     
/*ZZZZZZZZZZZZZZZZZZZKKKKKKKKK    KKKKKKKNNNNNNNN         NNNNNNN     OOOOOOOOO     XXXXXXX       XXXXXXX                      ...&@,....*@@@@@ ..,@@@@@@@@@@@@@&.                     
/*                                                                                                                                   ....,(&@@&..,,,/@&#*. .                             
/*                                                                                                                                    ......(&.,.,,/&@,.                                
/*                                                                                                                                      .....,%*.,*@%                               
/*                                                                                                                                    .#@@@&(&@*,,*@@%,..                               
/*                                                                                                                                    .##,,,**$.,,*@@@@@%.                               
/*                                                                                                                                     *(%%&&@(,,**@@@@@&                              
/*                                                                                                                                      . .  .#@((@@(*,**                                
/*                                                                                                                                             . (*. .                                   
/*                                                                                                                                              .*/
///* Copyright (C) 2025 - Renaud Dubois, Simon Masson - This file is part of ZKNOX project
///* License: This software is licensed under MIT License
///* This Code may be reused including this header, license and copyright notice.
///* See LICENSE file at the root folder of the project.
///* FILE: ZKNOX_falcon.sol
///* Description: Compute NIST compliant falcon verification
/**
 *
 */
// SPDX-License-Identifier: MIT
pragma solidity ^0.8.25;

import "./ZKNOX_common.sol";
import "./ZKNOX_IVerifier.sol";

import "./ZKNOX_falcon_utils.sol";
import {ZKNOX_NTT} from "./ZKNOX_NTT.sol";
import "./ZKNOX_falcon_core.sol";

//choose the XOF to use here
import "./ZKNOX_HashToPoint.sol";
import "./ZKNOX_falcon_encodings.sol";

import "../lib/account-abstraction/contracts/accounts/Simple7702Account.sol";

import "./Verifier.sol";

interface IZKVerifier {
  function Verify(bytes calldata proof, uint256[] calldata public_inputs) external view returns(bool);
}


/// @notice Contract designed for being delegated to by EOAs to authorize a IVerifier key to transact on their behalf.
contract ZKNOX_SimpleHybrid7702ZK is Simple7702Account {

    error AlreadyInitialized();
    error InvalidCaller();

    bytes32 constant SIMPLEHYBRID7702_STORAGE_POSITION = keccak256("zknox.hybrid.7702.zk");

    struct Storage {
        //address of the ZK Proof verifier
        address zkVerifier;
        //commitment to public key
        uint256 public_key_commitment;
   
        /// @notice Address of the contract storing the post quantum public key
        address authorized_PQPublicKey;
        /// @notice Address of the verification contract logic

        address CoreAddress; //address of the core verifier (FALCON, DILITHIUM, etc.), shall be the address of a ISigVerifier
        uint256 algoID;
    }

    function getStorage() internal pure returns (Storage storage ds) {
        bytes32 position = SIMPLEHYBRID7702_STORAGE_POSITION;
        assembly {
            ds.slot := position
        }
    }

    constructor() {}
    
    //input are AlgoIdentifier, Signature verification address, publickey storing contract
    function initialize(uint256 iAlgoID, address iCore, address iZkVerifier, uint256 iPublic_key_commitment, address iPublicPQKey) external {
        if (msg.sender != address(entryPoint().senderCreator())) {
            revert InvalidCaller();
        }
        /*
        if (getStorage().CoreAddress != address(0)) {
            revert AlreadyInitialized();
        }
        */
       
        getStorage().zkVerifier = iZkVerifier; // ZK Proof verifier
        getStorage().public_key_commitment=iPublic_key_commitment;  //public key commitment
        getStorage().CoreAddress = iCore; // Address of contract of Signature verification (FALCON, DILITHIUM)
        getStorage().algoID = iAlgoID;
        getStorage().authorized_PQPublicKey = iPublicPQKey;
    }

    // TODO : replace by proper encoding
    function _validateSignature(
        PackedUserOperation calldata userOp,
        bytes32 userOpHash
    ) internal virtual override returns (uint256 validationData) {
        if ((userOp.signature[0] == 0xff) && (userOp.signature[1] == 0xff)) {
            return SIG_VALIDATION_FAILED;
        }
        (bytes memory proof, uint256[] memory public_inputs, bytes memory sm) = abi.decode(userOp.signature, (bytes, uint256[], bytes));
        return isValid(userOpHash, proof, public_inputs, sm) ? SIG_VALIDATION_SUCCESS : SIG_VALIDATION_FAILED;
    }    


    //proof, public_inputs are input to Verifier, sm is the falcon signature
    function isValid(  
        bytes32 digest,
        bytes memory proof,
        uint256[] memory public_inputs, 
        bytes memory sm // the signature in the NIST KAT format, as output by test_falcon.js
        ) public returns (bool)
        {
            uint256 slen = (uint256(uint8(sm[0])) << 8) + uint256(uint8(sm[1]));
            uint256 mlen = sm.length - slen - 42;

            bytes memory message;
            bytes memory salt = new bytes(40);

        for (uint i = 0; i < 40; i++) {
          salt[i] = sm[i + 2];
         }
        message = new bytes(mlen);
        for (uint256 j = 0; j < mlen; j++) {
          message[j] = sm[j + 42];
        }

         if (sm[2 + 40 + mlen] != 0x29) {
             return false;
         }

        uint256[] memory s2 =_ZKNOX_NTT_Compact((_decompress_sig(sm, 2 + 40 + mlen + 1)));

         ISigVerifier Core = ISigVerifier(getStorage().CoreAddress);

         uint256[] memory nttpk;
         
         // verify commitment
         if (getStorage().public_key_commitment == uint256(0)) {
            return false;
         }
         if (public_inputs[public_inputs.length - 1] != getStorage().public_key_commitment) {
            return false;
         }
         // verify ZK proof
         IZKVerifier zkVerifier = IZKVerifier(getStorage().zkVerifier);
         if (!zkVerifier.Verify(proof, public_inputs)) {
            return false;
         }

         if (getStorage().authorized_PQPublicKey == address(0)) {
            return false;
         }
         
         nttpk = Core.GetPublicKey(getStorage().authorized_PQPublicKey);

         if (!Core.verify(abi.encodePacked(digest), salt, s2, nttpk)) {
            return false;
         }
    
         return true;

        }
    


    function GetPublicKey() public view returns (uint256[] memory res) {
        ISigVerifier Core = ISigVerifier(getStorage().CoreAddress);
        res = Core.GetPublicKey(getStorage().authorized_PQPublicKey);
    }

    function GetStorage() public view returns (address, address) {
        return (getStorage().CoreAddress, getStorage().authorized_PQPublicKey);
    }
    //receive() external payable {}
} //end contract