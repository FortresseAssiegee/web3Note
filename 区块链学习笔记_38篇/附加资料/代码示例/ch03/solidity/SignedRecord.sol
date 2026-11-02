// SPDX-License-Identifier: MIT
pragma solidity 0.8.30;

import {EIP712} from "@openzeppelin/contracts/utils/cryptography/EIP712.sol";
import {SignatureChecker} from "@openzeppelin/contracts/utils/cryptography/SignatureChecker.sol";

// Educational authorization recorder; it does not transfer money or prove a fact is true.
contract SignedRecord is EIP712 {
    struct Approval {
        address signer;
        address executor;
        bytes32 contentHash;
        uint256 nonce;
        uint256 deadline;
    }
    bytes32 public constant APPROVAL_TYPEHASH = keccak256(
        "Approval(address signer,address executor,bytes32 contentHash,uint256 nonce,uint256 deadline)"
    );
    mapping(address => uint256) public nonces;
    event Recorded(address indexed signer, bytes32 indexed contentHash, uint256 nonce);
    error InvalidApproval();

    constructor() EIP712("CryptoNotesRecord", "1") {}

    function digest(Approval calldata a) public view returns (bytes32) {
        return _hashTypedDataV4(keccak256(abi.encode(
            APPROVAL_TYPEHASH, a.signer, a.executor, a.contentHash, a.nonce, a.deadline
        )));
    }

    function record(Approval calldata a, bytes calldata signature) external {
        if (a.signer == address(0) || a.executor != msg.sender ||
            block.timestamp > a.deadline || a.nonce != nonces[a.signer]) {
            revert InvalidApproval();
        }
        if (!SignatureChecker.isValidSignatureNow(a.signer, digest(a), signature)) {
            revert InvalidApproval();
        }
        nonces[a.signer] = a.nonce + 1;
        emit Recorded(a.signer, a.contentHash, a.nonce);
    }
}
