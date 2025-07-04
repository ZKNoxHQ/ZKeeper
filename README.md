# ZKeeper: EIP7702 for Post quantum 


### Role-based access control to protect your Ethereum Accounts 

<p align="center">
  <img src="https://github.com/user-attachments/assets/fb63c8fc-f103-438d-b609-038cb448638f" alt="The PQ King" width="350"/>
</p>
<p align="center">
  
</p>

-----

## 🚀 Description

This projects demonstrates role based access control enforced by EIP-7702.
RBAC is a model for authorizing end-user access to systems. Zkeeper RBAC prevents hacks like Bybit, which are also possible bc. same keys are used for "sudo"  cmd (configuration, admin power) and basic commands.
By separating **physically** admin functions (using different devices), user are protected from misbehavior, or trapped UI.

### How does this work ?

The sudo account is protected by FALCON signatures, why standard commands are signed by ecdsa. At the entrance of Zkeeper, an analyzer estimates the level (sudo, standard) of the transaction. Then it is forwarded to be signed by the right role.

### What is serious and what is mocked ?

#### Accomplished work

#### Mocked parts

The analysis of transactions is mocked by a simple analysis of the amount of the transaction. In the future, a service like blockAID or similar, instead of being limited to Go/noGO shall provide the role required to execute the transaction.


-----

## ✨ What Will Be Demonstrated

Attendees will witness:

  * **Wristband-as-Signer:** How ETHPRAGUE wristbands, powered by ARX chips, function as direct Ethereum transaction signers.
  * **Post-Quantum Resilience:** The signature is transmitted to the wallet, which hybridates the signature with a **FALCON512 Post-Quantum signer** within a **7702 Smart Account**, showcasing practical quantum resistance.
  * **Hybrid Security:** A practical implementation of a hybrid account protecting funds with both legacy (ECDSA) and post-quantum cryptography.

-----

## 🛠️ How It Works

Our solution builds upon the new EIP-7702 standard to create a flexible smart account. When a transaction needs to be signed:

1.  **Wristband Interaction:** The wristband's ARX chip securely generates a signature using its embedded key.
2.  **Post-Quantum Signing:** The signature is then processed with the FALCON post-quantum algorithm (via JavaScript integration).
3.  **Hybrid Verification (On-Chain):** The `ZKNOX_hybrid.sol` smart contract on Ethereum verifies both the traditional ECDSA signature (from the wristband) and the FALCON post-quantum signature. This dual-verification ensures the account is protected against both classical and quantum-era threats.

-----

