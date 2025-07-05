For generating a json file with the pk commitment:
```
go run derive_pkcommit
```
For computing the proof from the previous json file:
```
go run minimal_plonk
```
This creates a solidity file for testing:
```
cd solidty/
forge test -vvv
```