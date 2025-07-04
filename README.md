# ZKeeper: EIP7702 for Post quantum 


### Role-based access control to protect your Ethereum Accounts 

<p align="center">
  <img src="https://github.com/user-attachments/assets/fb63c8fc-f103-438d-b609-038cb448638f" alt="The PQ King" width="350"/>
</p>
<p align="center">
  <small>(The PQ KING)</small>
</p>

-----

## 🚀 Description


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

