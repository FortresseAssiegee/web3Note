// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

// Teaching example: records claims; does not certify physical truth.
contract BatchProvenance {
    bytes32 public constant DOMAIN = keccak256("BatchProvenance.v1");
    address public immutable administrator;

    struct Head {
        bytes32 commitment;
        uint64 sequence;
        bool exists;
    }

    mapping(bytes32 => Head) public heads;
    mapping(bytes32 => mapping(address => bool)) public reporters;

    event BatchCreated(bytes32 indexed batchId);
    event ReporterSet(bytes32 indexed batchId, address indexed reporter, bool allowed);
    event Anchored(
        bytes32 indexed batchId,
        uint64 indexed sequence,
        address indexed reporter,
        bytes32 documentHash,
        bytes32 previousCommitment,
        bytes32 commitment,
        uint64 observedAt,
        uint256 recordedAt
    );

    constructor() {
        administrator = msg.sender;
    }

    modifier onlyAdministrator() {
        require(msg.sender == administrator, "not administrator");
        _;
    }

    function createBatch(bytes32 batchId) external onlyAdministrator {
        require(batchId != bytes32(0), "empty batch id");
        require(!heads[batchId].exists, "batch already exists");
        heads[batchId].exists = true;
        emit BatchCreated(batchId);
    }

    function setReporter(bytes32 batchId, address reporter, bool allowed)
        external
        onlyAdministrator
    {
        require(heads[batchId].exists, "unknown batch");
        require(reporter != address(0), "zero reporter");
        reporters[batchId][reporter] = allowed;
        emit ReporterSet(batchId, reporter, allowed);
    }

    function anchor(
        bytes32 batchId,
        bytes32 documentHash,
        bytes32 expectedPrevious,
        uint64 sequence,
        uint64 observedAt
    ) external returns (bytes32 commitment) {
        Head storage head = heads[batchId];
        require(head.exists, "unknown batch");
        require(reporters[batchId][msg.sender], "not reporter");
        require(documentHash != bytes32(0), "empty document hash");
        require(expectedPrevious == head.commitment, "stale previous");
        require(sequence == head.sequence + 1, "wrong sequence");
        require(observedAt <= block.timestamp, "future observation");

        commitment = keccak256(abi.encode(
            DOMAIN, block.chainid, address(this), batchId, sequence,
            documentHash, expectedPrevious, observedAt, msg.sender
        ));
        head.commitment = commitment;
        head.sequence = sequence;
        emit Anchored(
            batchId, sequence, msg.sender, documentHash,
            expectedPrevious, commitment, observedAt, block.timestamp
        );
    }
}
