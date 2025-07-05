# Zkipper: EIP7702 for Post quantum Role Access


### Role-based access control to protect your Ethereum Accounts 

<p align="center">
  <img src="https://github.com/user-attachments/assets/fb63c8fc-f103-438d-b609-038cb448638f" alt="The PQ King" width="350"/>
</p>
<p align="center">
  
</p>

-----

## 🚀 Description

This projects demonstrates role based access control enforced by EIP-7702.
RBAC is a model for authorizing end-user access to systems. Zkipper RBAC prevents hacks like Bybit, which are also possible bc. same keys are used for "sudo"  cmd (configuration, admin power) and basic commands.
By separating **physically** admin functions (using different devices), user are protected from misbehavior, or trapped UI.


-----

## ✨ What Will Be Demonstrated

Attendees will witness:

  * **Wristband-as-Signer:** How ETHCC wristbands, powered by ARX chips, function as direct Ethereum transaction signers.
  * **7702 Delegation:** The wristband/Ledger will delegate its eOA to the Zkipper contract 
  * **Post-Quantum Resilience:** The signature is transmitted to the wallet, which hybridates the signature with a **FALCON512 Post-Quantum signer** within a **7702 Smart Account**, showcasing practical quantum resistance.
  * **Role Based Transaction Signing:** Transaction are analyzed and routed to right account (admin or user).

-----

## 🛠️ How It Works

Our solution builds upon the new EIP-7702 standard to create a flexible smart account. When a transaction needs to be signed:

1.  **Role Identification:**  At entrance of contract, transactions are designed a role for the signature, according to their criticity.
2.  **ZkSafe:** The governance model is hidden by a ZkProof Verification of the ARX wrist signer.
3.  **Wristband Interaction:** The wristband's ARX chip securely generates a signature using its embedded key, it is used to generate the witnesses of the zkProof (ECDSA over k1).
4.  **Post-Quantum Signing:** The signature is then processed with the FALCON post-quantum algorithm (via JavaScript integration).

![image](https://github.com/user-attachments/assets/59332950-bed2-4a5b-8f8b-7c280d509c89)


The sudo account is protected by FALCON signatures, why standard commands are signed by ecdsa. At the entrance of Zkipper, an analyzer estimates the level (sudo, standard) of the transaction. Then it is forwarded to be signed by the right role.

### What is serious and what is mocked ?

#### Accomplished work
- mopro-gnark: gnark circuits have been binded in rust, and are used for the zkSafe module
- EIP7702: the smart Account integrates ZKNOX FALCON verification and the above verifier for the RBAC

#### Mocked parts

- The analysis of transactions is mocked by a simple analysis of the amount of the transaction. In the future, a service like blockAID or similar, instead of being limited to Go/noGO shall provide the role required to execute the transaction. For instance any delegate call could be detected and require admin (sudo) rights.
- The ZK verifier only takes one signer, in the future any k out of m circuit can be used instead.

-----

## 🔮 ZKproof
![image](https://github.com/user-attachments/assets/c2f0f078-b434-4b17-9f2b-4ac7a634a116)

The current assessment proved is "I know a preimage of commitment = H(Kpub, Nonce) with the same key related to the verification of the input message hash", to commit the public key in the contract without revealing it. This will allow to increase the number of shares, pick a threshold in the future. Currently it is hiding the public key value, and provide a resistance against a trapped HW. In case of loss of the nonce, sudo shall be used to restore a new ZK contract.

private input: nonce, kpub, signature (r,s)

public input: message hash

incircuit verification: h(kpub, nonce)=commitment && ecdsaVerify(kpub, messagehash, r,s)=true


-----
## 🚀 Deployments

- Zircuit: https://explorer.garfield-testnet.zircuit.com/address/0xd70bb0f082FCf522B25592fC8dE8D396e8289544?activeTab=3

-----


