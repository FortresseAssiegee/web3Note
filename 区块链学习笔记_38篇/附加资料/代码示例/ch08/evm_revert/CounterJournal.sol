// SPDX-License-Identifier: MIT
pragma solidity 0.8.30;

// Educational state/log rollback example. It holds no funds and has no roles.
contract CounterJournal {
    uint256 public value;

    error DeliberateFailure(uint256 attempted);
    event Changed(address indexed actor, uint256 previousValue, uint256 nextValue);

    function set(uint256 nextValue, bool failAfterWrite) external {
        uint256 previousValue = value;
        value = nextValue;
        emit Changed(msg.sender, previousValue, nextValue);
        if (failAfterWrite) revert DeliberateFailure(nextValue);
    }
}

contract CallerJournal {
    uint256 public attempts;
    event Attempted(bool success);

    function trySet(CounterJournal target, uint256 nextValue, bool failAfterWrite) external {
        attempts += 1;
        try target.set(nextValue, failAfterWrite) {
            emit Attempted(true);
        } catch {
            emit Attempted(false);
        }
    }
}
