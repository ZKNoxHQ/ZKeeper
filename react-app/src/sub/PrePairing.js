import Button from "./Button.tsx";
import {useState, useRef} from "react";
import {execHaloCmdWeb} from "@arx-research/libhalo/api/web";
import {computeAddress, JsonRpcProvider} from "ethers";
import {EIP155_CHAINS} from "../logic/EIP155Chains";

import { createSmartAccountClient } from "permissionless"
import { createPimlicoClient } from "permissionless/clients/pimlico"
import { http, createPublicClient, zeroAddress, encodeFunctionData, formatEther, parseEther, hashTypedData, encodeAbiParameters, parseSignature, isAddress } from "viem"
import { generatePrivateKey, privateKeyToAccount } from "viem/accounts"
import { sepolia } from "viem/chains"
import { toSimpleSmartAccount } from "permissionless/accounts"
import { entryPoint08Address, getUserOperationTypedData } from "viem/account-abstraction"
import { hashAuthorization } from "ethers/hash";
import { Falcon } from "../falcon.js";

import { dmk } from "../dmk.js";
import { DeviceStatus, DeviceActionStatus } from "@ledgerhq/device-management-kit";
import { SignerEthBuilder } from "@ledgerhq/device-signer-kit-ethereum";

const DEFAULT_SIGNING_PATH = "44'/60'/0'/0'/0"

const TEST_CHAIN_ID = "eip155:11155111"
let hasCode = "";

const pimlicoUrl = "https://api.pimlico.io/v2/11155111/rpc?apikey=pim_YxNShDyzJSnL1Ur2fp6KSM";
const pimlicoSponsorshipPolicyId = "sp_flashy_vector";
const sepoliaUrl = "https://ethereum-sepolia-rpc.publicnode.com";
const abi = [
    {
      "inputs": [
        {
          "internalType": "address",
          "name": "secondary",
          "type": "address"
        }
      ],
      "name": "initSecondary",
      "outputs": [],
      "stateMutability": "nonpayable",
      "type": "function"
    }    
]

const falconContract = "0x13b79503ED87a507551160a9E57FdBf46e6Fa444";
const iPQPublicKeyContract = "0xeAb06b810F3ECa9f3D00bad3Fd286A04ab03B3Db";
const falconDelegateContract = "0x78898a02c0ef4B5Fa1a15929997FDcDA758EE815";
const falconShakeId = 0x216840110134321;
const abiInitFalconDelegate = [
{
    "inputs": [
      {
        "internalType": "uint256",
        "name": "iAlgoID",
        "type": "uint256"
      },
      {
        "internalType": "address",
        "name": "iCore",
        "type": "address"
      },
      {
        "internalType": "address",
        "name": "iAuthorized_ECDSA",
        "type": "address"
      },
      {
        "internalType": "address",
        "name": "iPublicPQKey",
        "type": "address"
      }
    ],
    "name": "initialize",
    "outputs": [],
    "stateMutability": "nonpayable",
    "type": "function"
  }    
]

const falconDelegateContractZK = "0xBFF9BBC799a93b344ededE632e5e6c8d5Ef3e7Cb";
const ZKVerifierContract = "0x720915F3843C5c72deaa5bb09a73757dd3E7592E";
const abiInitFalconDelegateZK = [
{
    "inputs": [
      {
        "internalType": "uint256",
        "name": "iAlgoID",
        "type": "uint256"
      },
      {
        "internalType": "address",
        "name": "iCore",
        "type": "address"
      },
      {
        "internalType": "address",
        "name": "iZkVerifier",
        "type": "address"
      },
      {
        "internalType": "uint256",
        "name": "iPublic_key_commitment",
        "type": "uint256"
      },      
      {
        "internalType": "address",
        "name": "iPublicPQKey",
        "type": "address"
      }
    ],
    "name": "initialize",
    "outputs": [],
    "stateMutability": "nonpayable",
    "type": "function"
  }    
]

// TODO user entry
const pqSeed = "";


function PrePairing({haloAddress, haloCode, onGetHalo, onGetHaloCode, onStartPairing, onResetWallet}) {

    let [destinationAddress, setDestinationAddress] = useState('');
    let [accountBalance, setAccountBalance] = useState('');
    let [accountCode, setAccountCode] = useState('');
    let [amount, setAmount] = useState('');
    let [ledgerAddress, setLedgerAddress] = useState('');
    let [ledgerSessionId, setLedgerSessionId] = useState('');
    let [zkMode, setZkMode] = useState('');
    let [publicKeyForProof, setPublicKeyForProof] = useState('');
    let [publicKeyCommitment, setPublicKeyCommitment] = useState('');
    let [proof, setProof] = useState('');
    let [proofData, setProofData] = useState('');
    let [devicePublicKey, setDevicePublicKey] = useState('');
    let [useropNonce, setUseropNonce] = useState('');

    let lastUserOp = useRef(null);
    let lastSmartAccountClient = useRef(null);
    let lastTypedDataHash = useRef(null);


    async function updateAccount(address) {
      const provider = new JsonRpcProvider(EIP155_CHAINS[TEST_CHAIN_ID].rpc);
      try {
        let code = await provider.getCode(address);devicePublicKey
        setAccountCode(code);
      } catch (e) {
        alert("error getCode");
        alert(e.toString());
      }
      try {
        let accountBalance = await provider.getBalance(address);
        setAccountBalance("" + accountBalance);
      } catch(e) {
        alert("error getBalance");
        alert(e.toString());
      }
    }

    async function getNonce(address) {
      const provider = new JsonRpcProvider(EIP155_CHAINS[TEST_CHAIN_ID].rpc);
      try {
        let nonce = await provider.getTransactionCount(address);
        return nonce;
      } catch (e) {
        alert("error getNonce");
        alert(e.toString());
      }
    }

    function evButtonDownDestinationAddress(ev) {
    }

    function evButtonDownAmount(ev) {        
    }

    function evButtonDownProofData(ev) {
    }

    function evButtonDownProof(ev) {        
    }

    function evButtonDownPublicKeyForProof(ev) {        
    }

    function evButtonDownPublicKeyCommitment(ev) {        
    }

    function btnPair() {
        onStartPairing();
    }

    function btnResetWallet() {
        onResetWallet();
    }

    async function btnZkMode() {
        setZkMode(!zkMode);
    }

    function isDelegated(code) {
        if (!code) {
            return false;
        }
        let expectedDelegate = (zkMode ? falconDelegateContractZK : falconDelegateContract);
        return code.toUpperCase() == ("0xef0100" + expectedDelegate.substring(2)).toUpperCase();
    }

    async function ledgerSign7702(signer, path, nonce, chainId, delegate) {
        return new Promise(function(resolve, reject) {
            signer.signDelegationAuthorization(path, chainId, delegate, nonce).observable.subscribe((getDelegationDAState) => {
                switch(getDelegationDAState.status) {
                    case DeviceActionStatus.Completed:
                        console.log(getDelegationDAState.output);
                        resolve(getDelegationDAState.output);
                        break;
                    case DeviceActionStatus.Error:
                        console.log(getDelegationDAState.error);
                        reject(getDelegationDAState.error);
                        break;
                }
            });
        });
    }

    async function ledgerSignTypedData(signer, path, typedData) {
        return new Promise(function(resolve, reject) {
            signer.signTypedData(path, typedData).observable.subscribe((getDelegationDAState) => {
                switch(getDelegationDAState.status) {
                    case DeviceActionStatus.Completed:
                        console.log(getDelegationDAState.output);
                        resolve(getDelegationDAState.output);
                        break;
                    case DeviceActionStatus.Error:
                        console.log(getDelegationDAState.error);
                        reject(getDelegationDAState.error);
                        break;
                }
            });
        });
    }

    async function btnSendTx() {

        // Recover partial ZK stage signature if left pending

        if (lastUserOp.current && zkMode) {

            if (!proof) {
                alert("Missing proof");
                return;
            }

            console.log("Finalizing ZK signature");

            const pqSignature = await falconSign(pqSeed, lastTypedDataHash.current)

            let splitProof = proof.split(" ");
            let signatureFull = encodeAbiParameters(
                [
                    { name: 'proof', type: 'bytes'},
                    { name: 'public_inputs', type: 'int256[]'},
                    { name: 'pq', type: 'bytes' }
                ],
                [ splitProof[0], JSON.parse(splitProof[1]), pqSignature ]
            )

            const userOpHash = await lastSmartAccountClient.current.sendUserOperation({
                ...lastUserOp.current,
                signature: signatureFull
            })

            console.log("User Operation Hash:")
            console.log(userOpHash)

            const transactionHash =
                await lastSmartAccountClient.current.waitForUserOperationReceipt({
                    hash: userOpHash
                })

            console.log("Transaction Hash:")
            console.log(transactionHash.receipt.transactionHash)

            let src;
            if (ledgerAddress) {
                src = ledgerAddress;
            }
            else {
                src = haloAddress;
            }

            updateAccount(src);

            return;
        }

        lastUserOp.current = null;
        lastSmartAccountClient.current = null;
        lastTypedDataHash.current = null;

        // Verify parameters

        let usingLedger = false;
        let src;
        let signerLedger;
        if (ledgerAddress) {
            usingLedger = true;
            src = ledgerAddress;
            console.log("Using session id " + ledgerSessionId);
            signerLedger = new SignerEthBuilder({dmk: dmk, sessionId: ledgerSessionId, originToken:"test" }).build();                
        }
        else if (haloAddress) {
            src = haloAddress;
        }
        else {
            alert("No wallet connected");
            return;
        }
        if (!destinationAddress) {
            alert("No destination address");
            return;
        }   
        if (!isAddress(destinationAddress)) {
            alert("Invalid destination address");
            return;
        }
        if (!amount) {
            alert("No amount");
            return;
        }
        let amountNumber;
        try {
            amountNumber = parseEther(amount);
        }
        catch(e) {
            alert("Invalid amount");
            return;
        }
        if (amountNumber > accountBalance) {
            alert("Amount too big");
            return;
        }

        // Private key signature is overridden by the device

        const privateKey = generatePrivateKey()
        console.log("Using private key " + privateKey);

 const publicClient = createPublicClient({
        chain: sepolia,
        transport: http(sepoliaUrl)
    })

    const pimlicoClient = createPimlicoClient({
        transport: http(pimlicoUrl)
    })

    const owner = privateKeyToAccount(privateKey)
    owner.address = src;

    console.log(`Owner address: ${owner.address}`)

    const simpleSmartAccount = await toSimpleSmartAccount({
        owner,
        entryPoint: {
            address: entryPoint08Address,
            version: "0.8"
        },
        client: publicClient,
        address: owner.address
    })

    console.log(`Smart account address: ${simpleSmartAccount.address}`)

    // Create the smart account client
    const smartAccountClient = createSmartAccountClient({
        account: simpleSmartAccount,
        chain: sepolia,
        bundlerTransport: http(pimlicoUrl),
        paymaster: pimlicoClient,
        userOperation: {
            estimateFeesPerGas: async () => {
                return (await pimlicoClient.getUserOperationGasPrice()).fast
            }
        }
    })

    let initParameters;
    if (zkMode) {
        if (!isDelegated(accountCode) && !publicKeyCommitment) {
            alert("Missing Public Key Commitment");
            return;
        }
        initParameters = {
            abi: abiInitFalconDelegateZK,
            functionName: "initialize",
            args: [ falconShakeId, falconContract, ZKVerifierContract, (!publicKeyCommitment ? 0 : publicKeyCommitment), iPQPublicKeyContract ]                
        }
    }
    else {
        initParameters = {
            abi: abiInitFalconDelegate,
            functionName: "initialize",
            args: [ falconShakeId, falconContract, owner.address, iPQPublicKeyContract ]                
        }
    }

    const factoryData = encodeFunctionData(initParameters);

    let nonce = await getNonce(owner.address);
    let authorizationSignature;

    const authorization = {
        address: (zkMode ? falconDelegateContractZK : falconDelegateContract),
        chainId: sepolia.id,
        nonce: nonce        
    }

    if (!isDelegated(accountCode)) {
        if (usingLedger) {
            authorizationSignature = await ledgerSign7702(signerLedger, DEFAULT_SIGNING_PATH, nonce, sepolia.id, (zkMode ? falconDelegateContractZK : falconDelegateContract));
            // rename for viem
            authorizationSignature.yParity = authorizationSignature.v;
         }
        else {
            const authorizationHash = hashAuthorization(authorization)
            console.log(authorizationHash);

            let res;

            try {
                res = await execHaloCmdWeb({
                    "name": "sign",
                    "keyNo": 1,
                    "digest": authorizationHash.substring(2)
                });
            } catch (e) {
                alert(e);
                throw e;
            }

            authorizationSignature = parseSignature(res.signature.ether);
        }
    }
    else {
        authorizationSignature = {};
    }

    let userOpData = {
        calls: [
            {
                to: destinationAddress,
                data: "0x",
                value: amountNumber
            }
        ],
        paymasterContext: {
            sponsorshipPolicyId: pimlicoSponsorshipPolicyId
        },
        factory: '0x7702',
        factoryData: factoryData,
        paymasterVerificationGasLimit: 9000000,
        preVerificationGas : 5000000,
        verificationGasLimit : 9000000,
        authorization: {
            ...authorization,
            ...authorizationSignature
        }
    }

    if (isDelegated(accountCode)) {
        delete userOpData.factory;
        delete userOpData.factoryData;
        delete userOpData.authorization;
    }

    const userOp = await smartAccountClient.prepareUserOperation(userOpData);

    console.log("User Operation:")
    console.log(userOp)

    console.log("Signature direct " + await simpleSmartAccount.signUserOperation(userOp));
    const typedData = getUserOperationTypedData({
        chainId: publicClient.chain.id,
        entryPointAddress: entryPoint08Address,
        userOperation: {
            ...userOp,
            sender: await simpleSmartAccount.getAddress()
        }
    })
    const typedDataHash = hashTypedData(typedData);
    console.log("Typed Data hash " + typedDataHash)

    let signature1;

    // Handle ZK Mode

    // Compute the signature if no proof data is available

    if (!zkMode || (zkMode && !proof)) {

        if (usingLedger) {
            signature1 = await ledgerSignTypedData(signerLedger, DEFAULT_SIGNING_PATH, typedData);
            // rename for viem
            signature1.yParity = signature1.v;        
        }
        else {
            let res;

            try {
                res = await execHaloCmdWeb({
                    "name": "sign",
                    "keyNo": 1,
                    "digest": typedDataHash.substring(2)
                });
            } catch (e) {
                alert(e);
                throw e;
            }

            signature1 = parseSignature(res.signature.ether);
        }
    }

    // Provide information to provide the proof in ZK Mode 

    if (zkMode && !proof) {
        let tmp = typedDataHash.substring(2) + " " +
            signature1.r.substring(2) + " " +
            signature1.s.substring(2) + " " +
            devicePublicKey.substring(2, 2 + 32 + 32) + " " + 
            devicePublicKey.substring(2 + 32 + 32);
        setProofData(tmp);
        console.log("Proof data");
        console.log(tmp);
        lastUserOp.current = userOp;
        lastSmartAccountClient.current = smartAccountClient;
        lastTypedDataHash.current = typedDataHash;
        return;
    }

    console.log("Signature side 1");
    console.log(signature1);

    const pqSignature = await falconSign(pqSeed, typedDataHash)
    console.log("PQ signature " + pqSignature)

    let signatureFull;

    if (!zkMode) {
        signatureFull = encodeAbiParameters(
            [
                { name: 'v', type: 'uint8' },
                { name: 'r', type: 'bytes32' },
                { name: 's', type: 'bytes32' },
                { name: 'pq', type: 'bytes' }
            ],
            [ signature1.v, signature1.r, signature1.s, pqSignature ]
        )
    }

    const userOpHash = await smartAccountClient.sendUserOperation({
        ...userOp,
        signature: signatureFull
    })

    console.log("User Operation Hash:")
    console.log(userOpHash)

    const transactionHash =
        await smartAccountClient.waitForUserOperationReceipt({
            hash: userOpHash
        })

    console.log("Transaction Hash:")
    console.log(transactionHash.receipt.transactionHash)

    updateAccount(src);
}

// Global variables to store subscriptions and session info
let discoverySubscription;
let stateSubscription;
let currentSessionId;
 
function startDiscoveryAndConnect() {
  // Clear any previous discovery
  if (discoverySubscription) {
    discoverySubscription.unsubscribe();
  }
  cleanup();
 
  console.log("Starting device discovery...");
 
  // Start discovering - this will scan for any connected devices
  discoverySubscription = dmk.startDiscovering({}).subscribe({
    next: async (device) => {
      console.log(
        `Found device: ${device.id}, model: ${device.deviceModel.model}`,
      );
 
      // Connect to the first device we find
      try {
        // Pass the full device object, not just the ID
        currentSessionId = await dmk.connect({ device });
        setLedgerSessionId(currentSessionId);
        console.log(`Connected! Session ID: ${currentSessionId}`);
 
        // Stop discovering once we connect
        discoverySubscription.unsubscribe();
 
        // Get device information
        const connectedDevice = dmk.getConnectedDevice({
          sessionId: currentSessionId,
        });
        console.log(`Device name: ${connectedDevice.name}`);
        console.log(`Device model: ${connectedDevice.modelId}`);
 
        const signerEth = new SignerEthBuilder({dmk: dmk, sessionId: currentSessionId, originToken:"test" }).build();        
        signerEth.getAddress(DEFAULT_SIGNING_PATH).observable.subscribe((getAddressDAState) => {
            switch(getAddressDAState.status) {
                case DeviceActionStatus.Completed:
                    console.log(getAddressDAState.output);
                    setLedgerAddress(getAddressDAState.output.address);
                    setDevicePublicKey(getAddressDAState.output.publicKey);
                    setPublicKeyForProof(getAddressDAState.output.publicKey.substring(2, 2 + 32 + 32));
                    updateAccount(getAddressDAState.output.address);                    
                    break;
                case DeviceActionStatus.Error:
                    console.log(getAddressDAState.error);
                    break;
            }
        });
        /*
        // Start monitoring device state
        stateSubscription = monitorDeviceState(currentSessionId);
        */
      } catch (error) {
        console.error("Connection failed:", error);
      }
    },
    error: (error) => {
      console.error("Discovery error:", error);
    },
  });
}
 
function monitorDeviceState(sessionId) {
  return dmk.getDeviceSessionState({ sessionId }).subscribe({
    next: (state) => {
      console.log(`Device status: ${state.deviceStatus}`);
 
      // Check for specific status conditions
      if (state.deviceStatus === DeviceStatus.LOCKED) {
        console.log("Device is locked - please enter your PIN");
      }
 
      // Show battery level if available
      if (state.batteryStatus) {
        console.log(`Battery level: ${state.batteryStatus.level}%`);
      }
 
      // Show app information if available
      if (state.currentApp) {
        console.log(`Current app: ${state.currentApp.name}`);
        console.log(`App version: ${state.currentApp.version}`);
      }
 
      // Basic device model info
      console.log(`Device model: ${state.deviceModelId}`);
    },
    error: (error) => {
      console.error("State monitoring error:", error);
    },
  });
}
 
// Always clean up resources when done
async function cleanup() {
  // Unsubscribe from all observables
  if (discoverySubscription) {
    discoverySubscription.unsubscribe();
  }
 
  if (stateSubscription) {
    stateSubscription.unsubscribe();
  }
 
  // Disconnect from device if connected
  if (currentSessionId) {
    try {
      await dmk.disconnect({ sessionId: currentSessionId });
      console.log("Device disconnected successfully");
      currentSessionId = null;
      setLedgerSessionId('');
    } catch (error) {
      console.error("Disconnection error:", error);
    }
  }
  setLedgerAddress("");
  setAccountCode(null);
  setAccountBalance(null);
}

    async function btnConnectLedger() {
        startDiscoveryAndConnect();
    }

    async function btnDelegateLedger() {
        startDiscoveryAndConnect();
    }


  async function falconSign(seedHexInput, messageHexInput) {

    const seed = Buffer.from(seedHexInput, "hex");
    const message = Buffer.from(messageHexInput.startsWith("0x") ? messageHexInput.slice(2) : messageHexInput, "hex");

// --- Main FALCON Module execution ---
return Falcon().then((falcon) => {
  const pkLen = 897;
  const skLen = 1281;
  const sigMaxLen = 690;
  const seedLen = 32;

  // Allocate memory for key pair
  const pkPtr = falcon._malloc(pkLen);
  const skPtr = falcon._malloc(skLen);
  const seedPtr = falcon._malloc(seedLen);

  falcon.HEAPU8.set(seed, seedPtr);

  // Generate keypair using the provided seed
  falcon.ccall(
    'crypto_keypair',
    'number',
    ['number', 'number', 'number'],
    [pkPtr, skPtr, seedPtr]
  );

  const publicKey = Buffer.from(falcon.HEAPU8.subarray(pkPtr, pkPtr + pkLen));

  // --- Normal / Signature Output Path ---
  // Allocate memory for message
  const msgPtr = falcon._malloc(message.length);
  falcon.HEAPU8.set(message, msgPtr);


  // Conditional console logs for human readability
      console.log("🔑 Message (hex):", message.toString("hex"));
      const secretKey = Buffer.from(falcon.HEAPU8.subarray(skPtr, skPtr + skLen));
      console.log("🔑 Secret Key (hex):", secretKey.toString("hex"));
      console.log("🔑 Public Key (base64):", publicKey.toString("base64"));
      console.log("🔑 Public Key (hex):", publicKey.toString("hex")); // Full output includes hex PK


  // Sign the message
  const signedMsgMaxLen = message.length + sigMaxLen;
  const signedMsgPtr = falcon._malloc(signedMsgMaxLen);
  const signedMsgLenPtr = falcon._malloc(8); // 64-bit space

  const signRet = falcon._crypto_sign(
    signedMsgPtr,
    signedMsgLenPtr,
    msgPtr,
    BigInt(message.length),
    skPtr
  );

  if (signRet !== 0) {
    console.error("❌ Signing failed.");
    // Free memory before exiting on error
    [pkPtr, skPtr, msgPtr, seedPtr, signedMsgPtr, signedMsgLenPtr].forEach(ptr => falcon._free(ptr));
    return;
  }

  // Read 64-bit signature length (low + high)
  function readUint64(ptr) {
    const low = falcon.HEAPU32[ptr >> 2];
    const high = falcon.HEAPU32[(ptr >> 2) + 1];
    return BigInt(high) << 32n | BigInt(low);
  }

  const sigLen = Number(readUint64(signedMsgLenPtr));
  const signedMessage = Buffer.from(falcon.HEAPU8.subarray(signedMsgPtr, signedMsgPtr + sigLen));

      console.log("✅ Signature generated.");
      console.log("🔐 Sig+Msg (base64):", signedMessage.toString("base64"));

      console.log(signedMessage.toString("hex"));

  // Free all remaining memory
  [pkPtr, skPtr, msgPtr, seedPtr, signedMsgPtr, signedMsgLenPtr]
    .forEach(ptr => falcon._free(ptr));

  return "0x" + signedMessage.toString("hex");
  }
)}

  async function btnGetHalo() {
      let addr;

      let pkeys;

      try {
        pkeys = await execHaloCmdWeb({
            "name": "get_pkeys"
        });
      } catch (e) {
        alert(e.toString());
        return;
      }

      //alert(JSON.stringify(pkeys));

      let result;

      /*
      try {
        result = await execHaloCmdWeb({
            "name": "gen_key",
            "keyNo": 0x61,
            "entropy": "bea1b1544a407a1915b501e9590aafc303265817bd7f43f8dbdf8031513d60cc"

        });
      } catch (e) {
        alert(e.toString());
        return;
      }

      alert(JSON.stringify(result));
      

      try {
        result = await execHaloCmdWeb({
            "name": "sign",
            "keyNo": 0x62,
            "digest": "0000000000000000000000000000000000000000000000000000000000000000"

        });
      } catch (e) {
        alert(e.toString());
        return;
      }

      alert(JSON.stringify(result));
      */

      addr = computeAddress('0x' + pkeys.publicKeys[1]);
      setDevicePublicKey('0x' + pkeys.publicKeys[1]);
      setPublicKeyForProof(pkeys.publicKeys[1].substring(0, 32 + 32));
      onGetHalo(addr)
      updateAccount(addr);

      const provider = new JsonRpcProvider(EIP155_CHAINS[TEST_CHAIN_ID].rpc);
      try {
        let code = await provider.getCode(addr);
        onGetHaloCode(code)
      } catch (e) {
        alert(e.toString());
        return;
      }
   }

    let displayedAddress = "";
    let walletType = "";
    if (ledgerAddress) {
        displayedAddress = ledgerAddress;
        walletType = "Ledger"
    }
    else
    if (haloAddress) {
        displayedAddress = haloAddress;
        walletType = "HaLo";
    }
    else {
        displayedAddress = "None";
        walletType = "";
    }
    let delegationType;
    if (isDelegated(accountCode)) {
        delegationType = "DELEGATED";
    }
    else
    if (accountCode == "0x") {
        delegationType = "EOA";
    }
    else {
        delegationType = accountCode;
    }
    let balance = "";
    if (accountBalance) {
        balance = formatEther(parseInt(accountBalance)) + " ETH";
    }

    return (
        <div>
            <div style={{marginBottom: '40px'}}>
                <p className={"label-text"}>
                    Active Device:
                </p>
                <p style={{textTransform: 'none', color: 'white', fontFamily: 'monospace', fontSize: 12}}>
                    {displayedAddress} {walletType} {delegationType} {balance}
                </p>
            </div>

        <div style={{marginTop: '20px', marginBottom: '10px'}}>
            <p className={"label-text"}>Destination address:</p>
        </div>
          <input
              type="text"
              value={destinationAddress}
              className={"text-field"}
              onKeyDown={(ev) => evButtonDownDestinationAddress(ev)}
              onChange={(ev) => setDestinationAddress(ev.target.value)}
          />
        <div style={{marginTop: '20px', marginBottom: '10px'}}>
            <p className={"label-text"}>Amount (ETH):</p>
        </div>
          <input
              type="text"
              value={amount}
              className={"text-field"}
              onKeyDown={(ev) => evButtonDownAmount(ev)}
              onChange={(ev) => setAmount(ev.target.value)}
          />

            <Button onClick={() => btnSendTx()} fullWidth={true} className={"btn-pad"}>Send</Button>
            {!haloAddress ? 
                <Button onClick={() => btnGetHalo()} fullWidth={true} className={"btn-pad"}>Connect HaLo</Button> :
                <Button onClick={() => btnResetWallet()} fullWidth={true} className={"btn-pad"}>Reset HaLo</Button> 
            }
            <Button onClick={() => btnConnectLedger()} fullWidth={true} className={"btn-pad"}>Connect Ledger</Button>

        <div style={{marginTop: '40px', marginBottom: '10px'}}>
            <Button onClick={() => btnZkMode()} fullWidth={true} className={"btn-pad"}>ZK MODE</Button>
        </div>

        <div style={{display: (zkMode ? 'block' : 'none')}}>

        <div style={{marginTop: '20px', marginBottom: '10px'}}>
            <p className={"label-text"}>Public Key for proof :</p>
        </div>
          <input
              type="text"
              value={publicKeyForProof}
              className={"text-field"}
              onKeyDown={(ev) => evButtonDownPublicKeyForProof(ev)}
              onChange={(ev) => setPublicKeyForProof(ev.target.value)}
          />
        <div style={{marginTop: '20px', marginBottom: '10px'}}>
            <p className={"label-text"}>Public Key Commitment :</p>
        </div>
          <input
              type="text"
              value={publicKeyCommitment}
              className={"text-field"}
              onKeyDown={(ev) => evButtonDownPublicKeyCommitment(ev)}
              onChange={(ev) => setPublicKeyCommitment(ev.target.value)}
          />
        
        <div style={{marginTop: '20px', marginBottom: '10px'}}>
            <p className={"label-text"}>Proof data:</p>
        </div>
          <input
              type="text"
              value={proofData}
              className={"text-field"}
              onKeyDown={(ev) => evButtonDownProofData(ev)}
              onChange={(ev) => setProofData(ev.target.value)}
          />
        <div style={{marginTop: '20px', marginBottom: '10px'}}>
            <p className={"label-text"}>Proof:</p>
        </div>
          <input
              type="text"
              value={proof}
              className={"text-field"}
              onKeyDown={(ev) => evButtonDownProof(ev)}
              onChange={(ev) => setProof(ev.target.value)}
          />
        </div>
        
        </div>

    );
}

export default PrePairing;

